package languageutility

type CellResult struct {
	UseCaseID      string `json:"use_case_id"`
	StageID        string `json:"stage_id"`
	ProofChoice    string `json:"proof_choice"`
	State          string `json:"state"`
	ClaimStatus    string `json:"claim_status"`
	Resolution     string `json:"resolution"`
	Producer       string `json:"producer"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceKey    string `json:"evidence_key,omitempty"`
	EvidencePath   string `json:"evidence_path,omitempty"`
	EvidenceDigest string `json:"evidence_digest,omitempty"`
}

type UseCaseSummary struct {
	ID             string `json:"id"`
	ClosedCells    int    `json:"closed_cells"`
	TotalCells     int    `json:"total_cells"`
	RemainingCells int    `json:"remaining_cells"`
	Complete       bool   `json:"complete"`
}

type Summary struct {
	ClosedCells               int  `json:"closed_cells"`
	OpenCells                 int  `json:"open_cells"`
	UnknownCells              int  `json:"unknown_cells"`
	RefutedCells              int  `json:"refuted_cells"`
	CellsTotal                int  `json:"cells_total"`
	RemainingCells            int  `json:"remaining_cells"`
	ProgressBasisPoints       int  `json:"progress_basis_points"`
	CompleteUseCases          int  `json:"complete_use_cases"`
	UseCasesTotal             int  `json:"use_cases_total"`
	EvidenceArtifacts         int  `json:"evidence_artifacts"`
	ObservationIssues         int  `json:"observation_issues"`
	ClaimsOpen                int  `json:"claims_open"`
	ClaimsDischarged          int  `json:"claims_discharged"`
	ClaimsRefuted             int  `json:"claims_refuted"`
	ClosedFloor               int  `json:"closed_floor"`
	ClosedDeltaFromFloor      int  `json:"closed_delta_from_floor"`
	CompleteUseCaseFloor      int  `json:"complete_use_case_floor"`
	CompleteUseCaseFloorDelta int  `json:"complete_use_case_floor_delta"`
	ObservationComplete       bool `json:"observation_complete"`
	UtilityComplete           bool `json:"utility_complete"`
	PromotionComplete         bool `json:"promotion_complete"`
	RepositoryWrites          int  `json:"repository_writes"`
}
