package coupling

type wireSnapshot struct {
	Source   string `json:"source"`
	Semantic string `json:"semantic"`
}
type wireProfile struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}
type wireConfig struct {
	ToolchainDigest string                `json:"toolchain_digest"`
	Profile         wireProfile           `json:"profile"`
	ResourceBinding ResourceBindingConfig `json:"resource_binding"`
}
type wireRule struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}
type wireControls struct {
	CatalogDigest string      `json:"catalog_digest,omitempty"`
	PolicyDigest  string      `json:"policy_digest,omitempty"`
	Profile       wireProfile `json:"profile"`
}
type wireEvidenceRef struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}
type wireEvidence struct {
	ID           string       `json:"id"`
	Digest       string       `json:"digest"`
	Before       wireSnapshot `json:"before"`
	After        wireSnapshot `json:"after"`
	SourceBacked bool         `json:"source_backed"`
	Independent  bool         `json:"independent"`
	Controls     wireControls `json:"controls"`
}
type wireRecord struct {
	RecordID  string            `json:"record_id"`
	SubjectID string            `json:"subject_id"`
	ObjectID  string            `json:"object_id"`
	Rule      wireRule          `json:"rule"`
	Phase     string            `json:"phase"`
	Ordinal   uint64            `json:"ordinal"`
	Before    wireSnapshot      `json:"before"`
	After     wireSnapshot      `json:"after"`
	Authority wireAuthority     `json:"authority"`
	Evidence  []wireEvidenceRef `json:"evidence"`
	Controls  wireControls      `json:"controls"`
}
type wireAuthority struct {
	Layer  string `json:"layer"`
	Effect string `json:"effect"`
}
type wireEdge struct {
	RecordID          string            `json:"record_id"`
	SubjectID         string            `json:"subject_id"`
	ObjectID          string            `json:"object_id"`
	Rule              wireRule          `json:"rule"`
	Phase             string            `json:"phase"`
	Ordinal           uint64            `json:"ordinal"`
	Before            wireSnapshot      `json:"before"`
	After             wireSnapshot      `json:"after"`
	Authority         wireAuthority     `json:"authority"`
	Evidence          []wireEvidenceRef `json:"evidence"`
	Controls          wireControls      `json:"controls"`
	Kind              string            `json:"kind"`
	SourceRoots       []string          `json:"source_roots"`
	AcceptanceReceipt string            `json:"acceptance_receipt,omitempty"`
}
