package languagediagnosticprovenance

import "github.com/kimjooyoon/meta-ontology-go/internal/lsp"

type StepReceipt struct {
	Ordinal       int    `json:"ordinal"`
	Stage         string `json:"stage"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Status        string `json:"status"`
	Effects       int    `json:"effects"`
}

type Trace struct {
	Origin        string         `json:"origin"`
	Stage         string         `json:"stage"`
	Code          string         `json:"code"`
	Hardness      string         `json:"hardness"`
	Physical      Span           `json:"physical"`
	Logical       Span           `json:"logical"`
	SemanticID    string         `json:"semantic_id,omitempty"`
	SemanticKind  string         `json:"semantic_kind,omitempty"`
	Semantic      *Span          `json:"semantic,omitempty"`
	Diagnostic    lsp.Diagnostic `json:"lsp_diagnostic"`
	Steps         []StepReceipt  `json:"steps"`
	TraceDigest   string         `json:"trace_digest"`
}
