package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/mattn/go-runewidth"
)

func TestParseBuildProgress(t *testing.T) {
	for _, tc := range []struct {
		line     string
		n, total int64
		ok       bool
	}{
		{"[42/180] Building object", 42, 180, true},
		{"[180/180] Linking", 180, 180, true},
		{"[0/1] Checking globs", 0, 1, true},
		{"[2/1] invalid", 2, 1, false},
		{"[0/0] invalid", 0, 0, false},
		{"Warning: compiler diagnostic", 0, 0, false},
	} {
		n, total, ok := parseBuildProgress(tc.line)
		if ok != tc.ok || (ok && (n != tc.n || total != tc.total)) {
			t.Fatalf("%q: got %d/%d %v", tc.line, n, total, ok)
		}
	}
}

func TestStageKeepsDiagnosticsAndRawLog(t *testing.T) {
	var screen, log bytes.Buffer
	stage := newStageOutput(&screen, "classes · Build", false, false)
	// Enable rendering without a ticker to make the display deterministic.
	stage.animate = true
	stage.size = func() (int, int) { return 120, 24 }
	stage.log = &log
	parts := []string{"[4", "/10] Compiling test.F90\r", "file.F90:12: Warn", "ing: mismatch\n", "final diagnostic"}
	for _, part := range parts {
		if _, err := stage.Write([]byte(part)); err != nil {
			t.Fatal(err)
		}
	}
	if stage.completed != 4 || stage.total != 10 {
		t.Fatal("fragmented Ninja progress not recognized")
	}
	stage.finish(nil)
	if log.String() != strings.Join(parts, "") {
		t.Fatal("raw log modified")
	}
	for _, text := range []string{"Warning: mismatch", "final diagnostic", "✓ classes"} {
		if !strings.Contains(screen.String(), text) {
			t.Fatalf("missing %q in %q", text, screen.String())
		}
	}
	if strings.Contains(screen.String(), "[4/10]") {
		t.Fatal("progress flooded the terminal")
	}
}

func TestStagePlainAndQuietOutput(t *testing.T) {
	for _, silent := range []bool{false, true} {
		var screen bytes.Buffer
		stage := newStageOutput(&screen, "Build", false, silent)
		stage.Write([]byte("[1/2] Compile\nWarning: example\n"))
		stage.finish(nil)
		if strings.ContainsAny(screen.String(), "\r\x1b") {
			t.Fatal("plain output has terminal controls")
		}
		if silent && screen.Len() != 0 {
			t.Fatal("quiet output was not silent")
		}
		if !silent && !strings.Contains(screen.String(), "Warning: example") {
			t.Fatal("diagnostic lost")
		}
	}
	var screen bytes.Buffer
	stage := newStageOutput(&screen, "Build", false, true)
	stage.Write([]byte("Error: example\n"))
	stage.finish(errors.New("failed"))
	if !strings.Contains(screen.String(), "Error: example") {
		t.Fatal("quiet failure lost diagnostics")
	}
}

func TestStageBoundsPendingLine(t *testing.T) {
	stage := newStageOutput(io.Discard, "Build", false, true)
	stage.Write(bytes.Repeat([]byte{'x'}, 100*1024))
	if len(stage.pending) >= 16*1024 || len(stage.tail) > 20 {
		t.Fatal("unbounded output buffering")
	}
	stage.finish(nil)
}

type trackedStream struct {
	io.Reader
	closes int
}

func (s *trackedStream) Close() error { s.closes++; return nil }

func TestDownloadTracksBytesAndClosesStream(t *testing.T) {
	for _, total := range []int64{0, 10} {
		stage := newStageOutput(io.Discard, "Download", false, true)
		stream := &trackedStream{Reader: strings.NewReader("abc")}
		reader := (&ProgressBar{stage: stage}).TrackProgress("fixture", 2, total, stream)
		if _, err := io.Copy(io.Discard, reader); err != nil {
			t.Fatal(err)
		}
		if err := reader.Close(); err != nil {
			t.Fatal(err)
		}
		reader.Close()
		if stream.closes != 1 || stage.completed != 5 || stage.total != total {
			t.Fatalf("closes=%d progress=%d/%d", stream.closes, stage.completed, stage.total)
		}
		stage.finish(nil)
	}
}

