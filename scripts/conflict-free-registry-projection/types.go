package main

const (
	manifestSchema   = "gooo/language-concept-manifest/v1"
	projectionSchema = "gooo/manual-source-registration-edit-free-registry-projection/v1"
	digestSchema     = "gooo/manual-source-registration-edit-free-registry-projection-digest/v1"
	evidenceSchema   = "gooo/manual-source-registration-edit-free-registry-projection-evidence/v1"
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
	Path   string `json:"path"`
	Role   string `json:"role"`
	Digest string `json:"digest"`
}

type BindingRegistryEntry struct {
	MetricID              string `json:"metric_id"`
	RawSourceAddress      string `json:"raw_source_address"`
	SemanticDigest        string `json:"semantic_digest"`
	ConsumerEntryPoint    string `json:"consumer_entry_point"`
	ObservedOutputAddress string `json:"observed_output_address"`
	ObservedOutputDigest  string `json:"observed_output_digest"`
}

type Denominator struct {
	ID     string         `json:"id"`
	Values map[string]int `json:"values"`
}

type Manifest struct {
	Schema                 string                 `json:"schema"`
	StableID               string                 `json:"stable_id"`
	Concept                Concept                `json:"concept"`
	CodeBindings           []string               `json:"code_bindings"`
	MetricBindings         []string               `json:"metric_bindings"`
	BindingRegistry        []BindingRegistryEntry `json:"binding_registry"`
	UseCases               []UseCase              `json:"use_cases"`
	VerificationStrategies []string               `json:"verification_strategies"`
	Corpus                 []ResourceRef          `json:"corpus"`
	Registry               []ResourceRef          `json:"registry"`
	Denominators           []Denominator          `json:"denominators"`
	Documentation          []ResourceRef          `json:"documentation"`
	Comments               []string               `json:"comments"`
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
	StableID               string                 `json:"stable_id"`
	SourceManifest         string                 `json:"source_manifest"`
	Problem                string                 `json:"problem"`
	PositiveEffect         string                 `json:"positive_effect"`
	MetaOperation          string                 `json:"meta_operation"`
	Rarity                 string                 `json:"rarity"`
	Stage                  string                 `json:"stage"`
	NoveltyClaim           bool                   `json:"novelty_claim"`
	CodeBindings           []string               `json:"code_bindings"`
	MetricBindings         []string               `json:"metric_bindings"`
	BindingRegistry        []BindingRegistryEntry `json:"binding_registry"`
	UseCases               []UseCase              `json:"use_cases"`
	VerificationStrategies []string               `json:"verification_strategies"`
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
	ExistingSharedSourceTouchpoints RatioMetric `json:"existing_shared_source_touchpoints"`
	GeneratorChangedSharedOutputs   RatioMetric `json:"generator_changed_shared_outputs"`
	IndependentConformanceConsumer  RatioMetric `json:"independent_conformance_consumer"`
	ConceptLocalTouchpoints         RatioMetric `json:"concept_local_touchpoints"`
	ManualSourceRegistrationEdits   RatioMetric `json:"manual_source_registration_edits_required"`
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

type DenominatorReconciliation struct {
	StableID      string         `json:"stable_id"`
	DenominatorID string         `json:"denominator_id"`
	Declared      map[string]int `json:"declared"`
	Calculated    map[string]int `json:"calculated"`
	Decision      string         `json:"decision"`
	Reason        string         `json:"reason"`
}

type PredicateObservation struct {
	ID                 string `json:"id"`
	ObservedPredicate  string `json:"observed_predicate"`
	TargetAddress      string `json:"target_address"`
	TargetDigest       string `json:"target_digest"`
	Observed           bool   `json:"observed"`
	Decision           string `json:"decision"`
	PredicateTruth     string `json:"predicate_truth"`
	ExitCode           int    `json:"exit_code"`
	DiagnosticJSON     string `json:"diagnostic_json,omitempty"`
	DiagnosticDigest   string `json:"diagnostic_json_digest,omitempty"`
	RawInputDigest     string `json:"raw_input_digest,omitempty"`
	ContentDigest      string `json:"content_digest,omitempty"`
	ProvenanceArtifact string `json:"provenance_artifact,omitempty"`
	Stage              string `json:"stage"`
	Step               string `json:"step"`
	Reason             string `json:"reason"`
}

type SourceDigestComparison struct {
	Path   string `json:"path"`
	Before string `json:"before"`
	After  string `json:"after"`
	Equal  bool   `json:"equal"`
}

type StrategyResult struct {
	Name      string   `json:"name"`
	Selected  bool     `json:"selected"`
	Decision  string   `json:"decision"`
	Reason    string   `json:"reason"`
	Scenarios []string `json:"scenarios"`
}

type ClaimTransition struct {
	State             string `json:"state"`
	ObservedPredicate string `json:"observed_predicate"`
	PredicateTruth    string `json:"predicate_truth"`
	PredicateDigest   string `json:"predicate_digest"`
	TargetAddress     string `json:"target_address"`
	TargetDigest      string `json:"target_digest"`
	Stage             string `json:"stage"`
	Step              string `json:"step"`
	Reason            string `json:"reason"`
}

type Claim struct {
	ID                string            `json:"id"`
	Proposition       string            `json:"proposition"`
	ObservedPredicate string            `json:"observed_predicate"`
	TargetAddress     string            `json:"target_address"`
	TargetDigest      string            `json:"target_digest"`
	Transitions       []ClaimTransition `json:"transitions"`
}

type RepositoryObservation struct {
	BeforePaths       []string                 `json:"before_paths"`
	AfterPaths        []string                 `json:"after_paths"`
	BeforeFiles       []RepositoryFileSnapshot `json:"before_files"`
	AfterFiles        []RepositoryFileSnapshot `json:"after_files"`
	NetStateEqual     bool                     `json:"net_state_equal"`
	NetState          string                   `json:"net_state"`
	NetStatePredicate string                   `json:"net_state_predicate"`
	TransientMutation string                   `json:"transient_mutation"`
	MutationAuthority string                   `json:"mutation_authority"`
}

type RepositoryFileSnapshot struct {
	Path    string `json:"path"`
	Digest  string `json:"digest"`
	Tracked bool   `json:"tracked"`
}

type PredicateMetric struct {
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	Decision    string `json:"decision"`
	Stage       string `json:"stage"`
	Step        string `json:"step"`
	Reason      string `json:"reason"`
}

type UseCaseReceiptObservation struct {
	SourceArtifact string `json:"source_artifact"`
	Status         string `json:"status"`
	Numerator      int    `json:"completed_numerator"`
	Denominator    int    `json:"denominator"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
}

type Evidence struct {
	Schema                       string                      `json:"schema"`
	Decision                     string                      `json:"decision"`
	Reason                       string                      `json:"reason"`
	BoundedSlice                 []string                    `json:"bounded_slice"`
	BaselineTouchpoints          int                         `json:"baseline_touchpoints"`
	BaselineObservation          []baselineObservation       `json:"baseline_observation"`
	Metrics                      IntegrationMetrics          `json:"integration_metrics"`
	MetricDeltas                 map[string]MetricDelta      `json:"metric_deltas"`
	ProjectionReplay             ScenarioResult              `json:"projection_replay"`
	ManifestOrderInvariant       ScenarioResult              `json:"manifest_order_invariant"`
	SemanticCausality            ScenarioResult              `json:"semantic_manifest_causality"`
	CommentInvariant             ScenarioResult              `json:"comment_only_invariant"`
	NewConceptFixture            ScenarioResult              `json:"new_concept_fixture"`
	FailureContracts             []ScenarioResult            `json:"failure_contracts"`
	DenominatorReconciliations   []DenominatorReconciliation `json:"denominator_reconciliations"`
	DenominatorMismatch          ScenarioResult              `json:"denominator_mismatch_contract"`
	StaleDenominatorReceipt      *DenominatorReconciliation  `json:"stale_denominator_receipt"`
	PredicateObservations        []PredicateObservation      `json:"predicate_observations"`
	SourceDigestPreservation     []SourceDigestComparison    `json:"source_digest_preservation"`
	GeneratedOutputChanges       []string                    `json:"generated_output_changes"`
	GeneratedOutputChangeCount   int                         `json:"generated_output_change_count"`
	GeneratedOutputDenominator   int                         `json:"generated_output_denominator"`
	ConformanceConsumer          PredicateMetric             `json:"conformance_consumer"`
	ProductionAdoption           PredicateMetric             `json:"production_adoption"`
	ClaimTransitions             PredicateMetric             `json:"claim_transitions"`
	FailurePredicates            PredicateMetric             `json:"failure_predicates"`
	RepositoryNetStatePredicates PredicateMetric             `json:"repository_net_state_predicates"`
	BindingPredicates            PredicateMetric             `json:"binding_predicates"`
	ProvenancePredicates         PredicateMetric             `json:"provenance_predicates"`
	UseCaseReceipt               UseCaseReceiptObservation   `json:"use_case_receipt"`
	Strategies                   []StrategyResult            `json:"strategies"`
	Claims                       []Claim                     `json:"claims"`
	RepositoryNetState           RepositoryObservation       `json:"repository_net_state"`
	GeneratedOutputs             []OutputMetadata            `json:"generated_outputs"`
	FixtureGeneratedOutputs      []OutputMetadata            `json:"fixture_generated_outputs"`
}
