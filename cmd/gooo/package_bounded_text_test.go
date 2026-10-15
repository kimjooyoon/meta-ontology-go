package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func boundedTextWorkspace(t *testing.T) (string, string, int) {
	t.Helper()
	fixture := "../../examples/bounded-text-search"
	root := t.TempDir()
	for _, name := range []string{"search.gooo.fixture", "app.gooo.fixture", "gooo.workspace.json", "cases.json"} {
		raw, err := os.ReadFile(filepath.Join(fixture, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(root, "cases.json")
	return filepath.Join(root, "gooo.workspace.json"), cases, boundedTextOracle(t, cases)
}

func boundedTextOracle(t *testing.T, path string) int {
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
				Needle int    `json:"needle"`
			} `json:"inputs"`
			Expected map[string]int `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &suite); err != nil {
		t.Fatal(err)
	}
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) == 0 {
		t.Fatal("bounded search example has no valid oracle cases")
	}
	for _, row := range suite.Cases {
		query, hasInput := row.Inputs["app/search:Main"]
		expected, hasOutput := row.Expected["app/search:Main"]
		want := -1
		if query.Needle >= 0 && query.Needle <= 255 {
			want = strings.IndexByte(query.Text, byte(query.Needle))
		}
		if !hasInput || !hasOutput || expected != want {
			t.Fatal("public fixture expectation differs from Go byte search", row, want)
		}
	}
	return len(suite.Cases)
}

func TestPackageBoundedTextSearchExecutesAndReplaysImportedGooo(t *testing.T) {
	manifest, cases, total := boundedTextWorkspace(t)
	var output, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", cases, manifest}, &output, &diagnostics); code != exitOK {
		t.Fatal("imported search failed", code, diagnostics.String(), output.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "PASS" || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != total ||
		receipt.Result.Runtime.FiniteTotal != total || !receipt.Result.Runtime.RuntimeReplayed || receipt.Result.Runtime.ModelCalls != 0 {
		t.Fatal("imported search did not match all Go oracle cases", receipt)
	}
	receiptPath := filepath.Join(t.TempDir(), "execution.json")
	if err := os.WriteFile(receiptPath, output.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if code := run([]string{"package", "replay", "--json", "--receipt", receiptPath, "--cases", cases, manifest}, &output, &diagnostics); code != exitOK {
		t.Fatal("saved imported search failed", code, diagnostics.String(), output.String())
	}
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.Result.Replay == nil ||
		receipt.Result.Runtime.FinitePassed != total || !receipt.Result.Runtime.RuntimeReplayed || receipt.Result.Replay.ModelCalls != 0 {
		t.Fatal("saved imported search did not retain its observed results", err)
	}
}
