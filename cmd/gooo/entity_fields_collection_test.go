package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestCollectionCLIGeneration7673(t *testing.T) {
	const source = `package cli7673
namespace cli7673
entity Batch id "urn:cli7673:batch" fields {
 field numbers id "urn:cli7673:numbers" type integer required many
 field names id "urn:cli7673:names" type string required many
 field flags id "urn:cli7673:flags" type boolean optional many
}
activity Pass(Batch) -> Batch id "urn:cli7673:pass" computes "return input"
activity Deliver(Batch) -> Batch id "urn:cli7673:deliver" computes "return input"
bind Pass.result -> Deliver.input
`
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("목록7673.gooo", source, syntax.EntityFieldsV5Support())
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, syntax.EntityFieldsV5Support())
	if err != nil {
		t.Fatal(err)
	}
	if err := bidir.CheckGetPutWithEntityFieldsSupport(document, syntax.EntityFieldsV5Support()); err != nil {
		t.Fatal("ordered field BX law", err)
	}
	if _, err := bidir.GetWithEntityFieldsSupport(document, syntax.EntityFieldsV4Support()); err == nil {
		t.Fatal("older single-field profile silently accepted many")
	}
	directory := t.TempDir()
	input, output := filepath.Join(directory, "목록7673.gooo"), filepath.Join(directory, "generated")
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := runCheck([]string{input}, OSFileReader{}, EntityFieldsCLIParser{}, &stdout, &stderr); code != exitOK {
		t.Fatal("CLI check", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := runGenerate([]string{input, "--out", output}, OSFileReader{}, EntityFieldsCLIParser{}, &stdout, &stderr); code != exitOK {
		t.Fatal("CLI generate", code, stdout.String(), stderr.String())
	}
	generated, err := os.ReadFile(filepath.Join(output, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	for _, shape := range []string{"[]int64", "[]string", "*[]bool"} {
		if !strings.Contains(string(generated), shape) {
			t.Fatal("collection shape missing from CLI Go projection", shape, string(generated))
		}
	}
	if _, err := os.Stat(filepath.Join(output, "runtime-plan.json")); err != nil {
		t.Fatal("collection binding plan missing", err)
	}
	after, err := os.ReadFile(input)
	if err != nil || string(after) != source {
		t.Fatal("CLI changed original Gooo source", err)
	}
	t.Log("7673: ordinary check/generate, required and optional collection Go types, binding plan and unchanged-source BX law")
}
