// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package internal

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// None of these tests start a compiler, CMake or a Fortran program.

func TestCheckBuildSystem(t *testing.T) {
	for _, system := range []string{"cmake", "make"} {
		if err := checkBuildSystem(&Pkg{Name: "p", BuildSystem: system}); err != nil {
			t.Errorf("%s rejected: %v", system, err)
		}
	}
	for _, system := range []string{"mason", "", "ninja"} {
		err := checkBuildSystem(&Pkg{Name: "p", BuildSystem: system})
		if err == nil || !strings.Contains(err.Error(), "unsupported buildSystem") {
			t.Errorf("%q accepted or unclear error: %v", system, err)
		}
	}
}

func TestPkgInstallRejectsUnsupportedBuildSystemBeforeSideEffects(t *testing.T) {
	root := t.TempDir()
	pkg := &Pkg{
		Name: "fixture", IsActive: true, BuildSystem: "mason",
		BuildDir:   filepath.Join(root, "build"),
		InstallDir: filepath.Join(root, "install"),
		SourceDir:  filepath.Join(root, "src"),
		Url:        filepath.Join(root, "missing-source"),
	}
	err := PkgInstall(pkg, root)
	if err == nil || !strings.Contains(err.Error(), `"mason"`) {
		t.Fatalf("unsupported build system was not reported: %v", err)
	}
	for _, dir := range []string{pkg.BuildDir, pkg.InstallDir, pkg.SourceDir} {
		if _, statErr := os.Stat(dir); !os.IsNotExist(statErr) {
			t.Errorf("%s was created before the build system was checked", dir)
		}
	}

	pkg.IsActive = false
	if err := PkgInstall(pkg, root); err != nil {
		t.Fatalf("inactive package should still be skipped: %v", err)
	}
}

func TestFortranCompiler(t *testing.T) {
	base := t.TempDir()
	bin := filepath.Join(base, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(base, "my-gfortran")
	if err := os.WriteFile(executable, nil, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EASIFEM_TEST_FC", "ifx")
	cases := []struct{ compiler, path, want string }{
		{"", "", ""},
		{"$EASIFEM_TEST_FC", "", "ifx"},
		{"ifx", bin, filepath.Join(bin, "ifx")},
		{"", "bin", filepath.Join(bin, "gfortran")},
		{"gfortran", "my-gfortran", executable},
	}
	for _, c := range cases {
		got, err := fortranCompiler(c.compiler, c.path, base)
		if err != nil || got != c.want {
			t.Errorf("fortranCompiler(%q, %q) = %q, %v; want %q", c.compiler, c.path, got, err, c.want)
		}
	}
}

func TestRunnerTargetSettings(t *testing.T) {
	base := t.TempDir()
	runner := Runner{
		TargetName:       "main.out",
		IncludePath:      []string{"inc"},
		ExtraIncludePath: []string{"/abs/extra"},
		Flags:            []string{"-Wall", "-DNAME=a;b"},
		LibraryPath:      []string{"lib"},
	}
	got, err := runnerTargetSettings(runner, base, []string{"/first"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"target_include_directories(main.out PRIVATE \"/first\")\n",
		"target_include_directories(main.out PRIVATE " + quoteTestCMake(filepath.Join(base, "inc")) + ")\n",
		"target_include_directories(main.out PRIVATE \"/abs/extra\")\n",
		"target_compile_options(main.out PRIVATE \"-Wall\")\n",
		"target_compile_options(main.out PRIVATE \"-DNAME=a\\;b\")\n",
		"target_link_directories(main.out PRIVATE " + quoteTestCMake(filepath.Join(base, "lib")) + ")\n",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("settings mismatch\n got: %q\nwant: %q", got, want)
	}

	empty, err := runnerTargetSettings(Runner{TargetName: "main.out"}, base, nil)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty runner produced settings %q, %v", empty, err)
	}
}

func TestLintArgsHonorsLinterToml(t *testing.T) {
	pwd := t.TempDir()
	bin := filepath.Join(pwd, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	linter := Linter{
		Compiler:     "gfortran-14",
		CompilerPath: "bin",
		IncludePath:  []string{"/user/include"},
		Flags:        []string{"-Wextra", "-fcheck=all"},
		LintDir:      "/lint",
	}
	got, err := lintArgs(linter, "my file.F90", pwd, []string{"/install/base"})
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != filepath.Join(bin, "gfortran-14") {
		t.Errorf("compiler = %q", got[0])
	}
	if !reflect.DeepEqual(got[1:5], []string{"-I", "/user/include", "-I", "/install/base/include"}) {
		t.Errorf("include order = %q", got[1:5])
	}
	flagsEnd := 5 + len(lintDefaultFlags)
	if !reflect.DeepEqual(got[5:flagsEnd], lintDefaultFlags) {
		t.Errorf("built-in flags missing or reordered: %q", got[5:flagsEnd])
	}
	wantTail := []string{"-Wextra", "-fcheck=all", "-J", "/lint/include", "-c", "my file.F90", "-o", "/lint/lib/my file.F90.o"}
	if !reflect.DeepEqual(got[flagsEnd:], wantTail) {
		t.Errorf("tail = %q; want %q", got[flagsEnd:], wantTail)
	}

	plain, err := lintArgs(Linter{LintDir: "/lint"}, "a.F90", pwd, nil)
	if err != nil || plain[0] != "gfortran" {
		t.Errorf("default compiler = %q, %v", plain[0], err)
	}
}
