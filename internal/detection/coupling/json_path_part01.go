package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type pathWire struct {
	Version  string         `json:"version"`
	Edges    []edgeWire     `json:"edges"`
	Claims   []claimWire    `json:"claims"`
	Evidence []evidenceWire `json:"evidence"`
}
type snapshotWire struct {
	Source   string `json:"source"`
	Semantic string `json:"semantic"`
}
type profileWire struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}
type controlsWire struct {
	CatalogDigest string      `json:"catalog_digest"`
	PolicyDigest  string      `json:"policy_digest"`
	Profile       profileWire `json:"profile"`
}
type ruleWire struct {
	ID      semantic.ID `json:"id"`
	Version string      `json:"version"`
	Digest  string      `json:"digest"`
}
type phaseWire struct {
	Phase   semantic.InferencePhase `json:"phase"`
	Ordinal uint64                  `json:"ordinal"`
}
type authorityWire struct {
	Layer  semantic.AuthorityLayer  `json:"layer"`
	Effect semantic.AuthorityEffect `json:"effect"`
}
type evidenceRefWire struct {
	ID     semantic.ID `json:"id"`
	Digest string      `json:"digest"`
}
type recordWire struct {
	RecordID  semantic.ID       `json:"record_id"`
	SubjectID semantic.ID       `json:"subject_id"`
	ObjectID  semantic.ID       `json:"object_id"`
	Rule      ruleWire          `json:"rule"`
	Phase     phaseWire         `json:"phase"`
	Before    snapshotWire      `json:"before"`
	After     snapshotWire      `json:"after"`
	Authority authorityWire     `json:"authority"`
	Evidence  []evidenceRefWire `json:"evidence"`
	Controls  controlsWire      `json:"controls"`
}
type edgeWire struct {
	Record            recordWire             `json:"record"`
	Kind              semantic.InferenceKind `json:"kind"`
	SourceRoots       []semantic.ID          `json:"source_roots"`
	AcceptanceReceipt semantic.ID            `json:"acceptance_receipt"`
}
type claimWire struct {
	Record         recordWire                  `json:"record"`
	Kind           semantic.SemanticChangeKind `json:"kind"`
	CanonicalDelta string                      `json:"canonical_delta"`
	DeltaDigest    string                      `json:"delta_digest"`
}
