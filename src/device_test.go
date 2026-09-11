package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseDeviceList 验证 go-ios 常见嵌套设备结构能够转换为稳定模型。
func TestParseDeviceList(t *testing.T) {
	output := `{"deviceList":[{"properties":{"SerialNumber":"00008110-001234567890801E","ConnectionType":"USB"},"values":{"ProductName":"测试 iPhone","ProductType":"iPhone15,4","ProductVersion":"18.6"}}]}`

	devices, err := parseDeviceList(output)
	if err != nil {
		t.Fatalf("parseDeviceList() error = %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("parseDeviceList() count = %d, want 1", len(devices))
	}

	device := devices[0]
	if device.UDID != "00008110-001234567890801E" {
		t.Errorf("UDID = %q", device.UDID)
	}
	if device.Name != "测试 iPhone" {
		t.Errorf("Name = %q", device.Name)
	}
	if device.ProductType != "iPhone15,4" {
		t.Errorf("ProductType = %q, want 'iPhone15,4'", device.ProductType)
	}
	if device.ProductVersion != "18.6" {
		t.Errorf("ProductVersion = %q", device.ProductVersion)
	}
	if device.ConnectionType != "USB" {
		t.Errorf("ConnectionType = %q", device.ConnectionType)
	}
}

// TestParseDeviceListRealIosDetails 验证真实 go-ios 1.3.2 list --details 输出不会将 ProductName (iPhone OS) 错误当成设备名称。
func TestParseDeviceListRealIosDetails(t *testing.T) {
	output := `{"deviceList":[{"Udid":"000081300006486A0E52001C","ProductName":"iPhone OS","ProductType":"iPad16,2","ProductVersion":"26.6.1","ConnectionType":"USB"}]}`

	devices, err := parseDeviceList(output)
	if err != nil {
		t.Fatalf("parseDeviceList() error = %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("parseDeviceList() count = %d, want 1", len(devices))
	}

	device := devices[0]
	if device.UDID != "000081300006486A0E52001C" {
		t.Errorf("UDID = %q", device.UDID)
	}
	// 不能为 "iPhone OS"，应回退为原始硬件标识符
	if device.Name == "iPhone OS" {
		t.Errorf("Name should not be %q", device.Name)
	}
	if device.Name != "iPad16,2" {
		t.Errorf("Name = %q, want 'iPad16,2'", device.Name)
	}
	if device.ProductType != "iPad16,2" {
		t.Errorf("ProductType = %q, want 'iPad16,2'", device.ProductType)
	}
	if device.ProductVersion != "26.6.1" {
		t.Errorf("ProductVersion = %q", device.ProductVersion)
	}
	if device.ConnectionType != "USB" {
		t.Errorf("ConnectionType = %q", device.ConnectionType)
	}
}

// TestParseDeviceName 验证 go-ios devicename 命令输出解析。
func TestParseDeviceName(t *testing.T) {
	output := "{\"devicename\":\"王铮的iPad\"}\n"
	name := parseDeviceName(output)
	if name != "王铮的iPad" {
		t.Errorf("parseDeviceName() = %q, want '王铮的iPad'", name)
	}

	// 验证 "iPhone OS" 会被拒绝
	invalidOutput := "{\"devicename\":\"iPhone OS\"}\n"
	if invalidName := parseDeviceName(invalidOutput); invalidName != "" {
		t.Errorf("parseDeviceName() for iPhone OS = %q, want empty", invalidName)
	}
}

// TestParseDeviceListSkipsInvalidAndDuplicateDevices 验证缺少 UDID 与重复设备不会进入列表。
func TestParseDeviceListSkipsInvalidAndDuplicateDevices(t *testing.T) {
	output := `{"deviceList":[{"name":"无标识设备"},{"udid":"device-1","name":"iPhone"},{"serialNumber":"device-1","name":"重复设备"}]}`

	devices, err := parseDeviceList(output)
	if err != nil {
		t.Fatalf("parseDeviceList() error = %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("parseDeviceList() count = %d, want 1", len(devices))
	}
	if devices[0].UDID != "device-1" {
		t.Errorf("UDID = %q", devices[0].UDID)
	}
}

// TestParseDeviceListPlainUDIDArray 验证纯字符串 UDID 数组能正确识别为已连接待解锁设备。
func TestParseDeviceListPlainUDIDArray(t *testing.T) {
	output := `{"deviceList":["00008130-001234567890801E"]}`
	devices, err := parseDeviceList(output)
	if err != nil {
		t.Fatalf("parseDeviceList() error = %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("parseDeviceList() count = %d, want 1", len(devices))
	}
	if devices[0].UDID != "00008130-001234567890801E" {
		t.Errorf("UDID = %q", devices[0].UDID)
	}
	if !strings.Contains(devices[0].Name, "解锁屏幕") {
		t.Errorf("Name = %q, want contain 解锁屏幕", devices[0].Name)
	}
}

// TestValidateIPAPathRejectsNonIPA 验证安装入口会拒绝非 IPA 文件。
func TestValidateIPAPathRejectsNonIPA(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-an-ipa.zip")
	if _, err := validateIPAPath(path); err == nil {
		t.Fatal("validateIPAPath() expected an error")
	}
}

// TestFriendlyDeviceToolErrorInvalidHostID 验证 InvalidHostID 会被转换为友好的中文信任指引。
func TestFriendlyDeviceToolErrorInvalidHostID(t *testing.T) {
	rawErr := fmt.Errorf(`{"time":"2026-09-08T21:39:02","level":"ERROR","msg":"failed getting values","err":"StartSession failed: {EnableSessionSSL:false Request: SessionID: Error:} error: failed to start new lockdown session: InvalidHostID"}`)
	err := friendlyDeviceToolError(rawErr)
	if err == nil || !strings.Contains(err.Error(), "信任") {
		t.Fatalf("friendlyDeviceToolError() = %v, want friendly message with trust instructions", err)
	}
}
