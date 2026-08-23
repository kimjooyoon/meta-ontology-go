package languageconcept

import (
	"reflect"
	"testing"
)

func TestToolchainUseCasesBindExecutableEvidence(t *testing.T) {
	item := Catalog()[12]
	wantCode := []string{"internal/meta/languagereadiness/toolchainusecases",
		"cmd/toolchain-usecase-witness", "examples/toolchain-executable-use-cases"}
	wantMetrics := []string{
		"gooo.metric.toolchain.executable-use-cases-readiness-bps.v1",
		"gooo.metric.toolchain.executable-use-cases-executed-cases.v1",
		"gooo.metric.toolchain.executable-use-cases-pass-paths.v1",
		"gooo.metric.toolchain.executable-use-cases-fail-closed-paths.v1",
		"gooo.metric.toolchain.executable-use-cases-unresolved.guardrail.v1",
		"gooo.metric.toolchain.executable-use-cases-repository-writes.guardrail.v1",
		"gooo.metric.toolchain.executable-use-cases-mutation-authority.guardrail.v1",
		"gooo.metric.toolchain.executable-use-cases-registry-drift.guardrail.v1",
	}
	if item.ID != "toolchain-executable-use-cases" || item.MetaOperation != "execute-versioned-use-cases" ||
		item.Stage != "OPERATING" || !reflect.DeepEqual(item.CodeBindings, wantCode) ||
		!reflect.DeepEqual(item.MetricBindings, wantMetrics) {
		t.Fatalf("concept = %#v", item)
	}
	if len(item.UseCases) != 1 || item.UseCases[0].ExpectedOutcome != "IMPROVED_12_TO_13_OF_24_WITH_3_OF_3_CASES" {
		t.Fatalf("use cases = %#v", item.UseCases)
	}
}