func TestAnimationDisabledForPipelinesAndCI(t *testing.T) {
	oldQuiet, oldVerbose := quiet, verbose
	t.Cleanup(func() { quiet, verbose = oldQuiet, oldVerbose })
	quiet, verbose = false, false
	t.Setenv("CI", "true")
	if animatedOutput() {
		t.Fatal("CI enabled animation")
	}
	t.Setenv("CI", "")
	t.Setenv("TERM", "dumb")
	if animatedOutput() {
		t.Fatal("dumb terminal enabled animation")
	}
}

func TestStageRenderShowsMeasuredProgress(t *testing.T) {
	var screen bytes.Buffer
	stage := newStageOutput(&screen, "Build", false, false)
	stage.completed, stage.total = 4, 10
	stage.size = func() (int, int) { return 120, 24 }
	stage.render()
	if !strings.Contains(screen.String(), "40%") || !strings.Contains(screen.String(), "4 / 10 tasks") {
		t.Fatalf("missing measured progress: %q", screen.String())
	}
	screen.Reset()
	stage.total = 0
	stage.render()
	if strings.Contains(screen.String(), "%") {
		t.Fatal("unknown total displayed a percentage")
	}
	stage.finish(nil)
}

type failedLog struct{}

func (failedLog) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestStagePropagatesLogFailure(t *testing.T) {
	stage := newStageOutput(io.Discard, "Build", false, true)
	stage.log = failedLog{}
	if _, err := stage.Write([]byte("diagnostic\n")); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("log write error lost: %v", err)
	}
	stage.finish(io.ErrClosedPipe)
}

func TestDashboardUsesTwoBoundedPanels(t *testing.T) {
	stage := &stageOutput{label: "classes · Build", started: time.Now(), completed: 42, total: 180,
		tail: []string{"mesh.F90:12: Warning: mismatch", strings.Repeat("long diagnostic ", 100)},
	}
	for _, size := range [][2]int{{80, 24}, {120, 40}, {60, 24}, {20, 8}, {2, 2}} {
		lines := stage.dashboardLines(size[0], size[1])
		if len(lines) > size[1]-1 {
			t.Fatalf("layout exceeded height %v", size)
		}
		for _, line := range lines {
			if runewidth.StringWidth(line) > size[0]-1 {
				t.Fatalf("layout exceeded width %v: %q", size, line)
			}
		}
		if size[0] >= 80 && !strings.Contains(lines[0], "PROGRESS") {
			t.Fatal("wide terminal missing progress panel")
		}
		if size[0] >= 80 && !strings.Contains(strings.Join(lines, "\n"), "│ │") {
			t.Fatal("wide terminal missing divider")
		}
	}
}

func TestWarningsDoNotScrollProgress(t *testing.T) {
	var screen bytes.Buffer
	stage := newStageOutput(&screen, "classes · Build", false, false)
	stage.animate = true
	stage.size = func() (int, int) { return 120, 24 }
	stage.render()
	rows := stage.renderedRows
	screen.Reset()
	for i := 0; i < 100; i++ {
		stage.Write([]byte(fmt.Sprintf("module.F90:%d: Warning: example\n", i)))
	}
	if screen.Len() != 0 {
		t.Fatal("compiler output printed outside dashboard")
	}
	if len(stage.tail) != 20 {
		t.Fatal("recent output was not bounded")
	}
	stage.render()
	if stage.renderedRows != rows {
		t.Fatal("warnings changed the dashboard height")
	}
	if !strings.Contains(screen.String(), "Warning: example") || !strings.Contains(screen.String(), fmt.Sprintf("\x1b[%dA", rows-1)) {
		t.Fatal("dashboard did not redraw warnings in the reserved area")
	}
	stage.finish(nil)
}

