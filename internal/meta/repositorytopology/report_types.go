package repositorytopology

type Summary struct {
	Coordinates Counter         `json:"coordinates"`
	Rows        RowSummary      `json:"rows"`
	Languages   LanguageSummary `json:"languages"`
	Meta        MetaSummary     `json:"meta"`
	Root        RootSummary     `json:"root"`
}

type Counter struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type RowSummary struct {
	FilesObserved       int `json:"files_observed"`
	FilesExact          int `json:"files_exact"`
	DirectoriesObserved int `json:"directories_observed"`
	DirectoriesExact    int `json:"directories_exact"`
	DuplicatePaths      int `json:"duplicate_paths"`
}

type LanguageSummary struct {
	GoFiles   int `json:"go_files"`
	GoooFiles int `json:"gooo_files"`
	GoLines   int `json:"go_lines"`
	GoooLines int `json:"gooo_lines"`
}

type MetaSummary struct {
	Indicators       int `json:"indicators"`
	BoundIndicators  int `json:"bound_indicators"`
	BindingWitnesses int `json:"binding_witnesses"`
	UnknownDecisions int `json:"unknown_decisions"`
	KnownFailClosed  int `json:"known_fail_closed"`
}

type RootSummary struct {
	TopologyExemptions int `json:"topology_exemptions"`
	READMEExemptions   int `json:"readme_exemptions"`
}

type Indicator struct {
	ID          string `json:"id"`
	Class       string `json:"class"`
	ProofChoice string `json:"proof_choice"`
	Satisfied   bool   `json:"satisfied"`
	Observed    int    `json:"observed"`
	Expected    int    `json:"expected"`
}

type Proof struct {
	Choice   string `json:"choice"`
	Claim    string `json:"claim"`
	Evidence string `json:"evidence"`
}

type AudienceView struct {
	Audience     string   `json:"audience"`
	Decision     string   `json:"decision"`
	Resolution   string   `json:"resolution"`
	Satisfied    int      `json:"satisfied"`
	Total        int      `json:"total"`
	BasisPoints  int      `json:"basis_points"`
	IndicatorIDs []string `json:"indicator_ids"`
}
