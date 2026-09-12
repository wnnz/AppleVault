package appstore

import (
	"archive/zip"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
	"howett.net/plist"
)

var (
	ErrLicenseRequired = errors.New("license is required")
)

type DownloadInput struct {
	Account           Account
	App               App
	OutputPath        string
	Progress          *progressbar.ProgressBar
	ExternalVersionID string
	Platform          Platform
}

type DownloadOutput struct {
	DestinationPath string
	Sinfs           []Sinf
}

func (t *appstore) Download(input DownloadInput) (DownloadOutput, error) {
	macAddr, err := t.machine.MacAddress()
	if err != nil {
		return DownloadOutput{}, fmt.Errorf("failed to get mac address: %w", err)
	}

	guid := strings.ReplaceAll(strings.ToUpper(macAddr), ":", "")

	externalVersionID := input.ExternalVersionID
	if externalVersionID == "" && (input.Platform == PlatformAppleTV || input.Platform == PlatformVisionOS) {
		externalVersionID, err = t.lookupLatestExternalVersionID(input.Account, input.App, input.Platform)
		if err != nil {
			return DownloadOutput{}, fmt.Errorf("failed to resolve platform version: %w", err)
		}
	}

	res, err := t.sendDownloadProduct(input.Account, input.App, guid, externalVersionID, input.Platform)
	if err != nil {
		return DownloadOutput{}, err
	}

	if res.Data.FailureType == FailureTypePasswordTokenExpired ||
		res.Data.FailureType == FailureTypeSignInRequired ||
		res.Data.FailureType == FailureTypeDeviceVerificationFailed ||
		res.Data.FailureType == FailureTypeLicenseAlreadyExists {
		return DownloadOutput{}, ErrPasswordTokenExpired
	}

	if res.Data.FailureType == FailureTypeLicenseNotFound {
		return DownloadOutput{}, ErrLicenseRequired
	}

	if res.Data.FailureType != "" && res.Data.CustomerMessage != "" {
		return DownloadOutput{}, NewErrorWithMetadata(fmt.Errorf("received error: %s", res.Data.CustomerMessage), res)
	}

	if res.Data.FailureType != "" {
		return DownloadOutput{}, NewErrorWithMetadata(fmt.Errorf("received error: %s", res.Data.FailureType), res)
	}

	if len(res.Data.Items) == 0 {
		return DownloadOutput{}, NewErrorWithMetadata(errors.New("invalid response"), res)
	}

	item := res.Data.Items[0]

	version := "unknown"

	// Read the version from the item metadata
	if itemVersion, ok := item.Metadata["bundleShortVersionString"]; ok {
		version = fmt.Sprintf("%v", itemVersion)
	}

	destination, err := t.resolveDestinationPath(input.App, version, input.OutputPath)
	if err != nil {
		return DownloadOutput{}, fmt.Errorf("failed to resolve destination path: %w", err)
	}

	tmpPath := fmt.Sprintf("%s.tmp", destination)

	err = t.downloadFile(item.URL, tmpPath, input.Progress)
	if err != nil {
		return DownloadOutput{}, fmt.Errorf("failed to download file: %w", err)
	}
	if err = validateFileMD5(tmpPath, item.HashMD5); err != nil {
		_ = t.os.Remove(tmpPath)
		return DownloadOutput{}, fmt.Errorf("failed to verify download checksum: %w", err)
	}

	err = t.applyPatches(item, input.Account, tmpPath, destination)
	if err != nil {
		return DownloadOutput{}, fmt.Errorf("failed to apply patches: %w", err)
	}
	if err = validateIPAZip(destination); err != nil {
		_ = t.os.Remove(destination)
		return DownloadOutput{}, fmt.Errorf("failed to validate IPA integrity: %w", err)
	}

	err = t.validatePackagePlatform(destination, input.Platform)
	if err != nil {
		return DownloadOutput{}, fmt.Errorf("failed to validate package platform: %w", err)
	}

	err = t.os.Remove(fmt.Sprintf("%s.tmp", destination))
	if err != nil {
		return DownloadOutput{}, fmt.Errorf("failed to remove file: %w", err)
	}

	return DownloadOutput{
		DestinationPath: destination,
		Sinfs:           item.Sinfs,
	}, nil
}

