// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package internal

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfigureArgsRunsScriptDirectlyOffWindows(t *testing.T) {
	src := filepath.Join("src", "lapack95")
	inst := filepath.Join("inst", "lapack95")
	got := configureArgs("linux", src, inst, []string{"--enable-shared"})
	want := []string{filepath.Join(src, "configure"), "--prefix=" + inst, "--enable-shared"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("configureArgs(linux) = %q, want %q", got, want)
	}
}

func TestConfigureArgsUsesShOnWindows(t *testing.T) {
	src := filepath.Join("src", "lapack95")
	inst := filepath.Join("inst", "lapack95")
	got := configureArgs("windows", src, inst, []string{"--enable-shared"})
	want := []string{
		"sh",
		filepath.ToSlash(filepath.Join(src, "configure")),
		"--prefix=" + filepath.ToSlash(inst),
		"--enable-shared",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("configureArgs(windows) = %q, want %q", got, want)
	}
}

func TestConfigureArgsWithoutBuildOptions(t *testing.T) {
	got := configureArgs("windows", "src", "inst", nil)
	want := []string{"sh", "src/configure", "--prefix=inst"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("configureArgs(windows, no options) = %q, want %q", got, want)
	}
}
