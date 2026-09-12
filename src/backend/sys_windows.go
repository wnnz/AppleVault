//go:build windows

package backend

import (
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"
)

func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}

func initWindowIcon() {
	go func() {
		// 稍微延时等待 Wails 初始化并展示主窗口
		time.Sleep(200 * time.Millisecond)

		user32 := syscall.NewLazyDLL("user32.dll")
		shell32 := syscall.NewLazyDLL("shell32.dll")
		procEnumWindows := user32.NewProc("EnumWindows")
		procGetWindowThreadProcessId := user32.NewProc("GetWindowThreadProcessId")
		procSendMessageW := user32.NewProc("SendMessageW")
		procIsWindowVisible := user32.NewProc("IsWindowVisible")
		procExtractIconW := shell32.NewProc("ExtractIconW")

		exePath, err := os.Executable()
		if err != nil {
			return
		}
		exePathPtr, err := syscall.UTF16PtrFromString(exePath)
		if err != nil {
			return
		}

		hIcon, _, _ := procExtractIconW.Call(0, uintptr(unsafe.Pointer(exePathPtr)), 0)
		if hIcon == 0 {
			return
		}

		currentPID := uint32(os.Getpid())
		var targetHwnd uintptr

		cb := syscall.NewCallback(func(hwnd, lparam uintptr) uintptr {
			var pid uint32
			procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
			if pid == currentPID {
				visible, _, _ := procIsWindowVisible.Call(hwnd)
				if visible != 0 {
					targetHwnd = hwnd
					return 0 // 找到了，停止枚举
				}
			}
			return 1 // 继续枚举
		})

		for i := 0; i < 15; i++ {
			targetHwnd = 0
			procEnumWindows.Call(cb, 0)
			if targetHwnd != 0 {
				const (
					WM_SETICON = 0x0080
					ICON_SMALL = 0
					ICON_BIG   = 1
				)
				procSendMessageW.Call(targetHwnd, WM_SETICON, ICON_SMALL, hIcon)
				procSendMessageW.Call(targetHwnd, WM_SETICON, ICON_BIG, hIcon)
				break
			}
			time.Sleep(150 * time.Millisecond)
		}
	}()
}
