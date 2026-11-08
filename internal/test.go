// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package internal

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/easifem/easifemGO/internal/testrunner"
	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	var recursive, list, failFast, rebuild, allowNoCheck bool
	var filters []string
	var logDir, jsonPath string
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "test [directory]",
		Short: "Build and run standalone Fortran examples, keeping per-test logs",
		Long: `Discover *_test_*.F90/.f90 programs and run each from its source directory.
Tests run sequentially with isolated CMake build directories. Only test
executables are compiled; installed EASIFEM libraries are not rebuilt.

PASS requires successful execution and passing Test_Method/TAP checks.
Programs without checks are NOCHECK and make the command fail unless
--allow-no-check is supplied. Logs and optional JSON include partial results
if execution is interrupted. Existing CLI configuration is still required.`,
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		Example:       "  easifem test ./src/modules/FEDOF/examples\n  easifem test ./src/modules --recursive --filter 'DG*' --json results.json\n  easifem test ./examples --list",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			if timeout < 0 {
				return fmt.Errorf("--timeout must be nonnegative")
			}
			if list && jsonPath != "" {
				return fmt.Errorf("--list and --json cannot be combined")
			}
			if jsonPath != "" && !strings.EqualFold(filepath.Ext(jsonPath), ".json") {
				return fmt.Errorf("--json output must have a .json extension")
			}
			tests, err := testrunner.Discover(root, recursive, filters)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if list {
				for _, test := range tests {
					if _, err := fmt.Fprintln(out, test.Relative); err != nil {
						return err
					}
				}
				return nil
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			summary, runErr := testrunner.Run(ctx, tests, exampleTestExecutor{rebuild: rebuild}, testrunner.Options{
				LogDir: logDir, Timeout: timeout, FailFast: failFast, AllowNoCheck: allowNoCheck,
				OnResult: func(result testrunner.Result) error { return testrunner.PrintResult(out, result) },
			})
			var reportErr error
			if jsonPath != "" {
				if err := testrunner.WriteJSON(jsonPath, summary); err != nil {
					reportErr = fmt.Errorf("write JSON report: %w", err)
				}
			}
			summaryErr := testrunner.PrintSummary(out, summary)
			if err := errors.Join(runErr, reportErr, summaryErr); err != nil {
				return err
			}
			if bad := summary.Bad(allowNoCheck); bad > 0 {
				return fmt.Errorf("%d unsuccessful or unrun programs; see %s", bad, summary.LogDir)
			}
			return nil
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&recursive, "recursive", "r", false, "Search subdirectories (skipping hidden/build directories and symlinks)")
	flags.StringArrayVar(&filters, "filter", nil, "Select basename or relative-path glob (repeat for alternatives)")
	flags.BoolVar(&list, "list", false, "List selected files without building or executing them")
	flags.StringVar(&logDir, "logdir", "", "Parent for a unique log directory (default: system temporary directory)")
	flags.StringVar(&jsonPath, "json", "", "Write a versioned JSON report to this .json file")
	flags.DurationVar(&timeout, "timeout", 0, "Per-test build plus execution timeout; zero means unlimited")
	flags.BoolVar(&failFast, "fail-fast", false, "Stop after the first unsuccessful program")
	flags.BoolVar(&rebuild, "rebuild", false, "Clean only each test's isolated build before compiling it")
	flags.BoolVar(&allowNoCheck, "allow-no-check", false, "Accept successful programs that report no checks (still labelled NOCHECK)")
	return cmd
}

func init() { rootCmd.AddCommand(newTestCmd()) }
