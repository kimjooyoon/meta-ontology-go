package selectiveci

type pathWire struct {
	Version  string         `json:"version"`
	Edges    []edgeWire     `json:"edges"`
	Claims   []claimWire    `json:"claims"`
	Evidence []evidenceWire `json:"evidence"`
}
type edgeWire struct {
	Record            recordWire `json:"record"`
	Kind              string     `json:"kind"`
	SourceRoots       []string   `json:"source_roots"`
	AcceptanceReceipt string     `json:"acceptance_receipt"`
}
type claimWire struct {
	Record         recordWire `json:"record"`
	Kind           string     `json:"kind"`
	CanonicalDelta string     `json:"canonical_delta"`
	DeltaDigest    string     `json:"delta_digest"`
}
type recordWire struct {
	RecordID  string            `json:"record_id"`
	SubjectID string            `json:"subject_id"`
	ObjectID  string            `json:"object_id"`
	Rule      ruleWire          `json:"rule"`
	Phase     phaseWire         `json:"phase"`
	Before    snapshotWire      `json:"before"`
	After     snapshotWire      `json:"after"`
	Authority authorityWire     `json:"authority"`
	Evidence  []evidenceRefWire `json:"evidence"`
	Controls  controlsWire      `json:"controls"`
}
type ruleWire struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}
type phaseWire struct {
	Phase   string `json:"phase"`
	Ordinal uint64 `json:"ordinal"`
}
type snapshotWire struct {
	Source   string `json:"source"`
	Semantic string `json:"semantic"`
}
type authorityWire struct {
	Layer  string `json:"layer"`
	Effect string `json:"effect"`
}
type evidenceRefWire struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}
type controlsWire struct {
	CatalogDigest string      `json:"catalog_digest"`
	PolicyDigest  string      `json:"policy_digest"`
	Profile       profileWire `json:"profile"`
}
type profileWire struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}
type evidenceWire struct {
	ID           string       `json:"id"`
	Digest       string       `json:"digest"`
	Before       snapshotWire `json:"before"`
	After        snapshotWire `json:"after"`
	SourceBacked bool         `json:"source_backed"`
	Independent  bool         `json:"independent"`
	Controls     controlsWire `json:"controls"`
}
