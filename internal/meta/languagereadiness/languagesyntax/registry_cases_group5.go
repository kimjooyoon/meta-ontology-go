package languagesyntax

func expectedSyntaxCaseGroup5() []CaseDefinition {
	return []CaseDefinition{
		{ID: "local-field-bridge-reader", Path: "examples/local-field-bridge/trace.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, OrderedCollections: true},
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
	}
}
