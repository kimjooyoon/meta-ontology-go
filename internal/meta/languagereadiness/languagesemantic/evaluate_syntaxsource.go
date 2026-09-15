package languagesemantic

type syntaxSource struct {
	ExpectedHeadSHA  string              `json:"expected_head_sha"`
	ObservationKnown bool                `json:"observation_known"`
	ConceptBound     bool                `json:"concept_bound"`
	GoooFiles        []GoooFile          `json:"gooo_files"`
	PackageUnits     []syntaxPackageUnit `json:"package_units"`
}

type syntaxPackageUnit struct {
	ID                   string   `json:"id"`
	Path                 string   `json:"path"`
	Members              []string `json:"members"`
	Entry                string   `json:"entry"`
	ReportSchema         string   `json:"report_schema"`
	MetaReducer          string   `json:"meta_reducer"`
	SourceFilesIndicator string   `json:"source_files_indicator"`
	ExecutionIndicator   string   `json:"execution_indicator"`
}
