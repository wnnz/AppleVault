package backend

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"howett.net/plist"
)

const maxIPAInfoPlistSize = 2 << 20

type ipaBundleInfo struct {
	DisplayName        string   `plist:"CFBundleDisplayName"`
	Name               string   `plist:"CFBundleName"`
	BundleID           string   `plist:"CFBundleIdentifier"`
	ShortVersion       string   `plist:"CFBundleShortVersionString"`
	BuildVersion       string   `plist:"CFBundleVersion"`
	MinimumOSVersion   string   `plist:"MinimumOSVersion"`
	SupportedPlatforms []string `plist:"CFBundleSupportedPlatforms"`
	DeviceFamily       []int    `plist:"UIDeviceFamily"`
}

func (a *App) InspectIPA(ipaPath, deviceProductType, deviceVersion string) (IPAInspectionResult, error) {
	absolutePath, err := validateIPAPath(ipaPath)
	if err != nil {
		return IPAInspectionResult{}, err
	}
	reader, err := zip.OpenReader(absolutePath)
	if err != nil {
		return IPAInspectionResult{}, fmt.Errorf("IPA 文件结构损坏: %w", err)
	}
	defer reader.Close()

	var info ipaBundleInfo
	appRoot := ""
	signed := false
	foundInfo := false
	for _, entry := range reader.File {
		parts := strings.Split(entry.Name, "/")
		if len(parts) >= 2 && parts[0] == "Payload" && strings.HasSuffix(parts[1], ".app") {
			root := strings.Join(parts[:2], "/") + "/"
			if appRoot == "" {
				appRoot = root
			}
			if root == appRoot && (entry.Name == appRoot+"_CodeSignature/CodeResources" || entry.Name == appRoot+"embedded.mobileprovision") {
				signed = true
			}
		}
		if len(parts) == 3 && parts[0] == "Payload" && strings.HasSuffix(parts[1], ".app") && parts[2] == "Info.plist" && !foundInfo {
			stream, openErr := entry.Open()
			if openErr != nil {
				return IPAInspectionResult{}, openErr
			}
			data, readErr := io.ReadAll(io.LimitReader(stream, maxIPAInfoPlistSize+1))
			closeErr := stream.Close()
			if readErr != nil {
				return IPAInspectionResult{}, readErr
			}
			if closeErr != nil {
				return IPAInspectionResult{}, closeErr
			}
			if len(data) > maxIPAInfoPlistSize {
				return IPAInspectionResult{}, fmt.Errorf("IPA 的 Info.plist 文件过大")
			}
			if _, err = plist.Unmarshal(data, &info); err != nil {
				return IPAInspectionResult{}, fmt.Errorf("无法解析 IPA 应用信息: %w", err)
			}
			foundInfo = true
		}
	}
	if !foundInfo || appRoot == "" {
		return IPAInspectionResult{}, fmt.Errorf("IPA 中未找到 Payload/*.app/Info.plist")
	}
	if err = verifyZIPEntries(reader.File); err != nil {
		return IPAInspectionResult{}, fmt.Errorf("IPA 完整性校验失败: %w", err)
	}

	deviceTypes := deviceFamilyNames(info.DeviceFamily)
	compatible, message := evaluateIPACompatibility(info, signed, deviceProductType, deviceVersion)
	name := strings.TrimSpace(info.DisplayName)
	if name == "" {
		name = strings.TrimSpace(info.Name)
	}
	stat, err := os.Stat(absolutePath)
	if err != nil {
		return IPAInspectionResult{}, err
	}
	return IPAInspectionResult{
		Success: true, AppName: name, BundleID: info.BundleID, Version: info.ShortVersion,
		BuildVersion: info.BuildVersion, MinimumOSVersion: info.MinimumOSVersion,
		SupportedPlatforms: append([]string(nil), info.SupportedPlatforms...), SupportedDeviceTypes: deviceTypes,
		Signed: signed, FileSize: stat.Size(), DisplayFileSize: formatBytes(stat.Size()), Compatible: compatible,
		CompatibilityMessage: message,
	}, nil
}

