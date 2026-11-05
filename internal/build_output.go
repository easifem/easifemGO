package internal

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/easifem/easifemGO/internal/testrunner"
	"github.com/mattn/go-isatty"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

var verbose bool

func animatedOutput() bool {
	return !quiet && !verbose && os.Getenv("CI") == "" && os.Getenv("TERM") != "dumb" && isatty.IsTerminal(os.Stderr.Fd())
}

// NO_COLOR disables styling, including when present with an empty value.
func colorsAllowed(interactive bool) bool {
	_, disabled := os.LookupEnv("NO_COLOR")
	return interactive && !disabled
}

type stagePalette struct{ header, border string }

func paletteForStage(stage string, finished, failed bool) stagePalette {
	if failed {
		return stagePalette{"\x1b[1;91m", "\x1b[91m"}
	}
	if finished {
		return stagePalette{"\x1b[1;92m", "\x1b[92m"}
	}
	switch strings.ToLower(stage) {
	case "download":
		return stagePalette{"\x1b[1;96m", "\x1b[96m"}
	case "configure", "config":
		return stagePalette{"\x1b[1;94m", "\x1b[94m"}
	case "build":
		return stagePalette{"\x1b[1;95m", "\x1b[95m"}
	case "install":
		return stagePalette{"\x1b[1;93m", "\x1b[93m"}
	default:
		return stagePalette{"\x1b[1;94m", "\x1b[94m"}
	}
}

// Styling is applied only during painting, never to the layout strings. Column
// calculations and changed-text detection therefore remain ANSI-independent.
func (s *stageOutput) rowStyle(line string) string {
	if !s.color {
		return ""
	}
	palette := paletteForStage(s.stage, s.finished, s.failed)
	if strings.HasPrefix(line, "╭") || strings.HasPrefix(line, "├") {
		return palette.header
	}
	if strings.HasPrefix(line, "╰") {
		return palette.border
	}
	if progressBarPattern.MatchString(line) {
		return palette.header
	}
	return ""
}

// Plain layout strings keep bar styling out of width and diff calculations.
var progressBarPattern = regexp.MustCompile(`([█░]{16}) ([ 0-9]+%)`)

func (s *stageOutput) paintText(line string, column int, text, style string) string {
	if !s.color {
		return text
	}
	match := progressBarPattern.FindStringSubmatchIndex(line)
	if match == nil {
		if style == "" {
			return text
		}
		return style + text + "\x1b[0m"
	}
	start := runewidth.StringWidth(line[:match[2]])
	end := runewidth.StringWidth(line[:match[3]])
	percentStart := runewidth.StringWidth(line[:match[4]])
	percentEnd := runewidth.StringWidth(line[:match[5]])
	accent := paletteForStage(s.stage, s.finished, s.failed).header
	var out strings.Builder
	active := ""
	for _, r := range text {
		next := ""
		if column >= start && column < end {
			if r == '█' {
				next = accent
			} else {
				next = "\x1b[2;37m"
			}
		} else if column >= percentStart && column < percentEnd {
			next = accent
		}
		if next != active {
			if active != "" {
				out.WriteString("\x1b[0m")
			}
			out.WriteString(next)
			active = next
		}
		out.WriteRune(r)
		column += runewidth.RuneWidth(r)
	}
	if active != "" {
		out.WriteString("\x1b[0m")
	}
	return out.String()
}

// stageOutput owns a bounded terminal dashboard. Compiler output is buffered
// into its right panel instead of scrolling the progress panel off screen.
type stageOutput struct {
	mu               sync.Mutex
	out              io.Writer
	label            string
	started          time.Time
	stageStarted     time.Time
	completedStages  []string
	animate, silent  bool
	done             chan struct{}
	stopped          chan struct{}
	frame            int
	completed, total int64
	bytes            bool
	initialBytes     int64
	pending          []byte
	tail             []string
	detail           string
	log              io.Writer
	renderedRows     int
	previous         []string
	previousStyles   []string
	stage            string
	color            bool
	finished         bool
	failed           bool
	size             func() (int, int)
}

