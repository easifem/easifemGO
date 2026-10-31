// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package testrunner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Test is a standalone Fortran program, executed in its own source directory.
type Test struct {
	Path     string `json:"path"`
	Dir      string `json:"directory"`
	Relative string `json:"name"`
}

var testName = regexp.MustCompile(`_test_.+\.[fF]90$`)

// Discover returns sorted *_test_*.F90/.f90 files. Symlinks are not followed
// below the explicitly supplied root. Filters match a basename or relative path.
func Discover(root string, recursive bool, filters []string) ([]Test, error) {
	for _, pattern := range filters {
		if _, err := filepath.Match(pattern, ""); err != nil {
			return nil, fmt.Errorf("invalid filter %q: %w", pattern, err)
		}
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	absolute, err = filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", absolute)
	}
	var tests []Test
	err = filepath.WalkDir(absolute, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != absolute && (!recursive || strings.HasPrefix(entry.Name(), ".") || entry.Name() == "build" || entry.Name() == "CMakeFiles") {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || !testName.MatchString(entry.Name()) {
			return nil
		}
		rel, relErr := filepath.Rel(absolute, path)
		if relErr != nil {
			return relErr
		}
		matched := len(filters) == 0
		for _, pattern := range filters {
			baseMatch, _ := filepath.Match(pattern, entry.Name())
			relMatch, _ := filepath.Match(pattern, rel)
			if baseMatch || relMatch {
				matched = true
				break
			}
		}
		if matched {
			tests = append(tests, Test{Path: path, Dir: filepath.Dir(path), Relative: filepath.ToSlash(rel)})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover tests: %w", err)
	}
	sort.Slice(tests, func(i, j int) bool { return tests[i].Relative < tests[j].Relative })
	if len(tests) == 0 {
		return nil, fmt.Errorf("no matching *_test_*.F90 programs in %s", absolute)
	}
	return tests, nil
}
