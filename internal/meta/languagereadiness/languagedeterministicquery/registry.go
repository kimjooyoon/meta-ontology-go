package languagedeterministicquery

import "fmt"

var fixedCodeBindings = []string{
	"internal/query",
	"internal/meta/languagereadiness/languagedeterministicquery",
	"internal/meta/languagereadiness/languagedeterministicquerybinding",
	"cmd/language-deterministic-query-witness",
	"cmd/language-deterministic-query-readiness-binding",
	"examples/language-deterministic-query",
}

var fixedMetricBindings = []string{
	"gooo.metric.language.deterministic-query-bps.v1",
	"gooo.metric.language.deterministic-query-binding-plans.v1",
	"gooo.metric.language.deterministic-query-law-plans.v1",
	"gooo.metric.language.deterministic-query-canonical-replays.v1",
	"gooo.metric.language.deterministic-query-permutation-replays.v1",
	"gooo.metric.language.deterministic-query-concept-bindings.v1",
	"gooo.metric.language.deterministic-query-code-bindings.v1",
	"gooo.metric.language.deterministic-query-metric-bindings.v1",
	"gooo.metric.language.deterministic-query-use-case-bindings.v1",
	"gooo.metric.language.deterministic-query-not-satisfied.guardrail.v1",
	"gooo.metric.language.deterministic-query-unresolved.guardrail.v1",
	"gooo.metric.language.deterministic-query-registry-drift.guardrail.v1",
	"gooo.metric.language.deterministic-query-candidate-promotions.guardrail.v1",
	"gooo.metric.language.deterministic-query-unknown-acceptances.guardrail.v1",
	"gooo.metric.language.deterministic-query-graph-mutations.guardrail.v1",
	"gooo.metric.language.deterministic-query-effectful-stages.guardrail.v1",
	"gooo.metric.language.deterministic-query-repository-writes.guardrail.v1",
	"gooo.metric.language.deterministic-query-mutation-authorities.guardrail.v1",
}

var fixedUseCases = []string{
	"query-meta-bindings",
	"replay-query-plan",
	"reject-query-unknowns",
}

func Registry() PlanRegistry {
	cases := bindingDefinitions()
	cases = append(cases, lawDefinitions()...)
	return PlanRegistry{Schema: RegistrySchema, Version: RegistryVersion, Cases: cases}
}

func bindingDefinitions() []Definition {
	bindings := []struct {
		class string
		items []string
	}{
		{BindingConcept, []string{ConceptID}},
		{BindingCode, fixedCodeBindings},
		{BindingMetric, fixedMetricBindings},
		{BindingUseCase, fixedUseCases},
	}
	definitions := make([]Definition, 0, FixedBindingPlans)
	for _, binding := range bindings {
		for _, item := range binding.items {
			definitions = append(definitions, bindingDefinition(len(definitions), binding.class, item))
		}
	}
	return definitions
}

func bindingDefinition(index int, class, binding string) Definition {
	return Definition{
		ID:            fmt.Sprintf("binding-%02d", index+1),
		Kind:          CaseBinding,
		BindingClass:  class,
		Binding:       binding,
		ProofChoice:   bindingProofChoice(class),
		MetaOperation: "query-reified-meta-binding",
	}
}

func bindingProofChoice(class string) string {
	if class == BindingMetric || class == BindingUseCase {
		return "COHERENCE"
	}
	return "FOUNDATION"
}

func lawDefinitions() []Definition {
	return []Definition{
		{ID: "law-candidate-non-authority", Kind: CaseLaw, ProofChoice: "REGRESSION", MetaOperation: "reject-candidate-promotion"},
		{ID: "law-unknown-layer", Kind: CaseLaw, ProofChoice: "REGRESSION", MetaOperation: "fail-closed-unknown-layer"},
		{ID: "law-unknown-endpoint", Kind: CaseLaw, ProofChoice: "REGRESSION", MetaOperation: "fail-closed-unknown-endpoint"},
		{ID: "law-read-only-graph", Kind: CaseLaw, ProofChoice: "REGRESSION", MetaOperation: "seal-query-effects"},
	}
}
