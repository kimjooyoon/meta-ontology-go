package evidencequorum

func CanonicalContract() Contract {
	return Contract{
		Schema:                   ContractSchema,
		Scope:                    Scope,
		SourcePath:               "examples/evidence-quorum/main.gooo",
		SourceEntry:              "ProduceEvidence",
		FixedCaseDenominator:     5,
		MinimumIndependentGroups: 3,
		RequiredRoles:            []string{"producer", "consumer", "meta-operation"},
		Claim: ClaimDefinition{
			ID:            "evidence-source-claim",
			Statement:     "the bounded evidence activity is justified by independent evidence",
			Producer:      "gooo-source-producer",
			Consumer:      "evidence-quorum-consumer",
			MetaOperation: "justify-claim-with-independent-evidence",
			ProofChoice:   "INDEPENDENT_PROVENANCE_QUORUM",
		},
		Cases: []CaseDefinition{
			{ID: "sufficient-independent", ExpectedDecision: DecisionPass, ExpectedResolution: ResolutionExact,
				ExpectedReason: "QUORUM_CLAIM_DISCHARGED", ExpectedStatus: StatusDischarged},
			{ID: "same-origin-replica", ExpectedDecision: DecisionClosed, ExpectedResolution: ResolutionLower,
				ExpectedReason: "QUORUM_INSUFFICIENT_INDEPENDENT_GROUPS", ExpectedStatus: StatusOpen},
			{ID: "conflicting-independent", ExpectedDecision: DecisionClosed, ExpectedResolution: ResolutionInvariant,
				ExpectedReason: "QUORUM_CONFLICT", ExpectedStatus: StatusRefuted},
			{ID: "insufficient-independent", ExpectedDecision: DecisionClosed, ExpectedResolution: ResolutionLower,
				ExpectedReason: "QUORUM_INSUFFICIENT_INDEPENDENT_GROUPS", ExpectedStatus: StatusOpen},
			{ID: "unknown-producer-receipt", ExpectedDecision: DecisionUnknown, ExpectedResolution: ResolutionLower,
				ExpectedReason: "QUORUM_EVIDENCE_UNKNOWN", ExpectedStatus: StatusUnknown, ProducerDecision: DecisionUnknown},
		},
	}
}
