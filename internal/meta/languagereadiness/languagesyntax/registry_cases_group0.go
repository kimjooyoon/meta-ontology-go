package languagesyntax

func expectedSyntaxCaseGroup0() []CaseDefinition {
	return []CaseDefinition{
		{ID: "billing", Path: "examples/billing/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-test-pass", Path: "examples/language-test/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-test-failing-assertion", Path: "examples/language-test/failing.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "bootstrap", Path: "examples/bootstrap/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "conformance", Path: "examples/conformance/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "compiler-self-improvement", Path: "examples/compiler-self-improvement/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "compiler-self-improvement-operation-envelope", Path: "examples/compiler-self-improvement/operation-envelope.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "causal-ci-selection", Path: "examples/causal-ci-selection/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "causal-ci-selection-semantic-intervention", Path: "examples/causal-ci-selection/semantic-intervention.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "causal-ci-selection-nonsemantic-intervention", Path: "examples/causal-ci-selection/nonsemantic-intervention.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "causal-ci-selection-contradiction-intervention", Path: "examples/causal-ci-selection/contradiction-intervention.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "external-capability-authorization-policy", Path: "examples/external-capability-execution/authorization/policy.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "activity-cardinality-resolution", Path: "examples/activity-cardinality-resolution/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "meta-actionability", Path: "examples/meta-actionability/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "meta-binding-coverage", Path: "examples/meta-binding-coverage/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "meta-operation-artifact-coverage", Path: "examples/meta-operation-artifact-coverage/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "meta-policy-compilation", Path: "examples/meta-policy-compilation/policy.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "metric-meta-program-closure", Path: "examples/metric-meta-program-closure/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "metric-meta-program", Path: "examples/metric-meta-program/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "root-readme-indicator", Path: "examples/root-readme-indicator/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
	}
}