func TestDashboardStripsControlsButKeepsRawLog(t *testing.T) {
	input := "\x1b[31mWarning: red\x1b[0m\x1b]0;title\x07\x1b]8;;https://example.com\x1b\\link\x1b]8;;\x1b\\"
	if got := terminalText(input); got != "Warning: redlink" {
		t.Fatalf("controls not removed: %q", got)
	}
	var raw bytes.Buffer
	stage := newStageOutput(io.Discard, "Build", false, true)
	stage.log = &raw
	stage.Write([]byte(input + "\n"))
	stage.finish(nil)
	if raw.String() != input+"\n" {
		t.Fatal("dashboard sanitizing changed the raw log")
	}
}

func TestDashboardHasClosedBoxBorders(t *testing.T) {
	stage := &stageOutput{label: "classes · Build", started: time.Now()}
	for _, width := range []int{80, 120} {
		lines := stage.dashboardLines(width, 24)
		if !strings.HasPrefix(lines[0], "╭") || !strings.HasSuffix(lines[0], "╮") || !strings.Contains(lines[0], "╮ ╭") {
			t.Fatal("missing rounded top borders", lines[0])
		}
		last := lines[len(lines)-1]
		if !strings.HasPrefix(last, "╰") || !strings.HasSuffix(last, "╯") {
			t.Fatal("missing bottom borders", last)
		}
		for _, line := range lines {
			if runewidth.StringWidth(line) != width-1 {
				t.Fatalf("uneven box width: %q", line)
			}
		}
	}
}

func TestDashboardUpdatesOnlyChangedText(t *testing.T) {
	var screen bytes.Buffer
	stage := &stageOutput{out: &screen}
	first := []string{"╭─────╮ ╭─────────────────╮", "│ ⠋   │ │ Warning: stable │", "╰─────╯ ╰─────────────────╯"}
	next := append([]string(nil), first...)
	next[1] = "│ ⠙   │ │ Warning: stable │"
	stage.paintDashboard(first, 80, 24)
	screen.Reset()
	stage.paintDashboard(next, 80, 24)
	output := screen.String()
	if strings.Contains(output, "Warning") || strings.Contains(output, "╭") || strings.Contains(output, "\x1b[2K") || strings.Contains(output, "\n") {
		t.Fatalf("refresh repainted static content or cleared rows: %q", output)
	}
	if !strings.Contains(output, "⠙") {
		t.Fatal("spinner did not update")
	}
	screen.Reset()
	stage.paintDashboard(next, 80, 24)
	if screen.Len() != 0 {
		t.Fatal("unchanged frame produced terminal output")
	}
}

func TestDashboardErasesShorterTextWithoutClearing(t *testing.T) {
	var screen bytes.Buffer
	stage := &stageOutput{out: &screen}
	stage.paintDashboard([]string{"long message"}, 80, 24)
	screen.Reset()
	stage.paintDashboard([]string{"short"}, 80, 24)
	if !strings.Contains(screen.String(), "short       ") || strings.Contains(screen.String(), "\x1b[2K") {
		t.Fatalf("stale text not overwritten: %q", screen.String())
	}
}

func TestCompletionPreservesBoxesWithoutReplayingDiagnostics(t *testing.T) {
	var screen bytes.Buffer
	stage := newStageOutput(&screen, "classes · Build", false, false)
	stage.animate = true
	stage.size = func() (int, int) { return 120, 24 }
	stage.Write([]byte("Warning: retained inside box\n"))
	stage.render()
	screen.Reset()
	stage.finish(nil)
	if strings.Contains(screen.String(), "Warning:") || strings.Contains(screen.String(), "\x1b[2K") {
		t.Fatal("completion replayed diagnostics or cleared display")
	}
	if !strings.Contains(screen.String(), "✓") || !strings.Contains(screen.String(), "\x1b[?25h") {
		t.Fatal("completion did not mark success and restore cursor")
	}
}

