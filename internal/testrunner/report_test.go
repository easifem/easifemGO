// Copyright (C) 2026 Vikas Sharma. Licensed under the repository LICENSE.
package testrunner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestJSONPreservesPartialRun(t *testing.T) {
	summary := Summary{Version: 1, Total: 3, NotRun: 2, Results: []Result{{Status: Cancelled}}}
	path := filepath.Join(t.TempDir(), "nested", "results.json")
	if err := WriteJSON(path, summary); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Summary
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.NotRun != 2 || decoded.Results[0].Status != Cancelled || decoded.Version != 1 {
		t.Fatalf("report lost partial results: %+v", decoded)
	}
}
