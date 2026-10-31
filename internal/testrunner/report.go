// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package testrunner

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func PrintResult(out io.Writer, r Result) error {
	_, err := fmt.Fprintf(out, "%-9s %-45s ok=%d not_ok=%d skip=%d (%s)\n", r.Status, r.Test.Relative, r.Passed, r.Failed, r.Skipped, r.Duration.Round(time.Millisecond))
	if err == nil && r.Error != "" {
		_, err = fmt.Fprintf(out, "          %s\n", r.Error)
	}
	return err
}

func PrintSummary(out io.Writer, s Summary) error {
	counts := make(map[Status]int)
	for _, r := range s.Results {
		counts[r.Status]++
	}
	if _, err := fmt.Fprintf(out, "\nTests: %d  completed: %d  not run: %d\n", s.Total, len(s.Results), s.NotRun); err != nil {
		return err
	}
	for _, status := range []Status{Pass, Fail, Compile, Crash, Timeout, Cancelled, NoCheck, Skip} {
		if _, err := fmt.Fprintf(out, "%s: %d  ", status, counts[status]); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(out, "\nLogs: %s\n", s.LogDir)
	return err
}

// WriteJSON writes a versioned report via a temporary file in its destination
// directory, so interrupted writes do not replace an existing complete report.
func WriteJSON(path string, summary Summary) (err error) {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".easifem-report-*.json")
	if err != nil {
		return err
	}
	defer func() { _ = file.Close(); _ = os.Remove(file.Name()) }()
	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
