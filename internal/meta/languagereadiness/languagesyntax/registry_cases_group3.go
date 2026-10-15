package languagesyntax

func expectedSyntaxCaseGroup3() []CaseDefinition {
	return []CaseDefinition{
		{ID: "unterminated-string", Path: "examples/language-syntax-roundtrip/unterminated-string.txt", Kind: KindInvalid, ExpectedDecision: DecisionClosed, ExpectedDiagnostic: "lex.unterminated-string", ProofChoice: "REGRESSION", MetaOperation: "reject-invalid-syntax", Scope: ScopeLanguageCapability},
		{ID: "source-execution-invalid", Path: "examples/language-source-execution/invalid.gooo", Kind: KindInvalid, ExpectedDecision: DecisionClosed, ExpectedDiagnostic: "parse.unexpected-token", ProofChoice: "REGRESSION", MetaOperation: "reject-invalid-syntax", Scope: ScopeLanguageCapability},
		{ID: "source-splitter-conformance", Path: "examples/source-splitter-conformance/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "rollback-integrity-activation", Path: "examples/rollback-integrity-activation/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "vertical-slice-closure-activation", Path: "examples/vertical-slice-closure-activation/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "workgraph", Path: "examples/workgraph/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "external-conformance-activation", Path: "examples/external-conformance-activation/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "gooo-release-publication", Path: "examples/gooo-release-publication/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "ci-plan", Path: "examples/ci-plan/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "ci-effort-observation", Path: "examples/ci-effort-observation/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "ci-time-causality", Path: "examples/ci-time-causality/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, ImplicitActivityPorts: true},
		{ID: "reproducibility-semantics", Path: "examples/reproducibility-semantics/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "entity-fields-v1", Path: "examples/entity-fields-v1/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: true},
		{ID: "temporal-transition-ticket", Path: "examples/temporal-transition-ticket/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "domain-observation-definition", Path: "examples/domain-observation/definition.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "domain-observation-runtime-binding", Path: "examples/domain-observation/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "domain-observation-incident", Path: "examples/domain-observation/incident.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "domain-observation-repair", Path: "examples/domain-observation/repair.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "repair-handoff", Path: "examples/repair-handoff/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "self-improvement-ci-continuation", Path: "examples/self-improvement-ci-continuation/continuation.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeGovernanceObservation},
	}
}
