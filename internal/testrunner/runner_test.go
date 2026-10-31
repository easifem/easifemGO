// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package testrunner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeExecutor struct {
	build   func(context.Context, Test, io.Writer) (Program, error)
	execute func(context.Context, Program, io.Writer) error
}

func (e fakeExecutor) Build(ctx context.Context, test Test, out io.Writer) (Program, error) {
	if e.build != nil {
		return e.build(ctx, test, out)
	}
	return Program{Path: test.Path + ".out", Dir: test.Dir}, nil
}
func (e fakeExecutor) Execute(ctx context.Context, p Program, out io.Writer) error {
	return e.execute(ctx, p, out)
}

func fakeTests(root string) []Test {
	return []Test{{Path: filepath.Join(root, "A_test_1.F90"), Dir: root, Relative: "A_test_1.F90"}, {Path: filepath.Join(root, "B_test_1.F90"), Dir: root, Relative: "B_test_1.F90"}}
}

func TestFailuresAreSeparatedAndLogsAreRetained(t *testing.T) {
	tests := fakeTests(t.TempDir())
	executions := 0
	executor := fakeExecutor{
		build: func(ctx context.Context, test Test, out io.Writer) (Program, error) {
			fmt.Fprintln(out, "ok 1 - build output is not an execution check")
			if test.Relative == "A_test_1.F90" {
				return Program{}, errors.New("compiler failed")
			}
			return Program{Path: test.Path + ".out", Dir: test.Dir}, nil
		},
		execute: func(ctx context.Context, p Program, out io.Writer) error {
			executions++
			fmt.Fprintln(out, "ok 1")
			return errors.New("program crashed after reporting success")
		},
	}
	summary, err := Run(context.Background(), tests, executor, Options{LogDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if executions != 1 || summary.Results[0].Status != Compile || summary.Results[1].Status != Crash {
		t.Fatalf("unexpected results: %+v", summary)
	}
	if summary.Results[0].Passed != 0 {
		t.Fatal("build output was parsed as checks")
	}
	for _, result := range summary.Results {
		data, err := os.ReadFile(result.Log)
		if err != nil || !strings.Contains(string(data), "=== BUILD") {
			t.Fatalf("missing log: %v", err)
		}
	}
	if summary.Bad(false) != 2 {
		t.Fatal("failures counted as successful")
	}
}

func TestTimeoutAndCancellation(t *testing.T) {
	tests := fakeTests(t.TempDir())
	executor := fakeExecutor{execute: func(ctx context.Context, p Program, out io.Writer) error { <-ctx.Done(); return ctx.Err() }}
	summary, err := Run(context.Background(), tests[:1], executor, Options{LogDir: t.TempDir(), Timeout: time.Millisecond})
	if err != nil || summary.Results[0].Status != Timeout {
		t.Fatalf("timeout: %+v, %v", summary, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	executor.build = func(ctx context.Context, test Test, out io.Writer) (Program, error) {
		cancel()
		return Program{}, ctx.Err()
	}
	summary, err = Run(ctx, tests, executor, Options{LogDir: t.TempDir()})
	if !errors.Is(err, context.Canceled) || len(summary.Results) != 1 || summary.Results[0].Status != Cancelled || summary.NotRun != 1 {
		t.Fatalf("cancellation: %+v, %v", summary, err)
	}
}

func TestNoChecksAndFailFastPolicy(t *testing.T) {
	tests := fakeTests(t.TempDir())
	executor := fakeExecutor{execute: func(ctx context.Context, p Program, out io.Writer) error { return nil }}
	summary, err := Run(context.Background(), tests, executor, Options{LogDir: t.TempDir(), FailFast: true})
	if err != nil || summary.NotRun != 1 || summary.Bad(false) != 2 {
		t.Fatalf("fail fast: %+v, %v", summary, err)
	}
	summary, err = Run(context.Background(), tests, executor, Options{LogDir: t.TempDir(), FailFast: true, AllowNoCheck: true})
	if err != nil || summary.NotRun != 0 || summary.Bad(true) != 0 || summary.Results[0].Status != NoCheck {
		t.Fatalf("allow unchecked: %+v, %v", summary, err)
	}
}
