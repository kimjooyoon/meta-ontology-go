package externalconformanceactivation

type Input struct {
	SubjectSHA  string
	Assurance   []byte
	Eligibility []byte
	Merge       []byte
}

type ArtifactBinding struct {
	Name           string `json:"name"`
	SourceURI      string `json:"source_uri"`
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactDigest string `json:"artifact_digest"`
	CapsuleDigest  string `json:"capsule_digest"`
	ObservedDigest string `json:"observed_digest"`
	Bytes          int    `json:"bytes"`
	Exact          bool   `json:"exact"`
}

type Transition struct {
	MetricID       string `json:"metric_id"`
	MetaOperation  string `json:"meta_operation"`
	FromStatus     string `json:"from_status"`
	FromResolution string `json:"from_resolution"`
	ToStatus       string `json:"to_status"`
	ToResolution   string `json:"to_resolution"`
}

type validation struct {
	Eligibility      eligibilityReport
	RawExact         int
	AssuranceExact   int
	EligibilityExact int
	MergeExact       int
	Reason           string
	Resolution       string
}
