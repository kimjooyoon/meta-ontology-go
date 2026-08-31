package toolchaincli

// Executor is the bounded process boundary consumed by the meta evaluator.
type Executor interface {
	BinaryDigest() (string, error)
	Invoke(arguments []string) (Observation, error)
}

// Session invokes one exact executable from one repository root.
type Session struct {
	Executable string
	Root       string
}

// Observation records the process boundary and runner-scoped resource evidence.
type Observation struct {
	Arguments        []string `json:"arguments"`
	ExitCode         int      `json:"exit_code"`
	Stdout           string   `json:"stdout"`
	Stderr           string   `json:"stderr"`
	Failure          string   `json:"failure,omitempty"`
	PeakRSSKiB       int64    `json:"peak_rss_kib"`
	TreeBeforeDigest string   `json:"tree_before_digest"`
	TreeAfterDigest  string   `json:"tree_after_digest"`
	RepositoryWrites int      `json:"repository_writes"`
}
