package main

const (
	manifestSchema   = "gooo/language-concept-manifest/v1"
	projectionSchema = "gooo/conflict-free-registry-projection/v1"
	digestSchema     = "gooo/conflict-free-registry-projection-digest/v1"
	evidenceSchema   = "gooo/conflict-free-registry-projection-evidence/v1"
	defaultOutput    = "internal/meta/registryprojection/generated"
)

type UseCase struct {
	ID              string `json:"id"`
	Trigger         string `json:"trigger"`
	ExpectedOutcome string `json:"expected_outcome"`
}

type Concept struct {
	Problem        string `json:"problem"`
	PositiveEffect string `json:"positive_effect"`
	MetaOperation  string `json:"meta_operation"`
	Rarity         string `json:"rarity"`
	Stage          string `json:"stage"`
	NoveltyClaim   bool   `json:"novelty_claim"`
}

type ResourceRef struct {
	Path string `json:"path"`
	Role string `json:"role"`
}

type Denominator struct {
	ID     string         `json:"id"`
	Values map[string]int `json:"values"`
}

type Manifest struct {
	Schema                 string        `json:"schema"`
	StableID               string        `json:"stable_id"`
	Concept                Concept       `json:"concept"`
	CodeBindings           []string      `json:"code_bindings"`
	MetricBindings         []string      `json:"metric_bindings"`
	UseCases               []UseCase     `json:"use_cases"`
	VerificationStrategies []string      `json:"verification_strategies"`
	Corpus                 []ResourceRef `json:"corpus"`
	Registry               []ResourceRef `json:"registry"`
	Denominators           []Denominator `json:"denominators"`
	Documentation          []ResourceRef `json:"documentation"`
	Comments               []string      `json:"comments"`
}

type LoadedManifest struct {
	Manifest   Manifest
	SourcePath string
	RawDigest  string
}

type ManifestDigest struct {
	StableID       string `json:"stable_id"`
	SourceManifest string `json:"source_manifest"`
	RawDigest      string `json:"raw_digest"`
	SemanticDigest string `json:"semantic_digest"`
}

type CatalogEntry struct {
	StableID               string    `json:"stable_id"`
	SourceManifest         string    `json:"source_manifest"`
	Problem                string    `json:"problem"`
	PositiveEffect         string    `json:"positive_effect"`
	MetaOperation          string    `json:"meta_operation"`
	Rarity                 string    `json:"rarity"`
	Stage                  string    `json:"stage"`
	NoveltyClaim           bool      `json:"novelty_claim"`
	CodeBindings           []string  `json:"code_bindings"`
	MetricBindings         []string  `json:"metric_bindings"`
	UseCases               []UseCase `json:"use_cases"`
	VerificationStrategies []string  `json:"verification_strategies"`
}

type ResourceSnapshot struct {
	StableID string `json:"stable_id"`
	Path     string `json:"path"`
	Role     string `json:"role"`
	Bytes    int    `json:"bytes"`
	Digest   string `json:"digest"`
}

type DenominatorEntry struct {
	StableID       string         `json:"stable_id"`
	ID             string         `json:"id"`
	SemanticDigest string         `json:"semantic_digest"`
	Values         map[string]int `json:"values"`
}

type Projection struct {
	Schema        string             `json:"schema"`
	Catalog       []CatalogEntry     `json:"catalog"`
	Corpus        []ResourceSnapshot `json:"corpus"`
	Registry      []ResourceSnapshot `json:"registry"`
	Denominator   []DenominatorEntry `json:"denominator"`
	Documentation []ResourceSnapshot `json:"documentation"`
}

type RatioMetric struct {
	Numerator   int `json:"numerator"`
	Denominator int `json:"denominator"`
	BasisPoints int `json:"basis_points"`
}

type MetricDelta struct {
	Before RatioMetric `json:"before"`
	After  RatioMetric `json:"after"`
}

type IntegrationMetrics struct {
	SharedRegistrationTouchpoints RatioMetric `json:"shared_registration_touchpoints"`
	ConceptLocalTouchpoints       RatioMetric `json:"concept_local_touchpoints"`
	ManualGlobalEditsRequired     RatioMetric `json:"manual_global_edits_required"`
	ProjectionConflictSurfaceBPS  RatioMetric `json:"projection_conflict_surface_bps"`
}

type DigestFile struct {
	Schema                 string           `json:"schema"`
	RawManifestDigest      string           `json:"raw_manifest_digest"`
	SemanticManifestDigest string           `json:"semantic_manifest_digest"`
	ProjectionDigest       string           `json:"projection_digest"`
	CombinedDigest         string           `json:"combined_digest"`
	Outputs                []OutputMetadata `json:"outputs"`
}

type OutputMetadata struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
	Bytes  int    `json:"bytes"`
}

type Diagnostic struct {
	Decision string `json:"decision"`
	Stage    string `json:"stage"`
	Step     string `json:"step"`
	Reason   string `json:"reason"`
}

type ScenarioResult struct {
	ID         string      `json:"id"`
	Decision   string      `json:"decision"`
	Expected   string      `json:"expected"`
	Diagnostic *Diagnostic `json:"diagnostic,omitempty"`
	Detail     string      `json:"detail"`
}

type StrategyResult struct {
	Name      string   `json:"name"`
	Selected  bool     `json:"selected"`
	Decision  string   `json:"decision"`
	Reason    string   `json:"reason"`
	Scenarios []string `json:"scenarios"`
}

type ClaimTransition struct {
	State        string `json:"state"`
	TargetDigest string `json:"target_digest"`
	Stage        string `json:"stage"`
	Step         string `json:"step"`
	Reason       string `json:"reason"`
}

type Claim struct {
	ID           string            `json:"id"`
	Proposition  string            `json:"proposition"`
	TargetDigest string            `json:"target_digest"`
	Transitions  []ClaimTransition `json:"transitions"`
}

type RepositoryObservation struct {
	BeforePaths       []string `json:"before_paths"`
	AfterPaths        []string `json:"after_paths"`
	NetStateEqual     bool     `json:"net_state_equal"`
	MutationObserved  bool     `json:"mutation_observed"`
	ObservedAuthority string   `json:"observed_authority"`
}

type Evidence struct {
	Schema                 string                 `json:"schema"`
	Decision               string                 `json:"decision"`
	Reason                 string                 `json:"reason"`
	BoundedSlice           []string               `json:"bounded_slice"`
	BaselineTouchpoints    int                    `json:"baseline_touchpoints"`
	BaselineObservation    []baselineObservation  `json:"baseline_observation"`
	Metrics                IntegrationMetrics     `json:"integration_metrics"`
	MetricDeltas           map[string]MetricDelta `json:"metric_deltas"`
	ProjectionReplay       ScenarioResult         `json:"projection_replay"`
	ManifestOrderInvariant ScenarioResult         `json:"manifest_order_invariant"`
	SemanticCausality      ScenarioResult         `json:"semantic_manifest_causality"`
	CommentInvariant       ScenarioResult         `json:"comment_only_invariant"`
	NewConceptFixture      ScenarioResult         `json:"new_concept_fixture"`
	FailureContracts       []ScenarioResult       `json:"failure_contracts"`
	Strategies             []StrategyResult       `json:"strategies"`
	Claims                 []Claim                `json:"claims"`
	RepositoryNetState     RepositoryObservation  `json:"repository_net_state"`
	GeneratedOutputs       []OutputMetadata       `json:"generated_outputs"`
	MutationAuthority      string                 `json:"mutation_authority"`
}