type platformPackageInfo struct {
	SupportedPlatforms []string `plist:"CFBundleSupportedPlatforms,omitempty"`
}

func (*appstore) validatePackagePlatform(path string, platform Platform) error {
	var expectedPlatform string

	switch platform {
	case PlatformAppleTV:
		expectedPlatform = "AppleTVOS"
	case PlatformVisionOS:
		expectedPlatform = "XROS"
	default:
		return nil
	}

	reader, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("failed to open zip reader: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if !isTopLevelAppInfoPlist(file.Name) {
			continue
		}

		infoFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open info plist: %w", err)
		}

		data, readErr := io.ReadAll(infoFile)
		closeErr := infoFile.Close()

		if readErr != nil {
			return fmt.Errorf("failed to read info plist: %w", readErr)
		}

		if closeErr != nil {
			return fmt.Errorf("failed to close info plist: %w", closeErr)
		}

		var info platformPackageInfo

		_, err = plist.Unmarshal(data, &info)
		if err != nil {
			return fmt.Errorf("failed to decode info plist: %w", err)
		}

		for _, supportedPlatform := range info.SupportedPlatforms {
			if supportedPlatform == expectedPlatform {
				return nil
			}
		}
	}

	return fmt.Errorf("downloaded package does not declare %s support", expectedPlatform)
}

func isTopLevelAppInfoPlist(path string) bool {
	parts := strings.Split(path, "/")

	return len(parts) == 3 && parts[0] == "Payload" && strings.HasSuffix(parts[1], ".app") && parts[2] == "Info.plist"
}

type downloadItemResult struct {
	HashMD5  string                 `plist:"md5,omitempty"`
	URL      string                 `plist:"URL,omitempty"`
	Sinfs    []Sinf                 `plist:"sinfs,omitempty"`
	Metadata map[string]interface{} `plist:"metadata,omitempty"`
}

type downloadResult struct {
	FailureType     string               `plist:"failureType,omitempty"`
	CustomerMessage string               `plist:"customerMessage,omitempty"`
	Items           []downloadItemResult `plist:"songList,omitempty"`
}

func (t *appstore) downloadFile(src, dst string, progress *progressbar.ProgressBar) error {
	req, err := t.httpClient.NewRequest("GET", src, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	file, err := t.os.OpenFile(dst, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer file.Close()

	stat, err := t.os.Stat(dst)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	resumeOffset := stat.Size()
	if req != nil && resumeOffset > 0 {
		req.Header.Add("range", fmt.Sprintf("bytes=%d-", stat.Size()))
	}

	res, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if resumeOffset > 0 && res.StatusCode == http.StatusRequestedRangeNotSatisfiable && validContentRangeComplete(res.Header.Get("Content-Range"), resumeOffset) {
		// A fully downloaded temporary file can remain after an interrupted post-processing step.
		// Let the caller's checksum validation decide whether it is safe to reuse.
		return nil
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download server returned HTTP %d", res.StatusCode)
	}
	if res.StatusCode == http.StatusPartialContent {
		if !validContentRangeStart(res.Header.Get("Content-Range"), resumeOffset) {
			return fmt.Errorf("download server returned an invalid Content-Range for offset %d", resumeOffset)
		}
	} else if resumeOffset > 0 {
		switch res.StatusCode {
		case http.StatusOK:
			if err = file.Truncate(0); err != nil {
				return fmt.Errorf("failed to restart download: %w", err)
			}
			resumeOffset = 0
		default:
			return fmt.Errorf("download server did not honor resume request (HTTP %d)", res.StatusCode)
		}
	}
	if _, err = file.Seek(resumeOffset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek download file: %w", err)
	}

	if progress != nil {
		if res.ContentLength >= 0 {
			progress.ChangeMax64(res.ContentLength + resumeOffset)
		}
		err = progress.Set64(resumeOffset)

		if err != nil {
			return fmt.Errorf("can not set bar progress: %w", err)
		}

		_, err = io.Copy(io.MultiWriter(file, progress), res.Body)
	} else {
		_, err = io.Copy(file, res.Body)
	}

	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func validContentRangeStart(value string, expected int64) bool {
	prefix := fmt.Sprintf("bytes %d-", expected)
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(value)), prefix)
}

func validContentRangeComplete(value string, expected int64) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	const prefix = "bytes */"
	if !strings.HasPrefix(value, prefix) {
		return false
	}
	total, err := strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(value, prefix)), 10, 64)
	return err == nil && total == expected
}

