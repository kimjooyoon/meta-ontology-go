package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestSourceRefactorCLIPrintsSourceAndReceiptWithoutExecution(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	path := filepath.Join(t.TempDir(), "main.gooo")
	source := `package p
namespace p
entity Integer id "urn:gooo:type:integer"
activity Fee(Integer) -> Integer computes "return input"
activity Total(Integer) -> Integer computes "return Fee(input)"
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	for _, asJSON := range []bool{false, true} {
		args := []string{"refactor", "rename-activity", "--from", "Fee", "--to", "요금"}
		if asJSON {
			args = append(args, "--json")
		}
		args = append(args, path)
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != exitOK {
			t.Fatal(code, stderr.String())
		}
		actual := stdout.String()
		if asJSON {
			var receipt bodycodegen.SourceRenameResult
			if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
				t.Fatal(err)
			}
			if receipt.ActivityID != "p://activity/fee" || receipt.CallSites != 1 {
				t.Fatal(receipt)
			}
			actual = receipt.Source
		}
		if !strings.Contains(actual, `return 요금(input)`) {
			t.Fatal(actual)
		}
		unchanged, err := os.ReadFile(path)
		if err != nil || string(unchanged) != source {
			t.Fatal("input file changed", err)
		}
	}
}

func TestSourceRefactorCLIHelpAndFailures(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"help", "refactor"}, &stdout, &stderr); code != exitOK || !strings.Contains(stdout.String(), "rename-activity") {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	for _, args := range [][]string{{"refactor"}, {"refactor", "rename-activity"}, {"refactor", "rename-activity", "--from", "A", "--to", "B", "missing.gooo"}} {
		stdout.Reset()
		stderr.Reset()
		if code := run(args, &stdout, &stderr); code == exitOK || stdout.Len() != 0 {
			t.Fatal(code, stdout.String())
		}
	}
}
