package languagesourcebindingpromotion

func CanonicalContract() Contract {
	return Contract{Schema: ContractSchema, Scope: Scope,
		Claims: []ClaimDefinition{
			{ID: "structural-source-execution", ProofChoice: "FOUNDATION", MetaOperation: "bind-structural-execution-receipt"},
			{ID: "independent-source-binding", ProofChoice: "COHERENCE", MetaOperation: "bind-independent-artifact-oracle"},
			{ID: "source-binding-promotion", ProofChoice: "REGRESSION", MetaOperation: "promote-source-binding-claim"},
		},
		Edges: []EdgeDefinition{{From: "structural-source-execution", To: "source-binding-promotion"},
			{From: "independent-source-binding", To: "source-binding-promotion"}},
		Cases: []CaseDefinition{
			{ID: "exact-promotion", ExpectedDecision: DecisionPass, ExpectedResolution: ResolutionExact,
				ExpectedReason: "SOURCE_BINDING_CLAIM_DISCHARGED", ExpectedPromotionStatus: "DISCHARGED"},
			{ID: "missing-oracle", ExpectedDecision: DecisionClosed, ExpectedResolution: ResolutionLower,
				ExpectedReason: "ARTIFACT_ORACLE_EVIDENCE_MISSING", ExpectedPromotionStatus: "OPEN"},
			{ID: "unknown-oracle", ExpectedDecision: DecisionClosed, ExpectedResolution: ResolutionLower,
				ExpectedReason: "ARTIFACT_ORACLE_DECISION_UNKNOWN", ExpectedPromotionStatus: "OPEN"},
			{ID: "link-mismatch", ExpectedDecision: DecisionClosed, ExpectedResolution: ResolutionInvariant,
				ExpectedReason: "SOURCE_BINDING_EVIDENCE_LINK_MISMATCH", ExpectedPromotionStatus: "REFUTED"},
			{ID: "unknown-producer", ExpectedDecision: DecisionClosed, ExpectedResolution: ResolutionLower,
				ExpectedReason: "SOURCE_EXECUTION_DECISION_UNKNOWN", ExpectedPromotionStatus: "OPEN"},
		}}
}
