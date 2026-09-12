//go:build !windows

package backend

import "os/exec"

func setSysProcAttr(cmd *exec.Cmd) {}

func initWindowIcon() {}
