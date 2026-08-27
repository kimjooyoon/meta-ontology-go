package metacircularboundaryconsumer

import contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"

// expectedCases fixes only the four case identities and their proof routes.
// Observed attempt values are read from the Gooo computes programs.
func expectedCases() []contract.CaseDefinition {
	return []contract.CaseDefinition{
		{ID: "description-only", ProofChoice: proofRegression, MetaOperation: "deny-description-authority-escalation"},
		{ID: "explicit-read-only-capability", ProofChoice: proofCoherence, MetaOperation: "accept-explicit-read-only-capability"},
		{ID: "forged-capability", ProofChoice: proofRegression, MetaOperation: "reject-forged-capability"},
		{ID: "write-capability-out-of-scope", ProofChoice: proofRegression, MetaOperation: "reject-out-of-scope-capability"},
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
