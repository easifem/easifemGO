// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func libraryPathOf(t *testing.T, env []string) []string {
	t.Helper()
	prefix := libraryPathVar() + "="
	var found []string
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			found = append(found, strings.TrimPrefix(entry, prefix))
		}
	}
	if len(found) != 1 {
		t.Fatalf("want exactly one %s entry, got %q", libraryPathVar(), found)
	}
	return filepath.SplitList(found[0])
}

func TestSelectedEnvironmentPrependsSelectedLibraries(t *testing.T) {
	sep := string(os.PathListSeparator)
	base := []string{"HOME=/home/u", libraryPathVar() + "=/release/base/lib" + sep + "/opt/lib"}
	env := selectedEnvironment(base, []string{"/dev/base/lib", "/dev/classes/lib", "/opt/lib"})
	got := libraryPathOf(t, env)
	want := []string{"/dev/base/lib", "/dev/classes/lib", "/opt/lib", "/release/base/lib"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("library path = %q, want %q", got, want)
	}
	if env[0] != "HOME=/home/u" {
		t.Fatalf("other variables must be kept, got %q", env)
	}
}

func TestSelectedEnvironmentWithoutExistingPath(t *testing.T) {
	t.Setenv("EASIFEM_TEST_PREFIX", "/x")
	env := selectedEnvironment([]string{"A=b"}, []string{"$EASIFEM_TEST_PREFIX/lib"})
	if got := libraryPathOf(t, env); len(got) != 1 || got[0] != "/x/lib" {
		t.Fatalf("library path = %q, want [/x/lib]", got)
	}
	if env := selectedEnvironment([]string{"A=b"}, nil); len(env) != 1 {
		t.Fatalf("no library directories must not add an empty variable, got %q", env)
	}
}

func TestEnvStampChanged(t *testing.T) {
	dir := t.TempDir()
	dev := envStamp("dev", []string{"/dev/install"})
	release := envStamp("release", []string{"/install"})

	if envStampChanged(dir, dev) {
		t.Fatal("an unconfigured build directory needs no fresh configuration")
	}
	if err := os.WriteFile(filepath.Join(dir, "CMakeCache.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if !envStampChanged(dir, dev) {
		t.Fatal("a configured build directory without a stamp has unknown origin")
	}
	if err := writeEnvStamp(dir, dev); err != nil {
		t.Fatal(err)
	}
	if envStampChanged(dir, dev) {
		t.Fatal("same environment must reuse the CMake cache")
	}
	if !envStampChanged(dir, release) {
		t.Fatal("a different environment must reconfigure from scratch")
	}
}

func TestWriteEnvStampMissingDirectory(t *testing.T) {
	if err := writeEnvStamp(filepath.Join(t.TempDir(), "missing"), "x"); err == nil {
		t.Fatal("expected an error for a missing build directory")
	}
}
