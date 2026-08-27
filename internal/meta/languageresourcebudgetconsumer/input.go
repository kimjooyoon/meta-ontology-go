package languageresourcebudgetconsumer

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
	SamplesPerOp int         `json:"samples_per_operation"`
	Indicators   int         `json:"indicators"`
	Operations   []Operation `json:"operations"`
	Limits       Limits      `json:"limits"`
	NotClaimed   []string    `json:"not_claimed"`
	References   []Reference `json:"references"`
}

type Runner struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Image        string `json:"image"`
	ImageVersion string `json:"image_version"`
	GoVersion    string `json:"go_version"`
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

type Producer struct {
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

type Input struct {
	Schema         string        `json:"schema"`
	ExpectedHead   string        `json:"expected_head"`
	EvidenceClass  string        `json:"evidence_class"`
	ContractDigest string        `json:"contract_digest"`
	Contract       Contract      `json:"contract"`
	Producer       Producer      `json:"producer"`
	Observations   []Observation `json:"observations"`
}
