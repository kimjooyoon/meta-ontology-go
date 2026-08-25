package languageconcept

import "testing"

func TestLanguageSyntaxConceptBindsReceiptMetaCode(t *testing.T) {
	item := Catalog()[13]
	if item.ID != "language-syntax-roundtrip" || item.MetaOperation != "prove-language-syntax-roundtrip" ||
		item.Stage != "OPERATING" || len(item.CodeBindings) != 6 || len(item.MetricBindings) != 16 {
		t.Fatalf("concept = %#v", item)
	}
	if len(item.UseCases) != 1 || item.UseCases[0].ExpectedOutcome != "IMPROVED_13_TO_14_OF_24_WITH_20_OF_20_CASES" {
		t.Fatalf("use cases = %#v", item.UseCases)
	}
}
