package languagesyntax

func expectedSyntaxCaseGroup5() []CaseDefinition {
	cases := expectedSyntaxCaseGroup5Part0()
	cases = append(cases, expectedSyntaxCaseGroup5Part1()...)
	return cases
}
func expectedSyntaxCaseGroup5Part0() []CaseDefinition {
	return []CaseDefinition{
		{ID: "local-field-bridge-reader", Path: "examples/local-field-bridge/trace.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, OrderedCollections: false, EntityFieldsProfile: "gooo.entityfields.go-projection.v7"},
		{ID: "field-source-map-reader", Path: "examples/field-source-map/inspect.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, OrderedCollections: true},
		{ID: "scoped-origin-coverage-reader", Path: "examples/scoped-origin-coverage/coverage.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, OrderedCollections: true},
		{ID: "length-aware-summary", Path: "examples/length-aware-summary/summary.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, OrderedCollections: true},
		{ID: "assembled-length-summary", Path: "examples/assembled-length-summary/summary.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, OrderedCollections: true},
		{ID: "record-choice-origin-reader", Path: "examples/record-choice-origins/inspect.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, OrderedCollections: true},
		{ID: "fill-candidate-origin-source", Path: "examples/fill-candidate-origins/source.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, OrderedCollections: true},
		{ID: "fill-candidate-origin-reader", Path: "examples/fill-candidate-origins/inspect.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, OrderedCollections: true},
		{ID: "derived-candidate-origin-source", Path: "examples/derived-candidate-origins/source.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, OrderedCollections: true},
		{ID: "derived-candidate-origin-reader", Path: "examples/derived-candidate-origins/inspect.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, OrderedCollections: true},
		{ID: "everyday-local-condition-origins", Path: "examples/everyday-program/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false},
		{ID: "compound-local-update-origins", Path: "examples/compound-updates/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false},
		{ID: "readable-origin-display-policy", Path: "internal/meta/originview/roles.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false},
		{ID: "local-version-identity-reader", Path: "examples/local-version-origins/inspect.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "same-value-local-version-producer", Path: "examples/local-version-origins/producer/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false},
		{ID: "json-text-escape-policy", Path: "internal/meta/textinput/escapes.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false},
		{ID: "json-text-raw-and-decoded-origins", Path: "examples/text-input-origins/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "recorded-local-use-relations", Path: "examples/recorded-use-origins/inspect.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "source-control-source", Path: "examples/source-control-reader/source.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "source-control-reader", Path: "examples/source-control-reader/inspect.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
	}
}
func expectedSyntaxCaseGroup5Part1() []CaseDefinition {
	cases := expectedSyntaxCaseGroup5Part1Part0()
	cases = append(cases, expectedSyntaxCaseGroup5Part1Part1()...)
	return cases
}
func expectedSyntaxCaseGroup5Part1Part0() []CaseDefinition {
	return []CaseDefinition{
		{ID: "source-owned-else-policy", Path: "internal/meta/bodysurface/else.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "else-layout-original-source", Path: "examples/else-layout/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "authored-input-names", Path: "examples/named-inputs/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "record-field-compound-origins", Path: "examples/compound-updates/fields.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "recorded-arithmetic-fault-reader", Path: "examples/fault-origins/describe.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "transport-run-origin-binding", Path: "internal/meta/selfimprovementtransport/originpolicy/source.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "explicit-consumer-continuation", Path: "internal/meta/executioncontinuation/policy/source.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "executable-metric-registry", Path: "internal/meta/metricprogram/registry.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "unresolved-operation-source-origin", Path: "internal/meta/metricprogram/unresolvedpolicy/policy.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false, EntityFieldsProfile: "gooo.entityfields.go-projection.v2"},
		{ID: "fixed-record-value-contract", Path: "examples/fixed-record-contract/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "advisory-origin-lens-policy", Path: "internal/meta/originlenspolicy/policy.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false},
		{ID: "source-text-field-byte-bound", Path: "examples/bounded-text/main.gooo", Kind: KindValid,
			ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax",
			Scope: ScopeLanguageCapability, EntityFieldsProfile: "gooo.entityfields.go-projection.v6"},
		{ID: "recorded-field-parent-origins", Path: "examples/field-parent-origins/parent.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "scope-fragment-projection", Path: "examples/scope-fragments/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false, EntityFieldsProfile: "gooo.entityfields.go-projection.v6"},
		{ID: "constant-helper-calls", Path: "examples/constant-helpers/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "partial-reuse-source-assessment", Path: "internal/meta/publicpartialreuse/policy.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "source-owned-collection-filter", Path: "examples/collection-filter/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "source-split-completion-policy", Path: "internal/meta/sourcecompletion/source.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},
		{ID: "shared-workspace-source-consumers", Path: "examples/shared-source-consumers/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false},
		{ID: "source-owned-record-collections", Path: "examples/record-collections/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false, EntityFieldsProfile: "gooo.entityfields.go-projection.v7"},
	}
}
func expectedSyntaxCaseGroup5Part1Part1() []CaseDefinition {
	return []CaseDefinition{
		{ID: "source-owned-nominal-record-assembly", Path: "examples/record-collections/assembly.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false, EntityFieldsProfile: "gooo.entityfields.go-projection.v7"},

		{ID: "source-owned-profile-compatibility-records", Path: "internal/meta/syntaxregistration/contract.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false, EntityFieldsProfile: "gooo.entityfields.go-projection.v7"},

		{ID: "openfga-source-owned-model-change", Path: "examples/openfga-model-change/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false, EntityFieldsProfile: "gooo.entityfields.go-projection.v7"},

		{ID: "source-owned-record-origin-query", Path: "examples/record-origin-reader/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false, EntityFieldsProfile: "gooo.entityfields.go-projection.v7"},

		{ID: "record-element-location-consumer", Path: "examples/record-element-location/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: true},

		{ID: "loop-carried-source-values", Path: "examples/loop-carried-values/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false, EntityFieldsProfile: "gooo.entityfields.go-projection.v7"},

		{ID: "indexed-record-source-values", Path: "examples/indexed-record-values/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false, EntityFieldsProfile: "gooo.entityfields.go-projection.v7"},

		{ID: "constructed-record-source-values", Path: "examples/constructed-record-values/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: false, ImplicitActivityPorts: false, OrderedCollections: false, EntityFieldsProfile: "gooo.entityfields.go-projection.v7"},
	}
}
