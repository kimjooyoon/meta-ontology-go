package languageconcept

import "testing"

func TestSemanticModelBindsMetaCodeMetricsAndUseCase(t *testing.T) {
	item := Catalog()[14]
	identity := item.ID == "language-semantic-model"
	identity = identity && item.MetaOperation == "prove-staged-semantic-model"
	identity = identity && item.Stage == "OPERATING"
	if !identity {
		t.Fatalf("concept = %#v", item)
	}
	if len(item.CodeBindings) != 6 || len(item.MetricBindings) != 19 {
		t.Fatalf("code=%v metrics=%v", item.CodeBindings, item.MetricBindings)
	}
	if len(item.UseCases) != 1 || item.UseCases[0].ExpectedOutcome != "IMPROVED_14_TO_15_OF_24_WITH_18_OF_18_CASES_AND_ZERO_EFFECTS" {
		t.Fatalf("use cases = %#v", item.UseCases)
	}
}