func newStageOutput(out io.Writer, label string, animate, silent bool) *stageOutput {
	s := &stageOutput{out: out, label: label, started: time.Now(), animate: animate && !silent, silent: silent, done: make(chan struct{}), stopped: make(chan struct{})}
	s.color = colorsAllowed(s.animate)
	s.stage = label
	if index := strings.LastIndex(label, " · "); index >= 0 {
		s.stage = label[index+len(" · "):]
	}
	if !silent && !s.animate {
		fmt.Fprintf(out, "  %s ...\n", label)
	}
	if s.animate {
		fmt.Fprint(s.out, "\x1b[?25l")
		s.render()
		go func() {
			defer close(s.stopped)
			tick := time.NewTicker(100 * time.Millisecond)
			defer tick.Stop()
			for {
				select {
				case <-s.done:
					return
				case <-tick.C:
					s.mu.Lock()
					s.render()
					s.mu.Unlock()
				}
			}
		}()
	} else {
		close(s.stopped)
	}
	return s
}

func terminalSize() (int, int) {
	w, h, err := term.GetSize(int(os.Stderr.Fd()))
	if err != nil || w < 2 || h < 2 {
		return 80, 24
	}
	return w, h
}

// terminalText strips compiler colors and control codes for stable column widths.
// Raw bytes remain unchanged in the log file.
var terminalEscape = regexp.MustCompile(`\x1b(?:\[[0-?]*[ -/]*[@-~]|\][^\x07\x1b]*(?:\x07|\x1b\\))`)

func terminalText(text string) string {
	text = terminalEscape.ReplaceAllString(text, "")
	return strings.Map(func(r rune) rune {
		if r == '\t' {
			return ' '
		}
		if r < 32 || (r >= 127 && r < 160) {
			return -1
		}
		return r
	}, text)
}

func panelCell(text string, width int) string {
	text = runewidth.Truncate(terminalText(text), width, "…")
	return text + strings.Repeat(" ", max(0, width-runewidth.StringWidth(text)))
}

func (s *stageOutput) progressLines() []string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	mark := frames[s.frame%len(frames)]
	if s.finished {
		mark = "✓"
		if s.failed {
			mark = "✗"
		}
	}
	lines := []string{fmt.Sprintf("%s %s", mark, s.label)}
	s.frame++
	if s.total > 0 {
		current := min(max(s.completed, 0), s.total)
		filled := int(float64(current) / float64(s.total) * 16)
		lines = append(lines, fmt.Sprintf("%s%s %3.0f%%", strings.Repeat("█", filled), strings.Repeat("░", 16-filled), float64(current)/float64(s.total)*100))
		if s.bytes {
			transferStart := s.stageStarted
			if transferStart.IsZero() {
				transferStart = s.started
			}
			speed := float64(current-s.initialBytes) / time.Since(transferStart).Seconds()
			lines = append(lines, fmt.Sprintf("%.1f / %.1f MB", float64(current)/1e6, float64(s.total)/1e6))
			if speed > 0 {
				eta := time.Duration(float64(s.total-current) / speed * float64(time.Second)).Round(time.Second)
				lines = append(lines, fmt.Sprintf("%.1f MB/s · ETA %s", speed/1e6, eta))
			}
		} else {
			lines = append(lines, fmt.Sprintf("%d / %d tasks", current, s.total))
		}
	}
	lines = append(lines, fmt.Sprintf("Elapsed: %s", time.Since(s.started).Round(time.Second)))
	if s.detail != "" {
		lines = append(lines, s.detail)
	}
	for _, stage := range s.completedStages {
		lines = append(lines, "✓ "+stage)
	}
	return lines
}

// dashboardLines is pure apart from the spinner frame, making narrow/wide
// layouts testable without a terminal or running a compiler.
func boxEdge(title string, width int, left, right string) string {
	inside := width - 2
	label := runewidth.Truncate(" "+title+" ", inside, "…")
	return left + label + strings.Repeat("─", max(0, inside-runewidth.StringWidth(label))) + right
}

