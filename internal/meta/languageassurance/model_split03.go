package languageassurance

type Summary struct {
	DenominatorTotal          int  `json:"denominator_total"`
	Operating                 int  `json:"operating"`
	NotImplemented            int  `json:"not_implemented"`
	ImplementationCoverageBPS int  `json:"implementation_coverage_bps"`
	EvidenceGroupsObserved    int  `json:"evidence_groups_observed"`
	EvidenceGroupsTotal       int  `json:"evidence_groups_total"`
	EvidenceCoverageBPS       int  `json:"evidence_coverage_bps"`
	SelfMintingPaths          *int `json:"self_minting_paths"`
	RoleConflictPaths         *int `json:"role_conflict_paths"`
	UnknownLaunderingPaths    *int `json:"unknown_laundering_paths"`
	UnknownTopDecisions       *int `json:"unknown_top_decisions"`
	UnresolvedIndicators      int  `json:"unresolved_indicators"`
	ViolatedGuardrails        int  `json:"violated_guardrails"`
	RepositoryWrites          int  `json:"repository_writes"`
}

type Report struct {
	Schema                   string                  `json:"schema"`
	SubjectSHA               string                  `json:"subject_sha"`
	TransactionDigest        string                  `json:"transaction_digest"`
	DenominatorID            string                  `json:"denominator_id"`
	DenominatorDigest        string                  `json:"denominator_digest"`
	AssuranceDecision        string                  `json:"assurance_decision"`
	CandidateDecision        string                  `json:"candidate_decision"`
	CandidateReason          string                  `json:"candidate_reason"`
	CandidateResolution      Resolution              `json:"candidate_resolution"`
	Denominator              []ObligationDefinition  `json:"denominator"`
	Obligations              []ObligationObservation `json:"obligations"`
	MetaOperations           []MetaOperation         `json:"meta_operations"`
	RoleConflictPairs        []RolePair              `json:"role_conflict_pairs"`
	UnknownLaunderingOutputs []Decision              `json:"unknown_laundering_outputs"`
	Transaction              Transaction             `json:"transaction"`
	Findings                 []Finding               `json:"findings"`
	Indicators               []Indicator             `json:"indicators"`
	Summary                  Summary                 `json:"summary"`
	ReportDigest             string                  `json:"report_digest"`
}
