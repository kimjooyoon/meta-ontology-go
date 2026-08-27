package languageresourcebudgetconsumer

type RawSource struct {
	Filename      string `json:"filename"`
	ContentBase64 string `json:"content_base64"`
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

type Contract struct {
	Schema       string      `json:"schema"`
	SourcePaths  []string    `json:"source_paths"`
	SamplesPerOp int         `json:"samples_per_operation"`
	Indicators   int         `json:"indicators"`
	Operations   []Operation `json:"operations"`
	Limits       Limits      `json:"limits"`
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
	Schema             string   `json:"schema"`
	Producer           string   `json:"producer"`
	Consumer           string   `json:"consumer"`
	BeforeTreeDigest   string   `json:"before_tree_digest"`
	AfterTreeDigest    string   `json:"after_tree_digest"`
	WriteSetDigest     string   `json:"write_set_digest"`
	ChangedPaths       []string `json:"changed_paths"`
	DiffExitCode       int      `json:"diff_exit_code"`
	UntrackedFileCount int      `json:"untracked_file_count"`
	RepositoryWrites   int      `json:"repository_writes"`
	MutationAuthority  bool     `json:"mutation_authority"`
	Reason             string   `json:"reason"`
}

type Producer struct {
	SourceReceiptBase64 string              `json:"source_receipt_base64"`
	ArtifactBase64      string              `json:"artifact_base64"`
	ReplayBase64        string              `json:"replay_base64"`
	SourceDigest        string              `json:"source_digest"`
	SourceFiles         []RawSource         `json:"source_files"`
	SourceFileCount     int                 `json:"source_file_count"`
	GoFiles             int                 `json:"go_files"`
	Runner              Runner              `json:"runner"`
	Effects             Effects             `json:"effects"`
	WriteSet            WriteSetObservation `json:"write_set"`
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

type Input struct {
	Schema        string        `json:"schema"`
	ExpectedHead  string        `json:"expected_head"`
	EvidenceClass string        `json:"evidence_class"`
	Contract      Contract      `json:"contract"`
	Producer      Producer      `json:"producer"`
	Observations  []Observation `json:"observations"`
}