func TestPackageStagesReuseDashboard(t *testing.T) {
	var screen bytes.Buffer
	display := newStageOutput(&screen, "classes · Configure", false, false)
	display.animate = true
	display.size = func() (int, int) { return 120, 24 }
	operation := &libraryOutput{pkg: "classes", animate: true, display: display}
	display.render()
	display.Write([]byte("Warning: preserved from configure\n"))
	display.pending = append(display.pending, []byte("unterminated diagnostic")...)
	display.completed, display.total = 3, 4
	display.detail = "configure task"
	display.bytes = true
	operation.complete("Configure", nil)
	screen.Reset()
	if got := operation.begin("Build"); got != display {
		t.Fatal("stage replaced the dashboard")
	}
	if display.completed != 0 || display.total != 0 || display.bytes || display.detail != "" {
		t.Fatal("stage counters were not reset")
	}
	if len(display.tail) != 2 {
		t.Fatal("stage transition discarded compiler output")
	}
	if display.finished {
		t.Fatal("stage transition finished the operation")
	}
	operation.complete("Build", nil)
	if got := operation.begin("Install"); got != display {
		t.Fatal("install replaced the dashboard")
	}
	if !reflect.DeepEqual(display.completedStages, []string{"Configure", "Build"}) {
		t.Fatal(display.completedStages)
	}
	// Neither transition reserves more rows, redraws borders, nor restores cursor.
	output := screen.String()
	for _, unwanted := range []string{"\n", "╭", "╯", "\x1b[?25h"} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("stage transition recreated the display (%q): %q", unwanted, output)
		}
	}
	screen.Reset()
	operation.complete("Install", nil)
	operation.finish(nil)
	if !display.finished || display.failed || !strings.Contains(screen.String(), "\x1b[?25h") {
		t.Fatal("operation did not finish and restore cursor")
	}
	if !reflect.DeepEqual(display.completedStages, []string{"Configure", "Build", "Install"}) {
		t.Fatal(display.completedStages)
	}
}

func TestPackageStageFailureFinishesSharedDisplay(t *testing.T) {
	var screen bytes.Buffer
	display := newStageOutput(&screen, "classes · Build", false, false)
	display.animate = true
	display.size = func() (int, int) { return 120, 24 }
	operation := &libraryOutput{pkg: "classes", animate: true, display: display}
	display.render()
	failure := errors.New("compiler failed")
	operation.complete("Build", failure)
	operation.finish(failure)
	if !display.finished || !display.failed || len(display.completedStages) != 0 {
		t.Fatal("failed stage marked complete")
	}
	if display.label != "classes · Build" {
		t.Fatal("failed stage name lost")
	}
	if !strings.Contains(screen.String(), "\x1b[?25h") {
		t.Fatal("failure left cursor hidden")
	}
}

func TestStageHeaderColorsChangeWithoutNewDashboard(t *testing.T) {
	var screen bytes.Buffer
	display := newStageOutput(&screen, "classes · Configure", false, false)
	display.animate, display.color = true, true
	display.size = func() (int, int) { return 120, 24 }
	operation := &libraryOutput{pkg: "classes", animate: true, display: display}
	display.render()
	if !strings.Contains(screen.String(), "\x1b[1;94m") {
		t.Fatal("configure header missing blue text")
	}
	for _, tc := range []struct{ stage, header string }{
		{"Build", "\x1b[1;95m"},
		{"Install", "\x1b[1;93m"},
	} {
		screen.Reset()
		operation.begin(tc.stage)
		if !strings.Contains(screen.String(), tc.header) {
			t.Fatalf("%s header text color missing", tc.stage)
		}
		if strings.Contains(screen.String(), "\n") || strings.Contains(screen.String(), "\x1b[2K") {
			t.Fatal("theme change recreated or cleared dashboard")
		}
	}
	screen.Reset()
	operation.finish(nil)
	if !strings.Contains(screen.String(), "\x1b[1;92m") {
		t.Fatal("completion missing green header")
	}
	if !strings.Contains(screen.String(), "\x1b[0m") {
		t.Fatal("header style leaked into terminal")
	}
}

