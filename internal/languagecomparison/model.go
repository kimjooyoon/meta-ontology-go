package languagecomparison

const (
	ReceiptSchema      = "gooo/language-comparison-receipt/v1"
	ContractID         = "billing-declaration-signature-go-ast-v1"
	SamplesPerLanguage = 5
	MaximumSamples     = 20
)

type Request struct {
	SubjectSHA       string
	ExecutableDigest string
	RunnerLabel      string
	GoooFilename     string
	GoooSource       string
	GoFilename       string
	GoSource         string
	Entry            string
	Samples          int
}

type Runner struct {
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"architecture"`
	CPUs      int    `json:"logical_cpus"`
	Label     string `json:"runner_label"`
}

type Measurement struct {
	WallNanoseconds int64  `json:"wall_nanoseconds"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
}

type Sample struct {
	Sequence         int         `json:"sequence"`
	GoooDecision     string      `json:"gooo_decision"`
	GoDecision       string      `json:"go_decision"`
	GoooOutputDigest string      `json:"gooo_output_digest"`
	GoOutputDigest   string      `json:"go_output_digest"`
	Gooo             Measurement `json:"gooo"`
	Go               Measurement `json:"go"`
	FirstMeasured    string      `json:"first_measured"`
}

type SideSummary struct {
	WallMinNanoseconds    int64  `json:"wall_min_nanoseconds"`
	WallMedianNanoseconds int64  `json:"wall_median_nanoseconds"`
	WallMaxNanoseconds    int64  `json:"wall_max_nanoseconds"`
	TotalAllocMinBytes    uint64 `json:"total_alloc_min_bytes"`
	TotalAllocMedianBytes uint64 `json:"total_alloc_median_bytes"`
	TotalAllocMaxBytes    uint64 `json:"total_alloc_max_bytes"`
}

type Summary struct {
	SamplesRequested         int         `json:"samples_requested"`
	SamplesObserved          int         `json:"samples_observed"`
	EquivalentOutputSamples  int         `json:"equivalent_output_samples"`
	GoooOutputDigestVariants int         `json:"gooo_output_digest_variants"`
	GoOutputDigestVariants   int         `json:"go_output_digest_variants"`
	Gooo                     SideSummary `json:"gooo"`
	Go                       SideSummary `json:"go"`
	GoToGoooWallRatioPPM     uint64      `json:"go_to_gooo_wall_ratio_ppm"`
	GoToGoooAllocRatioPPM    uint64      `json:"go_to_gooo_allocation_ratio_ppm"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Receipt struct {
	Schema           string   `json:"schema"`
	ContractID       string   `json:"contract_id"`
	SubjectSHA       string   `json:"subject_sha"`
	ExecutableDigest string   `json:"executable_digest"`
	Decision         string   `json:"decision"`
	Resolution       string   `json:"resolution"`
	Reason           string   `json:"reason"`
	Scope            string   `json:"scope"`
	GoooFilename     string   `json:"gooo_filename"`
	GoFilename       string   `json:"go_filename"`
	Entry            string   `json:"entry"`
	GoooSourceDigest string   `json:"gooo_source_digest"`
	GoSourceDigest   string   `json:"go_source_digest"`
	Runner           Runner   `json:"runner"`
	Samples          []Sample `json:"samples"`
	Summary          Summary  `json:"summary"`
	Effects          Effects  `json:"effects"`
	NotClaimed       []string `json:"not_claimed"`
	Digest           string   `json:"digest"`
}
