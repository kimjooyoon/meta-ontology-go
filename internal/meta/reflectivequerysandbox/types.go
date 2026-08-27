package reflectivequerysandbox

const (
	Schema             = "gooo/reflective-query-sandbox-observation/v3"
	ReceiptSchema      = "gooo/reflective-query-sandbox-receipt/v3"
	MetricID           = "gooo.metric.language.reflective-query-sandbox.v3"
	SourcePath         = "examples/reflective-query-sandbox/main.gooo"
	ProducerName       = "reflective-query-sandbox.producer"
	ConsumerName       = "reflective-query-sandbox.independent-verifier"
	ProducerImportPath = "github.com/kimjooyoon/meta-ontology-go/internal/meta/reflectivequerysandbox"
)

type Bucket struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

type Contract struct {
	Schema              string   `json:"schema"`
	MetricID            string   `json:"metric_id"`
	GoVersion           string   `json:"go_version"`
	Denominator         int      `json:"denominator"`
	Classes             []Bucket `json:"classes"`
	Proofs              []Bucket `json:"proofs"`
	SourceNodes         int      `json:"source_nodes"`
	SourceFacts         int      `json:"source_facts"`
	ClaimCount          int      `json:"claim_count"`
	AttemptCount        int      `json:"attempt_count"`
	ReflectiveQueries   int      `json:"reflective_queries"`
	SafeQueries         int      `json:"safe_queries"`
	DeniedMutations     int      `json:"denied_mutations"`
	UnknownTargets      int      `json:"unknown_targets"`
	RefutedAttempts     int      `json:"refuted_attempts"`
	TransitionCount     int      `json:"transition_count"`
	SatisfiedIndicators int      `json:"satisfied_indicators"`
}

type Effects struct {
	RepositoryStatusBefore []string `json:"repository_status_before"`
	RepositoryStatusAfter  []string `json:"repository_status_after"`
	NetRepositoryChanges   []string `json:"net_repository_changes"`
	MutationAuthority      bool     `json:"mutation_authority"`
	MutationAPI            string   `json:"mutation_api"`
	MutationOutcome        string   `json:"mutation_outcome"`
	MutationError          string   `json:"mutation_error,omitempty"`
}

type Snapshot struct {
	Path           string `json:"path"`
	SourceDigest   string `json:"source_digest"`
	SemanticDigest string `json:"semantic_digest"`
	GraphDigest    string `json:"graph_digest"`
	NodeCount      int    `json:"node_count"`
	FactCount      int    `json:"fact_count"`
	GoooLines      int    `json:"gooo_lines"`
}

type Attempt struct {
	ID                          string   `json:"id"`
	Class                       string   `json:"class"`
	Operation                   string   `json:"operation"`
	Root                        string   `json:"root"`
	Relation                    string   `json:"relation"`
	Target                      string   `json:"target"`
	MetaOperation               string   `json:"meta_operation"`
	Producer                    string   `json:"producer"`
	Consumer                    string   `json:"consumer"`
	ProofChoice                 string   `json:"proof_choice"`
	Stage                       string   `json:"stage"`
	Step                        string   `json:"step"`
	Decision                    string   `json:"decision"`
	Resolution                  string   `json:"resolution"`
	Reason                      string   `json:"reason"`
	MatchedFacts                int      `json:"matched_facts"`
	EvidenceClaimIDs            []string `json:"evidence_claim_ids"`
	API                         string   `json:"api,omitempty"`
	APIOutcome                  string   `json:"api_outcome,omitempty"`
	APIError                    string   `json:"api_error,omitempty"`
	APIErrorCode                string   `json:"api_error_code,omitempty"`
	ObservedMaterialDigest      string   `json:"observed_material_digest,omitempty"`
	SemanticDigestBefore        string   `json:"semantic_digest_before"`
	SemanticDigestAfter         string   `json:"semantic_digest_after"`
	GraphDigestBefore           string   `json:"graph_digest_before"`
	GraphDigestAfter            string   `json:"graph_digest_after"`
	OriginalSemanticDigestAfter string   `json:"original_semantic_digest_after"`
	OriginalGraphDigestAfter    string   `json:"original_graph_digest_after"`
	ReturnedSemanticDigest      string   `json:"returned_semantic_digest,omitempty"`
	ReturnedGraphDigest         string   `json:"returned_graph_digest,omitempty"`
}

type ClaimTransition struct {
	Sequence               int    `json:"sequence"`
	ClaimID                string `json:"claim_id"`
	Class                  string `json:"class"`
	ProofChoice            string `json:"proof_choice"`
	MetaOperation          string `json:"meta_operation"`
	PriorState             string `json:"prior_state"`
	EvidenceAttempt        string `json:"evidence_attempt"`
	PredicateID            string `json:"predicate_id"`
	Producer               string `json:"producer"`
	Consumer               string `json:"consumer"`
	Stage                  string `json:"stage"`
	Step                   string `json:"step"`
	Reason                 string `json:"reason"`
	From                   string `json:"from"`
	To                     string `json:"to"`
	PreviousDigest         string `json:"previous_digest"`
	Digest                 string `json:"digest"`
	ObservedMaterialDigest string `json:"observed_material_digest,omitempty"`
}

type Observation struct {
	Schema                string            `json:"schema"`
	SubjectSHA            string            `json:"subject_sha"`
	Contract              Contract          `json:"contract"`
	Source                Snapshot          `json:"source"`
	Attempts              []Attempt         `json:"attempts"`
	Claims                []ClaimTransition `json:"claims"`
	Effects               Effects           `json:"effects"`
	ReceiptMaterialDigest string            `json:"receipt_material_digest"`
	Producer              string            `json:"producer"`
	Digest                string            `json:"digest"`
}

type Score struct {
	Name      string `json:"name"`
	Satisfied int    `json:"satisfied"`
	Total     int    `json:"total"`
}

type Coordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type Receipt struct {
	Schema               string            `json:"schema"`
	SubjectSHA           string            `json:"subject_sha"`
	MetricID             string            `json:"metric_id"`
	Decision             string            `json:"decision"`
	Resolution           string            `json:"resolution"`
	SubjectResolution    string            `json:"subject_resolution"`
	Reason               string            `json:"reason"`
	Producer             string            `json:"producer"`
	Consumer             string            `json:"consumer"`
	Contract             Contract          `json:"contract"`
	Source               Snapshot          `json:"source"`
	Attempts             []Attempt         `json:"attempts"`
	Claims               []ClaimTransition `json:"claims"`
	Coordinates          Coordinates       `json:"coordinates"`
	Classes              []Score           `json:"classes"`
	Proofs               []Score           `json:"proofs"`
	Effects              Effects           `json:"effects"`
	SourceReconstruction Coordinates       `json:"source_reconstruction"`
	ProducerImports      Coordinates       `json:"producer_imports"`
	PromotionCreditBPS   int               `json:"promotion_credit_bps"`
	MutationAuthority    bool              `json:"mutation_authority"`
	NotClaimed           []string          `json:"not_claimed"`
	Digest               string            `json:"digest"`
}
