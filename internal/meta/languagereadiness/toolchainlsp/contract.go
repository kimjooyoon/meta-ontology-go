package toolchainlsp

const (
	CorpusSchema        = "gooo/toolchain-lsp-corpus/v1"
	ReportSchema        = "gooo/toolchain-lsp-report/v1"
	DecisionPass        = "PASS"
	DecisionFailClosed  = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionInvariant = "INVARIANT"
	MetaOperation       = "project-exact-language-state-to-editor-protocol"
)

type Case struct {
	ID       string `json:"id"`
	Group    string `json:"group"`
	Expected string `json:"expected"`
}

type Corpus struct {
	Schema string `json:"schema"`
	Cases  []Case `json:"cases"`
}

type CaseResult struct {
	ID             string `json:"id"`
	Group          string `json:"group"`
	Expected       string `json:"expected"`
	Observed       string `json:"observed"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}