func (s *stageOutput) dashboardLines(width, height int) []string {
	width = max(1, width-1)
	rows := min(11, max(1, height-1))
	left := s.progressLines()
	var lines []string
	if width >= 79 && rows >= 7 {
		leftWidth := min(40, width*2/5)
		rightWidth := width - leftWidth - 1
		lines = append(lines, boxEdge("PROGRESS", leftWidth, "╭", "╮")+" "+boxEdge("COMPILER OUTPUT", rightWidth, "╭", "╮"))
		visible := rows - 2
		recent := s.tail[max(0, len(s.tail)-visible):]
		for i := 0; i < visible; i++ {
			l, r := "", ""
			if i < len(left) {
				l = left[i]
			}
			if i < len(recent) {
				r = recent[i]
			}
			if len(recent) == 0 && i == 0 {
				r = "No messages"
			}
			lines = append(lines, "│ "+panelCell(l, leftWidth-4)+" │ │ "+panelCell(r, rightWidth-4)+" │")
		}
		lines = append(lines, "╰"+strings.Repeat("─", leftWidth-2)+"╯ ╰"+strings.Repeat("─", rightWidth-2)+"╯")
	} else if width >= 12 && rows >= 7 {
		// A single bordered card gives narrow terminals readable progress rows.
		lines = append(lines, boxEdge("PROGRESS", width, "╭", "╮"))
		for _, line := range left[:min(len(left), rows-4, 5)] {
			lines = append(lines, "│ "+panelCell(line, width-4)+" │")
		}
		lines = append(lines, boxEdge("OUTPUT", width, "├", "┤"))
		visible := rows - len(lines) - 1
		for _, line := range s.tail[max(0, len(s.tail)-visible):] {
			lines = append(lines, "│ "+panelCell(line, width-4)+" │")
		}
		for len(lines) < rows-1 {
			lines = append(lines, "│ "+panelCell("", width-4)+" │")
		}
		lines = append(lines, "╰"+strings.Repeat("─", width-2)+"╯")
	} else {
		lines = append(lines, panelCell(strings.Join(left, " · "), width))
		visible := rows - 1
		for _, line := range s.tail[max(0, len(s.tail)-visible):] {
			lines = append(lines, panelCell(line, width))
		}
		for len(lines) < rows {
			lines = append(lines, panelCell("", width))
		}
	}
	return lines
}

// changedText returns just the changed cell range. Matching suffixes (including
// the entire compiler panel when only a spinner changes) are never repainted.
func changedText(old, next string) (int, string) {
	a, b := []rune(old), []rune(next)
	prefix := 0
	for prefix < len(a) && prefix < len(b) && a[prefix] == b[prefix] {
		prefix++
	}
	endA, endB := len(a), len(b)
	if runewidth.StringWidth(old) == runewidth.StringWidth(next) {
		for endA > prefix && endB > prefix && a[endA-1] == b[endB-1] {
			endA--
			endB--
		}
	}
	return runewidth.StringWidth(string(b[:prefix])), string(b[prefix:endB])
}

func (s *stageOutput) render() {
	getSize := s.size
	if getSize == nil {
		getSize = terminalSize
	}
	width, height := getSize()
	lines := s.dashboardLines(width, height)
	s.paintDashboard(lines, width, height)
}

