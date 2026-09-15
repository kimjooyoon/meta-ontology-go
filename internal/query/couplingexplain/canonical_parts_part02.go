package couplingexplain

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type canonicalPath struct {
	PathID     string              `json:"path_id"`
	StartID    string              `json:"start_id"`
	EndID      string              `json:"end_id"`
	StepCount  int                 `json:"step_count"`
	PathDigest string              `json:"path_digest"`
	Steps      []canonicalPathStep `json:"steps,omitempty"`
}
type canonicalPathUnsigned struct {
	PathID    string              `json:"path_id"`
	StartID   string              `json:"start_id"`
	EndID     string              `json:"end_id"`
	StepCount int                 `json:"step_count"`
	Steps     []canonicalPathStep `json:"steps,omitempty"`
}

func toCanonicalPath(value PathSummary) canonicalPath {
	steps := make([]canonicalPathStep, 0, len(value.Steps))
	for _, step := range value.Steps {
		steps = append(steps, canonicalPathStep{FromID: step.FromID, ToID: step.ToID, Kind: step.Kind,
			Phase: step.Phase.Phase, PhaseOrdinal: step.Phase.Ordinal, RuleRef: step.RuleRef,
			InputDigest: step.InputDigest, OutputDigest: step.OutputDigest, EvidenceRef: step.EvidenceRef})
	}
	return canonicalPath{PathID: value.PathID, StartID: value.StartID, EndID: value.EndID,
		StepCount: value.StepCount, PathDigest: value.PathDigest, Steps: steps}
}
func pathDigest(value PathSummary) string {
	path := toCanonicalPath(value)
	data, err := json.Marshal(canonicalPathUnsigned{PathID: path.PathID, StartID: path.StartID, EndID: path.EndID,
		StepCount: path.StepCount, Steps: path.Steps})
	if err != nil {
		return ""
	}
	return DigestBytes(data)
}

type canonicalPathStep struct {
	FromID       string                  `json:"from_id"`
	ToID         string                  `json:"to_id"`
	Kind         semantic.InferenceKind  `json:"kind"`
	Phase        semantic.InferencePhase `json:"phase"`
	PhaseOrdinal uint64                  `json:"phase_ordinal"`
	RuleRef      string                  `json:"rule_ref,omitempty"`
	InputDigest  string                  `json:"input_digest"`
	OutputDigest string                  `json:"output_digest"`
	EvidenceRef  string                  `json:"evidence_ref,omitempty"`
}
type canonicalReceipt struct {
	ReceiptID      string                      `json:"receipt_id"`
	SurfaceID      string                      `json:"surface_id"`
	ChangeClaim    ChangeClaim                 `json:"change_claim"`
	ReceiptKind    semantic.SemanticChangeKind `json:"receipt_kind"`
	BeforeIRDigest string                      `json:"before_ir_digest"`
	AfterIRDigest  string                      `json:"after_ir_digest"`
	CanonicalDelta string                      `json:"canonical_delta,omitempty"`
	DeltaDigest    string                      `json:"delta_digest,omitempty"`
	ReceiptDigest  string                      `json:"receipt_digest"`
	OriginPathID   string                      `json:"origin_path_id"`
	EvidenceRefs   []string                    `json:"evidence_refs"`
}
