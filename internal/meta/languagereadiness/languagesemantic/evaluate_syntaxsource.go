package languagesemantic

type syntaxSource struct {
	ExpectedHeadSHA  string     `json:"expected_head_sha"`
	ObservationKnown bool       `json:"observation_known"`
	ConceptBound     bool       `json:"concept_bound"`
	GoooFiles        []GoooFile `json:"gooo_files"`
}
