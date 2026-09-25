package metrictransition

const (
	StateSchema  = "gooo/repository-metric-state/v1"
	LedgerSchema = "gooo/metric-transition-ledger/v1"
)

// RootPolicy makes observation independent from root topology enforcement.
type RootPolicy struct {
	Subject               string `json:"subject"`
	SubjectKind           string `json:"subject_kind"`
	CountsApplicability   string `json:"counts_applicability"`
	TopologyApplicability string `json:"topology_applicability"`
	TopologyReason        string `json:"topology_reason"`
	READMERequirement     string `json:"readme_requirement"`
}

// MetricPlane distinguishes logical source from physical storage topology.
type MetricPlane struct {
	Name        string           `json:"name"`
	Root        Counts           `json:"root"`
	Directories []DirectoryState `json:"directories"`
}

// RepositoryState is the canonical, host-path-independent metric state.
type RepositoryState struct {
	Schema        string         `json:"schema"`
	Repository    string         `json:"repository"`
	CommitSHA     string         `json:"commit_sha"`
	RootPolicy    RootPolicy     `json:"root_policy"`
	Source        MetricPlane    `json:"source"`
	Storage       MetricPlane    `json:"storage"`
	LanguageFiles []LanguageFile `json:"language_files"`
	Digest        string         `json:"digest"`
}
