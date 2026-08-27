package authorizationfoundation

type Indicator struct {
	MetricID       string `json:"metric_id"`
	Class          string `json:"class"`
	ProofChoice    string `json:"proof_choice"`
	Stage          string `json:"stage"`
	MetaOperation  string `json:"meta_operation"`
	Value          int    `json:"value"`
	Target         int    `json:"target"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest,omitempty"`
	UnknownReason  string `json:"unknown_reason,omitempty"`
}

type Claim struct {
	ClaimID        string `json:"claim_id"`
	Stage          string `json:"stage"`
	Statement      string `json:"statement"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest,omitempty"`
}

type Unknown struct {
	Stage       string `json:"stage"`
	IndicatorID string `json:"indicator_id"`
	Reason      string `json:"reason"`
}

type Proof struct {
	Mode       string `json:"mode"`
	Status     string `json:"status"`
	Completed  int    `json:"completed"`
	Total      int    `json:"total"`
	Resolution string `json:"resolution"`
}

type ReaderView struct {
	Reader      string `json:"reader"`
	Completed   int    `json:"completed"`
	Total       int    `json:"total"`
	BasisPoints int    `json:"basis_points"`
	Resolution  string `json:"resolution"`
}

type FoundationBinding struct {
	ArtifactID        int64  `json:"artifact_id"`
	ArtifactName      string `json:"artifact_name"`
	ArchiveDigest     string `json:"archive_digest"`
	ProducerRunID     int64  `json:"producer_run_id"`
	ProducerSubject   string `json:"producer_subject_sha"`
	PriorReceipt      string `json:"prior_receipt_digest"`
	EvidenceDigest    string `json:"evidence_digest"`
}
