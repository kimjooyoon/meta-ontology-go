package languagedelivery

type Audience string
type IndicatorClass string
type ProofChoice string
type EvidenceKind string
type SourceName string
type ResultStatus string

const (
	ContractSchema = "gooo/language-delivery-contract/v1"
	ManifestSchema = "gooo/language-delivery-source-manifest/v1"
	ReportSchema   = "gooo/language-delivery-scorecard/v1"
	ContractID     = "gooo-v0.1-observable-delivery"
	MetaOperation  = "reduce-fixed-language-delivery-contract"
)

const (
	AudienceUser       Audience = "USER"
	AudienceToolAuthor Audience = "TOOL_AUTHOR"
	AudienceGovernor   Audience = "GOVERNOR"
	ClassOutcome       IndicatorClass = "OUTCOME"
	ClassDriver        IndicatorClass = "DRIVER"
	ClassGuardrail     IndicatorClass = "GUARDRAIL"
	ProofFoundation    ProofChoice = "FOUNDATION"
	ProofCoherence     ProofChoice = "COHERENCE"
	ProofRegression    ProofChoice = "REGRESSION"
)

const (
	SourceUserJourney SourceName = "USER_JOURNEY"
	SourceConformance SourceName = "TOOLCHAIN_CONFORMANCE"
	SourceLSP         SourceName = "TOOLCHAIN_LSP"
	SourceRelease     SourceName = "CROSS_PLATFORM_RELEASE"
	SourceReadiness   SourceName = "LANGUAGE_READINESS"
	SourceNone        SourceName = "NONE"
)

const (
	EvidenceJourney       EvidenceKind = "JOURNEY"
	EvidenceIndicator     EvidenceKind = "INDICATOR"
	EvidenceSurface       EvidenceKind = "CONFORMANCE_SURFACE"
	EvidenceLSPCounter    EvidenceKind = "LSP_COUNTER"
	EvidenceConformance   EvidenceKind = "CONFORMANCE_COUNTER"
	EvidenceRelease       EvidenceKind = "RELEASE_COUNTER"
	EvidenceReadiness     EvidenceKind = "READINESS_OBLIGATION"
	EvidenceUnimplemented EvidenceKind = "UNIMPLEMENTED"
)

const (
	StatusSatisfied     ResultStatus = "SATISFIED"
	StatusNotImplemented ResultStatus = "NOT_IMPLEMENTED"
	StatusNotSatisfied  ResultStatus = "NOT_SATISFIED"
	StatusUnknown       ResultStatus = "UNKNOWN"
)

var audienceOrder = []Audience{AudienceUser, AudienceToolAuthor, AudienceGovernor}
var sourceOrder = []SourceName{SourceUserJourney, SourceConformance, SourceLSP, SourceRelease, SourceReadiness}
