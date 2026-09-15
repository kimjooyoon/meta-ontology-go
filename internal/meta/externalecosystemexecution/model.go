package externalecosystemexecution

type ReferenceReceipt struct {
	Available       bool   `json:"available"`
	BindingExact    bool   `json:"binding_exact"`
	ContractVersion string `json:"contract_version"`
	Decision        string `json:"decision"`
	Resolution      string `json:"resolution"`
	URL             string `json:"url"`
	Commit          string `json:"commit"`
	Tree            string `json:"tree"`
	ModuleGo        string `json:"module_go"`
	EvidencePath    string `json:"evidence_path"`
	EvidenceSHA256  string `json:"evidence_sha256"`
}

type RepositoryState struct {
	Available    bool   `json:"available"`
	Commit       string `json:"commit"`
	Tree         string `json:"tree"`
	Dirty        bool   `json:"dirty"`
	StatusSHA256 string `json:"status_sha256"`
}

type Outcome struct {
	Package string `json:"package"`
	Test    string `json:"test,omitempty"`
	Action  string `json:"action"`
}

type goEvent struct {
	Action     string `json:"Action"`
	Package    string `json:"Package"`
	Test       string `json:"Test"`
	ImportPath string `json:"ImportPath"`
	Output     string `json:"Output"`
	OutputType string `json:"OutputType"`
}

var knownEventActions = map[string]bool{"start": true, "run": true, "pause": true, "cont": true, "pass": true,
	"bench": true, "fail": true, "output": true, "skip": true, "build-output": true, "build-fail": true}

type RunObservation struct {
	Index            int       `json:"index"`
	ExitCode         int       `json:"exit_code"`
	Passed           bool      `json:"passed"`
	EventCount       int       `json:"event_count"`
	RawSHA256        string    `json:"raw_sha256"`
	StderrSHA256     string    `json:"stderr_sha256"`
	StderrLineCount  int       `json:"stderr_line_count"`
	NormalizedSHA256 string    `json:"normalized_sha256"`
	Outcomes         []Outcome `json:"outcomes"`
	UnknownEvents    []string  `json:"unknown_events"`
	Diagnostics      []string  `json:"diagnostics"`
}

type RegressionReceipt struct {
	Passed     int `json:"passed"`
	Total      int `json:"total"`
	Unresolved int `json:"unresolved"`
}

type Observation struct {
	Schema                string            `json:"schema"`
	Reference             ReferenceReceipt  `json:"reference"`
	GoVersion             string            `json:"go_version"`
	Runs                  []RunObservation  `json:"runs"`
	SourceBefore          RepositoryState   `json:"source_before"`
	SourceAfter           RepositoryState   `json:"source_after"`
	ExternalBefore        RepositoryState   `json:"external_before"`
	ExternalAfter         RepositoryState   `json:"external_after"`
	Regression            RegressionReceipt `json:"regression"`
	OfficialMutationCount int               `json:"official_mutation_count"`
	PromotionCount        int               `json:"promotion_count"`
}
