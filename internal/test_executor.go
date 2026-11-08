// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package internal

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/easifem/easifemGO/internal/testrunner"
)

// exampleTestExecutor deliberately builds in fresh, isolated directories. The
// source examples keep their own CMakeLists.txt and retain their runtime cwd.
type exampleTestExecutor struct{ rebuild bool }

var cmakeTargetName = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.+-]*$`)

func (e exampleTestExecutor) Build(ctx context.Context, test testrunner.Test, out io.Writer) (testrunner.Program, error) {
	var program testrunner.Program
	runner := Runner{
		BuildDir:  filepath.Join(os.TempDir(), "easifem-example-builds"),
		BuildType: "Debug", CMakeMinimumVersion: "3.20.0", Compiler: "gfortran",
		Language: "Fortran", ProjectName: "easifemApp", SourceDir: test.Dir, TargetName: "main.out",
	}
	config := filepath.Join(test.Dir, "runner.toml")
	if _, err := os.Stat(config); err == nil {
		if _, err := toml.DecodeFile(config, &runner); err != nil {
			return program, fmt.Errorf("read %s: %w", config, err)
		}
	} else if !os.IsNotExist(err) {
		return program, err
	}
	if runner.BuildDir == "" {
		return program, fmt.Errorf("runner.toml: BuildDir must not be empty")
	}
	if !cmakeTargetName.MatchString(runner.TargetName) {
		return program, fmt.Errorf("invalid executable target name %q", runner.TargetName)
	}
	if runner.Language != "Fortran" {
		return program, fmt.Errorf("test runner requires Language = Fortran")
	}
	sourceBase, err := expandTestPath(test.Dir, runner.SourceDir)
	if err != nil {
		return program, err
	}
	if info, err := os.Stat(sourceBase); err != nil || !info.IsDir() {
		return program, fmt.Errorf("runner SourceDir is not a readable directory: %s", sourceBase)
	}
	buildRoot, err := expandTestPath(test.Dir, runner.BuildDir)
	if err != nil {
		return program, err
	}
	if err := os.MkdirAll(buildRoot, 0700); err != nil {
		return program, err
	}
	workspace, err := os.MkdirTemp(buildRoot, "easifem-test-")
	if err != nil {
		return program, err
	}
	generated := filepath.Join(workspace, "source")
	buildDir := filepath.Join(workspace, "build")
	if err := os.Mkdir(generated, 0700); err != nil {
		return program, err
	}
	runner.FileName = quoteTestCMake(test.Path)
	for i, file := range runner.ExtraFileNames {
		resolved, err := expandTestPath(sourceBase, file)
		if err != nil {
			return program, err
		}
		runner.ExtraFileNames[i] = quoteTestCMake(resolved)
	}
	for i, path := range runner.ExtraCMakePrefixPath {
		runner.ExtraCMakePrefixPath[i], err = expandTestPath(sourceBase, path)
		if err != nil {
			return program, err
		}
	}
	if len(runner.TargetLibs) == 0 {
		runner.TargetLibs = []string{"easifemBase"}
	}
	for i, library := range runner.ExtraLibs {
		library = os.ExpandEnv(library)
		if !strings.HasPrefix(library, "-") && (strings.ContainsAny(library, "/\\") || strings.HasPrefix(library, ".")) {
			library, err = expandTestPath(sourceBase, library)
			if err != nil {
				return program, err
			}
		}
		runner.ExtraLibs[i] = library
	}
	cmake := makeCmakeFile(runner)
	includes := append([]string{test.Dir, sourceBase}, runner.IncludePath...)
	includes = append(includes, runner.ExtraIncludePath...)
	for _, include := range includes {
		resolved, err := expandTestPath(sourceBase, include)
		if err != nil {
			return program, err
		}
		cmake = append(cmake, fmt.Sprintf("target_include_directories(%s PRIVATE %s)\n", runner.TargetName, quoteTestCMake(resolved)))
	}
	for _, flag := range runner.Flags {
		cmake = append(cmake, fmt.Sprintf("target_compile_options(%s PRIVATE %s)\n", runner.TargetName, quoteTestCMake(flag)))
	}
	for _, directory := range runner.LibraryPath {
		resolved, err := expandTestPath(sourceBase, directory)
		if err != nil {
			return program, err
		}
		cmake = append(cmake, fmt.Sprintf("target_link_directories(%s PRIVATE %s)\n", runner.TargetName, quoteTestCMake(resolved)))
	}
	if err := os.WriteFile(filepath.Join(generated, "CMakeLists.txt"), []byte(strings.Join(cmake, "")), 0600); err != nil {
		return program, err
	}
	args := []string{"cmake", "-G", "Ninja", "-S", generated, "-B", buildDir, "-DCMAKE_BUILD_TYPE:STRING=" + runner.BuildType}
	compiler := os.ExpandEnv(runner.Compiler)
	if runner.CompilerPath != "" {
		compilerPath, err := expandTestPath(sourceBase, runner.CompilerPath)
		if err != nil {
			return program, err
		}
		if info, err := os.Stat(compilerPath); err == nil && info.IsDir() {
			compiler = filepath.Join(compilerPath, compiler)
		} else {
			compiler = compilerPath
		}
	}
	if compiler != "" {
		args = append(args, "-DCMAKE_Fortran_COMPILER:FILEPATH="+compiler)
	}
	if err := testCommand(ctx, test.Dir, out, args...); err != nil {
		return program, err
	}
	args = []string{"cmake", "--build", buildDir, "--target", runner.TargetName}
	if e.rebuild || runner.ReBuild {
		args = append(args, "--clean-first")
	}
	if err := testCommand(ctx, test.Dir, out, args...); err != nil {
		return program, err
	}
	executable := filepath.Join(buildDir, runner.TargetName)
	// A CMake Fortran target gets an executable suffix on Windows.
	if info, err := os.Stat(executable); err != nil || !info.Mode().IsRegular() {
		if info, exeErr := os.Stat(executable + ".exe"); exeErr == nil && info.Mode().IsRegular() {
			executable += ".exe"
		} else {
			return program, fmt.Errorf("build did not produce %s", executable)
		}
	}
	return testrunner.Program{Path: executable, Dir: test.Dir}, nil
}

func (e exampleTestExecutor) Execute(ctx context.Context, program testrunner.Program, out io.Writer) error {
	return testCommand(ctx, program.Dir, out, program.Path)
}

func testCommand(ctx context.Context, dir string, out io.Writer, args ...string) error {
	if _, err := fmt.Fprintf(out, "command: %q\nworking directory: %s\n", args, dir); err != nil {
		return err
	}
	return testrunner.RunCommand(ctx, dir, out, nil, args...)
}

func expandTestPath(base, path string) (string, error) {
	path = os.ExpandEnv(path)
	if !filepath.IsAbs(path) {
		path = filepath.Join(base, path)
	}
	return filepath.Abs(path)
}

func quoteTestCMake(text string) string {
	text = strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "$", "\\$", ";", "\\;").Replace(text)
	return "\"" + text + "\""
}
