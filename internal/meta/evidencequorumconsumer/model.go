package evidencequorumconsumer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumpolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorumwire"
)

const (
	ReportSchema        = "gooo/meta-evidence-quorum-consumer-report/v2"
	Scope               = "GOOO_CLAIM_JUSTIFICATION_ONLY"
	DecisionPass        = "PASS"
	DecisionClosed      = "FAIL_CLOSED"
	DecisionUnknown     = "UNKNOWN"
	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	ResolutionInvariant = "INVARIANT_ONLY"
	StatusDischarged    = "DISCHARGED"
	StatusOpen          = "OPEN"
	StatusRefuted       = "REFUTED"
	ObservationExact    = "EXACT"
	ObservationUnknown  = "UNKNOWN"
)

type CaseInput struct {
	ID       string
	Receipts [][]byte
}

type Input struct {
	Policy     evidencequorumpolicy.Policy
	HeadSHA    string
	SourcePath string
	Source     []byte
	Cases      []CaseInput
}

type Provenance struct {
	OriginGroup           string   `json:"origin_group"`
	LineageKey            string   `json:"lineage_key"`
	ExecutableDigest      string   `json:"executable_digest"`
	DependencyPaths       []string `json:"dependency_paths"`
	DependencyDigest      string   `json:"dependency_digest"`
	SubjectRawDigest      string   `json:"subject_raw_digest"`
	SubjectSemanticDigest string   `json:"subject_semantic_digest"`
	ObservationDigest     string   `json:"observation_digest"`
}

type GroupResult struct {
	OriginGroup     string     `json:"origin_group"`
	EvidenceIDs     []string   `json:"evidence_ids"`
	Roles           []string   `json:"roles"`
	Values          []string   `json:"values"`
	EvidenceClasses []string   `json:"evidence_classes"`
	Provenance      Provenance `json:"provenance"`
	Independent     bool       `json:"independent"`
}

type ClaimTransition struct {
	From            string       `json:"from"`
	To              string       `json:"to"`
	PreviousDigest  string       `json:"previous_digest"`
	EvidenceDigests []string     `json:"evidence_digests"`
	Provenance      []Provenance `json:"provenance"`
	Stage           string       `json:"stage"`
	Step            string       `json:"step"`
	Reason          string       `json:"reason"`
}

type ClaimResult struct {
	ID               string            `json:"id"`
	Producer         string            `json:"producer"`
	Consumer         string            `json:"consumer"`
	MetaOperation    string            `json:"meta_operation"`
	ProofChoice      string            `json:"proof_choice"`
	State            string            `json:"state"`
	SubjectDecision  string            `json:"subject_decision"`
	ObservationState string            `json:"observation_state"`
	Resolution       string            `json:"resolution"`
	Reason           string            `json:"reason"`
	Stage            string            `json:"stage"`
	Step             string            `json:"step"`
	EvidenceDigests  []string          `json:"evidence_digests"`
	Transitions      []ClaimTransition `json:"transitions"`
}

type CaseResult struct {
	ID                  string        `json:"id"`
	Status              string        `json:"status"`
	ConformanceDecision string        `json:"conformance_decision"`
	SubjectState        string        `json:"subject_state"`
	SubjectDecision     string        `json:"subject_decision"`
	ObservationState    string        `json:"observation_state"`
	Resolution          string        `json:"resolution"`
	Reason              string        `json:"reason"`
	Stage               string        `json:"stage"`
	Step                string        `json:"step"`
	RawReceipts         int           `json:"raw_receipts"`
	CurrentEvidence     int           `json:"current_evidence"`
	SyntheticEvidence   int           `json:"synthetic_evidence"`
	IndependentGroups   int           `json:"independent_groups"`
	CollapsedReplicas   int           `json:"collapsed_replicas"`
	ConflictGroups      int           `json:"conflict_groups"`
	Groups              []GroupResult `json:"groups"`
	Claims              []ClaimResult `json:"claims"`
}

type Summary struct {
	CasesSatisfied             int  `json:"cases_satisfied"`
	CasesTotal                 int  `json:"cases_total"`
	ClaimsTotal                int  `json:"claims_total"`
	DischargedClaims           int  `json:"discharged_claims"`
	OpenClaims                 int  `json:"open_claims"`
	RefutedClaims              int  `json:"refuted_claims"`
	CurrentEvidenceTotal       int  `json:"current_evidence_total"`
	SyntheticEvidenceTotal     int  `json:"synthetic_evidence_total"`
	RawReceiptsTotal           int  `json:"raw_receipts_total"`
	DistinctProvenanceGroups   int  `json:"distinct_provenance_groups"`
	CollapsedReplicas          int  `json:"collapsed_replicas"`
	ConflictCases              int  `json:"conflict_cases"`
	QuorumSatisfiedCases       int  `json:"quorum_satisfied_cases"`
	LowerResolutionCases       int  `json:"lower_resolution_cases"`
	UnknownObservationCases    int  `json:"unknown_observation_cases"`
	MinimumIndependentGroups   int  `json:"minimum_independent_groups"`
	SourceReconstructionCount  int  `json:"source_reconstruction_count"`
	SourceReconstructionTotal  int  `json:"source_reconstruction_total"`
	ProducerPackageImports     int  `json:"producer_package_imports"`
	ProducerPackageImportTotal int  `json:"producer_package_import_total"`
	ConfidenceAggregated       bool `json:"confidence_aggregated"`
	RepositoryWrites           int  `json:"repository_writes"`
	MutationAuthority          bool `json:"mutation_authority"`
}

type Intervention struct {
	Name                    string `json:"name"`
	BeforeDecision          string `json:"before_decision"`
	AfterDecision           string `json:"after_decision"`
	BeforeSemanticDigest    string `json:"before_semantic_digest"`
	AfterSemanticDigest     string `json:"after_semantic_digest"`
	BeforeObservationDigest string `json:"before_observation_digest"`
	AfterObservationDigest  string `json:"after_observation_digest"`
	QuorumResultChanged     bool   `json:"quorum_result_changed"`
	SemanticDigestChanged   bool   `json:"semantic_digest_changed"`
	EffectsBefore           string `json:"effects_before"`
	EffectsAfter            string `json:"effects_after"`
}

type Report struct {
	Schema               string         `json:"schema"`
	Scope                string         `json:"scope"`
	HeadSHA              string         `json:"head_sha"`
	SourcePath           string         `json:"source_path"`
	SourceEntry          string         `json:"source_entry"`
	SourceRawDigest      string         `json:"source_raw_digest"`
	SourceSemanticDigest string         `json:"source_semantic_digest"`
	PolicySemanticDigest string         `json:"policy_semantic_digest"`
	Decision             string         `json:"decision"`
	SubjectDecision      string         `json:"subject_decision"`
	Resolution           string         `json:"resolution"`
	Reason               string         `json:"reason"`
	ReceiptDigests       []string       `json:"receipt_digests"`
	Cases                []CaseResult   `json:"cases"`
	Summary              Summary        `json:"summary"`
	Interventions        []Intervention `json:"interventions"`
	NotClaimed           []string       `json:"not_claimed"`
	RepositoryWrites     int            `json:"repository_writes"`
	MutationAuthority    bool           `json:"mutation_authority"`
	Digest               string         `json:"digest"`
}

type classifiedEvidence struct {
	Receipt    evidencequorumwire.Receipt
	Digest     string
	Role       string
	Value      string
	Predicate  string
	Provenance Provenance
}
