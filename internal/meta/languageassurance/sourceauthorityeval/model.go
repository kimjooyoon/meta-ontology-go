package sourceauthorityeval

type Bundle struct {
	Schema         string      `json:"schema"`
	SubjectSHA     string      `json:"subject_sha"`
	ContractDigest string      `json:"contract_digest"`
	Sources        []Source    `json:"sources"`
	Authorities    []Authority `json:"authorities"`
	Facts          []Fact      `json:"facts"`
}

type Source struct {
	ID             string `json:"id"`
	URI            string `json:"uri"`
	SnapshotDigest string `json:"snapshot_digest"`
	Bytes          []byte `json:"bytes"`
}

type Authority struct {
	ID             string `json:"id"`
	SourceRef      string `json:"source_ref"`
	SnapshotDigest string `json:"snapshot_digest"`
	Start          int    `json:"start"`
	End            int    `json:"end"`
}

type Fact struct {
	ID                   string `json:"id"`
	State                string `json:"state"`
	Claim                []byte `json:"claim"`
	ClaimDigest          string `json:"claim_digest"`
	SourceRef            string `json:"source_ref"`
	SourceSnapshotDigest string `json:"source_snapshot_digest"`
	Span                 Span   `json:"span"`
	AuthorityRef         string `json:"authority_ref"`
}

type Span struct {
	Start  int    `json:"start"`
	End    int    `json:"end"`
	Digest string `json:"digest"`
}

type Report struct {
	Schema                 string        `json:"schema"`
	MetricID               string        `json:"metric_id"`
	MetaOperation          string        `json:"meta_operation"`
	ProofChoice            string        `json:"proof_choice"`
	SubjectSHA             string        `json:"subject_sha"`
	ContractDigest         string        `json:"contract_digest"`
	EvidenceContractDigest string        `json:"evidence_contract_digest"`
	Observation            string        `json:"observation"`
	Resolution             string        `json:"resolution"`
	Enforcement            string        `json:"enforcement"`
	Reason                 string        `json:"reason"`
	Summary                Summary       `json:"summary"`
	Receipts               []FactReceipt `json:"receipts"`
	ReceiptDigest          string        `json:"receipt_digest"`
}
