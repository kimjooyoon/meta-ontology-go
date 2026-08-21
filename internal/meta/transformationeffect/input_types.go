package transformationeffect

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type Options struct {
	Root           string
	MetricsPath    string
	PlanPath       string
	ExecutionPath  string
	ReceiptsPath   string
	ProvenancePath string
	ExpectedSHA    string
}

type ArtifactDigests struct {
	SourceMetrics string `json:"source_metrics"`
	Plan          string `json:"plan"`
	Execution     string `json:"execution"`
	Receipts      string `json:"receipts"`
	Provenance    string `json:"provenance"`
}

type inputSet struct {
	metrics    linecaps.LineMetricsReport
	plan       generation.Plan
	execution  generation.ExecutionManifest
	receipts   generation.ReceiptReport
	provenance generation.ArtifactProvenance
	digests    ArtifactDigests
}

type Result struct {
	Ledger     Ledger
	Patch      Patch
	Receipts   generation.ReceiptReport
	Provenance generation.ArtifactProvenance
}
