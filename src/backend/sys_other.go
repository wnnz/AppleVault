//go:build !windows

package backend

import "os/exec"

func setSysProcAttr(cmd *exec.Cmd) {}

func initWindowIcon() {}

func openExplorerWindows(targetPath string, statErr error, isDir bool) error {
	return nil
}
