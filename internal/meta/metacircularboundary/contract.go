package metacircularboundary

func DenominatorContract() Denominator {
	cases := []CaseDefinition{
		{ID: "description-only", ExpectedDecision: DecisionPass, ExpectedAuthorization: AuthorizationDenied, ExpectedExecution: ExecutionBlocked, ExpectedReason: ReasonDescriptionOnly, ProofChoice: ProofRegression, MetaOperation: "deny-description-authority-escalation"},
		{ID: "explicit-read-only-capability", ExpectedDecision: DecisionPass, ExpectedAuthorization: AuthorizationGranted, ExpectedExecution: ExecutionAllowed, ExpectedReason: ReasonExplicitCapability, ProofChoice: ProofCoherence, MetaOperation: "accept-explicit-read-only-capability"},
		{ID: "forged-capability", ExpectedDecision: DecisionPass, ExpectedAuthorization: AuthorizationDenied, ExpectedExecution: ExecutionBlocked, ExpectedReason: ReasonForgedCapability, ProofChoice: ProofRegression, MetaOperation: "reject-forged-capability"},
		{ID: "write-capability-out-of-scope", ExpectedDecision: DecisionPass, ExpectedAuthorization: AuthorizationDenied, ExpectedExecution: ExecutionBlocked, ExpectedReason: ReasonOutOfScopeCapability, ProofChoice: ProofRegression, MetaOperation: "reject-out-of-scope-capability"},
	}
	denominator := Denominator{ID: DenominatorID, Cases: cases}
	denominator.Digest = digestValue(denominator)
	return denominator
}

func MetaOperations() []MetaOperation {
	return []MetaOperation{
		{ID: "bind-self-description", Producer: "metacircularboundary.observeSource", Consumer: "metacircularboundary.IndependentJudge", ProofChoice: ProofFoundation},
		{ID: "deny-description-authority-escalation", Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundary.IndependentJudge", ProofChoice: ProofRegression},
		{ID: "accept-explicit-read-only-capability", Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundary.IndependentJudge", ProofChoice: ProofCoherence},
		{ID: "reject-forged-capability", Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundary.IndependentJudge", ProofChoice: ProofRegression},
		{ID: "reject-out-of-scope-capability", Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundary.IndependentJudge", ProofChoice: ProofRegression},
		{ID: "replay-independent-judge", Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundary.IndependentJudge", ProofChoice: ProofCoherence},
		{ID: "preserve-read-only-effect-ceiling", Producer: "metacircularboundary.Evaluate", Consumer: "meta-circular-boundary-ci", ProofChoice: ProofRegression},
	}
}

func CaseInput(id, sourceDigest string) (Attempt, bool) {
	base := Attempt{DescriptionDigest: sourceDigest, RequestExecution: true}
	switch id {
	case "description-only":
		return base, true
	case "explicit-read-only-capability":
		base.Capability = &Capability{Issuer: "external-authority", SubjectDigest: sourceDigest, Operation: MetaOperationID, Scope: ScopeReadOnly, Handle: capabilityHandle(sourceDigest)}
		return base, true
	case "forged-capability":
		base.Capability = &Capability{Issuer: "external-authority", SubjectDigest: sourceDigest, Operation: MetaOperationID, Scope: ScopeReadOnly, Handle: digestBytes([]byte("forged|" + sourceDigest))}
		return base, true
	case "write-capability-out-of-scope":
		base.Capability = &Capability{Issuer: "external-authority", SubjectDigest: sourceDigest, Operation: MetaOperationID, Scope: ScopeWrite, Handle: capabilityHandle(sourceDigest)}
		return base, true
	default:
		return Attempt{}, false
	}
}