func (s *stageOutput) paintDashboard(lines []string, width, height int) {
	oldRows := min(s.renderedRows, max(1, height-1))
	areaRows := max(oldRows, len(lines))
	var frame strings.Builder
	// Reserve rows only when the layout grows. Refreshes never emit newlines.
	if areaRows > oldRows {
		frame.WriteString(strings.Repeat("\n", areaRows-max(1, oldRows)))
	}
	frame.WriteString("\r")
	if areaRows > 1 {
		fmt.Fprintf(&frame, "\x1b[%dA", areaRows-1)
	}
	cursorRow := 0
	changed := false
	styles := make([]string, len(lines))
	for i := 0; i < areaRows; i++ {
		old, next := "", ""
		if i < len(s.previous) {
			old = s.previous[i]
		}
		if i < len(lines) {
			next = lines[i]
		}
		style, oldStyle := s.rowStyle(next), ""
		if i < len(styles) {
			styles[i] = style
		}
		if i < len(s.previousStyles) {
			oldStyle = s.previousStyles[i]
		}
		if old == next && style == oldStyle {
			continue
		}
		// Space padding erases stale cells without ever blanking a whole row.
		next += strings.Repeat(" ", max(0, min(width-1, runewidth.StringWidth(old))-runewidth.StringWidth(next)))
		column, text := changedText(old, next)
		if style != oldStyle {
			column, text = 0, next
		}
		if text == "" {
			continue
		}
		changed = true
		if i > cursorRow {
			fmt.Fprintf(&frame, "\x1b[%dB", i-cursorRow)
		}
		cursorRow = i
		frame.WriteString("\r")
		if column > 0 {
			fmt.Fprintf(&frame, "\x1b[%dC", column)
		}
		frame.WriteString(s.paintText(next, column, text, style))
	}
	if !changed && oldRows == len(lines) {
		return
	}
	last := len(lines) - 1
	if cursorRow < last {
		fmt.Fprintf(&frame, "\x1b[%dB", last-cursorRow)
	}
	if cursorRow > last {
		fmt.Fprintf(&frame, "\x1b[%dA", cursorRow-last)
	}
	frame.WriteString("\r")
	// One write per frame avoids visible clear-then-paint intermediate states.
	io.WriteString(s.out, frame.String())
	s.previous = append(s.previous[:0], lines...)
	s.previousStyles = styles
	s.renderedRows = len(lines)
}

var ninjaProgress = regexp.MustCompile(`^\[([0-9]+)/([0-9]+)\](?:\s|$)`)

func parseBuildProgress(line string) (int64, int64, bool) {
	match := ninjaProgress.FindStringSubmatch(line)
	if match == nil {
		return 0, 0, false
	}
	current, e1 := strconv.ParseInt(match[1], 10, 64)
	total, e2 := strconv.ParseInt(match[2], 10, 64)
	return current, total, e1 == nil && e2 == nil && total > 0 && current <= total
}

func (s *stageOutput) Write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.log != nil {
		if _, err := s.log.Write(data); err != nil {
			return 0, err
		}
	}
	for _, b := range data {
		if b == '\n' || b == '\r' {
			s.line()
		} else {
			s.pending = append(s.pending, b)
			// Do not allow an unbounded compiler line to accumulate in memory.
			if len(s.pending) >= 16*1024 {
				s.line()
			}
		}
	}
	return len(data), nil
}

func (s *stageOutput) line() {
	if len(s.pending) == 0 {
		return
	}
	line := string(s.pending)
	s.pending = s.pending[:0]
	if n, total, ok := parseBuildProgress(line); ok && s.animate {
		s.completed, s.total = n, total
		s.detail = strings.TrimSpace(ninjaProgress.ReplaceAllString(line, ""))
		return
	}
	if s.silent || s.animate {
		s.tail = append(s.tail, line)
		if len(s.tail) > 20 {
			s.tail = s.tail[1:]
		}
	}
	if !s.silent && !s.animate {
		fmt.Fprintln(s.out, line)
	}
}

func (s *stageOutput) finish(err error) {
	close(s.done)
	<-s.stopped
	s.mu.Lock()
	defer s.mu.Unlock()
	s.line()
	if s.silent && err != nil {
		for _, line := range s.tail {
			fmt.Fprintln(s.out, line)
		}
	}
	if s.animate {
		s.finished, s.failed = true, err != nil
		s.render()
		fmt.Fprint(s.out, "\x1b[?25h\n")
		return
	}
	if !s.silent {
		mark := "✓"
		if err != nil {
			mark = "✗"
		}
		// Plain mode deliberately contains no terminal escape sequences.
		fmt.Fprintf(s.out, "  %s %s · %s\n", mark, s.label, time.Since(s.started).Round(100*time.Millisecond))
	}
}

