package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// These native checks exercise the public Gooo fixtures in ordinary CI. The
// stored observation files remain historical records, not test expectations.
func TestSourceRelationalCallerPrograms(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "source-relational-callers")
	for _, tc := range []struct {
		name, entry, prefix, decision                   string
		budget, attempts, selected, criteria, evaluated int
	}{
		{"monotonic", "Nondecreasing", "", "COMPLETE_FINITE", 2, 2, 1, 1, 1},
		{"insufficient", "Nondecreasing", "", "PARTIAL_FINITE", 2, 2, 0, 1, 0},
		{"roundtrip", "RoundTrip", "roundtrip-", "COMPLETE_FINITE", 4, 2, 1, 2, 1},
		{"roundtrip-insufficient", "RoundTrip", "roundtrip-", "PARTIAL_FINITE", 4, 4, 0, 2, 0},
		{"conflicting", "Nondecreasing", "", "PARTIAL_FINITE", 2, 2, 0, 2, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := readRelationalFixture(t, filepath.Join(root, tc.name+".gooo.fixture"))
			construction, err := DecodeCompositionCases(readRelationalFixture(t, filepath.Join(root, tc.prefix+"construction-cases.json")))
			if err != nil {
				t.Fatal(err)
			}
			evaluation, err := DecodeCompositionCases(readRelationalFixture(t, filepath.Join(root, tc.prefix+"evaluation-cases.json")))
			if err != nil {
				t.Fatal(err)
			}
			built, observed, err := ConstructAndEvaluateJoint(context.Background(), tc.name+".gooo", source,
				construction, evaluation, JointOptions{EntryActivity: tc.entry, ProgramBudget: tc.budget, GoBinary: nativeTool()})
			if err != nil {
				t.Fatal(err)
			}
			if built.Decision != tc.decision || len(built.Attempts) != tc.attempts || built.SelectedAttempt != tc.selected {
				t.Fatal("relation selection or unresolved result changed", built.Decision, len(built.Attempts), built.SelectedAttempt)
			}
			if built.OriginalSourceSHA256 != digest(source) {
				t.Fatal("construction lost its original source identity")
			}
			assertRelationalProjection(t, source, []byte(built.SelectedSource))
			selected := built.Attempts[built.SelectedAttempt]
			if selected.CriteriaPassed != tc.criteria || selected.CriteriaTotal != tc.criteria || selected.CriteriaUnknown != 0 {
				t.Fatal("local source obligations lost", selected.CriteriaPassed, selected.CriteriaTotal, selected.CriteriaUnknown)
			}
			if observed.Runtime.FinitePassed != tc.evaluated || observed.Runtime.FiniteTotal != 1 || observed.NewModelCalls != 0 {
				t.Fatal("separate relation evaluation changed", observed.Runtime.FinitePassed, observed.Runtime.FiniteTotal, observed.NewModelCalls)
			}
			for _, attempt := range built.Attempts {
				for _, candidate := range attempt.Candidates {
					spec, err := bodycodegen.SourceAssembly(context.Background(), tc.name+".gooo", source, candidate.Activity)
					if err != nil || spec == nil {
						t.Fatal("candidate lost its source declaration", candidate.Activity, err)
					}
					contract, err := spec.Canonical()
					if err != nil || candidate.ContractSHA256 != digest([]byte(contract)) {
						t.Fatal("candidate changed its original source contract", candidate.Activity, err)
					}
				}
				if attempt.Runtime.ModelCalls != 0 {
					t.Fatal("unexpected model inference")
				}
			}
			if tc.name == "conflicting" {
				second := built.Attempts[1]
				if second.Runtime.FinitePassed != 1 || second.CriteriaPassed != 0 || second.CriteriaFailed != 2 {
					t.Fatal("caller success erased conflicting local conditions")
				}
			}
		})
	}
}

func readRelationalFixture(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// Parsing these recorded sources does not execute a program or load a model.
func TestSourceRelationalCallerProjection(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "source-relational-callers")
	for _, name := range []string{"monotonic", "insufficient", "roundtrip", "roundtrip-insufficient", "conflicting"} {
		t.Run(name, func(t *testing.T) {
			original := readRelationalFixture(t, filepath.Join(root, name+".gooo.fixture"))
			selected := readRelationalFixture(t, filepath.Join(root, name+".selected.gooo.fixture"))
			assertRelationalProjection(t, original, selected)
		})
	}
}

