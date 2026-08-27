package nonmonotonicrefutation

func CanonicalContract() Contract {
	return Contract{
		Schema: ContractSchema, FixedClaimTotal: 3, FixedTransitionTotal: 6,
		Cases: []CaseDefinition{
			{
				ID: "support-only", ClaimID: "claim://support-only", InitialStatus: StatusOpen,
				ExpectedFinalStatus: StatusDischarged, Producer: ProducerID, Consumer: ConsumerID,
				MetaOperation: MetaOperation, ProofChoice: ProofFoundation,
				Evidence: []Evidence{evidence("support-only-1", "claim://support-only", EvidenceSupport,
					"independent reproduction supports the claim", ProofFoundation,
					"accept-support", "SUPPORTING_EVIDENCE_ACCEPTED")},
			},
			{
				ID: "new-counterevidence", ClaimID: "claim://new-counterevidence", InitialStatus: StatusOpen,
				ExpectedFinalStatus: StatusRefuted, Producer: ProducerID, Consumer: ConsumerID,
				MetaOperation: MetaOperation, ProofChoice: ProofCoherence,
				Evidence: []Evidence{
					evidence("new-counterevidence-1", "claim://new-counterevidence", EvidenceSupport,
						"initial reproduction supports the claim", ProofCoherence,
						"accept-support", "SUPPORTING_EVIDENCE_ACCEPTED"),
					evidence("new-counterevidence-2", "claim://new-counterevidence", EvidenceRefute,
						"new counterexample contradicts the discharged claim", ProofCoherence,
						"accept-counterexample", "NEW_EVIDENCE_REFUTES_DISCHARGED"),
				},
			},
			{
				ID: "re-evaluation", ClaimID: "claim://re-evaluation", InitialStatus: StatusOpen,
				ExpectedFinalStatus: StatusDischarged, Producer: ProducerID, Consumer: ConsumerID,
				MetaOperation: MetaOperation, ProofChoice: ProofRegression,
				Evidence: []Evidence{
					evidence("re-evaluation-1", "claim://re-evaluation", EvidenceSupport,
						"baseline reproduction supports the claim", ProofRegression,
						"accept-support", "SUPPORTING_EVIDENCE_ACCEPTED"),
					evidence("re-evaluation-2", "claim://re-evaluation", EvidenceRefute,
						"new counterexample contradicts the discharged claim", ProofRegression,
						"accept-counterexample", "NEW_EVIDENCE_REFUTES_DISCHARGED"),
					evidence("re-evaluation-3", "claim://re-evaluation", EvidenceSupport,
						"later independent support defeats the refutation", ProofRegression,
						"reconsider-with-support", "LATER_SUPPORT_REDISCHARGES_REFUTED"),
				},
			},
		},
	}
}

func evidence(id, claimID, kind, basis, proof, step, reason string) Evidence {
	stage := "EVIDENCE"
	if reason == "LATER_SUPPORT_REDISCHARGES_REFUTED" {
		stage = "REASSESS"
	}
	return Evidence{ID: id, ClaimID: claimID, Kind: kind, Basis: basis,
		Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperation,
		ProofChoice: proof, Coordinate: Coordinate{Stage: stage, Step: step, Reason: reason}}
}
