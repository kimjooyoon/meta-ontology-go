package languagesyntax

func expectedSyntaxCaseGroup2() []CaseDefinition {
	return []CaseDefinition{
		{ID: "language-value-witness-equal", Path: "examples/language-value-witness/equal.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-and", Path: "examples/language-value-witness/and.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-runtime-binding", Path: "examples/language-runtime-binding/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-runtime-binding-and", Path: "examples/language-runtime-binding/boolean-and.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-runtime-binding-typed-chain", Path: "examples/language-runtime-binding/typed-chain.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-runtime-binding-suspend-resume", Path: "examples/language-runtime-binding/suspend-resume.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-record-binding", Path: "examples/language-record-binding/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability, EntityFields: true},
		{ID: "self-improvement-value-witness-execution-input", Path: "examples/self-improvement-value-witness-execution-input/contract.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-operation-catalog", Path: "examples/language-operation-catalog/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-operation-catalog-unknown", Path: "examples/language-operation-catalog/unknown.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "claim-resolution-tuple", Path: "cmd/gooo/testdata/claim-resolution/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "claim-dependency-causality", Path: "cmd/gooo/testdata/claim-dependency/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "roundtrip-minimal", Path: "internal/detection/roundtrip/testdata/minimal.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "directory-kind-ontology", Path: "internal/meta/directorykind/ontology.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "directory-partition-ontology", Path: "internal/meta/directorypartition/ontology.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "repository-projection-repair", Path: "examples/repository-projection-repair/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "opentofu-observation", Path: "examples/opentofu-observation/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "public-self-observation-discovery-policy", Path: "examples/self-improvement-discovery/discovery.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "public-self-observation-discovery-project", Path: "examples/self-improvement-discovery/project.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "unknown-keyword", Path: "examples/language-syntax-roundtrip/unknown-keyword.txt", Kind: KindInvalid, ExpectedDecision: DecisionClosed, ExpectedDiagnostic: "parse.unexpected-token", ProofChoice: "REGRESSION", MetaOperation: "reject-invalid-syntax", Scope: ScopeLanguageCapability},
	}
}
