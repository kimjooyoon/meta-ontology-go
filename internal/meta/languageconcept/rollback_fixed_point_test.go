package languageconcept

import "testing"

func TestRollbackFixedPointBindsProjectionAndRecovery(t *testing.T) {
	item := Catalog()[11]
	if item.ID != "rollback-fixed-point-recovery" ||
		item.MetaOperation != "recover-guarded-fixed-point" || item.Stage != "OPERATING" {
		t.Fatalf("concept = %#v", item)
	}
	if len(item.CodeBindings) != 4 || len(item.MetricBindings) != 16 {
		t.Fatalf("code=%v metrics=%v", item.CodeBindings, item.MetricBindings)
	}
	if item.MetricBindings[0] !=
		"gooo.metric.language.promotion-compatibility-readiness-bps.v1" ||
		item.MetricBindings[15] !=
			"gooo.metric.language.rollback-fixed-point-source-mutations.guardrail.v1" {
		t.Fatalf("metric boundary = %v", item.MetricBindings)
	}
	if len(item.UseCases) != 1 || item.UseCases[0].ExpectedOutcome !=
		"IMPROVED_11_TO_12_OF_24_PLUS_417_BPS_WITH_ZERO_WRITES" {
		t.Fatalf("use cases = %#v", item.UseCases)
	}
}
