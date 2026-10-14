package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func freshRegistryOriginSource28903(t *testing.T, units map[string]string) *goSource {
	t.Helper()
	inputs := map[string][]byte{}
	for name, body := range units {
		inputs[syntaxRoot+name] = []byte("package languagesyntax\n" + body)
	}
	source, err := parseSourceUnits(inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	source.activity = "gooo://syntax-registration/native-registry"
	return source
}

func TestRegistrySourceOrigins28903(t *testing.T) {
	for _, layout := range []struct {
		name  string
		units map[string]string
		tail  string
	}{
		{"inline-comment-comma", map[string]string{"owner.go": `func expectedRegistry() Registry { return Registry{Cases: []CaseDefinition{{ID: "one"} /* prose, comma */}} }`}, "owner.go"},
		{"inline", map[string]string{"owner.go": `func expectedRegistry() Registry { return Registry{Cases: []CaseDefinition{{ID: "one"}}} }`}, "owner.go"},
		{"renamed-and-relocated", map[string]string{
			"entry_renamed.go": `//line fictional.go:81
func expectedRegistry() Registry { return Registry{Cases: collected()} }`,
			"assembly_renamed.go": `func collected() []CaseDefinition { rows := first(); rows = append(rows, tail()...); return rows }`,
			"first_renamed.go":    `func first() []CaseDefinition { return []CaseDefinition{{ID: "first"}} }`,
			"tail_renamed.go": `func tail() []CaseDefinition { return []CaseDefinition{{ID: "tail"}} }
func unrelated() []CaseDefinition { return []CaseDefinition{{ID: "unrelated"}} }`,
		}, "tail_renamed.go"},
		{"nested", map[string]string{
			"entry.go": `func expectedRegistry() Registry { return Registry{Cases: all()} }`,
			"assembly.go": `func all() []CaseDefinition { rows := first(); rows = append(rows, remainder()...); return rows }
func remainder() []CaseDefinition { return last() }`,
			"leaves.go": `func first() []CaseDefinition { return []CaseDefinition{{ID: "first"}} }
func last() []CaseDefinition { return []CaseDefinition{{ID: "last"}} }`,
		}, "leaves.go"},
	} {
		t.Run(layout.name, func(t *testing.T) {
			source := freshRegistryOriginSource28903(t, layout.units)
			request := Request{Case: languagesyntax.CaseDefinition{ID: "fresh-origin-28903", Path: "examples/fresh-origin28903/main.gooo"}}
			if err := generateRegistry(source, request); err != nil {
				t.Fatal(err)
			}
			changes, err := source.finish()
			if err != nil {
				t.Fatal(err)
			}
			if len(changes) != 1 || !bytes.Contains(changes[syntaxRoot+layout.tail], []byte(request.Case.ID)) {
				t.Fatalf("wrong source owner or extra writes: %v", sortedPaths(changes))
			}
			if bytes.Contains(changes[syntaxRoot+layout.tail], []byte(`ID: "unrelated", Path:`)) {
				t.Fatal("unrelated literal acquired registration ownership")
			}
		})
	}
}

func TestRegistrySourceOriginsRejectUnknown28903(t *testing.T) {
	for _, item := range []struct{ name, expression, helpers string }{
		{"cycle", "a()", `func a() []CaseDefinition { return b() }; func b() []CaseDefinition { return a() }`},
		{"duplicate-leaf", "all()", `func all() []CaseDefinition { rows := leaf(); rows = append(rows, leaf()...); return rows }; func leaf() []CaseDefinition { return []CaseDefinition{{ID: "one"}} }`},
		{"branch", "all()", `func all() []CaseDefinition { if enabled { return []CaseDefinition{{ID: "one"}} }; return []CaseDefinition{{ID: "two"}} }`},
		{"external", "external.Cases()", ""},
		{"argument", "leaf(1)", `func leaf(n int) []CaseDefinition { return []CaseDefinition{{ID: "one"}} }`},
		{"named-return", "leaf()", `func leaf() (rows []CaseDefinition) { return []CaseDefinition{{ID: "one"}} }`},
		{"non-slice-append", "all()", `func all() []CaseDefinition { rows := leaf(); rows = append(rows, CaseDefinition{ID: "two"}); return rows }; func leaf() []CaseDefinition { return []CaseDefinition{{ID: "one"}} }`},
		{"side-effect", "leaf()", `func leaf() []CaseDefinition { mutate(); return []CaseDefinition{{ID: "one"}} }`},
		{"empty-tail", "leaf()", `func leaf() []CaseDefinition { return []CaseDefinition{} }`},
		{"shadowed-append", "all()", `func append(a,b []CaseDefinition) []CaseDefinition { return a }; func all() []CaseDefinition { rows := leaf(); rows = append(rows, tail()...); return rows }; func leaf() []CaseDefinition { return []CaseDefinition{{ID: "one"}} }; func tail() []CaseDefinition { return []CaseDefinition{{ID: "two"}} }`},
		{"shadowed-helper", "all()", `func all() []CaseDefinition { leaf := first(); leaf = append(leaf, leaf()...); return leaf }; func leaf() []CaseDefinition { return []CaseDefinition{{ID: "two"}} }; func first() []CaseDefinition { return []CaseDefinition{{ID: "one"}} }`},
		{"ambiguous-helper", "leaf()", `func leaf() []CaseDefinition { return []CaseDefinition{{ID: "one"}} }; func leaf() []CaseDefinition { return []CaseDefinition{{ID: "two"}} }`},
	} {
		t.Run(item.name, func(t *testing.T) {
			source := freshRegistryOriginSource28903(t, map[string]string{
				"entry.go":   "func expectedRegistry() Registry { return Registry{Cases: " + item.expression + "} }",
				"helpers.go": item.helpers,
			})
			if err := generateRegistry(source, Request{}); err == nil || len(source.edits) != 0 {
				t.Fatalf("unknown shape emitted edits: %v / %d", err, len(source.edits))
			}
		})
	}
}

func TestRegistryFullCandidateOrigins28903(t *testing.T) {
	data, request := fixture(t)
	delete(data, request.Case.Path)
	request.Case.ID = "source-origin-assembly28903"
	request.Case.Path = "examples/source-origin-assembly28903/main.gooo"
	data[request.Case.Path] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(fixtureSource, "registrationfixture", "sourceorigin28903"))}
	pin(t, data, &request)
	before := digestValue(data)
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := plan.Generate(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.ValidateCandidate(data, candidate); err != nil {
		t.Fatal(err)
	}
	if digestValue(data) != before || candidate.ApplyAuthorized || candidate.RepositoryWrites != 0 || len(candidate.Artifacts) != 9 {
		t.Fatal("source or nine role boundaries changed")
	}
	var beforeCorpus, afterCorpus languagesyntax.Registry
	if err := json.Unmarshal(data[corpusPath].Data, &beforeCorpus); err != nil {
		t.Fatal(err)
	}
	for _, member := range candidate.Members {
		if member.Path == corpusPath {
			if err := json.Unmarshal(member.Content, &afterCorpus); err != nil {
				t.Fatal(err)
			}
		}
		if member.Path == syntaxRoot+"registry_cases_group5.go" && !bytes.Contains(member.Content, []byte(request.Case.ID)) {
			t.Fatal("tail source did not receive new member")
		}
		if member.Path == syntaxRoot+"registry.go" {
			t.Fatal("nonpromotion rewrote the root owner")
		}
	}
	if len(afterCorpus.Cases) != 111 || !reflect.DeepEqual(afterCorpus.Cases[:110], beforeCorpus.Cases) || afterCorpus.Cases[110] != request.Case {
		t.Fatal("candidate lost existing order, flags or the new member")
	}
	promoted, promotedRequest := promotedMetaFixture(t)
	promotedRequest.Case.ID = "promoted-origin28903"
	pin(t, promoted, &promotedRequest)
	promotedPlan, err := Compile(promoted, promotedRequest)
	if err != nil {
		t.Fatal(err)
	}
	promotedCandidate, err := promotedPlan.Generate(promoted)
	if err != nil {
		t.Fatal(err)
	}
	if len(promotedCandidate.Artifacts) != 9 || promotedCandidate.Required != 10 {
		t.Fatal("split promotion lost roles or owners")
	}
	if err := promotedPlan.ValidateCandidate(promoted, promotedCandidate); err != nil {
		t.Fatal(err)
	}
	var nativePaths []string
	for _, artifact := range candidate.Artifacts {
		if artifact.OutputID == plan.binding.outputs[2] {
			nativePaths = artifact.Paths
		}
	}
	if !reflect.DeepEqual(nativePaths, []string{syntaxRoot + "registry_cases_group5.go"}) {
		t.Fatalf("native source origin binding changed: %v", nativePaths)
	}
	if out := os.Getenv("GOOO_FRESH_SOURCE_OUTPUT"); out != "" {
		raw, err := json.MarshalIndent(candidate, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(out+"/candidate28903.json", raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
