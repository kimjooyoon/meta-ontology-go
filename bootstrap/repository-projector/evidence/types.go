package evidence

type Entry struct {
	Logical    string `json:"logical"`
	Backing    string `json:"backing"`
	ObjectSHA  string `json:"object_sha256"`
	ContentSHA string `json:"content_sha256"`
	Kind       string `json:"kind"`
	Language   string `json:"language,omitempty"`
	Mode       uint32 `json:"mode"`
	Lines      int    `json:"lines,omitempty"`
}

type Indicator struct {
	ID        string `json:"id"`
	Value     int    `json:"value"`
	Limit     int    `json:"limit"`
	Blocking  bool   `json:"blocking"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Proof     string `json:"proof_choice"`
}

type Subject struct {
	Indicator           string `json:"indicator"`
	Logical             string `json:"logical"`
	Physical            string `json:"physical,omitempty"`
	Value               int    `json:"value"`
	Limit               int    `json:"limit"`
	Consumer            string `json:"consumer"`
	Operation           string `json:"meta_operation"`
	Applicability       string `json:"applicability,omitempty"`
	ApplicabilityReason string `json:"applicability_reason,omitempty"`
}

type Topology struct {
	ObservedDirect int
	Direct         int
	ExemptDirect   int
	Mixed          int
	Subjects       []Subject
}

type Report struct {
	Schema       string      `json:"schema"`
	SourceSHA    string      `json:"source_sha"`
	TrackedFiles int         `json:"tracked_files"`
	Objects      int         `json:"stored_objects"`
	Indicators   []Indicator `json:"indicators"`
	Subjects     []Subject   `json:"subjects"`
}

type indicator = Indicator
type subject = Subject
type topologyEvidence = Topology
type evidence = Report
