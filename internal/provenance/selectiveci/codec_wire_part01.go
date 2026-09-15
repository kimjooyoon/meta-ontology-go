package selectiveci

type wireSnapshot struct {
	Source   string `json:"source"`
	Semantic string `json:"semantic"`
}
type wireProfile struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}
type wireEvidenceRef struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}
type wireRecord struct {
	RecordID        string            `json:"record_id"`
	SubjectID       string            `json:"subject_id"`
	ObjectID        string            `json:"object_id"`
	RuleID          string            `json:"rule_id"`
	RuleVersion     string            `json:"rule_version"`
	RuleDigest      string            `json:"rule_digest"`
	Phase           string            `json:"phase"`
	PhaseOrdinal    uint64            `json:"phase_ordinal"`
	Before          wireSnapshot      `json:"before"`
	After           wireSnapshot      `json:"after"`
	AuthorityLayer  string            `json:"authority_layer"`
	AuthorityEffect string            `json:"authority_effect"`
	Evidence        []wireEvidenceRef `json:"evidence"`
	CatalogDigest   string            `json:"catalog_digest"`
	PolicyDigest    string            `json:"policy_digest"`
	Profile         wireProfile       `json:"profile"`
}
type wireEdge struct {
	Record            wireRecord `json:"record"`
	Kind              string     `json:"kind"`
	SourceRoots       []string   `json:"source_roots"`
	AcceptanceReceipt string     `json:"acceptance_receipt"`
}
type wireClaim struct {
	Record         wireRecord `json:"record"`
	Kind           string     `json:"kind"`
	CanonicalDelta string     `json:"canonical_delta"`
	DeltaDigest    string     `json:"delta_digest"`
}
type wireInferenceEvidence struct {
	ID            string       `json:"id"`
	Digest        string       `json:"digest"`
	Before        wireSnapshot `json:"before"`
	After         wireSnapshot `json:"after"`
	SourceBacked  bool         `json:"source_backed"`
	Independent   bool         `json:"independent"`
	CatalogDigest string       `json:"catalog_digest"`
	PolicyDigest  string       `json:"policy_digest"`
	Profile       wireProfile  `json:"profile"`
}
type wireInferencePath struct {
	Version  string                  `json:"version"`
	Edges    []wireEdge              `json:"edges"`
	Claims   []wireClaim             `json:"claims"`
	Evidence []wireInferenceEvidence `json:"evidence"`
}
type wirePath struct {
	PathID        string   `json:"path_id"`
	RootID        string   `json:"root_id"`
	ObligationID  string   `json:"obligation_id"`
	CommandID     string   `json:"command_id"`
	ReceiptID     string   `json:"receipt_id"`
	RecordIDs     []string `json:"record_ids"`
	ExpectedKinds []string `json:"expected_kinds"`
}
