package toolchainconformance

const (
	Schema                 = "gooo/toolchain-conformance-report/v1"
	CorpusSchema           = "gooo/toolchain-conformance-corpus/v1"
	DecisionPass           = "PASS"
	DecisionFailClosed     = "FAIL_CLOSED"
	ResolutionExact        = "EXACT"
	ResolutionLower        = "LOWER_RESOLUTION"
	ExpectedConceptID      = "toolchain-conformance"
	ExpectedMetaOperation  = "close-toolchain-conformance-ledger"
	ExpectedSurfaceCount   = 9
	ExpectedCaseCount      = 156
	ExpectedIndicatorCount = 151
	ExpectedProofCount     = 27
	ExpectedTamperCount    = 13
	ExpectedMetricCount    = 28
)

type SurfaceDefinition struct {
	ID         string `json:"id"`
	Schema     string `json:"schema"`
	Cases      int    `json:"cases"`
	Indicators int    `json:"indicators"`
	Proofs     int    `json:"proofs"`
}

type TamperDefinition struct {
	ID       string `json:"id"`
	Mutation string `json:"mutation"`
	Target   string `json:"target"`
}

type Corpus struct {
	Schema      string              `json:"schema"`
	Surfaces    []SurfaceDefinition `json:"surfaces"`
	TamperCases []TamperDefinition  `json:"tamper_cases"`
}

type Input struct {
	ExpectedHeadSHA string
	ConceptArtifact []byte
	RegistryRaw     []byte
	Artifacts       map[string][]byte
}
