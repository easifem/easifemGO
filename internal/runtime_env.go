// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Programs built by `easifem run` and `easifem test` must load the shared
// libraries of the environment selected for this invocation (`--env` or
// `envName`), not whatever LD_LIBRARY_PATH the user's shell exported when it
// last sourced easifemvar.fish. LD_LIBRARY_PATH also takes precedence over the
// RUNPATH CMake embeds, so the child's library path must be set explicitly.

// libraryPathVar is the dynamic-loader search variable for this platform.
func libraryPathVar() string {
	if runtime.GOOS == "darwin" {
		return "DYLD_LIBRARY_PATH"
	}
	return "LD_LIBRARY_PATH"
}

// selectedEnvironment returns a copy of base whose library search path starts
// with libDirs (in order, environment variables expanded, duplicates removed),
// followed by the entries base already had that are not in libDirs.
func selectedEnvironment(base []string, libDirs []string) []string {
	name := libraryPathVar()
	prefix := name + "="
	var existing string
	env := make([]string, 0, len(base)+1)
	for _, entry := range base {
		if strings.HasPrefix(entry, prefix) {
			existing = strings.TrimPrefix(entry, prefix)
			continue
		}
		env = append(env, entry)
	}
	seen := map[string]bool{}
	var dirs []string
	add := func(dir string) {
		if dir == "" {
			return
		}
		key := filepath.Clean(dir)
		if seen[key] {
			return
		}
		seen[key] = true
		dirs = append(dirs, dir)
	}
	for _, dir := range libDirs {
		add(os.ExpandEnv(dir))
	}
	for _, dir := range filepath.SplitList(existing) {
		add(dir)
	}
	if len(dirs) == 0 {
		return env
	}
	return append(env, prefix+strings.Join(dirs, string(os.PathListSeparator)))
}

// runtimeEnv is the environment for commands launched by `run` and `test`.
func runtimeEnv() []string {
	return selectedEnvironment(os.Environ(), easifem_cache.LD_LIBRARY_PATH)
}

//----------------------------------------------------------------------------
//                                                     build directory stamp
//----------------------------------------------------------------------------

// easifem_env_stamp_file records, inside a reused `easifem run` build
// directory, which environment and prefixes configured it. CMake caches the
// result of find_package (e.g. easifemBase_DIR), so a build directory
// configured for one environment would keep linking that environment's
// libraries after `--env` changes unless it is reconfigured from scratch.
const easifem_env_stamp_file = ".easifem-env"

func envStamp(envName string, prefixes []string) string {
	return fmt.Sprintf("envName = %q\nprefixes = %q\n", envName, prefixes)
}

// envStampChanged reports whether buildDir was configured with a different
// stamp. A build directory without a CMake cache never needs a fresh start.
func envStampChanged(buildDir, stamp string) bool {
	if _, err := os.Stat(filepath.Join(buildDir, "CMakeCache.txt")); err != nil {
		return false
	}
	old, err := os.ReadFile(filepath.Join(buildDir, easifem_env_stamp_file))
	if err != nil {
		return true // configured by an older CLI or by hand: origin unknown
	}
	return string(old) != stamp
}

func writeEnvStamp(buildDir, stamp string) error {
	if err := os.WriteFile(filepath.Join(buildDir, easifem_env_stamp_file), []byte(stamp), 0o644); err != nil {
		return fmt.Errorf("write environment stamp: %w", err)
	}
	return nil
}
