package capabilityscopedexpansion

import (
	"fmt"
	"strings"
)

const (
	Schema              = "gooo/capability-scoped-expansion/v1"
	MetaOperation       = "expand-capability-scoped-meta-code"
	Producer            = "capabilityscopedexpansion.Evaluate"
	Consumer            = "ci-capability-expansion-gate"
	GoVersion           = "go1.27.0"
	StageExpand         = "expand"
	StepAuthorize       = "authorize-before-expand"
	DecisionAllow       = "ALLOW"
	DecisionDeny        = "DENY"
	DecisionUnknown     = "UNKNOWN"
	ResolutionExact     = "EXACT"
	ResolutionReject    = "INVARIANT_ONLY"
	ResolutionUnknown   = "UNKNOWN"
	EffectNone          = "NONE"
	EffectBlock         = "BLOCK"
	FixedIndicatorTotal = 12
	FixedCaseTotal      = 8
)

const (
	KindFile          = "file"
	KindTime          = "time"
	KindEnvironment   = "environment"
	KindNetwork       = "network"
	OperationRead     = "read"
	FileTarget        = "source"
	TimeTarget        = "logical-clock"
	EnvironmentTarget = "GOOO_EXPANSION_PROFILE"
	NetworkTarget     = "https://example.invalid/gooo/pinned-schema"
)

type CapabilityDeclaration struct {
	Kind      string `json:"kind"`
	Operation string `json:"operation"`
	Target    string `json:"target"`
	Declared  bool   `json:"declared"`
}

type CapabilityValue struct {
	ValueID   string `json:"value_id"`
	Kind      string `json:"kind"`
	Operation string `json:"operation"`
	Target    string `json:"target"`
}

type Evidence struct {
	ValueID        string `json:"value_id"`
	Observed       string `json:"observed"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Request struct {
	CaseID                      string            `json:"case_id"`
	SubjectSHA                  string            `json:"subject_sha"`
	Stage                       string            `json:"stage"`
	Step                        string            `json:"step"`
	Toolchain                   string            `json:"toolchain"`
	Capabilities                []CapabilityValue `json:"capabilities"`
	Evidence                    []Evidence        `json:"evidence"`
	RequestedRepositoryWrites   int               `json:"requested_repository_writes"`
	RequestedMutationAuthority  bool              `json:"requested_mutation_authority"`
	RequestedPromotionAuthority bool              `json:"requested_promotion_authority"`
}

type Unknown struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type Authority struct {
	CapabilitiesRequested       int  `json:"capabilities_requested"`
	CapabilitiesDeclared        int  `json:"capabilities_declared"`
	CapabilitiesAuthorized      int  `json:"capabilities_authorized"`
	CapabilitiesDenied          int  `json:"capabilities_denied"`
	CapabilitiesUnknown         int  `json:"capabilities_unknown"`
	RequestedRepositoryWrites   int  `json:"requested_repository_writes"`
	RequestedMutationAuthority  bool `json:"requested_mutation_authority"`
	RequestedPromotionAuthority bool `json:"requested_promotion_authority"`
	RepositoryWrites            int  `json:"repository_writes"`
	MutationAuthority           bool `json:"mutation_authority"`
	PromotionAuthority          bool `json:"promotion_authority"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	Status        string `json:"status"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Observed      int    `json:"observed"`
	Target        int    `json:"target"`
}

type Claim struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	ProofChoice string `json:"proof_choice"`
	Evidence    string `json:"evidence"`
}

type Receipt struct {
	Schema             string                  `json:"schema"`
	MetaOperation      string                  `json:"meta_operation"`
	Producer           string                  `json:"producer"`
	Consumer           string                  `json:"consumer"`
	SubjectSHA         string                  `json:"subject_sha"`
	GoVersion          string                  `json:"go_version"`
	SourceDigest       string                  `json:"source_digest"`
	CaseID             string                  `json:"case_id"`
	Stage              string                  `json:"stage"`
	Step               string                  `json:"step"`
	Decision           string                  `json:"decision"`
	Resolution         string                  `json:"resolution"`
	EnforcementEffect  string                  `json:"enforcement_effect"`
	Reason             string                  `json:"reason"`
	Declarations       []CapabilityDeclaration `json:"declarations"`
	Capabilities       []CapabilityValue       `json:"capabilities"`
	Evidence           []Evidence              `json:"evidence"`
	Unknown            *Unknown                `json:"unknown,omitempty"`
	Authority          Authority               `json:"authority"`
	Claims             []Claim                 `json:"claims"`
	Indicators         []Indicator             `json:"indicators"`
	RepositoryWrites   int                     `json:"repository_writes"`
	MutationAuthority  bool                    `json:"mutation_authority"`
	PromotionAuthority bool                    `json:"promotion_authority"`
	ReportDigest       string                  `json:"report_digest"`
}

type Case struct {
	ID                 string  `json:"id"`
	Request            Request `json:"request"`
	ExpectedDecision   string  `json:"expected_decision"`
	ExpectedResolution string  `json:"expected_resolution"`
}

type CaseResult struct {
	CaseID             string `json:"case_id"`
	ExpectedDecision   string `json:"expected_decision"`
	ObservedDecision   string `json:"observed_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ObservedResolution string `json:"observed_resolution"`
	ReceiptDigest      string `json:"receipt_digest"`
	IndependentJudge   string `json:"independent_judge"`
	IndependentReason  string `json:"independent_reason"`
}

