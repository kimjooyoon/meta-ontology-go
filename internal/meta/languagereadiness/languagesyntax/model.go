package languagesyntax

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"

const (
	RegistrySchema  = "gooo/language-syntax-roundtrip-corpus/v1"
	ReportSchema    = "gooo/language-syntax-roundtrip/v1"
	DecisionPass    = "PASS"
	DecisionClosed  = "FAIL_CLOSED"
	ResolutionExact = "EXACT"
	ResolutionLower = "LOWER_RESOLUTION"
	KindValid       = "VALID"
	KindInvalid     = "INVALID"
	totalCases      = 22
	validCases      = 19
	invalidCases    = 3
	invalidDigest   = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
)

type Registry struct {
	Schema string           `json:"schema"`
	Cases  []CaseDefinition `json:"cases"`
}

type CaseDefinition struct {
	ID                 string `json:"id"`
	Path               string `json:"path"`
	Kind               string `json:"kind"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedDiagnostic string `json:"expected_diagnostic,omitempty"`
	ProofChoice        string `json:"proof_choice"`
	MetaOperation      string `json:"meta_operation"`
}

type Source struct {
	ExpectedHeadSHA         string                   `json:"expected_head_sha"`
	ConceptArtifactDigest   string                   `json:"concept_artifact_digest"`
	CatalogDigest           string                   `json:"catalog_digest"`
	RegistryDigest          string                   `json:"registry_digest"`
	CorpusDigest            string                   `json:"corpus_digest"`
	ObservationKnown        bool                     `json:"observation_known"`
	ConceptBound            bool                     `json:"concept_bound"`
	GoooFiles               []replay.FileObservation `json:"gooo_files"`
	UnregisteredGooo        []string                 `json:"unregistered_gooo"`
	MissingRegistered       []string                 `json:"missing_registered"`
	ConceptRepositoryWrites int                      `json:"concept_repository_writes"`
}

type CaseResult struct {
	Definition     CaseDefinition `json:"definition"`
	Evidence       replay.Result  `json:"evidence"`
	Status         string         `json:"status"`
	EvidenceDigest string         `json:"evidence_digest"`
}
