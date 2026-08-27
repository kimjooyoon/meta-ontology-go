package languageutility

type CellObservation struct {
	UseCaseID     string `json:"use_case_id"`
	StageID       string `json:"stage_id"`
	State         string `json:"state"`
	Producer      string `json:"producer"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	EvidenceKey   string `json:"evidence_key,omitempty"`
	EvidencePath  string `json:"evidence_path,omitempty"`
	EvidenceDigest string `json:"evidence_digest,omitempty"`
}

type Observation struct {
	Schema           string            `json:"schema"`
	ContractID       string            `json:"contract_id"`
	SubjectSHA       string            `json:"subject_sha"`
	RepositoryWrites int               `json:"repository_writes"`
	Cells            []CellObservation `json:"cells"`
}

const (
	StateClosed  = "CLOSED"
	StateOpen    = "OPEN"
	StateUnknown = "UNKNOWN"
	StateRefuted = "REFUTED"
)
