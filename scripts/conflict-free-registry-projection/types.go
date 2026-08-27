package main

const (
	manifestSchema   = "gooo/language-concept-manifest/v1"
	projectionSchema = "gooo/manual-source-registration-edit-free-registry-projection/v1"
	digestSchema     = "gooo/manual-source-registration-edit-free-registry-projection-digest/v1"
	evidenceSchema   = "gooo/manual-source-registration-edit-free-registry-projection-evidence/v1"
	defaultOutput    = "internal/meta/registryprojection/generated"
)

var expectedConformancePredicateIDs = []string{
	"independent-manifest-order",
	"independent-resource-digests",
	"independent-denominator-reconciliation",
	"independent-binding-registry",
	"independent-conformance-consumer",
}

var expectedFailurePredicateIDs = []string{
	"consumer-malformed-manifest",
	"consumer-missing-manifest",
	"consumer-cross-directory-manifest",
	"consumer-missing-binding",
	"consumer-stale-denominator",
	"consumer-stale-generated-projection",
	"consumer-duplicate-stable-id",
	"consumer-binding-self-search",
	"consumer-binding-output-digest-mismatch",
	"consumer-binding-comment-only",
	"consumer-binding-unused-string",
	"consumer-binding-cross-package-same-name",
	"consumer-binding-shadowed-local",
	"consumer-binding-unused-declaration",
	"consumer-binding-unrelated-use",
	"consumer-binding-unresolved-import",
	"consumer-binding-unrelated-type-error",
	"consumer-binding-metric-row-swap",
	"consumer-binding-different-metric-literal",
	"consumer-binding-unrelated-call",
	"consumer-receipt-occurrence-address-swap",
	"consumer-receipt-occurrence-digest-swap",
	"consumer-receipt-occurrence-pair-swap",
	"consumer-receipt-output-row-cross-swap",
	"consumer-receipt-unknown-field",
	"consumer-receipt-duplicate-metric-id",
	"classifier-success-exit-counterexample",
}

