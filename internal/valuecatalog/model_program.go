package valuecatalog

type ProgramResult struct {
	Activity            string       `json:"activity"`
	Program             string       `json:"program"`
	CompileReason       string       `json:"compile_reason"`
	SemanticFingerprint string       `json:"semantic_fingerprint"`
	Cases               []CaseResult `json:"cases"`
	Passed              int          `json:"passed"`
}

type CaseResult struct {
	Input    int64  `json:"input"`
	Expected int64  `json:"expected"`
	Actual   int64  `json:"actual"`
	Passed   bool   `json:"passed"`
	Reason   string `json:"reason"`
}

type Coordinate struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type Improvement struct {
	ID             string     `json:"id"`
	Before         Coordinate `json:"before"`
	After          Coordinate `json:"after"`
	BeforeEvidence string     `json:"before_evidence"`
	AfterEvidence  string     `json:"after_evidence"`
}
