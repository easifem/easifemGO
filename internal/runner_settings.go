// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

// Shared runner.toml/linter.toml handling for `easifem run`, `easifem test`
// and `easifem lint`, so the three commands interpret the same keys alike.

// fortranCompiler resolves Compiler and CompilerPath. CompilerPath may name
// the compiler executable or the directory containing Compiler (gfortran when
// Compiler is empty); relative paths are resolved from base. An empty result
// means that no compiler was configured.
func fortranCompiler(compiler, compilerPath, base string) (string, error) {
	compiler = os.ExpandEnv(compiler)
	if compilerPath == "" {
		return compiler, nil
	}
	resolved, err := expandTestPath(base, compilerPath)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(resolved); err == nil && info.IsDir() {
		if compiler == "" {
			compiler = "gfortran"
		}
		return filepath.Join(resolved, compiler), nil
	}
	return resolved, nil
}

// runnerTargetSettings returns CMake commands applying include directories
// (includes, then IncludePath, then ExtraIncludePath), Flags and LibraryPath
// to the runner's executable target. Relative paths are resolved from base.
func runnerTargetSettings(runner Runner, base string, includes []string) ([]string, error) {
	var cmake []string
	all := append(append(append([]string{}, includes...), runner.IncludePath...), runner.ExtraIncludePath...)
	for _, include := range all {
		resolved, err := expandTestPath(base, include)
		if err != nil {
			return nil, err
		}
		cmake = append(cmake, fmt.Sprintf("target_include_directories(%s PRIVATE %s)\n", runner.TargetName, quoteTestCMake(resolved)))
	}
	for _, flag := range runner.Flags {
		cmake = append(cmake, fmt.Sprintf("target_compile_options(%s PRIVATE %s)\n", runner.TargetName, quoteTestCMake(flag)))
	}
	for _, directory := range runner.LibraryPath {
		resolved, err := expandTestPath(base, directory)
		if err != nil {
			return nil, err
		}
		cmake = append(cmake, fmt.Sprintf("target_link_directories(%s PRIVATE %s)\n", runner.TargetName, quoteTestCMake(resolved)))
	}
	return cmake, nil
}
