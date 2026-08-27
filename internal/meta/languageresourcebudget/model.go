package languageresourcebudget

const (
	ContractSchema     = "gooo/meta-resource-budget-contract/v1"
	InputSchema        = "gooo/meta-resource-budget-input/v1"
	ReportSchema       = "gooo/meta-resource-budget-report/v1"
	ObservationSchema  = "gooo/meta-resource-budget-observation/v1"
	ExpectedIndicators = 22
)

type Limits struct {
	WallTimeMS     int64 `json:"wall_time_ms"`
	PeakRSSKiB     int64 `json:"peak_rss_kib"`
	ReceiptBytes   int64 `json:"receipt_bytes"`
	GeneratedBytes int64 `json:"generated_bytes"`
}

type Operation struct {
	ID            string `json:"id"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Output        string `json:"output"`
}

type Reference struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type Contract struct {
	Schema       string      `json:"schema"`
	ID           string      `json:"id"`
	SourcePaths  []string    `json:"source_paths"`
	Entry        string      `json:"entry"`
	SamplesPerOp int         `json:"samples_per_operation"`
	Indicators   int         `json:"indicators"`
	Operations   []Operation `json:"operations"`
	Limits       Limits      `json:"limits"`
	NotClaimed   []string    `json:"not_claimed"`
	References   []Reference `json:"references"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Observation struct {
	Schema         string `json:"schema"`
	SubjectSHA     string `json:"subject_sha"`
	Producer       string `json:"producer"`
	Consumer       string `json:"consumer"`
	Operation      string `json:"operation"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	MetaOperation  string `json:"meta_operation"`
	ProofChoice    string `json:"proof_choice"`
	Reason         string `json:"reason"`
	Sequence       int    `json:"sequence"`
	ExitCode       int    `json:"exit_code"`
	WallTimeNS     int64  `json:"wall_time_ns"`
	PeakRSSKiB     int64  `json:"peak_rss_kib"`
	ReceiptBytes   int64  `json:"receipt_bytes"`
	GeneratedBytes int64  `json:"generated_bytes"`
	OutputDigest   string `json:"output_digest"`
}

type ProducerEvidence struct {
	SourceReceiptBase64 string  `json:"source_receipt_base64"`
	ArtifactBase64      string  `json:"artifact_base64"`
	ReplayBase64        string  `json:"replay_base64"`
	SourceDigest        string  `json:"source_digest"`
	SourceFiles         int     `json:"source_files"`
	GoFiles             int     `json:"go_files"`
	Effects             Effects `json:"effects"`
}

type Input struct {
	Schema       string           `json:"schema"`
	ExpectedHead string           `json:"expected_head"`
	Contract     Contract         `json:"contract"`
	Producer     ProducerEvidence `json:"producer"`
	Observations []Observation    `json:"observations"`
}

type Semantic struct {
	Decision       string `json:"decision"`
	Resolution     string `json:"resolution"`
	Reason         string `json:"reason"`
	SourceDigest   string `json:"source_digest"`
	ArtifactDigest string `json:"artifact_digest"`
	ReplayDigest   string `json:"replay_digest"`
}

type ResourceSummary struct {
	Operation        string `json:"operation"`
	Samples          int    `json:"samples"`
	WallMinNS        int64  `json:"wall_min_ns"`
	WallMedianNS     int64  `json:"wall_median_ns"`
	WallMaxNS        int64  `json:"wall_max_ns"`
	PeakRSSMaxKiB    int64  `json:"peak_rss_max_kib"`
	ReceiptMax       int64  `json:"receipt_max_bytes"`
	GeneratedMax     int64  `json:"generated_max_bytes"`
	BudgetViolations int    `json:"budget_violations"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Observed      int64  `json:"observed"`
	Expected      int64  `json:"expected"`
	Satisfied     bool   `json:"satisfied"`
}

type Counter struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type ClaimTransition struct {
	Sequence       int    `json:"sequence"`
	ClaimID        string `json:"claim_id"`
	From           string `json:"from"`
	To             string `json:"to"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	Evidence       string `json:"evidence"`
	PreviousDigest string `json:"previous_digest"`
	Digest         string `json:"digest"`
}

type CaseResult struct {
	Name       string `json:"name"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Reason     string `json:"reason"`
	Impact     string `json:"impact"`
}

type Summary struct {
	Coordinates Counter           `json:"coordinates"`
	Operations  int               `json:"operations"`
	Samples     int               `json:"samples"`
	Resources   []ResourceSummary `json:"resources"`
	Semantic    Semantic          `json:"semantic"`
	Effects     Effects           `json:"effects"`
	SourceFiles int               `json:"source_files"`
	GoFiles     int               `json:"go_files"`
	Unknowns    int               `json:"unknowns"`
}

type Report struct {
	Schema             string            `json:"schema"`
	Case               string            `json:"case"`
	Decision           string            `json:"decision"`
	Resolution         string            `json:"resolution"`
	Reason             string            `json:"reason"`
	Interpretation     string            `json:"interpretation"`
	ResourceResolution string            `json:"resource_resolution"`
	Semantic           Semantic          `json:"semantic"`
	Summary            Summary           `json:"summary"`
	Indicators         []Indicator       `json:"indicators"`
	Cases              []CaseResult      `json:"cases"`
	Transitions        []ClaimTransition `json:"claim_transitions"`
	NotClaimed         []string          `json:"not_claimed"`
	Effects            Effects           `json:"effects"`
	FactsDigest        string            `json:"facts_digest"`
	Digest             string            `json:"digest"`
}
