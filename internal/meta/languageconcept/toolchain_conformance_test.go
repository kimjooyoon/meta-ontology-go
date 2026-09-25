package languageconcept

import "testing"

func TestToolchainConformanceConceptBindsMetaLedger(t *testing.T) {
	item := Catalog()[21]
	if item.ID != "toolchain-conformance" ||
		item.MetaOperation != "close-toolchain-conformance-ledger" ||
		item.Stage != "OPERATING" || item.NoveltyClaim ||
		len(item.CodeBindings) != 5 || len(item.MetricBindings) != 28 ||
		len(item.UseCases) != 3 {
		t.Fatalf("concept = %#v", item)
	}
}
