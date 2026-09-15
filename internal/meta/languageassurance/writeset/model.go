package writeset

const (
	SnapshotSchema = "gooo/write-set-snapshot/v1"
	ReceiptSchema  = "gooo/write-set-receipt/v1"
	MetricID       = "gooo.metric.effects.write-set-exactness.v1"
	MetaOperation  = "observe-exact-write-set"
)

type Entry struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Size   int64  `json:"size"`
	Digest string `json:"digest"`
}

type Snapshot struct {
	Schema     string  `json:"schema"`
	RootDigest string  `json:"root_digest"`
	Entries    []Entry `json:"entries"`
}

type Summary struct {
	DeclaredPaths int `json:"declared_paths"`
	ObservedPaths int `json:"observed_paths"`
	MismatchPaths int `json:"mismatch_paths"`
	ExactnessBPS  int `json:"exactness_bps"`
}

type Receipt struct {
	Schema            string   `json:"schema"`
	MetricID          string   `json:"metric_id"`
	MetaOperation     string   `json:"meta_operation"`
	ProofChoice       string   `json:"proof_choice"`
	ObserverID        string   `json:"observer_id"`
	SubjectSHA        string   `json:"subject_sha"`
	DenominatorDigest string   `json:"denominator_digest"`
	BeforeDigest      string   `json:"before_digest,omitempty"`
	AfterDigest       string   `json:"after_digest,omitempty"`
	DeclaredPaths     []string `json:"declared_paths,omitempty"`
	ObservedPaths     []string `json:"observed_paths,omitempty"`
	MismatchPaths     []string `json:"mismatch_paths,omitempty"`
	Decision          string   `json:"decision"`
	Reason            string   `json:"reason"`
	Resolution        string   `json:"resolution"`
	Summary           Summary  `json:"summary"`
	ReceiptDigest     string   `json:"receipt_digest"`
}