func TestColorThemesDoNotRepaintHeadersOnSpinnerTick(t *testing.T) {
	var screen bytes.Buffer
	display := &stageOutput{out: &screen, color: true, stage: "Build"}
	first := []string{"╭ PROGRESS ─╮", "│ ⠋ Build  │", "╰───────────╯"}
	display.paintDashboard(first, 80, 24)
	screen.Reset()
	next := append([]string(nil), first...)
	next[1] = "│ ⠙ Build  │"
	display.paintDashboard(next, 80, 24)
	if strings.Contains(screen.String(), "\x1b[1;") || strings.Contains(screen.String(), "╭") {
		t.Fatal("spinner tick repainted header")
	}
	if !strings.Contains(screen.String(), "⠙") {
		t.Fatal("spinner did not advance")
	}
}

func TestStagePaletteAndNoColor(t *testing.T) {
	headers := map[string]bool{}
	for _, stage := range []string{"Download", "Configure", "Build", "Install"} {
		palette := paletteForStage(stage, false, false)
		if headers[palette.header] {
			t.Fatal("stages share a header text color")
		}
		headers[palette.header] = true
	}
	if paletteForStage("Build", true, true).header != "\x1b[1;91m" {
		t.Fatal("failure did not override stage color")
	}
	// Presence of NO_COLOR, including an empty value, disables all styling.
	for _, value := range []string{"", "1"} {
		t.Setenv("NO_COLOR", value)
		if colorsAllowed(true) {
			t.Fatal("NO_COLOR ignored")
		}
	}
	if colorsAllowed(false) {
		t.Fatal("noninteractive styling enabled")
	}
	var screen bytes.Buffer
	display := &stageOutput{out: &screen, color: false, stage: "Build"}
	display.paintDashboard([]string{"╭ PROGRESS ─╮", "╰───────────╯"}, 80, 24)
	if regexp.MustCompile(`\x1b\[[0-9;]*m`).MatchString(screen.String()) {
		t.Fatal("monochrome render emitted a color sequence")
	}
}

func TestStylingPreservesLayoutAndRawLogs(t *testing.T) {
	var screen, raw bytes.Buffer
	display := newStageOutput(&screen, "classes · Build", false, false)
	display.animate, display.color = true, true
	display.size = func() (int, int) { return 120, 24 }
	display.log = &raw
	diagnostic := "Warning: original compiler output\n"
	display.Write([]byte(diagnostic))
	for _, line := range display.dashboardLines(120, 24) {
		if strings.Contains(line, "\x1b") || runewidth.StringWidth(line) != 119 {
			t.Fatal("styling corrupted layout measurements")
		}
	}
	display.render()
	display.finish(errors.New("build failed"))
	if raw.String() != diagnostic {
		t.Fatal("dashboard colors polluted raw logs")
	}
	if !strings.Contains(screen.String(), "\x1b[1;91m") {
		t.Fatal("failure header missing red text")
	}
}

func TestBoldProgressBarHasMeasuredFill(t *testing.T) {
	for _, tc := range []struct {
		done    int64
		filled  int
		percent string
	}{
		{0, 0, "0%"}, {5, 8, "50%"}, {10, 16, "100%"},
	} {
		display := &stageOutput{started: time.Now(), completed: tc.done, total: 10}
		lines := display.progressLines()
		bar := lines[1]
		if strings.Count(bar, "█") != tc.filled || strings.Count(bar, "░") != 16-tc.filled || !strings.Contains(bar, tc.percent) {
			t.Fatalf("incorrect progress bar: %q", bar)
		}
		if strings.Contains(bar, "\x1b") {
			t.Fatal("layout bar contains styling")
		}
	}
}

