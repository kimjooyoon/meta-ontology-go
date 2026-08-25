package languageprofile

import "github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"

const (
	ReceiptSchema          = "gooo/language-profile-receipt/v1"
	RunnerScopedResolution = "RUNNER_SCOPED"
	MaximumSamples         = 20
)

type Request struct {
	Filename string
	Source   string
	Entry    string
	Samples  int
}

type Runner struct {
	GoVersion    string `json:"go_version"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
}

type Measurement struct {
	WallNanoseconds int64  `json:"wall_nanoseconds"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
}

type Sample struct {
	Sequence        int    `json:"sequence"`
	Decision        string `json:"decision"`
	ExecutionDigest string `json:"execution_digest"`
	WallNanoseconds int64  `json:"wall_nanoseconds"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
}

type Summary struct {
	SamplesRequested        int    `json:"samples_requested"`
	SamplesObserved         int    `json:"samples_observed"`
	SuccessfulExecutions    int    `json:"successful_executions"`
	ExecutionDigestVariants int    `json:"execution_digest_variants"`
	WallObservations        int    `json:"wall_observations"`
	AllocationObservations  int    `json:"allocation_observations"`
	WallMinNanoseconds      int64  `json:"wall_min_nanoseconds"`
	WallMedianNanoseconds   int64  `json:"wall_median_nanoseconds"`
	WallMaxNanoseconds      int64  `json:"wall_max_nanoseconds"`
	TotalAllocMinBytes      uint64 `json:"total_alloc_min_bytes"`
	TotalAllocMedianBytes   uint64 `json:"total_alloc_median_bytes"`
	TotalAllocMaxBytes      uint64 `json:"total_alloc_max_bytes"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Receipt struct {
	Schema         string                `json:"schema"`
	Decision       string                `json:"decision"`
	Resolution     string                `json:"resolution"`
	Reason         string                `json:"reason"`
	Filename       string                `json:"filename"`
	Entry          string                `json:"entry"`
	SourceDigest   string                `json:"source_digest"`
	SemanticDigest string                `json:"semantic_digest,omitempty"`
	ProfiledEntry  sourceexecution.Entry `json:"profiled_entry"`
	Runner         Runner                `json:"runner"`
	Samples        []Sample              `json:"samples"`
	Summary        Summary               `json:"summary"`
	Effects        Effects               `json:"effects"`
	NotClaimed     []string              `json:"not_claimed"`
	Digest         string                `json:"digest"`
}
