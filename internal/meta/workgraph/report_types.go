package workgraph

type Cell struct {
	ID             string `json:"id"`
	State          string `json:"state"`
	Resolution     string `json:"resolution"`
	Activity       string `json:"activity"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceKey    string `json:"evidence_key"`
	EvidenceDigest string `json:"evidence_digest,omitempty"`
	ProofChoice    string `json:"proof_choice"`
}

type ClaimSnapshot struct {
	Status        string `json:"status"`
	State         string `json:"state"`
	Resolution    string `json:"resolution"`
	Stage         string `json:"stage,omitempty"`
	Step          string `json:"step,omitempty"`
	Reason        string `json:"reason"`
	NextOperation string `json:"next_operation"`
}

type ClaimLifecycle struct {
	ID                string        `json:"id"`
	Entity            string        `json:"entity"`
	Before            ClaimSnapshot `json:"before"`
	After             ClaimSnapshot `json:"after"`
	TraceRetained     bool          `json:"trace_retained"`
	PredecessorDigest string        `json:"predecessor_digest,omitempty"`
}

type Summary struct {
	TotalGates       int `json:"total_gates"`
	ClosedGates      int `json:"closed_gates"`
	UnknownGates     int `json:"unknown_gates"`
	RefutedGates     int `json:"refuted_gates"`
	ActiveClaims     int `json:"active_claims"`
	DischargedClaims int `json:"discharged_claims"`
	RepositoryWrites int `json:"repository_writes"`
}
