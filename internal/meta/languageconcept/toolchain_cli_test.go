package languageconcept

import "testing"

func TestToolchainCLIConceptBindsExecutableMetaEvidence(t *testing.T) {
	item := Catalog()[19]
	if item.ID != "toolchain-cli" || item.MetaOperation != "evaluate-toolchain-cli-contract" ||
		item.Stage != "OPERATING" || item.NoveltyClaim || len(item.CodeBindings) != 5 ||
		len(item.MetricBindings) != 18 || len(item.UseCases) != 3 {
		t.Fatalf("concept = %#v", item)
	}
}