func verifyZIPEntries(entries []*zip.File) error {
	for _, entry := range entries {
		if entry.FileInfo().IsDir() {
			continue
		}
		stream, err := entry.Open()
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(io.Discard, stream)
		closeErr := stream.Close()
		if copyErr != nil {
			return fmt.Errorf("%s: %w", entry.Name, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("%s: %w", entry.Name, closeErr)
		}
	}
	return nil
}

func deviceFamilyNames(values []int) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		switch value {
		case 1:
			result = append(result, "iPhone/iPod")
		case 2:
			result = append(result, "iPad")
		case 3:
			result = append(result, "Apple TV")
		case 7:
			result = append(result, "Apple Vision")
		}
	}
	return result
}

func evaluateIPACompatibility(info ipaBundleInfo, signed bool, productType, osVersion string) (bool, string) {
	if !signed {
		return false, "未检测到 IPA 代码签名文件"
	}
	productType = strings.ToLower(strings.TrimSpace(productType))
	if isAppleVersion(osVersion) && isAppleVersion(info.MinimumOSVersion) && compareAppleVersions(osVersion, info.MinimumOSVersion) < 0 {
		return false, fmt.Sprintf("设备系统 %s 低于应用要求的 %s", osVersion, info.MinimumOSVersion)
	}
	if productType == "" {
		return true, "IPA 结构与签名检查通过；选择设备后可进一步检查兼容性"
	}
	recognizedDevice := strings.Contains(productType, "iphone") || strings.Contains(productType, "ipod") ||
		strings.Contains(productType, "ipad") || strings.Contains(productType, "appletv") ||
		strings.Contains(productType, "reality") || strings.Contains(productType, "vision")
	if !recognizedDevice {
		return true, "IPA 完整性与签名结构检查通过；无法识别设备型号，未检查设备类型兼容性"
	}
	expectedPlatform := "iPhoneOS"
	expectedFamily := 1
	if strings.Contains(productType, "ipad") {
		expectedFamily = 2
	} else if strings.Contains(productType, "appletv") {
		expectedPlatform, expectedFamily = "AppleTVOS", 3
	} else if strings.Contains(productType, "reality") || strings.Contains(productType, "vision") {
		expectedPlatform, expectedFamily = "XROS", 7
	}
	if len(info.SupportedPlatforms) > 0 && !containsFold(info.SupportedPlatforms, expectedPlatform) {
		return false, fmt.Sprintf("IPA 不支持目标设备平台 %s", expectedPlatform)
	}
	if len(info.DeviceFamily) > 0 && !containsInt(info.DeviceFamily, expectedFamily) {
		return false, "IPA 不支持所选设备类型"
	}
	return true, "IPA 完整性、签名结构和设备兼容性检查通过"
}

func isAppleVersion(value string) bool {
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if numericPrefix(part) == "" {
			return false
		}
	}
	return true
}

func compareAppleVersions(left, right string) int {
	leftParts, rightParts := strings.Split(left, "."), strings.Split(right, ".")
	length := len(leftParts)
	if len(rightParts) > length {
		length = len(rightParts)
	}
	for index := 0; index < length; index++ {
		var leftValue, rightValue int
		if index < len(leftParts) {
			leftValue, _ = strconv.Atoi(numericPrefix(leftParts[index]))
		}
		if index < len(rightParts) {
			rightValue, _ = strconv.Atoi(numericPrefix(rightParts[index]))
		}
		if leftValue < rightValue {
			return -1
		}
		if leftValue > rightValue {
			return 1
		}
	}
	return 0
}

func numericPrefix(value string) string {
	end := 0
	for end < len(value) && value[end] >= '0' && value[end] <= '9' {
		end++
	}
	return value[:end]
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
