package operationconformance

const (
	SplitGoV1ContractID  = "gooo/source-splitter/split-go-conformance/v1"
	SplitGoV1Schema      = "gooo/operation-conformance-contract/v1"
	SplitGoV1Denominator = 6
)

type IndicatorRole string

const (
	RoleOutcome   IndicatorRole = "OUTCOME"
	RoleDriver    IndicatorRole = "DRIVER"
	RoleGuardrail IndicatorRole = "GUARDRAIL"
)

type ProofRoute string

const (
	RouteFoundation ProofRoute = "FOUNDATION"
	RouteCoherence  ProofRoute = "COHERENCE"
	RouteRegression ProofRoute = "REGRESSION"
)

type UnknownPolicy string

const PreserveUnknownAndBlock UnknownPolicy = "PRESERVE_UNKNOWN_AND_BLOCK"

type IndicatorDefinition struct {
	ID             string
	Role           IndicatorRole
	ProofRoute     ProofRoute
	AuthorityURI   string
	EvidenceSchema string
}

type Contract struct {
	Schema        string
	ID            string
	Version       int
	Denominator   int
	UnknownPolicy UnknownPolicy
	Indicators    []IndicatorDefinition
}

