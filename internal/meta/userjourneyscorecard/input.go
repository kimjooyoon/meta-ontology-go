package userjourneyscorecard

type Contract struct {
	Schema            string              `json:"schema"`
	Version           int                 `json:"version"`
	SamplesPerJourney int                 `json:"samples_per_journey"`
	WallMSLimit       int64               `json:"wall_ms_limit"`
	MaxRSSKiBLimit    int64               `json:"max_rss_kib_limit"`
	BinarySizeLimit   int64               `json:"binary_size_bytes_limit"`
	Source            string              `json:"source"`
	Journeys          []JourneyDefinition `json:"journeys"`
}

type JourneyDefinition struct {
	ID            string   `json:"id"`
	Operation     string   `json:"operation"`
	Arguments     []string `json:"arguments"`
	ProofChoice   string   `json:"proof_choice"`
	MetaOperation string   `json:"meta_operation"`
}

type Profile struct {
	Schema       string     `json:"schema"`
	SubjectSHA   string     `json:"subject_sha"`
	Runner       Runner     `json:"runner"`
	Executable   Executable `json:"executable"`
	SourcePath   string     `json:"source_path"`
	SourceDigest string     `json:"source_digest"`
	Samples      []Sample   `json:"samples"`
}

type Runner struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Image        string `json:"image"`
	ImageVersion string `json:"image_version"`
	GoVersion    string `json:"go_version"`
}

type Executable struct {
	Digest    string `json:"digest"`
	SizeBytes int64  `json:"size_bytes"`
}

type Sample struct {
	Operation    string   `json:"operation"`
	Arguments    []string `json:"arguments"`
	Sequence     int      `json:"sequence"`
	ExitCode     int      `json:"exit_code"`
	WallMS       int64    `json:"wall_ms"`
	MaxRSSKiB    int64    `json:"max_rss_kib"`
	StdoutBytes  int64    `json:"stdout_bytes"`
	StderrBytes  int64    `json:"stderr_bytes"`
	StdoutDigest string   `json:"stdout_digest"`
	StderrDigest string   `json:"stderr_digest"`
}
