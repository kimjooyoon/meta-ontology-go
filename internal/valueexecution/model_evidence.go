package valueexecution

type CaseResult struct {
	ID            string `json:"id"`
	Input         int64  `json:"input"`
	Expected      int64  `json:"expected"`
	Actual        int64  `json:"actual"`
	Replay        int64  `json:"replay"`
	Passed        bool   `json:"passed"`
	ReplayMatched bool   `json:"replay_matched"`
}

type CounterexampleResult struct {
	ID             string `json:"id"`
	ExpectedReason string `json:"expected_reason"`
	ActualReason   string `json:"actual_reason"`
	ReplayReason   string `json:"replay_reason"`
	Passed         bool   `json:"passed"`
	ReplayMatched  bool   `json:"replay_matched"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type View struct {
	Audience     string   `json:"audience"`
	Resolution   string   `json:"resolution"`
	Satisfied    int      `json:"satisfied"`
	Total        int      `json:"total"`
	BasisPoints  int      `json:"basis_points"`
	IndicatorIDs []string `json:"indicator_ids"`
}

type Proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}
