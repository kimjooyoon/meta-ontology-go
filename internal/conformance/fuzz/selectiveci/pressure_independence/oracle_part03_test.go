package pressureindependence

import (
	"strings"
	"testing"
)

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
