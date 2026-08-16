package pressureindependence

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCorpus(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if corpus.Schema != CorpusSchemaV1 || len(corpus.Cases) < 11 {
		t.Fatalf("corpus schema/count = %q/%d", corpus.Schema, len(corpus.Cases))
	}
	if corpus.CanonicalDigest != CorpusDigest(corpus) {
		t.Logf("corpus canonical_digest=%q", CorpusDigest(corpus))
		t.Errorf("corpus digest mismatch")
	}
	seen := make(map[string]struct{}, len(corpus.Cases))
	for _, row := range corpus.Cases {
		if row.Name == "" {
			t.Errorf("empty corpus case name")
		}
		if _, exists := seen[row.Name]; exists {
			t.Errorf("duplicate corpus case %q", row.Name)
		}
		seen[row.Name] = struct{}{}
		got := Evaluate(row.Input)
		if row.Expected.CanonicalOutputDigest == "" {
			encoded, _ := json.Marshal(got)
			t.Logf("%s expected=%s", row.Name, encoded)
			continue
		}
		if !reflect.DeepEqual(got, row.Expected) {
			t.Errorf("%s output = %#v, want %#v", row.Name, got, row.Expected)
		}
	}
}

func TestStrictInputJSON(t *testing.T) {
	base := mustCorpusInput(t, "two-independent-groups-pass")
	data, err := json.Marshal(base)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeInput(append(data, []byte(` {"schema":"extra"}`)...)); err == nil {
		t.Fatal("trailing JSON value accepted")
	}
	duplicate := strings.Replace(string(data), `"schema":"gooo/pressure-independence/v1"`,
		`"schema":"gooo/pressure-independence/v1","schema":"gooo/pressure-independence/v1"`, 1)
	if _, err := DecodeInput([]byte(duplicate)); err == nil {
		t.Fatal("duplicate JSON key accepted")
	}
	unknown := strings.TrimSuffix(string(data), "}") + `,"unknown":true}`
	if _, err := DecodeInput([]byte(unknown)); err == nil {
		t.Fatal("unknown JSON field accepted")
	}
}

func TestPermutationAndExpectedMutationInvariants(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	input := mustCorpusInput(t, "two-independent-groups-pass")
	permuted := input
	permuted.PressureRecords = append([]PressureRecord(nil), input.PressureRecords...)
	permuted.RequiredPressureIDs = append([]string(nil), input.RequiredPressureIDs...)
	permuted.GuardIDs = append([]string(nil), input.GuardIDs...)
	permuted.FinitePathIDs = append([]string(nil), input.FinitePathIDs...)
	reverseRecords(permuted.PressureRecords)
	reverseStrings(permuted.RequiredPressureIDs)
	reverseStrings(permuted.GuardIDs)
	reverseStrings(permuted.FinitePathIDs)
	if CanonicalInputDigest(input) != CanonicalInputDigest(permuted) ||
		!reflect.DeepEqual(Evaluate(input), Evaluate(permuted)) {
		t.Fatal("permutation changed input digest or output")
	}
	left := Evaluate(input)
	right := left
	right.Decision, right.Reason = DecisionFailClosed, ReasonPredicateFalse
	if CanonicalInputDigest(input) != CanonicalInputDigest(input) || left.InputDigest != right.InputDigest {
		t.Fatal("expected-only mutation changed input digest")
	}
	if CanonicalOutputDigest(left) == CanonicalOutputDigest(right) {
		t.Fatal("output mutation did not change output digest")
	}
	mutatedExpected := corpus.Cases[0]
	mutatedExpected.Expected.Decision = DecisionFailClosed
	if !reflect.DeepEqual(Evaluate(mutatedExpected.Input), Evaluate(corpus.Cases[0].Input)) {
		t.Fatal("expected-only mutation changed oracle result")
	}
}

