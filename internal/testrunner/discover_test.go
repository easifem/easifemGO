// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package testrunner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiscoverSelectionAndWorkingDirectories(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"Z_test_2.F90", "A_test_1.f90", "helper.F90", "nested/B_test_1.F90", "build/Old_test_1.F90", ".hidden/Old_test_1.F90"} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		recursive bool
		filters   []string
		want      []string
	}{
		{false, nil, []string{"A_test_1.f90", "Z_test_2.F90"}},
		{true, nil, []string{"A_test_1.f90", "Z_test_2.F90", "nested/B_test_1.F90"}},
		{true, []string{"B*", "A*"}, []string{"A_test_1.f90", "nested/B_test_1.F90"}},
	} {
		tests, err := Discover(root, tc.recursive, tc.filters)
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, test := range tests {
			names = append(names, test.Relative)
			if test.Dir != filepath.Dir(test.Path) || !filepath.IsAbs(test.Path) {
				t.Fatalf("invalid test paths: %+v", test)
			}
		}
		if !reflect.DeepEqual(names, tc.want) {
			t.Fatalf("got %v, want %v", names, tc.want)
		}
	}
	if _, err := Discover(root, true, []string{"["}); err == nil {
		t.Fatal("invalid glob was accepted")
	}
	if _, err := Discover(root, true, []string{"missing*"}); err == nil {
		t.Fatal("empty selection was accepted")
	}
}
