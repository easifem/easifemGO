//go:build unix

// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package testrunner

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func configureCancellation(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
}
