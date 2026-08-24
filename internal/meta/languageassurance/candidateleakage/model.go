package candidateleakage

type Candidate struct {
	SubjectSHA    string `json:"subject_sha"`
	Digest        string `json:"digest"`
	Decision      string `json:"decision"`
	Resolution    string `json:"resolution"`
	MetaOperation string `json:"meta_operation"`
}

type Promotion struct {
	SubjectSHA     string `json:"subject_sha"`
	CandidateDigest string `json:"candidate_digest"`
	EvidenceDigest  string `json:"evidence_digest"`
	Decision        string `json:"decision"`
	Resolution      string `json:"resolution"`
	MetaOperation   string `json:"meta_operation"`
}

type Official struct {
	SubjectSHA    string `json:"subject_sha"`
	Status        string `json:"status"`
	Decision      string `json:"decision"`
	Resolution    string `json:"resolution"`
	MetaOperation string `json:"meta_operation"`
}

type Input struct {
	Schema     string    `json:"schema"`
	SubjectSHA string    `json:"subject_sha"`
	Candidate  Candidate `json:"candidate"`
	Promotion  Promotion `json:"promotion"`
	Official   Official  `json:"official"`
}

type Definition struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedEffect     string `json:"expected_effect"`
	ExpectedReason     string `json:"expected_reason"`
	ExpectedLeakage    int    `json:"expected_leakage_paths"`
	ExpectedUnknown    int    `json:"expected_unknown_paths"`
}

type Denominator struct {
	ID     string       `json:"id"`
	Cases  []Definition `json:"cases"`
	Digest string       `json:"digest"`
}

type MetaOperation struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
}
