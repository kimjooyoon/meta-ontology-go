package formattercontract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type manifest struct {
	Schema            string           `json:"schema"`
	Stage             string           `json:"stage"`
	EvidenceStatus    string           `json:"evidence_status"`
	PromotionEligible bool             `json:"promotion_eligible"`
	Positive          positiveFixture  `json:"positive_fixture"`
	Counterexamples   []counterexample `json:"counterexamples"`
}

type positiveFixture struct {
	Input            string       `json:"input"`
	ExpectedOutput   string       `json:"expected_output"`
	ExpectedMeasures measurements `json:"expected_measurements"`
	ExpectedResult   string       `json:"expected_result"`
}

type measurements struct {
	InputBytes          int `json:"input_bytes"`
	ExpectedOutputBytes int `json:"expected_output_bytes"`
	SemanticNodes       int `json:"semantic_nodes"`
	UsedRelations       int `json:"prov_used_relations"`
	GeneratedRelations  int `json:"prov_wasGeneratedBy_relations"`
	SemanticRelations   int `json:"semantic_relations"`
	UnrelatedRewrites   int `json:"unrelated_rewrites"`
	DiagnosticCount     int `json:"diagnostic_count"`
}

type counterexample struct {
	ID                 string `json:"id"`
	Input              any    `json:"input"`
	ExpectedDiagnostic string `json:"expected_diagnostic"`
	ExpectedOutput     string `json:"expected_output"`
	ExpectedResult     string `json:"expected_result"`
}

func TestFormatterContractManifestIsDeferredAndComplete(t *testing.T) {
	root := fixtureRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "contract.json"))
	if err != nil {
		t.Fatal(err)
	}
	var contract manifest
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatal(err)
	}
	if contract.Schema != "gooo/formatter-contract/v1" || contract.Stage != "formatter-contract-only" {
		t.Fatalf("unexpected contract identity: %#v", contract)
	}
	if contract.EvidenceStatus != "deferred" || contract.PromotionEligible {
		t.Fatalf("unimplemented contract was marked promotable: %#v", contract)
	}
	fixtures := []string{contract.Positive.Input, contract.Positive.ExpectedOutput,
		"positive.drifted.gooo", "negative-unknown-reference.gooo", "adapter-document.json"}
	for _, name := range fixtures {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatalf("missing fixture %q: %v", name, err)
		}
	}
	if contract.Positive.ExpectedResult != "semantic_fingerprint_equal" {
		t.Fatalf("unexpected positive result: %q", contract.Positive.ExpectedResult)
	}
	measures := contract.Positive.ExpectedMeasures
	if measures.InputBytes != 166 || measures.ExpectedOutputBytes != 164 || measures.SemanticNodes != 3 ||
		measures.UsedRelations != 1 || measures.GeneratedRelations != 1 || measures.SemanticRelations != 2 ||
		measures.UnrelatedRewrites != 0 || measures.DiagnosticCount != 0 {
		t.Fatalf("unexpected positive measurements: %#v", measures)
	}
	if len(contract.Counterexamples) != 4 {
		t.Fatalf("counterexample count = %d, want 4", len(contract.Counterexamples))
	}
}

func TestFormatterContractMeasurementsMatchFixtures(t *testing.T) {
	root := fixtureRoot(t)
	source := readFixture(t, root, "positive.gooo")
	golden := readFixture(t, root, "positive.golden.gooo")
	drifted := readFixture(t, root, "positive.drifted.gooo")
	if len([]byte(source)) != 166 || len([]byte(golden)) != 164 || len([]byte(drifted)) != 172 {
		t.Fatalf("fixture byte measurements changed: source=%d golden=%d drifted=%d", len(source), len(golden), len(drifted))
	}
	if countPrefix(source, "entity ") != 2 || countPrefix(source, "activity ") != 1 {
		t.Fatalf("fixture declaration measurements changed")
	}
	if !strings.Contains(source, "activity Transform( Input ) -> Output") {
		t.Fatal("positive fixture no longer contains the formatting perturbation")
	}
	if strings.Contains(golden, "( Input )") || !strings.HasSuffix(golden, "\n") {
		t.Fatal("golden fixture is not canonical")
	}
	t.Logf("formatter fixture bytes: input=%d golden=%d drifted=%d; nodes=3 relations=2",
		len(source), len(golden), len(drifted))
}

func fixtureRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate fixture test")
	}
	return filepath.Join(filepath.Dir(file), "..", "testdata", "contract", "formatter")
}

func readFixture(t *testing.T, root, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

func countPrefix(source, prefix string) int {
	count := 0
	for _, line := range strings.Split(source, "\n") {
		if strings.HasPrefix(line, prefix) {
			count++
		}
	}
	return count
}