func TestBarColorsStayWithinProgressContent(t *testing.T) {
	display := &stageOutput{color: true, stage: "Build"}
	line := "│ ████████░░░░░░░░  50% │ │ Warning: unchanged │"
	painted := display.paintText(line, 0, line, display.rowStyle(line))
	if terminalText(painted) != line {
		t.Fatal("coloring changed bar text or width")
	}
	if !strings.Contains(painted, "\x1b[1;95m████████") || !strings.Contains(painted, "\x1b[2;37m░░░░░░░░") {
		t.Fatal("missing bold fill or subdued track")
	}
	if !strings.Contains(painted, "\x1b[0m │ │ Warning: unchanged │") {
		t.Fatal("bar colors leaked into compiler panel")
	}
	if regexp.MustCompile(`\x1b\[[0-9;]*(?:4[0-9]|10[0-7])m`).MatchString(painted) {
		t.Fatal("bar introduced a background color")
	}
	display.color = false
	if got := display.paintText(line, 0, line, ""); got != line {
		t.Fatal("monochrome bar contains styling")
	}
}

func TestBarIncrementKeepsDiffRendering(t *testing.T) {
	var screen bytes.Buffer
	display := &stageOutput{out: &screen, color: true, stage: "Build"}
	display.paintDashboard([]string{"│ ████████░░░░░░░░  50% │ │ Warning: stable │"}, 80, 24)
	screen.Reset()
	display.paintDashboard([]string{"│ █████████░░░░░░░  56% │ │ Warning: stable │"}, 80, 24)
	if strings.Contains(screen.String(), "Warning") || strings.Contains(screen.String(), "\n") || strings.Contains(screen.String(), "\x1b[2K") {
		t.Fatal("bar increment repainted compiler output or cleared the display")
	}
	if !strings.Contains(screen.String(), "\x1b[1;95m█") {
		t.Fatal("increment lost its bold stage color")
	}
}

func TestLibraryLogLocationAndStageRetention(t *testing.T) {
	buildDir := filepath.Join(t.TempDir(), "package-build")
	operation := newLibraryOutput("classes", buildDir, "build")
	operation.animate = false
	for _, text := range []string{"configure output\n", "build output\n"} {
		file, err := operation.openLog()
		if err != nil {
			t.Fatal(err)
		}
		if file.Name() != filepath.Join(buildDir, "easifem-build-log.txt") {
			t.Fatal("unexpected log location", file.Name())
		}
		if _, err := file.WriteString(text); err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(buildDir, "easifem-build-log.txt"))
	if err != nil || string(data) != "configure output\nbuild output\n" {
		t.Fatalf("stage output was lost: %q %v", data, err)
	}
	if len(operation.logs) != 1 {
		t.Fatal("same log listed more than once", operation.logs)
	}
}

func TestNewLibraryInvocationReplacesOnlyItsOwnLog(t *testing.T) {
	buildDir := t.TempDir()
	for _, mode := range []string{"build", "install"} {
		operation := newLibraryOutput("classes", buildDir, mode)
		file, err := operation.openLog()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.WriteString("old " + mode + " output\n"); err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	operation := newLibraryOutput("classes", buildDir, "build")
	file, err := operation.openLog()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("latest build output\n"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"easifem-build-log.txt": "latest build output\n", "easifem-install-log.txt": "old install output\n"} {
		data, err := os.ReadFile(filepath.Join(buildDir, name))
		if err != nil || string(data) != want {
			t.Fatalf("unexpected %s contents: %q %v", name, data, err)
		}
	}
}

func TestLibraryLogSetupErrors(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocker, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"", blocker, filepath.Join(blocker, "build")} {
		operation := newLibraryOutput("classes", dir, "install")
		if file, err := operation.openLog(); err == nil {
			file.Close()
			t.Fatalf("invalid log directory accepted: %q", dir)
		}
		if operation.logStarted {
			t.Fatal("failed log setup was marked started")
		}
	}
}
