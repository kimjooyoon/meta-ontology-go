package languageprofileexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/languageprofile"

type Input struct {
	SubjectSHA       string                  `json:"subject_sha"`
	ExecutableDigest string                  `json:"executable_digest"`
	Contract         Contract                `json:"contract"`
	First            languageprofile.Receipt `json:"first"`
	Replay           languageprofile.Receipt `json:"replay"`
	UnknownEntry     languageprofile.Receipt `json:"unknown_entry"`
}

type Counter struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Observed      int    `json:"observed"`
	Expected      int    `json:"expected"`
	Satisfied     bool   `json:"satisfied"`
}

type ResourceSummary struct {
	WallObservations       int    `json:"wall_observations"`
	AllocationObservations int    `json:"allocation_observations"`
	WallMinNanoseconds     int64  `json:"wall_min_nanoseconds"`
	WallMedianNanoseconds  int64  `json:"wall_median_nanoseconds"`
	WallMaxNanoseconds     int64  `json:"wall_max_nanoseconds"`
	TotalAllocMinBytes     uint64 `json:"total_alloc_min_bytes"`
	TotalAllocMedianBytes  uint64 `json:"total_alloc_median_bytes"`
	TotalAllocMaxBytes     uint64 `json:"total_alloc_max_bytes"`
}

type CompilerSummary struct {
	ExecutableDigest string `json:"executable_digest"`
	Go127Runtimes    int    `json:"go127_runtimes"`
}

type EffectSummary struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Summary struct {
	Coordinates             Counter         `json:"coordinates"`
	Profiles                int             `json:"profiles"`
	Samples                 int             `json:"samples"`
	SuccessfulExecutions    int             `json:"successful_executions"`
	SourceCoherence         int             `json:"source_coherence"`
	ExecutionDigestVariants int             `json:"execution_digest_variants"`
	Resources               ResourceSummary `json:"resources"`
	Compiler                CompilerSummary `json:"compiler"`
	UnknownEntryRejections  int             `json:"unknown_entry_rejections"`
	Effects                 EffectSummary   `json:"effects"`
	NotClaimed              int             `json:"not_claimed"`
	Unknowns                int             `json:"unknowns"`
}

type View struct {
	Audience     string   `json:"audience"`
	Resolution   string   `json:"resolution"`
	Satisfied    int      `json:"satisfied"`
	Total        int      `json:"total"`
	BasisPoints  int      `json:"basis_points"`
	IndicatorIDs []string `json:"indicator_ids"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Report struct {
	Schema                  string      `json:"schema"`
	Decision                string      `json:"decision"`
	Resolution              string      `json:"resolution"`
	Reason                  string      `json:"reason"`
	Interpretation          string      `json:"interpretation"`
	SubjectSHA              string      `json:"subject_sha"`
	ContractID              string      `json:"contract_id"`
	ResourceObservationMode string      `json:"resource_observation_mode"`
	Summary                 Summary     `json:"summary"`
	Indicators              []Indicator `json:"indicators"`
	Views                   []View      `json:"views"`
	Proofs                  []Proof     `json:"proofs"`
	NotClaimed              []string    `json:"not_claimed"`
	RepositoryWrites        int         `json:"repository_writes"`
	MutationAuthority       bool        `json:"mutation_authority"`
	FactsDigest             string      `json:"facts_digest"`
	Digest                  string      `json:"digest"`
}