type SuiteSummary struct {
	CasesTotal              int  `json:"cases_total"`
	CasesPassed             int  `json:"cases_passed"`
	AllowCases              int  `json:"allow_cases"`
	DenyCases               int  `json:"deny_cases"`
	UnknownCases            int  `json:"unknown_cases"`
	CapabilityRequests      int  `json:"capability_requests"`
	CapabilityAuthorized    int  `json:"capability_authorized"`
	CapabilityDenied        int  `json:"capability_denied"`
	CapabilityUnknown       int  `json:"capability_unknown"`
	BlockedWriteAttempts    int  `json:"blocked_write_attempts"`
	BlockedMutationAttempts int  `json:"blocked_mutation_attempts"`
	RepositoryWrites        int  `json:"repository_writes"`
	MutationAuthority       bool `json:"mutation_authority"`
	PromotionAuthority      bool `json:"promotion_authority"`
}

type Suite struct {
	Schema             string       `json:"schema"`
	MetaOperation      string       `json:"meta_operation"`
	SubjectSHA         string       `json:"subject_sha"`
	SourceDigest       string       `json:"source_digest"`
	Decision           string       `json:"decision"`
	Resolution         string       `json:"resolution"`
	Summary            SuiteSummary `json:"summary"`
	Cases              []CaseResult `json:"cases"`
	IndependentJudge   string       `json:"independent_judge"`
	RepositoryWrites   int          `json:"repository_writes"`
	MutationAuthority  bool         `json:"mutation_authority"`
	PromotionAuthority bool         `json:"promotion_authority"`
	SuiteDigest        string       `json:"suite_digest"`
}

func ValidateShape(source []byte) error {
	if len(strings.TrimSpace(string(source))) == 0 {
		return fmt.Errorf("Gooo source is empty")
	}
	for _, marker := range []string{
		"package capabilityscopedexpansion",
		"namespace capabilityscopedexpansion",
		"activity DeclareFileReadCapability",
		"activity DeclareLogicalTimeCapability",
		"activity DeclareEnvironmentReadCapability",
		"activity DeclarePinnedNetworkCapability",
		"activity BindCapabilityEvidence",
		"activity ExpandWithCapabilityEvidence",
		"activity DenyByDefault",
	} {
		if !strings.Contains(string(source), marker) {
			return fmt.Errorf("Gooo source is missing %q", marker)
		}
	}
	return nil
}
