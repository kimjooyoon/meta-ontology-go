package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type ObligationBinding struct {
	ID         string   `json:"id"`
	Subject    string   `json:"subject"`
	CommandIDs []string `json:"command_ids"`
}

func (b ObligationBinding) SemanticBinding() semanticbinding.Obligation {
	return semanticbinding.Obligation{ID: b.ID, Subject: b.Subject}
}

type Registry struct {
	SchemaVersion       string              `json:"schema_version"`
	Digest              string              `json:"registry_digest"`
	PolicyDigest        string              `json:"policy_digest"`
	Nodes               []impactgraph.Node  `json:"nodes"`
	DependencyEdges     []DependencyEdge    `json:"dependency_edges"`
	Obligations         []ObligationBinding `json:"obligations"`
	Commands            []Command           `json:"commands"`
	GlobalGuardCommands []Command           `json:"global_guard_commands"`
}
type Receipt struct {
	CommandID      string                    `json:"command_id"`
	SnapshotDigest string                    `json:"snapshot_digest"`
	Envelope       resourceenvelope.Envelope `json:"envelope"`
}
type PathRequirement struct {
	PathID        string   `json:"path_id"`
	RecordIDs     []string `json:"record_ids"`
	ExpectedKinds []string `json:"expected_kinds"`
	StartID       string   `json:"start_id"`
	EndID         string   `json:"end_id"`
}
type ProvenancePath struct {
	CommandID   string                   `json:"command_id"`
	Path        semantic.InferencePathV1 `json:"path"`
	Requirement PathRequirement          `json:"requirement"`
}
type Input struct {
	SchemaVersion   string           `json:"schema_version"`
	Base            SnapshotManifest `json:"base"`
	Head            SnapshotManifest `json:"head"`
	Registry        Registry         `json:"registry"`
	CPUCapacity     uint64           `json:"cpu_capacity"`
	Receipts        []Receipt        `json:"resource_receipts"`
	ProvenancePaths []ProvenancePath `json:"provenance_paths"`
}
type PlanResult struct {
	SchemaVersion           string   `json:"schema_version"`
	Status                  Status   `json:"status"`
	ReasonCode              string   `json:"reason_code"`
	BaseSnapshotDigest      string   `json:"base_snapshot_digest"`
	HeadSnapshotDigest      string   `json:"head_snapshot_digest"`
	ChangedSemanticIDs      []string `json:"changed_semantic_ids"`
	SelectedCommandIDs      []string `json:"selected_command_ids"`
	SelectedGuardCommandIDs []string `json:"selected_guard_command_ids"`
	SelectedWorkIDs         []string `json:"selected_work_ids"`
	ResourceReceiptDigests  []string `json:"resource_receipt_digests"`
	ProvenancePathIDs       []string `json:"provenance_path_ids"`
	CanonicalDigest         string   `json:"canonical_digest"`
	Digest                  string   `json:"-"`
}
type Output = PlanResult
