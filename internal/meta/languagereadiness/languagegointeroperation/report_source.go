package languagegointeroperation

type Source struct {
	ExpectedHeadSHA       string `json:"expected_head_sha"`
	ConceptID             string `json:"concept_id"`
	Producer              string `json:"producer"`
	Consumer              string `json:"consumer"`
	MetaOperation         string `json:"meta_operation"`
	ConceptArtifactDigest string `json:"concept_artifact_digest"`
	CatalogDigest         string `json:"catalog_digest"`
	RegistryDigest        string `json:"registry_digest"`
	Toolchain             string `json:"toolchain"`
	GoReleaseNotes        string `json:"go_release_notes"`
	MacroReference        string `json:"macro_reference"`
	ConceptBound          bool   `json:"concept_bound"`
}

type StageReceipt struct {
	Ordinal       int    `json:"ordinal"`
	Stage         string `json:"stage"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Status        string `json:"status"`
	Effects       int    `json:"effects"`
}
