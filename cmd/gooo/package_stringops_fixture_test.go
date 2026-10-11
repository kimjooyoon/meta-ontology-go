package main

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func stringOpsWorkspace(t *testing.T) (string, string, int) {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"strings.gooo.fixture", "app.gooo.fixture", "gooo.workspace.json", "cases.json"} {
		raw, err := os.ReadFile(filepath.Join("../../examples/bounded-string-operations", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(root, "cases.json")
	return filepath.Join(root, "gooo.workspace.json"), cases, stringOpsOracle(t, cases)
}

func stringOpsOracle(t *testing.T, path string) int {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var suite struct {
		Schema string `json:"schema"`
		Cases  []struct {
			Inputs map[string]struct {
				Text   string `json:"text"`
				Needle string `json:"needle"`
			} `json:"inputs"`
			Expected map[string]map[string]int `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &suite); err != nil {
		t.Fatal(err)
	}
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) != 20 {
		t.Fatal("string operations fixture case inventory changed")
	}
	for _, row := range suite.Cases {
		query, hasInput := row.Inputs["app/stringops:Main"]
		expected, hasOutput := row.Expected["app/stringops:Main"]
		want := map[string]int{"index": strings.Index(query.Text, query.Needle),
			"count": strings.Count(query.Text, query.Needle), "runes": utf8.RuneCountInString(query.Text)}
		if !hasInput || !hasOutput || !maps.Equal(expected, want) {
			t.Fatal("stored string operations differ from Go reference", row, want)
		}
	}
	return len(suite.Cases)
}
