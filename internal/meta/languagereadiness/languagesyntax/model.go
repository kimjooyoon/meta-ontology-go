package languagesyntax

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"

const (
	RegistrySchema             = "gooo/language-syntax-roundtrip-corpus/v2"
	ReportSchema               = "gooo/language-syntax-roundtrip/v1"
	DecisionPass               = "PASS"
	DecisionClosed             = "FAIL_CLOSED"
	ResolutionExact            = "EXACT"
	ResolutionLower            = "LOWER_RESOLUTION"
	KindValid                  = "VALID"
	KindInvalid                = "INVALID"
	ScopeLanguageCapability    = "LANGUAGE_CAPABILITY"
	ScopeGovernanceObservation = "GOVERNANCE_OBSERVATION"
	totalCases                 = 49
	validCases                 = 46
	invalidCases               = 3
	FixedTotal                 = totalCases
	FixedCapabilityTotal       = 48
	FixedGovernanceTotal       = 1
	invalidDigest              = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
)

type Registry struct {
	Schema       string              `json:"schema"`
	Cases        []CaseDefinition    `json:"cases"`
	PackageUnits []PackageDefinition `json:"package_units"`
	MetaSources  []string            `json:"meta_sources"`
}

type CaseDefinition struct {
	ID                    string `json:"id"`
	Path                  string `json:"path"`
	Kind                  string `json:"kind"`
	ExpectedDecision      string `json:"expected_decision"`
	ExpectedDiagnostic    string `json:"expected_diagnostic,omitempty"`
	ProofChoice           string `json:"proof_choice"`
	MetaOperation         string `json:"meta_operation"`
	Scope                 string `json:"scope"`
	EntityFields          bool   `json:"entity_fields,omitempty"`
	ImplicitActivityPorts bool   `json:"implicit_activity_ports,omitempty"`
}

type PackageDefinition struct {
	ID                   string   `json:"id"`
	Path                 string   `json:"path"`
	Members              []string `json:"members"`
	Entry                string   `json:"entry"`
	ReportSchema         string   `json:"report_schema"`
	MetaReducer          string   `json:"meta_reducer"`
	SourceFilesIndicator string   `json:"source_files_indicator"`
	ExecutionIndicator   string   `json:"execution_indicator"`
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
	PackageUnits            []PackageDefinition      `json:"package_units"`
	ConceptRepositoryWrites int                      `json:"concept_repository_writes"`
}

type CaseResult struct {
	Definition     CaseDefinition `json:"definition"`
	Evidence       replay.Result  `json:"evidence"`
	Status         string         `json:"status"`
	EvidenceDigest string         `json:"evidence_digest"`
}
