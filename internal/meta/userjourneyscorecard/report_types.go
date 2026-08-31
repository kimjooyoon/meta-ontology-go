package userjourneyscorecard

type Summary struct {
	Coordinates Counter           `json:"coordinates"`
	Functional  FunctionalSummary `json:"functional"`
	Resources   ResourceSummary   `json:"resources"`
	Meta        MetaSummary       `json:"meta"`
	Effects     EffectSummary     `json:"effects"`
}

type Counter struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type FunctionalSummary struct {
	UpstreamCases      int `json:"upstream_cases"`
	PositiveJourneys   int `json:"positive_journeys"`
	OutputReplays      int `json:"output_replays"`
	StructuredOutputs  int `json:"structured_outputs"`
	LanguageOperations int `json:"language_operations"`
	DeclaredCommands   int `json:"declared_commands"`
}

type ResourceSummary struct {
	SamplesObserved      int   `json:"samples_observed"`
	SamplesExpected      int   `json:"samples_expected"`
	EnvelopesPassed      int   `json:"envelopes_passed"`
	WallViolations       int   `json:"wall_violations"`
	RSSViolations        int   `json:"rss_violations"`
	BinarySizeBytes      int64 `json:"binary_size_bytes"`
	BinarySizeLimit      int64 `json:"binary_size_limit"`
	BinarySizeViolations int   `json:"binary_size_violations"`
}

type MetaSummary struct {
	Bindings int `json:"bindings"`
	Unknowns int `json:"unknowns"`
}

type EffectSummary struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type JourneyStats struct {
	ID, Operation                      string
	Arguments                          []string `json:"arguments"`
	Samples, Successful                int
	OutputReplay, EnvelopePassed       bool
	WallMinMS, WallMedianMS, WallMaxMS int64
	RSSMinKiB, RSSMedianKiB, RSSMaxKiB int64
	StdoutMaxBytes, StderrMaxBytes     int64
}

type Indicator struct {
	ID, Class, ProofChoice, MetaOperation string
	Satisfied                             bool
	Observed, Expected                    int64
}

type AudienceView struct {
	Audience, Decision, Resolution string
	Satisfied, Total, BasisPoints  int
	IndicatorIDs                   []string `json:"indicator_ids"`
}

type Proof struct {
	Choice, Claim, Evidence string
	Passed                  bool
}
