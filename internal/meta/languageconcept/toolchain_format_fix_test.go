package languageconcept

import "testing"

func TestToolchainFormatFixConceptBindsMetaApplication(t *testing.T) {
	item := Catalog()[20]
	if item.ID != "toolchain-format-fix" ||
		item.MetaOperation != "evaluate-toolchain-format-fix" ||
		item.Stage != "OPERATING" || item.NoveltyClaim || len(item.CodeBindings) != 6 ||
		len(item.MetricBindings) != 18 || len(item.UseCases) != 3 {
		t.Fatalf("concept = %#v", item)
	}
}
