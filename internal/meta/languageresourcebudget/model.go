package languageresourcebudget

const (
	ContractSchema     = "gooo/meta-resource-budget-contract/v1"
	InputSchema        = "gooo/meta-resource-budget-input/v1"
	ReportSchema       = "gooo/meta-resource-budget-report/v1"
	ObservationSchema  = "gooo/meta-resource-budget-observation/v1"
	ExpectedIndicators = 19
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

type RawSource struct {
	Filename      string `json:"filename"`
	ContentBase64 string `json:"content_base64"`
}

type RawOutput struct {
	Operation     string `json:"operation"`
	Sequence      int    `json:"sequence"`
	Kind          string `json:"kind"`
	PayloadBase64 string `json:"payload_base64"`
}

type Reference struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type Contract struct {
	Schema       string      `json:"schema"`
	ID           string      `json:"id"`
	SourcePaths  []string    `json:"source_paths"`
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

type WriteSetObservation struct {
	Operation            string   `json:"operation"`
	Schema               string   `json:"schema"`
	Producer             string   `json:"producer"`
	Consumer             string   `json:"consumer"`
	BeforeTreeDigest     string   `json:"before_tree_digest"`
	AfterTreeDigest      string   `json:"after_tree_digest"`
	BeforeStatusBase64   string   `json:"before_status_base64"`
	AfterStatusBase64    string   `json:"after_status_base64"`
	BeforeStatusObserved bool     `json:"before_status_observed"`
	AfterStatusObserved  bool     `json:"after_status_observed"`
	WriteSetDigest       string   `json:"write_set_digest"`
	ChangedPaths         []string `json:"changed_paths"`
	DiffExitCode         int      `json:"diff_exit_code"`
	UntrackedFileCount   int      `json:"untracked_file_count"`
	RepositoryWrites     int      `json:"repository_writes"`
	MutationAuthority    bool     `json:"mutation_authority"`
	AuthorityObserved    bool     `json:"authority_observed"`
	SampleStart          int      `json:"sample_start"`
	SampleEnd            int      `json:"sample_end"`
	Reason               string   `json:"reason"`
}

type ImportScan struct {
	Schema                         string `json:"schema"`
	ConsumerPackageReducerImported bool   `json:"consumer_package_reducer_imported"`
	ConsumerCommandReducerImported bool   `json:"consumer_command_reducer_imported"`
	ConsumerPackageFilesScanned    int    `json:"consumer_package_files_scanned"`
	ConsumerCommandFilesScanned    int    `json:"consumer_command_files_scanned"`
	Numerator                      int    `json:"numerator"`
	Denominator                    int    `json:"denominator"`
	Digest                         string `json:"digest"`
}

type Runner struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Image        string `json:"image"`
	ImageVersion string `json:"image_version"`
	GoVersion    string `json:"go_version"`
}

type Observation struct {
	Schema               string `json:"schema"`
	SubjectSHA           string `json:"subject_sha"`
	Producer             string `json:"producer"`
	Consumer             string `json:"consumer"`
	Operation            string `json:"operation"`
	Stage                string `json:"stage"`
	Step                 string `json:"step"`
	MetaOperation        string `json:"meta_operation"`
	ProofChoice          string `json:"proof_choice"`
	Reason               string `json:"reason"`
	Sequence             int    `json:"sequence"`
	ExitCode             int    `json:"exit_code"`
	WallTimeNS           int64  `json:"wall_time_ns"`
	PeakRSSKiB           int64  `json:"peak_rss_kib"`
	ReceiptBytes         int64  `json:"receipt_bytes"`
	GeneratedBytes       int64  `json:"generated_bytes"`
	OutputDigest         string `json:"output_digest"`
	SourceRawDigest      string `json:"source_raw_digest"`
	SourceSemanticDigest string `json:"source_semantic_digest"`
	EntryDigest          string `json:"entry_digest"`
	TargetDigest         string `json:"target_digest"`
}

type ProducerEvidence struct {
	SourceReceiptBase64 string                `json:"source_receipt_base64"`
	ArtifactBase64      string                `json:"artifact_base64"`
	ReplayBase64        string                `json:"replay_base64"`
	SourceDigest        string                `json:"source_digest"`
	SourceFiles         []RawSource           `json:"source_files"`
	RawOutputs          []RawOutput           `json:"raw_outputs"`
	SourceFileCount     int                   `json:"source_file_count"`
	GoFiles             int                   `json:"go_files"`
	Runner              Runner                `json:"runner"`
	Effects             Effects               `json:"effects"`
	WriteSets           []WriteSetObservation `json:"write_sets"`
	ImportScan          ImportScan            `json:"import_scan"`
}

type Input struct {
	Schema         string           `json:"schema"`
	ExpectedHead   string           `json:"expected_head"`
	EvidenceClass  string           `json:"evidence_class"`
	ContractDigest string           `json:"contract_digest"`
	Contract       Contract         `json:"contract"`
	Producer       ProducerEvidence `json:"producer"`
	Observations   []Observation    `json:"observations"`
}

type Semantic struct {
	Decision       string `json:"decision"`
	Resolution     string `json:"resolution"`
	ClaimState     string `json:"claim_state"`
	Reason         string `json:"reason"`
	SourceDigest   string `json:"source_digest"`
	SemanticDigest string `json:"semantic_digest"`
	TargetDigest   string `json:"target_digest"`
	ArtifactDigest string `json:"artifact_digest"`
	ReplayDigest   string `json:"replay_digest"`
}

type ResourceSummary struct {
	Operation        string `json:"operation"`
	Samples          int    `json:"samples"`
	MissingSamples   int    `json:"missing_samples"`
	InvalidSamples   int    `json:"invalid_samples"`
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
	Status        string `json:"status"`
	Satisfied     bool   `json:"satisfied"`
}

type Counter struct {
	Satisfied   int `json:"satisfied"`
	Refuted     int `json:"refuted"`
	Unknown     int `json:"unknown"`
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
	Name          string `json:"name"`
	EvidenceClass string `json:"evidence_class"`
	Decision      string `json:"decision"`
	Resolution    string `json:"resolution"`
	Reason        string `json:"reason"`
	Impact        string `json:"impact"`
}

type Summary struct {
	Coordinates    Counter               `json:"coordinates"`
	Operations     int                   `json:"operations"`
	Samples        int                   `json:"samples"`
	Resources      []ResourceSummary     `json:"resources"`
	Semantic       Semantic              `json:"semantic"`
	Effects        Effects               `json:"effects"`
	SourceFiles    int                   `json:"source_files"`
	GoFiles        int                   `json:"go_files"`
	Runner         Runner                `json:"runner"`
	WriteSets      []WriteSetObservation `json:"write_sets"`
	EvidenceDigest string                `json:"evidence_digest"`
	Unknowns       int                   `json:"unknowns"`
}

type Report struct {
	Schema             string                `json:"schema"`
	Case               string                `json:"case"`
	EvidenceClass      string                `json:"evidence_class"`
	Decision           string                `json:"decision"`
	Resolution         string                `json:"resolution"`
	Reason             string                `json:"reason"`
	Interpretation     string                `json:"interpretation"`
	ResourceResolution string                `json:"resource_resolution"`
	ReadOnlyResolution string                `json:"read_only_resolution"`
	Semantic           Semantic              `json:"semantic"`
	Summary            Summary               `json:"summary"`
	Indicators         []Indicator           `json:"indicators"`
	Cases              []CaseResult          `json:"cases"`
	Transitions        []ClaimTransition     `json:"claim_transitions"`
	NotClaimed         []string              `json:"not_claimed"`
	Effects            Effects               `json:"effects"`
	WriteSets          []WriteSetObservation `json:"write_sets"`
	FactsDigest        string                `json:"facts_digest"`
	ProvenanceDigest   string                `json:"provenance_digest"`
	Digest             string                `json:"digest"`
}
