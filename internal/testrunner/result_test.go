// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package testrunner

import (
	"strings"
	"testing"
)

func TestFlatCheckClassification(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		want         Status
	}{
		{"passing", "1..2\nok 1 - first\nok 2 - second\n", Pass},
		{"assertion", "ok 1\nnot ok 2 - incorrect\n", Fail},
		{"incomplete plan", "1..2\nok 1\n", Fail},
		{"duplicate check", "ok 1\nok 1\n", Fail},
		{"multiple plans", "1..1\nok 1\n1..1\n", Fail},
		{"bail out", "ok 1\nBail out! no mesh\n", Fail},
		{"runtime diagnostic", "ok 1\nFortran runtime error: bounds\n", Crash},
		{"unchecked", "simulation completed\n", NoCheck},
		{"skipped plan", "1..0 # SKIP unavailable\n", Skip},
		{"skipped checks", "ok 1 # SKIP optional\nnot ok 2 # TODO later\n", Skip},
		{"ANSI and final line", "\x1b[32mok 1 - value\x1b[0m", Pass},
		{"large data line", strings.Repeat("x", 100000) + "\nok 1\n", Pass},
		{"oversized check", "ok 1 - " + strings.Repeat("x", 100000), Fail},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var checks checkOutput
			// Split output into small chunks to exercise streaming boundaries.
			for _, b := range []byte(tc.output) {
				_, _ = checks.Write([]byte{b})
			}
			checks.finish()
			got, reason := checks.status()
			if got != tc.want {
				t.Fatalf("got %s (%s), want %s", got, reason, tc.want)
			}
		})
	}
}
