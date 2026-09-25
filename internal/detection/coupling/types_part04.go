package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type CountDimension struct {
	Known bool   `json:"known"`
	Value uint64 `json:"value"`
}
type ObservationVector struct {
	ChangedSurfaces   CountDimension `json:"changed_surfaces"`
	Receipts          CountDimension `json:"receipts"`
	InferenceRecords  CountDimension `json:"inference_records"`
	InferencePaths    CountDimension `json:"inference_paths"`
	DeterministicWork CountDimension `json:"deterministic_work"`
	ResourceWork      CountDimension `json:"resource_work"`
	CPU               CountDimension `json:"cpu"`
	Memory            CountDimension `json:"memory"`
}
type Result struct {
	Schema             string            `json:"schema"`
	Status             Status            `json:"status"`
	AcceptedSurfaceIDs []semantic.ID     `json:"accepted_surface_ids"`
	Reasons            []Reason          `json:"reasons"`
	Observation        ObservationVector `json:"observation"`
	FullSuiteRequired  bool              `json:"full_suite_required"`
	InputDigest        string            `json:"input_digest"`
	Digest             string            `json:"digest"`
}
