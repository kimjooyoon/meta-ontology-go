package metacircularboundary

func DenominatorContract() Denominator {
	cases := contractCases()
	denominator := Denominator{ID: DenominatorID, Cases: cases}
	denominator.Digest = digestValue(denominator)
	return denominator
}

func MetaOperations() []MetaOperation {
	return []MetaOperation{
		{ID: "bind-self-description", Producer: "metacircularboundary.observeSource", Consumer: "metacircularboundaryconsumer.Judge", ProofChoice: ProofFoundation},
		{ID: "deny-description-authority-escalation", Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge", ProofChoice: ProofRegression},
		{ID: "accept-explicit-read-only-capability", Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge", ProofChoice: ProofCoherence},
		{ID: "reject-forged-capability", Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge", ProofChoice: ProofRegression},
		{ID: "reject-out-of-scope-capability", Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge", ProofChoice: ProofRegression},
		{ID: "replay-independent-judge", Producer: "metacircularboundary.Evaluate", Consumer: "metacircularboundaryconsumer.Judge", ProofChoice: ProofCoherence},
		{ID: "preserve-read-only-effect-ceiling", Producer: "metacircularboundary.Evaluate", Consumer: "meta-circular-boundary-ci", ProofChoice: ProofRegression},
	}
}
