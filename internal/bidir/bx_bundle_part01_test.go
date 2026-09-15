package bidir

import (
	_ "embed"
)

//go:embed testdata/bx_billing_evidence_bundle.json
var bxBillingEvidenceBundle []byte

type bxBundleArtifact struct {
	Hash  string `json:"hash"`
	Count int    `json:"count"`
}
type bxBundle struct {
	SchemaVersion string                         `json:"schema_version"`
	BundleKind    string                         `json:"bundle_kind"`
	Status        string                         `json:"status"`
	FeatureGreen  bool                           `json:"feature_green"`
	Base          map[string]bxBundleArtifact    `json:"base"`
	RoundTrip     bxBundleRoundTrip              `json:"round_trip"`
	Source        bxBundleSource                 `json:"source"`
	Deltas        map[string]bxBundleDelta       `json:"deltas"`
	Transactions  map[string]bxBundleTransaction `json:"transactions"`
	Policy        bxBundlePolicy                 `json:"policy"`
}
type bxBundleSource struct {
	IntegrationBase string `json:"integration_base"`
	ParentHead      string `json:"parent_head"`
	Fixture         string `json:"fixture"`
}
type bxBundleRoundTrip struct {
	GetPut             bool `json:"get_put"`
	PutGet             bool `json:"put_get"`
	SemanticEquivalent bool `json:"semantic_equivalence"`
}
type bxBundleDelta struct {
	SequenceHash       string           `json:"sequence_hash"`
	OrderHash          string           `json:"order_hash"`
	PortOrderHash      string           `json:"port_order_hash"`
	RelationOrderHash  string           `json:"relation_order_hash"`
	PortSequence       []string         `json:"port_sequence"`
	RelationSequence   []string         `json:"relation_sequence"`
	Candidates         []string         `json:"candidates"`
	Closure            bxBundleClosure  `json:"closure"`
	Evidence           bxBundleEvidence `json:"evidence"`
	PartialObservation bool             `json:"partial_observation"`
}
type bxBundleClosure struct {
	Touched  []string `json:"touched"`
	Affected []string `json:"affected"`
	Members  []string `json:"members"`
	Hash     string   `json:"hash"`
}
type bxBundleEvidence struct {
	IDs                 []string                 `json:"ids"`
	FactKeys            []string                 `json:"fact_keys"`
	Spans               []string                 `json:"spans"`
	Records             []bxBundleEvidenceRecord `json:"records"`
	IDCount             int                      `json:"id_count"`
	SpanCount           int                      `json:"span_count"`
	Hash                string                   `json:"hash"`
	EvidenceIDAuthority string                   `json:"evidence_id_authority"`
}
type bxBundleEvidenceRecord struct {
	EvidenceID string `json:"evidence_id"`
	FactKey    string `json:"fact_key"`
	Span       string `json:"span"`
	HasSpan    bool   `json:"has_span"`
}
