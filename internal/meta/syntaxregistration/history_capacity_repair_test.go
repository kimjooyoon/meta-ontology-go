package syntaxregistration

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func parseHistoryCapacityFixture(t *testing.T, raw []byte) *goSource {
	t.Helper()
	source, err := parseSourceUnits(map[string][]byte{closureRoot + "history.go": raw}, closureRoot, "verticalsliceclosureshadow", false)
	if err != nil {
		t.Fatal(err)
	}
	source.activity = "gooo://syntax-registration/activity/denominator-admission"
	return source
}

func historyCapacityLegacyFixture() []byte {
	var text strings.Builder
	text.WriteString("package verticalsliceclosureshadow\nfunc expectedBoundarySpecsForVersion(version int) []boundarySpec {\nexpected := expectedBoundarySpecs()\nif version <= 46 { expected[5].Target=20 }\nswitch version {\n")
	for version := 101; version <= 173; version++ {
		fmt.Fprintf(&text, "case %d: expected[0].Target=%d\n", version, version+811)
	}
	text.WriteString("}\nreturn expected\n}\n")
	return []byte(text.String())
}

func TestHistoryCapacityGrowthPreservesEveryDeclaredTarget(t *testing.T) {
	raw := historyCapacityLegacyFixture()
	for step := range 53 {
		source := parseHistoryCapacityFixture(t, raw)
		if err := pinPreviousSyntaxTarget(source, 174+step, 985+step); err != nil {
			t.Fatal(err)
		}
		changes, err := source.finish()
		if err != nil {
			t.Fatal(err)
		}
		raw = changes[closureRoot+"history.go"]
		next := parseHistoryCapacityFixture(t, raw)
		history, err := readHistoryOrigins(next)
		if err != nil {
			t.Fatal(err)
		}
		if len(history.targets) != 74+step {
			t.Fatal("history population changed", len(history.targets))
		}
		for version := 101; version <= 174+step; version++ {
			if history.targets[version] != version+811 {
				t.Fatal("prior target changed", version)
			}
		}
		for _, declaration := range next.file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if lines := strings.Count(next.text(function), "\n") + 1; lines > 60 {
				t.Fatalf("unbounded history declaration %s at step %d: %d", function.Name, step, lines)
			}
		}
		if !strings.Contains(string(raw), "expected[5].Target = 20") {
			t.Fatal("release history changed")
		}
	}
	t.Log("53 fresh append steps; 126 declared historical targets retained; each generated function <=60 lines; producer/model 0")
}

func TestHistoryCapacityRejectsAmbiguousOwnershipAndCycles(t *testing.T) {
	source := parseHistoryCapacityFixture(t, historyCapacityLegacyFixture())
	history, err := readHistoryOrigins(source)
	if err != nil {
		t.Fatal(err)
	}
	writeHistoryOrigins(source, history)
	changes, err := source.finish()
	if err != nil {
		t.Fatal(err)
	}
	raw := string(changes[closureRoot+"history.go"])
	variants := []string{
		strings.Replace(raw, "case 102:", "case 101:", 1),
		strings.Replace(raw, "return 912, true", "return 912, false", 1),
		strings.Replace(raw, "return historicalSyntaxTargetChunk1(version)", "return historicalSyntaxTargetChunk0(version)", 1),
		strings.Replace(raw, "; known {", "; !known {", 1),
		raw + "\nfunc historicalSyntaxTargetChunk99(version int)(int,bool){return 0,false}\n",
	}
	for index, text := range variants {
		if text == raw {
			t.Fatal("mutation did not change fixture", index)
		}
		candidate := parseHistoryCapacityFixture(t, []byte(text))
		if err := pinPreviousSyntaxTarget(candidate, 174, 985); err == nil || len(candidate.edits) != 0 {
			t.Fatal("unsupported history generated edits", index, err)
		}
	}
	mismatch := parseHistoryCapacityFixture(t, []byte(raw))
	if err := pinPreviousSyntaxTarget(mismatch, 101, 1); err == nil || len(mismatch.edits) != 0 {
		t.Fatal("historical target rewritten")
	}
}

func TestHistoryCapacityCurrentSourceNormalizationProposal(t *testing.T) {
	root := "../../.."
	paths, err := filepath.Glob(filepath.Join(root, closureRoot, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	inputs := map[string][]byte{}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		inputs[filepath.ToSlash(relative)] = raw
	}
	source, err := parseSourceUnits(inputs, closureRoot, "verticalsliceclosureshadow", false)
	if err != nil {
		t.Fatal(err)
	}
	source.activity = "gooo://syntax-registration/activity/denominator-admission"
	history, err := readHistoryOrigins(source)
	if err != nil {
		t.Fatal(err)
	}
	writeHistoryOrigins(source, history)
	changes, err := source.finish()
	if err != nil {
		t.Fatal(err)
	}
	updated := historyNormalizationUnits(t, source, history, changes)
	after, err := parseSourceUnits(updated, closureRoot, "verticalsliceclosureshadow", false)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := readHistoryOrigins(after)
	if err != nil || !reflect.DeepEqual(history.targets, observed.targets) {
		t.Fatal("historical target changed", err)
	}
	if output := os.Getenv("GOOO_HISTORY_REPAIR_PROPOSAL"); output != "" {
		members := []Member{}
		for path, raw := range changes {
			members = append(members, Member{Path: path, BeforeDigest: digest(inputs[path]), AfterDigest: digest(raw), Content: raw})
		}
		raw, err := json.MarshalIndent(members, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(output, append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("normalization: %d historical targets preserved; %d helpers; no corpus or denominator evidence edits", len(observed.targets), len(observed.helpers))
}