func validateFileMD5(path, expected string) error {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return nil
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := md5.New()
	if _, err = io.Copy(hash, file); err != nil {
		return err
	}
	want, err := decodeMD5(expected)
	if err != nil {
		return err
	}
	got := hash.Sum(nil)
	if !strings.EqualFold(hex.EncodeToString(got), hex.EncodeToString(want)) {
		return fmt.Errorf("MD5 mismatch: got %s", hex.EncodeToString(got))
	}
	return nil
}

func decodeMD5(value string) ([]byte, error) {
	if decoded, err := hex.DecodeString(value); err == nil && len(decoded) == md5.Size {
		return decoded, nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil && len(decoded) == md5.Size {
		return decoded, nil
	}
	return nil, errors.New("invalid MD5 value from download response")
}

func validateIPAZip(path string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer reader.Close()
	foundInfo := false
	for _, entry := range reader.File {
		if isTopLevelAppInfoPlist(entry.Name) {
			foundInfo = true
		}
		if entry.FileInfo().IsDir() {
			continue
		}
		stream, err := entry.Open()
		if err != nil {
			return fmt.Errorf("open %s: %w", entry.Name, err)
		}
		_, copyErr := io.Copy(io.Discard, stream)
		closeErr := stream.Close()
		if copyErr != nil {
			return fmt.Errorf("verify %s: %w", entry.Name, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", entry.Name, closeErr)
		}
	}
	if !foundInfo {
		return errors.New("archive does not contain Payload/*.app/Info.plist")
	}
	return nil
}

func fileName(app App, version string) string {
	var parts []string

	if app.BundleID != "" {
		parts = append(parts, app.BundleID)
	}

	if app.ID != 0 {
		parts = append(parts, strconv.FormatInt(app.ID, 10))
	}

	if version != "" {
		parts = append(parts, version)
	}

	return fmt.Sprintf("%s.ipa", strings.Join(parts, "_"))
}

func (t *appstore) resolveDestinationPath(app App, version string, path string) (string, error) {
	file := fileName(app, version)

	if path == "" {
		workdir, err := t.os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}

		return fmt.Sprintf("%s/%s", workdir, file), nil
	}

	isDir, err := t.isDirectory(path)
	if err != nil {
		return "", fmt.Errorf("failed to determine whether path is a directory: %w", err)
	}

	if isDir {
		return fmt.Sprintf("%s/%s", path, file), nil
	}

	return path, nil
}

func (t *appstore) isDirectory(path string) (bool, error) {
	info, err := t.os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to read file metadata: %w", err)
	}

	if info == nil {
		return false, nil
	}

	return info.IsDir(), nil
}

func (t *appstore) applyPatches(item downloadItemResult, acc Account, src, dst string) error {
	srcZip, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip reader: %w", err)
	}
	defer srcZip.Close()

	dstFile, err := t.os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer dstFile.Close()

	dstZip := zip.NewWriter(dstFile)
	defer dstZip.Close()

	err = t.replicateZip(srcZip, dstZip)
	if err != nil {
		return fmt.Errorf("failed to replicate zip: %w", err)
	}

	err = t.writeMetadata(item.Metadata, acc, dstZip)
	if err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func (t *appstore) writeMetadata(metadata map[string]interface{}, acc Account, zip *zip.Writer) error {
	metadata["apple-id"] = acc.Email
	metadata["userName"] = acc.Email

	metadataFile, err := zip.Create("iTunesMetadata.plist")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	data, err := plist.Marshal(metadata, plist.BinaryFormat)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	_, err = metadataFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}
