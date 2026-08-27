package valuecatalog

import "github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"

const (
	ReportSchema             = "gooo.language.operation-catalog/v2"
	DecisionBaselineObserved = "CATALOG_BASELINE_OBSERVED"
	DecisionExtensionProven  = "SOURCE_ONLY_EXTENSION_PROVEN"
	DecisionFailClosed       = "FAIL_CLOSED"
	ReasonBaselineObserved   = "SOURCE_ONLY_EXTENSION_NOT_PRESENT"
	ReasonExtensionExact     = "SOURCE_ONLY_EXTENSION_EXACT"
	ReasonSourceReadFailed   = "CATALOG_SOURCE_READ_FAILED"
	ReasonObservationFailed  = "CATALOG_OBSERVATION_FAILED"
	ResolutionCoreValue      = "CORE_IR_ACTIVITY_VALUE_PROGRAM"
	ResolutionSyntaxOnly     = "SYNTAX_ONLY"
	BaselineActivity         = "IncrementOne"
	ExtensionActivity        = "IncrementTwo"
	CatalogIndicatorCount    = 22
	OperationSpecAxisTotal   = 9
	CatalogMetricID          = "gooo.metric.language-operation-catalog.extension.v1"
	OperationSpecMetricID    = "gooo.metric.language-operation-spec.os9.v1"
)

type Report struct {
	Schema               string        `json:"schema"`
	Decision             string        `json:"decision"`
	Reason               string        `json:"reason"`
	Resolution           string        `json:"resolution"`
	HeadSHA              string        `json:"head_sha"`
	SourcePath           string        `json:"source_path"`
	SourceDigest         string        `json:"source_digest"`
	BeforeSourceDigest   string        `json:"before_source_digest"`
	Diagnostic           string        `json:"diagnostic"`
	SourceLines          int           `json:"source_lines"`
	ActivitiesObserved   int           `json:"activities_observed"`
	CoreIRFingerprint    string        `json:"core_ir_fingerprint"`
	BaselineCoreProgram  string        `json:"baseline_core_program"`
	ExtensionCoreProgram string        `json:"extension_core_program"`
	OperationSpecs       []valueexecution.OperationSpec `json:"operation_specs"`
	OperationSpecMetrics OperationSpecMetrics            `json:"operation_spec_metrics"`
	Claims               []Claim                         `json:"claims"`
	ProcessCoordinate    ProcessCoordinate               `json:"coordinate"`
	Baseline             ProgramResult `json:"baseline"`
	Extension            ProgramResult `json:"extension"`
	Improvement          Improvement   `json:"improvement"`
	Summary              Summary       `json:"summary"`
	Indicators           []Indicator   `json:"indicators"`
	Views                []View        `json:"views"`
	Proofs               []Proof       `json:"proofs"`
	NonClaims            []string      `json:"non_claims"`
	Authority            Authority     `json:"authority"`
	Digest               string        `json:"digest"`
}
