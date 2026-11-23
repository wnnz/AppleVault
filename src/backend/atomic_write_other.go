//go:build !windows

package backend

import "os"

func replaceFileAtomic(tmpPath, dstPath string) error {
	return os.Rename(tmpPath, dstPath)
}
