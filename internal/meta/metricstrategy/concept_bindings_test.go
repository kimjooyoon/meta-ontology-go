package metricstrategy

import (
	"os"
	"strings"
	"testing"
)

func TestLanguageConceptsGovernStrategyBindings(t *testing.T) {
	source := []Binding{
		{IndicatorID: "f1", Family: "FOUNDATION", Trilemma: "AXIOM", MetaOperation: "bind-exact-source-metrics"},
		{IndicatorID: "f2", Family: "FOUNDATION", Trilemma: "AXIOM", MetaOperation: "exempt-project-root-topology"},
		{IndicatorID: "f3", Family: "FOUNDATION", Trilemma: "AXIOM", MetaOperation: "interpret-dimension-registry"},
		{IndicatorID: "c", Family: "COHERENCE", Trilemma: "COHERENCE", MetaOperation: "project-algebraic-root-state"},
		{IndicatorID: "r1", Family: "REGRESSION", Trilemma: "REGRESS", MetaOperation: "observe-counterfactual-boundary"},
		{IndicatorID: "r2", Family: "REGRESSION", Trilemma: "REGRESS", MetaOperation: "preserve-repository-workspace"},
		{IndicatorID: "r3", Family: "REGRESSION", Trilemma: "REGRESS", MetaOperation: "replay-counterfactual"},
	}
	bindings, err := bindLanguageConcepts(os.DirFS("../../.."), source)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{"outcome": 0, "driver": 0, "guardrail": 0, "operation": 0}
	for _, binding := range bindings {
		for class := range counts {
			if strings.HasPrefix(binding.IndicatorID, "gooo.strategy.concept."+class+".") || class == "operation" && strings.HasPrefix(binding.IndicatorID, "gooo.concept.operation.") {
				counts[class]++
			}
		}
	}
	if len(bindings) != 23 || counts["outcome"] != 1 || counts["driver"] != 3 || counts["guardrail"] != 3 || counts["operation"] != 9 {
		t.Fatalf("concept governance counts=%v bindings=%d", counts, len(bindings))
	}
}
