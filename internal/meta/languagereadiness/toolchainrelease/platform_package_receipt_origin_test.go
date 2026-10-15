package toolchainrelease

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func TestPackageReceiptOriginPresence3233(t *testing.T) {
	const source = `package ticket3233
namespace ticket3233
entity Integer id "urn:gooo:type:integer"
activity Fee(Integer) -> Integer id "urn:ticket3233:fee" computes "if input > 97 { return input - 31 }; return input + 11"
`
	program, err := workspaceexecution.Prepare(packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "ticket3233", Activity: "Fee"},
		Packages: []packageruntime.PackageSpec{{Path: "ticket3233", Name: "ticket3233",
			Sources: []packageruntime.Source{{Filename: "receipt3233.gooo", Content: source}}}}})
	if err != nil || len(program.ConditionOrigins) != 1 {
		t.Fatal(err, len(program.ConditionOrigins))
	}
	current, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	checkReceiptOriginForms3233(t, current, program)
	t.Log("fresh3233: original and absent origins accepted; recorded null, empty, changed origin and unknown member rejected through raw receipt comparison; source preparation only; model0")
}

func checkReceiptOriginForms3233(t *testing.T, current json.RawMessage, program workspaceexecution.Program) {
	t.Helper()
	for _, form := range []struct {
		key, replacement string
		omit, want       bool
	}{
		{"", "", false, true}, {"condition_origins", "", true, true},
		{"condition_origins", "null", false, false}, {"condition_origins", "[]", false, false},
		{"condition_origins", `[{"input_source":{"filename":"other3233.gooo"}}]`, false, false},
		{"unrecognized_origin", `"claimed"`, false, false},
	} {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(current, &fields); err != nil {
			t.Fatal(err)
		}
		if form.omit {
			delete(fields, form.key)
		} else if form.key != "" {
			fields[form.key] = json.RawMessage(form.replacement)
		}
		raw, err := json.Marshal(map[string]any{"result": map[string]any{"program": fields}})
		if err != nil || computationPackageReceiptProgramEqual(raw, program) != form.want {
			t.Fatalf("raw receipt origin %s=%s omitted=%v: %v", form.key, form.replacement, form.omit, err)
		}
	}
}
