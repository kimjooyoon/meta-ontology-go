package metainvocation

type VerificationReceipt struct {
	Schema          string         `json:"schema"`
	SubjectDigest   string         `json:"subject_digest"`
	Decision        string         `json:"decision"`
	Resolution      string         `json:"resolution"`
	EvidenceDigests []string       `json:"evidence_digests"`
	Unknowns        []UnknownCause `json:"unknowns"`
	Digest          string         `json:"digest"`
}

type Report struct {
	Schema       string              `json:"schema"`
	Decision     string              `json:"decision"`
	Resolution   string              `json:"resolution"`
	CaseID       string              `json:"case_id"`
	Entry        string              `json:"entry"`
	SourceDigest string              `json:"source_digest"`
	InputDigest  string              `json:"input_digest"`
	Plan         CheckPlan           `json:"plan"`
	Receipt      VerificationReceipt `json:"receipt"`
	Unknowns     []UnknownCause      `json:"unknowns"`
	Claims       []Claim             `json:"claims"`
	Effects      Effects             `json:"effects"`
	NotClaimed   []string            `json:"not_claimed"`
	ReportDigest string              `json:"report_digest"`
}
