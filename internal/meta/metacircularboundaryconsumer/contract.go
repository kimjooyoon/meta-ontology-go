package metacircularboundaryconsumer

import contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"

func expectedCases(sourceDigest string) []contract.CaseDefinition {
	return []contract.CaseDefinition{
		{ID: "description-only", ExpectedDecision: decisionPass, ExpectedAuthorization: authorizationDeny, ExpectedExecution: executionBlocked, ExpectedReason: reasonDescription, ProofChoice: proofRegression, MetaOperation: "deny-description-authority-escalation"},
		{ID: "explicit-read-only-capability", ExpectedDecision: decisionPass, ExpectedAuthorization: authorizationGrant, ExpectedExecution: executionAllowed, ExpectedReason: reasonExplicit, ProofChoice: proofCoherence, MetaOperation: "accept-explicit-read-only-capability"},
		{ID: "forged-capability", ExpectedDecision: decisionPass, ExpectedAuthorization: authorizationDeny, ExpectedExecution: executionBlocked, ExpectedReason: reasonForged, ProofChoice: proofRegression, MetaOperation: "reject-forged-capability"},
		{ID: "write-capability-out-of-scope", ExpectedDecision: decisionPass, ExpectedAuthorization: authorizationDeny, ExpectedExecution: executionBlocked, ExpectedReason: reasonOutOfScope, ProofChoice: proofRegression, MetaOperation: "reject-out-of-scope-capability"},
	}
}

func expectedAttempt(id, sourceDigest string) (contract.Attempt, bool) {
	attempt := contract.Attempt{DescriptionDigest: sourceDigest, RequestExecution: true}
	switch id {
	case "description-only":
		return attempt, true
	case "explicit-read-only-capability":
		attempt.Capability = &contract.Capability{Issuer: "external-authority", SubjectDigest: sourceDigest, Operation: metaOperationID, Scope: scopeReadOnly, Handle: capabilityHandle(sourceDigest)}
		return attempt, true
	case "forged-capability":
		attempt.Capability = &contract.Capability{Issuer: "external-authority", SubjectDigest: sourceDigest, Operation: metaOperationID, Scope: scopeReadOnly, Handle: digestBytes([]byte("forged|" + sourceDigest))}
		return attempt, true
	case "write-capability-out-of-scope":
		attempt.Capability = &contract.Capability{Issuer: "external-authority", SubjectDigest: sourceDigest, Operation: metaOperationID, Scope: scopeWrite, Handle: capabilityHandle(sourceDigest)}
		return attempt, true
	default:
		return contract.Attempt{}, false
	}
}

func expectedMetaOperations() []contract.MetaOperation {
	consumer := "metacircularboundaryconsumer.Judge"
	return []contract.MetaOperation{
		{ID: "bind-self-description", Producer: "metacircularboundary.observeSource", Consumer: consumer, ProofChoice: proofFoundation},
		{ID: "deny-description-authority-escalation", Producer: "metacircularboundary.Evaluate", Consumer: consumer, ProofChoice: proofRegression},
		{ID: "accept-explicit-read-only-capability", Producer: "metacircularboundary.Evaluate", Consumer: consumer, ProofChoice: proofCoherence},
		{ID: "reject-forged-capability", Producer: "metacircularboundary.Evaluate", Consumer: consumer, ProofChoice: proofRegression},
		{ID: "reject-out-of-scope-capability", Producer: "metacircularboundary.Evaluate", Consumer: consumer, ProofChoice: proofRegression},
		{ID: "replay-independent-judge", Producer: "metacircularboundary.Evaluate", Consumer: consumer, ProofChoice: proofCoherence},
		{ID: "preserve-read-only-effect-ceiling", Producer: "metacircularboundary.Evaluate", Consumer: "meta-circular-boundary-ci", ProofChoice: proofRegression},
	}
}
