package internal

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestLegacyCommandsPropagateFailures(t *testing.T) {
	t.Setenv("EASIFEM_TEST_HELPER_EXIT", "1")
	for _, name := range []string{"install", "dev", "lint"} {
		t.Run(name, func(t *testing.T) {
			operation := newLibraryOutput("fixture", filepath.Join(t.TempDir(), "build"), name)
			operation.animate = false
			for _, args := range [][]string{{os.Args[0], "-test.run=^TestRunCommandExitHelper$"}, {filepath.Join(t.TempDir(), "missing-command")}} {
				var err error
				if name == "lint" {
					err = lintRunCmd(args, "fixture", "build")
				} else {
					err = operation.run(args, "Build", "")
				}
				if err == nil {
					t.Fatal("subprocess failure was discarded")
				}
				if name != "lint" {
					if len(operation.logs) != 1 || !strings.Contains(err.Error(), operation.logs[0]) {
						t.Fatal("failure did not identify its persistent log", err)
					}
					data, readErr := os.ReadFile(operation.logs[0])
					if readErr != nil || !strings.Contains(string(data), "=== Build ·") {
						t.Fatalf("failure log unavailable: %q %v", data, readErr)
					}
				}
			}
		})
	}
}

func TestCleanRejectsCurrentDirectoryAndAncestors(t *testing.T) {
	root := t.TempDir()
	current := filepath.Join(root, "work")
	if err := os.Mkdir(current, 0700); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"", current, root} {
		if err := validateCleanPath(target, current); err == nil {
			t.Fatalf("accepted unsafe target %q", target)
		}
	}
	if err := validateCleanPath(filepath.Join(root, "other"), current); err != nil {
		t.Fatal(err)
	}
	// The safe build directory must survive when the install directory is unsafe.
	old := easifem_pkgs
	t.Cleanup(func() { easifem_pkgs = old })
	safeBuild := filepath.Join(root, "build")
	if err := os.Mkdir(safeBuild, 0700); err != nil {
		t.Fatal(err)
	}
	easifem_pkgs = map[string]*Pkg{"fixture": {BuildDir: safeBuild, InstallDir: root}}
	if err := cleanPkgs("fixture", current); err == nil {
		t.Fatal("unsafe cleanup succeeded")
	}
	if _, err := os.Stat(safeBuild); err != nil {
		t.Fatal("cleanup removed a rejected directory")
	}
	if err := cleanPkgs("missing", current); err == nil {
		t.Fatal("unknown package accepted")
	}
}

func TestExternalPackageFallback(t *testing.T) {
	old := easifem_pkgs
	t.Cleanup(func() { easifem_pkgs = old; viper.Reset() })
	viper.Reset()
	easifem_pkgs = map[string]*Pkg{
		"base": {IsActive: true}, "classes": {IsActive: true},
		"z": {IsActive: true, IsExtPkg: true}, "a": {IsActive: true, IsExtPkg: true},
		"inactive": {IsExtPkg: true},
	}
	if got := pkgGetExtPkgs(); !reflect.DeepEqual(got, []string{"a", "z"}) {
		t.Fatal(got)
	}
	viper.Set(easifem_current_env_name+".extpkgs", []string{"z"})
	if got := pkgGetExtPkgs(); !reflect.DeepEqual(got, []string{"z"}) {
		t.Fatal(got)
	}
}

func TestInstallGroupOrderAndFailure(t *testing.T) {
	old := easifem_pkgs
	t.Cleanup(func() { easifem_pkgs = old })
	easifem_pkgs = map[string]*Pkg{"base": {}, "classes": {}}
	var names []string
	if err := installEasifemWith(".", func(name, _ string) error { names = append(names, name); return nil }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(names, []string{"base", "classes"}) {
		t.Fatal(names)
	}
	sentinel := errors.New("build failed")
	names = nil
	err := installEasifemWith(".", func(name, _ string) error { names = append(names, name); return sentinel })
	if !errors.Is(err, sentinel) || len(names) != 1 {
		t.Fatalf("did not stop after failure: %v %v", names, err)
	}
}

func TestExplicitRunFlagsOverrideConfig(t *testing.T) {
	oldNoRun, oldCache, oldRebuild := noRun, cacheClean, reBuild
	t.Cleanup(func() { noRun, cacheClean, reBuild = oldNoRun, oldCache, oldRebuild })
	noRun, cacheClean, reBuild = true, false, false
	runner := Runner{IsExecute: false, CacheClean: true, ReBuild: true}
	applyRunFlags(&runner, func(string) bool { return true })
	if !runner.IsExecute || runner.CacheClean || runner.ReBuild {
		t.Fatalf("flags not applied: %+v", runner)
	}
	runner = Runner{IsExecute: true, CacheClean: true, ReBuild: true}
	applyRunFlags(&runner, func(string) bool { return false })
	if !runner.IsExecute || !runner.CacheClean || !runner.ReBuild {
		t.Fatal("implicit defaults overrode config")
	}
}

func TestEnvironmentEntries(t *testing.T) {
	t.Cleanup(viper.Reset)
	viper.Reset()
	viper.Set("fixture.env", []string{"TOKEN=a=b=", "EMPTY="})
	got, err := pkgGetEnvVarsFromViper("fixture")
	if err != nil || got["TOKEN"] != "a=b=" || got["EMPTY"] != "" {
		t.Fatalf("%v %v", got, err)
	}
	for _, entry := range []string{"MISSING_EQUALS", "=missing-key"} {
		viper.Set("fixture.env", []string{entry})
		if _, err := pkgGetEnvVarsFromViper("fixture"); err == nil {
			t.Fatalf("accepted %q", entry)
		}
	}
}