const (
	expectedConformancePredicateCount = 5
	expectedFailurePredicateCount     = 27
	expectedClaimCount                = 32
	expectedFailureContractCount      = 8
	expectedBindingReceiptCount       = 9
	embeddedOutputAddress             = "embedded://consumer_output_artifact/raw_bytes"
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
	MetricID               string `json:"metric_id"`
	RawSourceAddress       string `json:"raw_source_address"`
	RegistrationUseAddress string `json:"registration_use_address"`
	SemanticDigest         string `json:"semantic_digest"`
	ConsumerEntryPoint     string `json:"consumer_entry_point"`
	ObservedOutputAddress  string `json:"observed_output_address"`
	ObservedOutputDigest   string `json:"observed_output_digest"`
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
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	BasisPoints int    `json:"basis_points"`
	Decision    string `json:"decision,omitempty"`
	Stage       string `json:"stage,omitempty"`
	Step        string `json:"step,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type MetricDelta struct {
	Before RatioMetric `json:"before"`
	After  RatioMetric `json:"after"`
}

type IntegrationMetrics struct {
	ExistingSharedSourceTouchpoints RatioMetric `json:"existing_shared_source_touchpoints"`
	GeneratorChangedSharedOutputs   RatioMetric `json:"generator_changed_shared_outputs"`
	IndependentConformanceConsumer  RatioMetric `json:"independent_conformance_consumer"`
	ProducerPackageImports          RatioMetric `json:"producer_package_imports"`
	RawSourceReconstruction         RatioMetric `json:"raw_source_reconstruction"`
	SeparateExecutable              RatioMetric `json:"separate_executable"`
	AlgorithmicIndependence         RatioMetric `json:"algorithmic_independence"`
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
	ID                string             `json:"id"`
	ObservedPredicate string             `json:"observed_predicate"`
	TargetAddress     string             `json:"target_address"`
	TargetDigest      string             `json:"target_digest"`
	Observed          bool               `json:"observed"`
	Decision          string             `json:"decision"`
	PredicateTruth    string             `json:"predicate_truth"`
	ExitCode          int                `json:"exit_code"`
	DiagnosticJSON    string             `json:"diagnostic_json,omitempty"`
	DiagnosticDigest  string             `json:"diagnostic_json_digest,omitempty"`
	RawInputDigest    string             `json:"raw_input_digest,omitempty"`
	RawInputBytes     string             `json:"raw_input_bytes,omitempty"`
	RawInputArtifacts []RawInputArtifact `json:"raw_input_artifacts,omitempty"`
	ContentDigest     string             `json:"content_digest,omitempty"`
	Stage             string             `json:"stage"`
	Step              string             `json:"step"`
	Reason            string             `json:"reason"`
}

type RawInputArtifact struct {
	Path   string `json:"path"`
	Bytes  []byte `json:"bytes"`
	Digest string `json:"digest"`
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

type InventoryReceipt struct {
	Expected []string `json:"expected"`
	Observed []string `json:"observed"`
	Decision string   `json:"decision"`
	Stage    string   `json:"stage"`
	Step     string   `json:"step"`
	Reason   string   `json:"reason"`
}

type ObservedOutputArtifact struct {
	ObservedPath string `json:"observed_path"`
	Digest       string `json:"digest"`
	Bytes        int    `json:"bytes"`
	RawBytes     string `json:"raw_bytes"`
}

type BindingOutputReceipt struct {
	MetricID                string `json:"metric_id"`
	RawSourceAddress        string `json:"raw_source_address"`
	RegistrationUseAddress  string `json:"registration_use_address"`
	SemanticDigest          string `json:"semantic_digest"`
	MetricOccurrenceAddress string `json:"metric_occurrence_address"`
	MetricOccurrenceDigest  string `json:"metric_occurrence_digest"`
	ConsumerEntryPoint      string `json:"consumer_entry_point"`
	OutputAddress           string `json:"output_address"`
	OutputDigest            string `json:"output_digest"`
	OutputBytes             int    `json:"output_bytes"`
	OutputRowAddress        string `json:"output_row_address"`
	OutputRowDigest         string `json:"output_row_digest"`
}

type Evidence struct {
	Schema                        string                      `json:"schema"`
	Decision                      string                      `json:"decision"`
	Reason                        string                      `json:"reason"`
	BoundedSlice                  []string                    `json:"bounded_slice"`
	BaselineTouchpoints           int                         `json:"baseline_touchpoints"`
	BaselineObservation           []baselineObservation       `json:"baseline_observation"`
	Metrics                       IntegrationMetrics          `json:"integration_metrics"`
	MetricDeltas                  map[string]MetricDelta      `json:"metric_deltas"`
	ProjectionReplay              ScenarioResult              `json:"projection_replay"`
	ManifestOrderInvariant        ScenarioResult              `json:"manifest_order_invariant"`
	SemanticCausality             ScenarioResult              `json:"semantic_manifest_causality"`
	SemanticMetricChange          ScenarioResult              `json:"semantic_metric_change"`
	CommentInvariant              ScenarioResult              `json:"comment_only_invariant"`
	CommentPositionInvariant      ScenarioResult              `json:"comment_position_invariant"`
	NewConceptFixture             ScenarioResult              `json:"new_concept_fixture"`
	FailureContracts              []ScenarioResult            `json:"failure_contracts"`
	DenominatorReconciliations    []DenominatorReconciliation `json:"denominator_reconciliations"`
	DenominatorMismatch           ScenarioResult              `json:"denominator_mismatch_contract"`
	StaleDenominatorReceipt       *DenominatorReconciliation  `json:"stale_denominator_receipt"`
	PredicateObservations         []PredicateObservation      `json:"predicate_observations"`
	SourceDigestPreservation      []SourceDigestComparison    `json:"source_digest_preservation"`
	GeneratedOutputChanges        []string                    `json:"generated_output_changes"`
	GeneratedOutputChangeCount    int                         `json:"generated_output_change_count"`
	GeneratedOutputDenominator    int                         `json:"generated_output_denominator"`
	ConformanceConsumer           PredicateMetric             `json:"conformance_consumer"`
	ProductionAdoption            PredicateMetric             `json:"production_adoption"`
	ClaimTransitions              PredicateMetric             `json:"claim_transitions"`
	FailurePredicates             PredicateMetric             `json:"failure_predicates"`
	RepositoryNetStatePredicates  PredicateMetric             `json:"repository_net_state_predicates"`
	BindingPredicates             PredicateMetric             `json:"binding_predicates"`
	ProvenancePredicates          PredicateMetric             `json:"provenance_predicates"`
	UseCaseReceipt                UseCaseReceiptObservation   `json:"use_case_receipt"`
	Strategies                    []StrategyResult            `json:"strategies"`
	Claims                        []Claim                     `json:"claims"`
	RepositoryNetState            RepositoryObservation       `json:"repository_net_state"`
	GeneratedOutputs              []OutputMetadata            `json:"generated_outputs"`
	FixtureGeneratedOutputs       []OutputMetadata            `json:"fixture_generated_outputs"`
	PredicateInventory            InventoryReceipt            `json:"predicate_inventory"`
	FailureInventory              InventoryReceipt            `json:"failure_inventory"`
	ClaimInventory                InventoryReceipt            `json:"claim_inventory"`
	ProvenanceInventory           InventoryReceipt            `json:"provenance_inventory"`
	ASTResolvedBindings           PredicateMetric             `json:"ast_resolved_bindings"`
	MetricOccurrences             PredicateMetric             `json:"metric_occurrences"`
	UniqueSemanticRelationDigests PredicateMetric             `json:"unique_semantic_relation_digests"`
	OutputRowAddresses            PredicateMetric             `json:"output_row_addresses"`
	ConsumerOutputArtifact        ObservedOutputArtifact      `json:"consumer_output_artifact"`
	BindingOutputReceipts         []BindingOutputReceipt      `json:"binding_output_receipts"`
}
