// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package testrunner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// Program separates the artifact path from the working directory for fixtures.
type Program struct{ Path, Dir string }

// Executor separates build errors from runtime errors. It must return process
// failures and complete all output writes before returning.
type Executor interface {
	Build(context.Context, Test, io.Writer) (Program, error)
	Execute(context.Context, Program, io.Writer) error
}

type Options struct {
	// LogDir is a parent directory; each invocation gets a unique subdirectory.
	LogDir       string
	Timeout      time.Duration // Per program, including its build; zero disables it.
	FailFast     bool
	AllowNoCheck bool // Also controls whether fail-fast stops on NOCHECK.
	OnResult     func(Result) error
}

var unsafeLogName = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)

// Run executes programs sequentially, retaining partial results on interruption.
// Program failures are results; setup/reporting failures are returned as errors.
func Run(ctx context.Context, tests []Test, executor Executor, opts Options) (summary Summary, err error) {
	started := time.Now()
	summary = Summary{Version: 1, Started: started.UTC(), Total: len(tests), NotRun: len(tests), Results: []Result{}}
	defer func() { summary.Duration = time.Since(started) }()
	if executor == nil {
		return summary, errors.New("no test executor")
	}
	if len(tests) == 0 {
		return summary, errors.New("no tests selected")
	}
	if opts.Timeout < 0 {
		return summary, errors.New("timeout cannot be negative")
	}
	if err := ctx.Err(); err != nil {
		return summary, err
	}
	if opts.LogDir != "" {
		if err := os.MkdirAll(opts.LogDir, 0700); err != nil {
			return summary, err
		}
	}
	logs, err := os.MkdirTemp(opts.LogDir, "easifem-tests-")
	if err != nil {
		return summary, fmt.Errorf("create logs: %w", err)
	}
	summary.LogDir, err = filepath.Abs(logs)
	if err != nil {
		return summary, err
	}
	for i, test := range tests {
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		logPath := filepath.Join(summary.LogDir, fmt.Sprintf("%03d-%s.log", i+1, unsafeLogName.ReplaceAllString(filepath.Base(test.Path), "_")))
		result, err := runOne(ctx, test, executor, opts.Timeout, logPath)
		if err != nil {
			return summary, err
		}
		summary.Results = append(summary.Results, result)
		summary.NotRun = summary.Total - len(summary.Results)
		if opts.OnResult != nil {
			if err := opts.OnResult(result); err != nil {
				return summary, err
			}
		}
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		if opts.FailFast && result.Status != Pass && result.Status != Skip && !(opts.AllowNoCheck && result.Status == NoCheck) {
			break
		}
	}
	return summary, nil
}

func runOne(parent context.Context, test Test, executor Executor, timeout time.Duration, logPath string) (result Result, err error) {
	started := time.Now()
	result = Result{Test: test, Log: logPath}
	defer func() { result.Duration = time.Since(started) }()
	log, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return result, fmt.Errorf("open test log: %w", err)
	}
	defer func() {
		if closeErr := log.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	output := &checkedWriter{out: log}
	ctx := parent
	cancel := func() {}
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(parent, timeout)
	}
	defer cancel()
	_, _ = fmt.Fprintf(output, "=== BUILD %s ===\n", test.Path)
	program, buildErr := executor.Build(ctx, test, output)
	var checks checkOutput
	runErr := error(nil)
	if buildErr == nil && ctx.Err() == nil {
		result.Executable = program.Path
		_, _ = fmt.Fprintln(output, "\n=== EXECUTE ===")
		runErr = executor.Execute(ctx, program, io.MultiWriter(output, &checks))
		checks.finish()
	}
	if output.err != nil {
		return result, fmt.Errorf("write log %s: %w", logPath, output.err)
	}
	result.Passed, result.Failed, result.Skipped = checks.passed, checks.failed, checks.skipped
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		result.Status, result.Error = Timeout, ctx.Err().Error()
	case ctx.Err() != nil:
		result.Status, result.Error = Cancelled, ctx.Err().Error()
	case buildErr != nil:
		result.Status, result.Error = Compile, buildErr.Error()
	case runErr != nil:
		result.Status, result.Error = Crash, runErr.Error()
	default:
		result.Status, result.Error = checks.status()
	}
	return result, nil
}

// checkedWriter keeps log I/O failures distinct from a failing Fortran program.
type checkedWriter struct {
	out io.Writer
	err error
}

func (w *checkedWriter) Write(data []byte) (int, error) {
	n, err := w.out.Write(data)
	if err != nil && w.err == nil {
		w.err = err
	}
	return n, err
}
