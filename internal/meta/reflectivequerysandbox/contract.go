package reflectivequerysandbox

func CanonicalContract() Contract {
	return Contract{
		Schema: Schema, MetricID: MetricID, GoVersion: ExpectedGoVersion,
		Denominator:   ExpectedDenominator,
		Classes:       []Bucket{{Name: "OUTCOME", Total: 4}, {Name: "DRIVER", Total: 4}, {Name: "GUARDRAIL", Total: 4}},
		Proofs:        []Bucket{{Name: "FOUNDATION", Total: 4}, {Name: "COHERENCE", Total: 4}, {Name: "REGRESSION", Total: 4}},
		ExpectedNodes: 9, ExpectedFacts: 8, ExpectedAttempts: 5,
		ExpectedSafe: 3, ExpectedDenied: 1, ExpectedUnknown: 1,
		ExpectedTransitions: ExpectedTransitionCount,
	}
}

type claimSpec struct {
	ID, Class, ProofChoice, MetaOperation, Producer, Consumer string
}

func claimSpecs() []claimSpec {
	producer := "reflective-query-sandbox.producer"
	consumer := "reflective-query-sandbox.independent-verifier"
	return []claimSpec{
		{"outcome.structure", "OUTCOME", "FOUNDATION", "query-self-structure", producer, consumer},
		{"outcome.claims", "OUTCOME", "COHERENCE", "query-self-claims", producer, consumer},
		{"outcome.metrics", "OUTCOME", "REGRESSION", "query-self-metrics", producer, consumer},
		{"outcome.mutation-denied", "OUTCOME", "FOUNDATION", "deny-mutation-request", producer, consumer},
		{"driver.semantic-snapshot", "DRIVER", "COHERENCE", "bind-semantic-snapshot", producer, consumer},
		{"driver.query-projection", "DRIVER", "REGRESSION", "project-read-only-query-view", producer, consumer},
		{"driver.query-receipts", "DRIVER", "FOUNDATION", "seal-query-receipts", producer, consumer},
		{"driver.claim-ledger", "DRIVER", "COHERENCE", "transition-claim-ledger", producer, consumer},
		{"guardrail.unknown-closed", "GUARDRAIL", "REGRESSION", "preserve-unknown-target", producer, consumer},
		{"guardrail.graph-unchanged", "GUARDRAIL", "FOUNDATION", "compare-query-graph-digest", producer, consumer},
		{"guardrail.repository-write-set", "GUARDRAIL", "COHERENCE", "observe-repository-write-set", producer, consumer},
		{"guardrail.mutation-authority", "GUARDRAIL", "REGRESSION", "bind-mutation-authority", producer, consumer},
	}
}