func TestGroupMutationChangesResult(t *testing.T) {
	input := mustCorpusInput(t, "two-independent-groups-pass")
	mutated := input
	mutated.PressureRecords = append([]PressureRecord(nil), input.PressureRecords...)
	mutated.PressureRecords[1].IndependenceGroupID = mutated.PressureRecords[0].IndependenceGroupID
	left, stale := Evaluate(input), Evaluate(mutated)
	if left.Decision != DecisionPass || stale.Decision != DecisionUnknown || stale.Reason != ReasonStaleDigest {
		t.Fatalf("group mutation left=%#v stale=%#v", left, stale)
	}
	contracts, ok := readArtifactContracts()
	if !ok {
		t.Fatal("artifact contracts invalid")
	}
	mutated.RegistryDigest = registryBindingDigest(mutated.PressureRecords, contracts.registry)
	right := Evaluate(mutated)
	if right.Decision != DecisionUnknown || right.Reason != ReasonIndependentGroupShortfall ||
		left.InputDigest == right.InputDigest {
		t.Fatalf("rebound group mutation left=%#v right=%#v", left, right)
	}
}

func TestRegistryBindingTracksCategoryAndApplicability(t *testing.T) {
	input := mustCorpusInput(t, "two-independent-groups-pass")
	mutations := []struct {
		name   string
		mutate func(*PressureRecord)
	}{
		{name: "category", mutate: func(record *PressureRecord) { record.CategoryID = "category-mutated" }},
		{name: "applicability", mutate: func(record *PressureRecord) { record.ApplicabilityRuleID = "rule-v2" }},
	}
	for _, mutation := range mutations {
		mutated := input
		mutated.PressureRecords = append([]PressureRecord(nil), input.PressureRecords...)
		mutation.mutate(&mutated.PressureRecords[1])
		got := Evaluate(mutated)
		if got.Decision != DecisionUnknown || got.Reason != ReasonStaleDigest {
			t.Fatalf("%s mutation with old registry = %#v", mutation.name, got)
		}
	}
}

func TestAmbiguousApplicabilityAndNonCompensatingResources(t *testing.T) {
	input := mustCorpusInput(t, "two-independent-groups-pass")
	ambiguous := input
	ambiguous.PressureRecords = append([]PressureRecord(nil), input.PressureRecords...)
	ambiguous.PressureRecords[1].ApplicabilityRuleID = "rule-v2"
	contracts, ok := readArtifactContracts()
	if !ok {
		t.Fatal("artifact contracts invalid")
	}
	ambiguous.RegistryDigest = registryBindingDigest(ambiguous.PressureRecords, contracts.registry)
	if got := Evaluate(ambiguous); got.Decision != DecisionUnknown || got.Reason != ReasonInputAmbiguous {
		t.Fatalf("ambiguous applicability = %#v", got)
	}
	resource := input
	resource.ResourceCeilings = input.ResourceCeilings
	resource.ResourceCeilings.MemoryBytes = 1
	resource.ResourceCeilings.CPUCoreNS = 1000
	if got := Evaluate(resource); got.Decision != DecisionUnknown || got.Reason != ReasonInvalidResourceReceipt {
		t.Fatalf("compensated resource receipt = %#v", got)
	}
}

func TestBaselineBudgetAndNoUniqueBenefit(t *testing.T) {
	comparison := Compare(mustCorpusInput(t, "two-independent-groups-pass"))
	if !comparison.OutcomeMatch || !comparison.ReasonMatch ||
		!comparison.LocalizationMatch || !comparison.ResearchBudgetOK {
		t.Fatalf("comparison = %#v", comparison)
	}
	if comparison.Finding != NoUniqueBenefit || comparison.ResearchWorkUnits > comparison.BaselineWorkUnits {
		t.Fatalf("comparison finding/budget = %#v", comparison)
	}
}

func TestOutputHasNoPromotionAuthorization(t *testing.T) {
	data, err := EncodeOutputJSON(Evaluate(mustCorpusInput(t, "two-independent-groups-pass")))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "promotion_authorized") {
		t.Fatal("promotion authorization escaped the research output")
	}
}

func mustCorpusInput(t testing.TB, name string) Input {
	t.Helper()
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range corpus.Cases {
		if row.Name == name {
			return row.Input
		}
	}
	t.Fatalf("corpus case %q not found", name)
	return Input{}
}

func reverseRecords(values []PressureRecord) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func reverseStrings(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
