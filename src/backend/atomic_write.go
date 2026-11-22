package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+"-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err = tmp.Chmod(perm); err == nil {
		if _, err = tmp.Write(data); err == nil {
			err = tmp.Sync()
		}
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}

	if err = os.Rename(tmpPath, path); err == nil {
		return nil
	}
	// Windows does not replace an existing destination with os.Rename.
	backupPath := path + ".bak"
	_ = os.Remove(backupPath)
	if _, statErr := os.Stat(path); statErr == nil {
		if renameErr := os.Rename(path, backupPath); renameErr != nil {
			return err
		}
	}
	if renameErr := os.Rename(tmpPath, path); renameErr != nil {
		_ = os.Rename(backupPath, path)
		return renameErr
	}
	_ = os.Remove(backupPath)
	return nil
}

// backupCorruptFile 把无法解析的文件重命名为 <path>.corrupt-<20060102-150405>，返回新路径。
func backupCorruptFile(path string) (string, error) {
	stamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.corrupt-%s", path, stamp)
	if _, err := os.Stat(backupPath); err == nil {
		backupPath = fmt.Sprintf("%s.corrupt-%s-%d", path, stamp, time.Now().UnixNano())
	}
	if err := os.Rename(path, backupPath); err != nil {
		return "", err
	}
	return backupPath, nil
}
