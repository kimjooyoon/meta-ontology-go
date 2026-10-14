package languagesyntax

func expectedSyntaxCaseGroup1() []CaseDefinition {
	return []CaseDefinition{
		{ID: "self-improvement", Path: "examples/self-improvement/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "self-improvement-minimal-loop", Path: "examples/self-improvement-minimal-loop/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "self-improvement-operation-envelope", Path: "examples/self-improvement-minimal-loop/operation-envelope.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "self-improvement-candidate", Path: "examples/self-improvement/candidate.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "self-improvement-candidate-authorization", Path: "examples/self-improvement/authorization.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "self-improvement-transport", Path: "examples/self-improvement/transport.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "self-improvement-execution-contract", Path: "examples/self-improvement-execution-contract/contract.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "self-improvement-execution-grant", Path: "examples/self-improvement-execution-grant/grant.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness", Path: "examples/language-value-witness/main.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-sign", Path: "examples/language-value-witness/sign.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-mod", Path: "examples/language-value-witness/mod.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-neg", Path: "examples/language-value-witness/neg.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-abs", Path: "examples/language-value-witness/abs.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-add", Path: "examples/language-value-witness/add.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-mul", Path: "examples/language-value-witness/mul.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-div", Path: "examples/language-value-witness/div.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-max", Path: "examples/language-value-witness/max.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-min", Path: "examples/language-value-witness/min.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-iszero", Path: "examples/language-value-witness/iszero.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
		{ID: "language-value-witness-not", Path: "examples/language-value-witness/not.gooo", Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", Scope: ScopeLanguageCapability},
	}
}
