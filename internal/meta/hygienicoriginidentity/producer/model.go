package producer

type Report struct {
	SourcePath     string   `json:"source_path"`
	RawDigest      string   `json:"raw_digest"`
	SemanticDigest string   `json:"semantic_digest"`
	Records        []Record `json:"records"`
}

type Record struct {
	CaseID           string `json:"case_id"`
	Spelling         string `json:"spelling"`
	OriginIdentity   string `json:"origin_identity"`
	DefinitionScope  string `json:"definition_scope"`
	UseScope         string `json:"use_scope"`
	ResolvedUseScope string `json:"resolved_use_scope"`
	ResolvedIdentity string `json:"resolved_identity"`
	Captured         bool   `json:"captured"`
}
