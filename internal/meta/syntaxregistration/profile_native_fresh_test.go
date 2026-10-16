package syntaxregistration

import (
	"encoding/json"
	"io/fs"
	"os"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestProfileNativePopulationFresh1846(t *testing.T) {
	repository := os.DirFS("../../..")
	paths, err := fs.Glob(repository, syntaxRoot+"*.go")
	if err != nil {
		t.Fatal(err)
	}
	inputs := map[string][]byte{}
	for _, path := range paths {
		raw, err := fs.ReadFile(repository, path)
		if err != nil {
			t.Fatal(err)
		}
		inputs[path] = raw
	}
	source, err := parseSourceUnits(inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := readProfileNativeCases(source)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := fs.ReadFile(repository, corpusPath)
	if err != nil {
		t.Fatal(err)
	}
	var registry languagesyntax.Registry
	if err := json.Unmarshal(raw, &registry); err != nil {
		t.Fatal(err)
	}
	definitions := make([]languagesyntax.CaseDefinition, len(observed))
	for index, row := range observed {
		definitions[index] = row.definition
	}
	if !reflect.DeepEqual(definitions, registry.Cases) || len(source.edits) != 0 {
		t.Fatal("native/JSON source identity population differs")
	}
	t.Logf("PROFILE_POPULATION_OBSERVATION cases=%d changed_cases=0 parsed_gooo_programs=0 native_program_executions=0", len(observed))
}

func TestProfileNativeCounterexamplesFresh1847(t *testing.T) {
	for _, item := range []struct{ name, fields, declarations, reason string }{
		{"duplicate-field", `ID:"a",ID:"b",Path:"p"`, "", "PROFILE_NATIVE_CASE_FIELD_AMBIGUOUS"},
		{"unknown-field", `ID:"a",Path:"p",Other:"x"`, "", "PROFILE_NATIVE_CASE_FIELD_UNKNOWN"},
		{"typed-mismatch", `ID:true,Path:"p"`, "", "PROFILE_NATIVE_VALUE_TYPE_MISMATCH"},
		{"variable", `ID:caseID,Path:"p"`, `var caseID = "a"`, "PROFILE_NATIVE_CONSTANT_UNBOUND_OR_AMBIGUOUS"},
		{"cycle", `ID:a,Path:"p"`, `const a=b;const b=a`, "PROFILE_NATIVE_CONSTANT_CYCLE_OR_DEPTH"},
		{"shadow-bool", `ID:"a",Path:"p",EntityFields:true`, `func true(){}`, "PROFILE_NATIVE_CONSTANT_SHADOWED"},
	} {
		t.Run(item.name, func(t *testing.T) {
			inputs := map[string][]byte{"owner1847.go": []byte("package languagesyntax\n" + item.declarations +
				"\nfunc expectedRegistry() Registry {return Registry{Cases:[]CaseDefinition{{" + item.fields + "}}}}\n")}
			source, err := parseSourceUnits(inputs, "./", "languagesyntax", false)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := readProfileNativeCases(source); err == nil || err.Error() != item.reason {
				t.Fatalf("exact refusal: %v, want %s", err, item.reason)
			}
		})
	}
}
