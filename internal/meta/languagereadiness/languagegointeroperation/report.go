package languagegointeroperation

type Evidence struct {
	ActualOutcome       string `json:"actual_outcome"`
	FailureStage        string `json:"failure_stage,omitempty"`
	ErrorCode           string `json:"error_code,omitempty"`
	SourceDigest        string `json:"source_digest,omitempty"`
	ReplaySourceDigest  string `json:"replay_source_digest,omitempty"`
	CanonicalDigest     string `json:"canonical_digest,omitempty"`
	ReplayCanonical     string `json:"replay_canonical_digest,omitempty"`
	APIDigest           string `json:"api_digest,omitempty"`
	ReplayAPIDigest     string `json:"replay_api_digest,omitempty"`
	SourceMapDigest     string `json:"source_map_digest,omitempty"`
	ReplaySourceMap     string `json:"replay_source_map_digest,omitempty"`
	SourceMapMappings   int    `json:"source_map_mappings"`
	ExportedObjects     int    `json:"exported_objects"`
	GenericMethods      int    `json:"generic_methods"`
	AliasNodes          int    `json:"alias_nodes"`
	ASTReifications     int    `json:"ast_reifications"`
	CanonicalReplay     bool   `json:"canonical_replay"`
	TypeIdentityReplay  bool   `json:"type_identity_replay"`
	Rejected            bool   `json:"rejected"`
	InvalidAccepted     bool   `json:"invalid_accepted"`
	UnknownAccepted     bool   `json:"unknown_accepted"`
	ImportAccepted      bool   `json:"import_accepted"`
	Effects             int    `json:"effects"`
}

type CaseResult struct {
	Definition Definition `json:"definition"`
	Evidence   Evidence   `json:"evidence"`
	Status     CaseStatus `json:"status"`
	Digest     string     `json:"evidence_digest"`
}
