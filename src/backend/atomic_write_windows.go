//go:build windows

package backend

import (
	"os"

	"golang.org/x/sys/windows"
)

func replaceFileAtomic(tmpPath, dstPath string) error {
	srcPtr, srcErr := windows.UTF16PtrFromString(tmpPath)
	dstPtr, dstErr := windows.UTF16PtrFromString(dstPath)
	if srcErr == nil && dstErr == nil {
		err := windows.MoveFileEx(srcPtr, dstPtr, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
		if err == nil {
			return nil
		}
	}

	if err := os.Rename(tmpPath, dstPath); err == nil {
		return nil
	} else {
		// Fallback two-step replacement when MoveFileEx/Rename cannot replace an existing target.
		backupPath := dstPath + ".bak"
		_ = os.Remove(backupPath)
		if _, statErr := os.Stat(dstPath); statErr == nil {
			if renameErr := os.Rename(dstPath, backupPath); renameErr != nil {
				return err
			}
		}
		if renameErr := os.Rename(tmpPath, dstPath); renameErr != nil {
			_ = os.Rename(backupPath, dstPath)
			return renameErr
		}
		_ = os.Remove(backupPath)
		return nil
	}
}
