package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagetest"
)

type languageTestReader map[string][]byte

func (reader languageTestReader) ReadFile(filename string) ([]byte, error) {
	value, ok := reader[filename]
	if !ok {
		return nil, fmt.Errorf("missing %s", filename)
	}
	return value, nil
}

func TestRunLanguageTestJSON(t *testing.T) {
	reader := languageTestReader{"main.gooo": []byte(`package p
namespace p
entity Input id "p://input"
entity Output id "p://output"
entity BuildsOutput id "gooo://test/activity/Build/output/Output"
activity Build(Input) -> Output
`)}
	var stdout, stderr bytes.Buffer
	code := runLanguageTest([]string{"--json", "main.gooo"}, reader, &stdout, &stderr)
	var receipt languagetest.Receipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if code != exitOK || receipt.Decision != languagetest.DecisionPass || stderr.Len() != 0 {
		t.Fatalf("code=%d receipt=%+v stderr=%q", code, receipt, stderr.String())
	}
}
