package metarecognition

// BaselineConfig contains only explicit assertions and stable IDs. It has no
// display-name, prose, inference, timing, or external command surface.
type BaselineConfig struct {
	Subject          Subject
	StableID         string
	BoundID          string
	ExpectedFile     string
	ObservedFile     string
	ExpectedBlob     string
	ObservedBlob     string
	DeclarationName  string
	DirectivePresent bool
	WorkspaceRoot    string
	SourcePath       string
	RegistryPresent  bool
	SourceMapPresent bool
	Ambiguous        bool

	UnknownIDs []string
	MissedIDs  []string
	Roots      []string
	Commands   []CommandAssertion
	Obligation ObligationAssertion
	Path       PathAssertion
	Resource   ResourceAssertion
	External   ExternalAssertion

	FullCommands     int
	SelectedCommands int
	ProvRecords      int
	ProvPaths        int
}
type CommandAssertion struct {
	ID             string `json:"id"`
	FullStatus     Status `json:"full_status"`
	SelectedStatus Status `json:"selected_status"`
	FullDigest     string `json:"full_digest"`
	SelectedDigest string `json:"selected_digest"`
	Selected       bool   `json:"selected"`
	GlobalGuard    bool   `json:"global_guard"`
	Impacted       bool   `json:"impacted"`
	FullFailure    bool   `json:"full_failure"`
}
type ObligationAssertion struct {
	ID        string    `json:"id"`
	Authority Authority `json:"authority"`
	Impacted  bool      `json:"impacted"`
	Selected  bool      `json:"selected"`
}
type PathAssertion struct {
	IDs       []string `json:"ids"`
	Duplicate bool     `json:"duplicate"`
	Conflict  bool     `json:"conflict"`
}
type ResourceAssertion struct {
	Valid    bool `json:"valid"`
	Overflow bool `json:"overflow"`
}
type ExternalAssertion struct {
	Authenticity bool `json:"authenticity"`
	Provider     bool `json:"provider"`
	Phase        bool `json:"phase"`
	Observer     bool `json:"observer"`
}
type Expected struct {
	State        State    `json:"state"`
	Reason       Reason   `json:"reason"`
	LocalizedIDs []string `json:"localized_ids"`
}
