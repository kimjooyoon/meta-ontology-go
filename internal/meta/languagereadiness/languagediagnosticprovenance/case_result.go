package languagediagnosticprovenance

type Evidence struct {
	ActualOutcome      string `json:"actual_outcome"`
	FailureCode        string `json:"failure_code,omitempty"`
	Traced             bool   `json:"traced"`
	PhysicalBound      bool   `json:"physical_bound"`
	LogicalBound       bool   `json:"logical_bound"`
	SemanticBound      bool   `json:"semantic_bound"`
	LSPProjected       bool   `json:"lsp_projected"`
	CanonicalReplay    bool   `json:"canonical_replay"`
	OrderedDiagnostics bool   `json:"ordered_diagnostics"`
	LineDirectiveRemap bool   `json:"line_directive_remap"`
	TypeClassified     bool   `json:"type_classified"`
	Rejected           bool   `json:"rejected"`
	UnknownAccepted    bool   `json:"unknown_accepted"`
	MissingMapAccepted bool   `json:"missing_map_accepted"`
	AmbiguousAccepted  bool   `json:"ambiguous_map_accepted"`
	InvalidAccepted    bool   `json:"invalid_accepted"`
	ProvenanceSteps    int    `json:"provenance_steps"`
	Effects            int    `json:"effects"`
}

type CaseResult struct {
	Definition Definition `json:"definition"`
	Evidence   Evidence   `json:"evidence"`
	Trace      *Trace     `json:"trace,omitempty"`
	Status     CaseStatus `json:"status"`
	Digest     string     `json:"evidence_digest"`
}
