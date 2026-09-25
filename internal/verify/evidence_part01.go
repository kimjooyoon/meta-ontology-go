package verify

const (
	EvidenceSchemaVersion = "gooo/evidence/v1"
	EvidenceProducerGo    = "go"
	EvidenceProducerGooo  = "gooo"
)

// ConformanceStage names the staged verifier trust boundary.
type ConformanceStage uint8

const (
	StageGoBaseline ConformanceStage = iota
	StageDualEvidence
	StageGoooFallback
	StageGoooAuthoritative
)

// EvidenceFact is a canonical, stable-ID result from one verifier.
type EvidenceFact struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// EvidenceBundle is the producer-independent payload compared by CI.
type EvidenceBundle struct {
	Schema   string           `json:"schema"`
	Stage    ConformanceStage `json:"stage"`
	Fixture  string           `json:"fixture"`
	Decision string           `json:"decision"`
	Facts    []EvidenceFact   `json:"facts"`
}

// EvidenceArtifact identifies the implementation that emitted a comparable
// bundle. Producer identity is excluded from the comparison payload.
type EvidenceArtifact struct {
	Producer string         `json:"producer"`
	Bundle   EvidenceBundle `json:"bundle"`
}

// EvidenceManifest records the independently verifiable payload digest.
type EvidenceManifest struct {
	Schema        string           `json:"schema"`
	Producer      string           `json:"producer"`
	Stage         ConformanceStage `json:"stage"`
	Fixture       string           `json:"fixture"`
	Decision      string           `json:"decision"`
	PayloadSHA256 string           `json:"payload_sha256"`
}
