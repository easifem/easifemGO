// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package testrunner

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

// RunCommand invokes arguments directly, without a shell. It preserves process
// errors and uses the supplied working directory and streams. A nil input
// provides EOF rather than consuming the test runner's terminal input.
func RunCommand(ctx context.Context, dir string, output io.Writer, input io.Reader, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir, cmd.Stdin, cmd.Stdout, cmd.Stderr = dir, input, output, output
	cmd.WaitDelay = 3 * time.Second
	configureCancellation(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", args[0], err)
	}
	return nil
}