// libraryOutput owns one dashboard across every stage of a package operation.
// It is passed explicitly so nested CMake/make/download paths share the display.
type libraryOutput struct {
	pkg     string
	animate bool
	display *stageOutput
	logs    []string
}

func newLibraryOutput(pkg string) *libraryOutput {
	return &libraryOutput{pkg: pkg, animate: animatedOutput()}
}

func (o *libraryOutput) begin(stage string) *stageOutput {
	if !o.animate {
		return nil
	}
	if o.display == nil {
		o.display = newStageOutput(os.Stderr, o.pkg+" · "+stage, true, false)
	} else {
		s := o.display
		s.mu.Lock()
		s.line()
		s.label = o.pkg + " · " + stage
		s.stage = stage
		s.completed, s.total, s.initialBytes = 0, 0, 0
		s.bytes = false
		s.detail = ""
		s.stageStarted = time.Now()
		s.render()
		s.mu.Unlock()
	}
	return o.display
}

func (o *libraryOutput) complete(stage string, err error) {
	if o.display == nil {
		return
	}
	s := o.display
	s.mu.Lock()
	defer s.mu.Unlock()
	s.line()
	s.log = nil
	if err == nil {
		s.completedStages = append(s.completedStages, stage)
	}
}

func (o *libraryOutput) finish(err error) {
	if o.display == nil {
		return
	}
	if err == nil {
		o.display.mu.Lock()
		o.display.label = o.pkg + " · Complete"
		o.display.mu.Unlock()
	}
	o.display.finish(err)
	if len(o.logs) > 0 {
		fmt.Fprintln(os.Stderr, "  Stage logs:")
		for _, path := range o.logs {
			fmt.Fprintln(os.Stderr, "   ", path)
		}
	}
}

func (o *libraryOutput) run(args []string, stage, dir string) error {
	return runBuildStageWithOutput(args, o.pkg, stage, dir, o)
}

func runBuildStage(args []string, pkg, stage, dir string) error {
	return runBuildStageWithOutput(args, pkg, stage, dir, nil)
}

func runBuildStageWithOutput(args []string, pkg, stage, dir string, operation *libraryOutput) error {
	if len(args) == 0 {
		return fmt.Errorf("empty %s command", stage)
	}
	var display *stageOutput
	if operation != nil {
		display = operation.begin(stage)
	}
	logFile, err := os.CreateTemp("", "easifem-"+strings.ToLower(stage)+"-*.log")
	if err != nil {
		return fmt.Errorf("%s: create %s log: %w", pkg, stage, err)
	}
	shared := display != nil
	if !shared {
		display = newStageOutput(os.Stderr, pkg+" · "+stage, animatedOutput(), quiet)
	}
	display.mu.Lock()
	display.log = logFile
	display.mu.Unlock()
	if verbose && !quiet {
		fmt.Fprintf(os.Stderr, "  command: %q\n  directory: %s\n", args, dir)
	}
	_, logErr := fmt.Fprintf(logFile, "command: %q\nworking directory: %s\n", args, dir)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if logErr == nil {
		err = testrunner.RunCommand(ctx, dir, display, nil, args...)
	} else {
		err = logErr
	}
	if closeErr := logFile.Close(); err == nil {
		err = closeErr
	}
	if shared {
		operation.complete(stage, err)
		operation.logs = append(operation.logs, logFile.Name())
	} else {
		display.finish(err)
	}
	if err != nil {
		return fmt.Errorf("%s: %s failed: %w\nFull log: %s", pkg, stage, err, logFile.Name())
	}
	if !shared && (verbose || (display.animate && len(display.tail) > 0)) && !quiet {
		fmt.Fprintln(os.Stderr, "  Log:", logFile.Name())
	}
	return nil
}
