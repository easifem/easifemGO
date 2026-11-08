//go:build !unix

// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package testrunner

import "os/exec"

// CommandContext's default cancellation terminates the immediate process.
func configureCancellation(cmd *exec.Cmd) {}
