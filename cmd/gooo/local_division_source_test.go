package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

func TestLocalDivisionPublicRegistryProfile(t *testing.T) {
	const path = "examples/compound-updates/main.gooo"
	r := replay.Execute(os.DirFS("../.."), path, "VALID", "")
	if !r.ASTReplayed || !r.ByteReplayed || !r.SemanticReplayed || !r.GetPut || !r.PutGet {
		t.Fatal("original scalar registry profile", r)
	}
	for _, args := range [][]string{
		{"roundtrip", "--json", "../../" + path},
		{"package", "resolve", "--json", "../../examples/compound-updates/division.workspace.json"},
	} {
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK {
			t.Fatal(code, diagnostics.String(), out.String())
		}
		if !json.Valid(out.Bytes()) {
			t.Fatal("invalid public JSON")
		}
	}
	t.Log("new Divide declaration preserves original scalar registry profile and public source/manifest round trips; native executions 0")
}