func assertRelationalProjection(t *testing.T, original, selected []byte) {
	t.Helper()
	before, diagnostics := bodycodegen.ParseBodyFile("original.gooo", original)
	if diagnostics.HasErrors() || before == nil {
		t.Fatal(diagnostics)
	}
	after, diagnostics := bodycodegen.ParseBodyFile("selected.gooo", selected)
	if diagnostics.HasErrors() || after == nil {
		t.Fatal(diagnostics)
	}
	if before.Package.Name != after.Package.Name || before.Namespace.Name != after.Namespace.Name ||
		len(before.Declarations) != len(after.Declarations) {
		t.Fatal("source projection changed declaration scope")
	}
	assembled, fixed := 0, 0
	for i, declaration := range before.Declarations {
		switch a := declaration.(type) {
		case *syntax.EntityDecl:
			b, ok := after.Declarations[i].(*syntax.EntityDecl)
			if !ok || a.Name != b.Name || a.ID != b.ID {
				t.Fatal("entity identity changed")
			}
		case *syntax.ActivityDecl:
			b, ok := after.Declarations[i].(*syntax.ActivityDecl)
			if !ok || a.Name != b.Name || a.ID != b.ID || a.IDPresent != b.IDPresent ||
				a.Output != b.Output || len(a.Inputs) != len(b.Inputs) {
				t.Fatal("activity identity or signature changed")
			}
			for j := range a.Inputs {
				if a.Inputs[j].Name != b.Inputs[j].Name {
					t.Fatal("activity input type changed")
				}
			}
			if b.Assembly != nil || !b.ValueProgramPresent {
				t.Fatal("selected source is not a fixed execution projection")
			}
			if a.Assembly == nil {
				fixed++
				if a.ValueProgram != b.ValueProgram {
					t.Fatal("fixed relation or predicate changed", a.Name)
				}
			} else {
				assembled++
			}
		default:
			t.Fatalf("unexpected relation declaration %T", declaration)
		}
	}
	if assembled == 0 || fixed < 2 {
		t.Fatal("fixture lost its assembly or fixed checks")
	}
}

func TestSourceRelationalCallerObservationBindings(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "source-relational-callers")
	var saved struct {
		Variants map[string]struct {
			OriginalSourceSHA256   string `json:"original_source_sha256"`
			RecordedSelectedSource string `json:"recorded_selected_source"`
			SelectedSourceSHA256   string `json:"selected_source_sha256"`
			SelectedFixtureSHA256  string `json:"selected_fixture_sha256"`
			Attempts               []struct {
				Choices []struct {
					Activity       string `json:"activity"`
					ContractSHA256 string `json:"contract_sha256"`
				} `json:"choices"`
			} `json:"attempts"`
		} `json:"variants"`
	}
	if err := json.Unmarshal(readRelationalFixture(t, filepath.Join(root, "observation.json")), &saved); err != nil {
		t.Fatal(err)
	}
	if len(saved.Variants) != 5 {
		t.Fatal("missing source observations")
	}
	for name, variant := range saved.Variants {
		source := readRelationalFixture(t, filepath.Join(root, name+".gooo.fixture"))
		if variant.OriginalSourceSHA256 != digest(source) {
			t.Fatal("observation source differs", name)
		}

		selected := readRelationalFixture(t, filepath.Join(root, name+".selected.gooo.fixture"))
		if digest([]byte(variant.RecordedSelectedSource)) != "sha256:"+variant.SelectedSourceSHA256 ||
			digest(selected) != variant.SelectedFixtureSHA256 {
			t.Fatal("selected source digest differs", name)
		}
		lines := strings.Split(variant.RecordedSelectedSource, "\n")
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], " \t")
		}
		if string(selected) != strings.Join(lines, "\n") {
			t.Fatal("selected fixture changed beyond whitespace", name)
		}
		assertRelationalProjection(t, source, selected)
		if len(variant.Attempts) == 0 {
			t.Fatal("missing attempts", name)
		}
		for _, attempt := range variant.Attempts {
			if len(attempt.Choices) == 0 {
				t.Fatal("missing source choices", name)
			}
			for _, choice := range attempt.Choices {
				spec, err := bodycodegen.SourceAssembly(context.Background(), name+".gooo", source, choice.Activity)
				if err != nil || spec == nil {
					t.Fatal("missing source contract", name, choice.Activity, err)
				}
				canonical, err := spec.Canonical()
				if err != nil || choice.ContractSHA256 != digest([]byte(canonical)) {
					t.Fatal("observation contract differs", name, choice.Activity, err)
				}
			}
		}
	}
}
