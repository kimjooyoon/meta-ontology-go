package reflectivequerysandbox

const (
	Schema                  = "gooo/reflective-query-sandbox-observation/v1"
	ReceiptSchema           = "gooo/reflective-query-sandbox-receipt/v1"
	MetricID                = "gooo.metric.language.reflective-query-sandbox.v1"
	ExpectedGoVersion       = "1.27.0"
	ExpectedSourcePath      = "examples/reflective-query-sandbox/main.gooo"
	ExpectedDenominator     = 12
	ExpectedTransitionCount = 24
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
	ExpectedNodes       int      `json:"expected_nodes"`
	ExpectedFacts       int      `json:"expected_facts"`
	ExpectedAttempts    int      `json:"expected_attempts"`
	ExpectedSafe        int      `json:"expected_safe_queries"`
	ExpectedDenied      int      `json:"expected_denied_mutations"`
	ExpectedUnknown     int      `json:"expected_unknown_targets"`
	ExpectedTransitions int      `json:"expected_transitions"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Snapshot struct {
	Path           string `json:"path"`
	SourceDigest   string `json:"source_digest"`
	SemanticDigest string `json:"semantic_digest"`
	NodeCount      int    `json:"node_count"`
	FactCount      int    `json:"fact_count"`
	GoooLines      int    `json:"gooo_lines"`
}

type Attempt struct {
	ID                   string `json:"id"`
	Class                string `json:"class"`
	Operation            string `json:"operation"`
	Root                 string `json:"root"`
	Relation             string `json:"relation"`
	Target               string `json:"target"`
	MetaOperation        string `json:"meta_operation"`
	Producer             string `json:"producer"`
	Consumer             string `json:"consumer"`
	ProofChoice          string `json:"proof_choice"`
	Stage                string `json:"stage"`
	Step                 string `json:"step"`
	Decision             string `json:"decision"`
	Resolution           string `json:"resolution"`
	Reason               string `json:"reason"`
	MatchedFacts         int    `json:"matched_facts"`
	SemanticDigestBefore string `json:"semantic_digest_before"`
	SemanticDigestAfter  string `json:"semantic_digest_after"`
	GraphDigestBefore    string `json:"graph_digest_before"`
	GraphDigestAfter     string `json:"graph_digest_after"`
}

type ClaimTransition struct {
	Sequence       int    `json:"sequence"`
	ClaimID        string `json:"claim_id"`
	Class          string `json:"class"`
	ProofChoice    string `json:"proof_choice"`
	MetaOperation  string `json:"meta_operation"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	From           string `json:"from"`
	To             string `json:"to"`
	PreviousDigest string `json:"previous_digest"`
	Digest         string `json:"digest"`
}

type Observation struct {
	Schema     string            `json:"schema"`
	SubjectSHA string            `json:"subject_sha"`
	Contract   Contract          `json:"contract"`
	Source     Snapshot          `json:"source"`
	Attempts   []Attempt         `json:"attempts"`
	Claims     []ClaimTransition `json:"claims"`
	Effects    Effects           `json:"effects"`
	Producer   string            `json:"producer"`
	Digest     string            `json:"digest"`
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
	Schema             string            `json:"schema"`
	SubjectSHA         string            `json:"subject_sha"`
	MetricID           string            `json:"metric_id"`
	Decision           string            `json:"decision"`
	Resolution         string            `json:"resolution"`
	Reason             string            `json:"reason"`
	Producer           string            `json:"producer"`
	Consumer           string            `json:"consumer"`
	Contract           Contract          `json:"contract"`
	Source             Snapshot          `json:"source"`
	Attempts           []Attempt         `json:"attempts"`
	Claims             []ClaimTransition `json:"claims"`
	Coordinates        Coordinates       `json:"coordinates"`
	Classes            []Score           `json:"classes"`
	Proofs             []Score           `json:"proofs"`
	Effects            Effects           `json:"effects"`
	PromotionCreditBPS int               `json:"promotion_credit_bps"`
	RepositoryWrites   int               `json:"repository_writes"`
	MutationAuthority  bool              `json:"mutation_authority"`
	NotClaimed         []string          `json:"not_claimed"`
	Digest             string            `json:"digest"`
}
