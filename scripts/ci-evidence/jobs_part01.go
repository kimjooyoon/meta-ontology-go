package main

const evidenceSchema = "gooo/ci-evidence/v3"

var canonicalJobs = []string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy"}

type apiJob struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
	RunID      int64  `json:"run_id"`
}
type jobEvidence struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
	RunID      int64  `json:"run_id"`
	RunAttempt int64  `json:"run_attempt"`
}
type digests struct {
	SourceSHA256           string `json:"source_sha256"`
	IRSHA256               string `json:"ir_sha256"`
	GeneratorFixtureSHA256 string `json:"generator_fixture_sha256"`
	GeneratedOutputSHA256  string `json:"generated_output_sha256"`
	SourceMapSHA256        string `json:"source_map_sha256"`
	PolicySHA256           string `json:"policy_sha256"`
	ToolchainSHA256        string `json:"toolchain_sha256"`
	BundleSHA256           string `json:"bundle_sha256"`
}
type evidence struct {
	Schema                  string             `json:"schema"`
	Repository              string             `json:"repository"`
	Event                   string             `json:"event"`
	EventRef                string             `json:"event_ref"`
	CheckoutRef             string             `json:"checkout_ref"`
	BaseRef                 string             `json:"base_ref"`
	BaseSHA                 string             `json:"base_sha"`
	HeadSHA                 string             `json:"head_sha"`
	RunID                   int64              `json:"run_id"`
	RunAttempt              int64              `json:"run_attempt"`
	WorkflowSHA             string             `json:"workflow_sha"`
	Toolchain               string             `json:"toolchain"`
	SlotPreservation        bool               `json:"slot_preservation"`
	NoWriteOutsideGenerated bool               `json:"no_write_outside_generated"`
	Jobs                    []jobEvidence      `json:"jobs"`
	ArtifactProvenance      artifactProvenance `json:"artifact_provenance"`
	Digests                 digests            `json:"digests"`
}
type metadata struct {
	Repository  string
	Event       string
	EventRef    string
	CheckoutRef string
	BaseRef     string
	BaseSHA     string
	HeadSHA     string
	RunID       int64
	RunAttempt  int64
	WorkflowSHA string
	Toolchain   string
}
