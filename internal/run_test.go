// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package internal

import (
	"os"
	"testing"
)

// This starts only a Go test helper, never CMake or a Fortran example.
func TestRunRunCmdPropagatesExitFailure(t *testing.T) {
	t.Setenv("EASIFEM_TEST_HELPER_EXIT", "1")
	err := runRunCmd([]string{os.Args[0], "-test.run=^TestRunCommandExitHelper$"}, true)
	if err == nil {
		t.Fatal("subprocess failure was discarded")
	}
}

func TestRunCommandExitHelper(t *testing.T) {
	if os.Getenv("EASIFEM_TEST_HELPER_EXIT") == "1" {
		os.Exit(7)
	}
}
