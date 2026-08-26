package languageconcept

import (
	"reflect"
	"testing"
)

func TestGuardedPromotionBindsCapabilityMetrics(t *testing.T) {
	item := Catalog()[10]
	wantCode := []string{
		"internal/meta/languagereadiness/guardedpromotion",
		"internal/meta/languagereadiness/guardedcapability",
		"cmd/language-readiness-witness/guarded-capability",
	}
	wantMetrics := []string{
		"gooo.metric.language.guarded-capability-capability-readiness-bps.v1",
		"gooo.metric.language.guarded-capability-foundation-receipts.v1",
		"gooo.metric.language.guarded-capability-implementation-tree-equivalence-bps.v1",
		"gooo.metric.language.guarded-capability-foundation-ancestor-bps.v1",
		"gooo.metric.language.guarded-capability-unresolved-evidence.v1",
		"gooo.metric.language.guarded-capability-implementation-tree-drift.v1",
		"gooo.metric.language.guarded-capability-observer-writes.v1",
		"gooo.metric.language.guarded-capability-mutation-authority.v1",
	}
	if item.ID != "guarded-exact-promotion" ||
		item.MetaOperation != "bind-guarded-capability-foundation" {
		t.Fatalf("concept = %#v", item)
	}
	if !reflect.DeepEqual(item.CodeBindings, wantCode) ||
		!reflect.DeepEqual(item.MetricBindings, wantMetrics) {
		t.Fatalf("code=%v metrics=%v", item.CodeBindings, item.MetricBindings)
	}
	if len(item.UseCases) != 1 || item.UseCases[0].ExpectedOutcome !=
		"IMPROVED_10_TO_11_OF_24_PLUS_417_BPS_WITH_EVENT_FAILURE_PRESERVED" {
		t.Fatalf("use cases = %#v", item.UseCases)
	}
}
