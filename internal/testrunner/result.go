// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package testrunner

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Status string

const (
	Pass      Status = "PASS"
	Fail      Status = "FAIL"
	Compile   Status = "COMPILE"
	Crash     Status = "CRASH"
	Timeout   Status = "TIMEOUT"
	Cancelled Status = "CANCELLED"
	NoCheck   Status = "NOCHECK"
	Skip      Status = "SKIP"
)

type Result struct {
	Test       Test          `json:"test"`
	Status     Status        `json:"status"`
	Passed     int           `json:"passed_checks"`
	Failed     int           `json:"failed_checks"`
	Skipped    int           `json:"skipped_checks"`
	Log        string        `json:"log"`
	Executable string        `json:"executable,omitempty"`
	Duration   time.Duration `json:"duration_ns"`
	Error      string        `json:"error,omitempty"`
}

type Summary struct {
	Version  int           `json:"version"`
	Started  time.Time     `json:"started"`
	Duration time.Duration `json:"duration_ns"`
	Total    int           `json:"total"`
	NotRun   int           `json:"not_run"`
	LogDir   string        `json:"log_directory"`
	Results  []Result      `json:"results"`
}

// Bad counts unsuccessful programs; explicitly skipped programs are acceptable.
func (s Summary) Bad(allowNoCheck bool) int {
	bad := s.NotRun
	for _, r := range s.Results {
		if r.Status != Pass && r.Status != Skip && !(allowNoCheck && r.Status == NoCheck) {
			bad++
		}
	}
	return bad
}

var checkLine = regexp.MustCompile(`^(not ok|ok)(?:\s|$)`)
var checkNumber = regexp.MustCompile(`^(?:not ok|ok)\s+([0-9]+)(?:\s|$)`)
var planLine = regexp.MustCompile(`^1\.\.([0-9]+)(?:\s|$)`)
var skipDirective = regexp.MustCompile(`(?i)#\s*SKIP(?:\s|$)`)
var todoDirective = regexp.MustCompile(`(?i)#\s*TODO(?:\s|$)`)
var ansiSequence = regexp.MustCompile("\x1b\\[[0-?]*[ -/]*[@-~]")

// checkOutput parses flat TAP/Test_Method results from execution only.
// Full output goes to disk; at most 64 KiB of a single line is retained here.
type checkOutput struct {
	mu                               sync.Mutex
	line                             []byte
	overflow                         bool
	passed, failed, skipped, records int
	plan                             int
	hasPlan                          bool
	skipAll                          bool
	problem                          string
	runtimeError                     bool
}

func (p *checkOutput) Write(data []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, b := range data {
		if b == '\n' {
			p.consume()
			continue
		}
		if len(p.line) < 64*1024 {
			p.line = append(p.line, b)
		} else {
			p.overflow = true
		}
	}
	return len(data), nil
}

func (p *checkOutput) finish() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.line) > 0 || p.overflow {
		p.consume()
	}
	if p.hasPlan && p.plan != p.records {
		p.problem = fmt.Sprintf("TAP plan expected %d checks; received %d", p.plan, p.records)
	}
}

func (p *checkOutput) consume() {
	line := strings.TrimSpace(ansiSequence.ReplaceAllString(string(p.line), ""))
	p.line = p.line[:0]
	overflow := p.overflow
	p.overflow = false
	if strings.Contains(line, "Fortran runtime error:") || strings.HasPrefix(line, "Program received signal") || strings.HasPrefix(line, "ERROR STOP") {
		p.runtimeError = true
	}
	if strings.HasPrefix(line, "Bail out!") {
		p.problem = line
		return
	}
	if match := planLine.FindStringSubmatch(line); match != nil {
		n, err := strconv.Atoi(match[1])
		if err != nil {
			p.problem = "invalid TAP plan"
			return
		}
		if p.hasPlan {
			p.problem = "multiple TAP plans are unsupported"
			return
		}
		p.plan, p.hasPlan = n, true
		p.skipAll = n == 0 && skipDirective.MatchString(line)
		return
	}
	match := checkLine.FindStringSubmatch(line)
	if match == nil {
		return
	}
	p.records++
	if number := checkNumber.FindStringSubmatch(line); number != nil {
		n, err := strconv.Atoi(number[1])
		if err != nil || n != p.records {
			p.problem = "TAP check numbers are not sequential"
		}
	}
	if overflow {
		p.problem = "TAP check line exceeded 64 KiB"
	}
	if skipDirective.MatchString(line) || todoDirective.MatchString(line) {
		p.skipped++
	} else if match[1] == "not ok" {
		p.failed++
	} else {
		p.passed++
	}
}

func (p *checkOutput) status() (Status, string) {
	if p.runtimeError {
		return Crash, "runtime-error diagnostic in execution output"
	}
	if p.problem != "" {
		return Fail, p.problem
	}
	if p.failed > 0 {
		return Fail, ""
	}
	if p.passed > 0 {
		return Pass, ""
	}
	if p.skipped > 0 || p.skipAll {
		return Skip, ""
	}
	return NoCheck, "program exited successfully but reported no checks"
}
