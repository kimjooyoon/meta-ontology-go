package bodycodegen

// RecordAssemblyPolicy is an explicit Gooo program that controls whether a
// scored candidate is followed by another candidate in the existing ranking.
type RecordAssemblyPolicy struct {
	Source   string `json:"source"`
	Activity string `json:"activity"`
}

type RecordAssemblyControl struct {
	Schema          string                 `json:"schema"`
	Policy          RecordAssemblyPolicy   `json:"policy"`
	SourceSHA256    string                 `json:"source_sha256"`
	GeneratedSHA256 string                 `json:"generated_sha256"`
	ActivityID      string                 `json:"activity_id"`
	Decisions       []RecordPolicyDecision `json:"decisions"`
}

type RecordPolicyCounts struct {
	Matched int `json:"matched"`
	Total   int `json:"total"`
	Best    int `json:"best"`
	Scored  int `json:"scored"`
	Budget  int `json:"budget"`
}

type RecordPolicyDecision struct {
	AttemptIndex int                `json:"attempt_index"`
	Mask         uint16             `json:"mask"`
	Input        RecordPolicyCounts `json:"input"`
	State        string             `json:"state"`
	Operation    string             `json:"next_operation"`
	Message      string             `json:"message"`
	Continue     bool               `json:"continue"`
}
