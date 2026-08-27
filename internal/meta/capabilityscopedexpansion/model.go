package capabilityscopedexpansion

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	Schema              = "gooo/capability-scoped-expansion/v2"
	MetaOperation       = "expand-capability-scoped-meta-code"
	Producer            = "capabilityscopedexpansion.Evaluate"
	Consumer            = "capabilityscopedexpansion.verify.Judge"
	GoVersion           = "go1.27.0"
	DecisionAllow       = "ALLOW"
	DecisionDeny        = "DENY"
	DecisionUnknown     = "UNKNOWN"
	SuitePass           = "PASS"
	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	EffectNone          = "NONE"
	EffectBlock         = "BLOCK"
	CurrentEvidence     = "CURRENT_EVIDENCE"
	HistoricalFixture   = "HISTORICAL_FIXTURE"
	ClaimOpen           = "OPEN"
	ClaimDischarged     = "DISCHARGED"
	ClaimRefuted        = "REFUTED"
	PolicyDefaultDeny   = "DENY"
	PolicyExactCurrent  = "exact-current"
	PolicyDenyAll       = "deny-all"
	EffectPolicyNone    = "NONE"
	ProviderSchema      = "gooo/capability-scoped-expansion/provider/v2"
	OperationScheme     = "capability.operation"
	PolicyScheme        = "capability.policy"
	DeclarationScheme   = "capability.declare"
	CaseScheme          = "capability.case"
	FixedIndicatorTotal = 12
	FixedCaseTotal      = 8
)

type SemanticValue struct {
	Scheme string            `json:"scheme"`
	Fields map[string]string `json:"fields"`
	NodeID string            `json:"node_id"`
	Raw    string            `json:"raw"`
}

type Policy struct {
	ID                  string `json:"id"`
	DefaultDecision     string `json:"default_decision"`
	AuthorizationMode   string `json:"authorization_mode"`
	Effects             string `json:"effects"`
	PriorClaimState     string `json:"prior_claim_state"`
	SemanticValueNodeID string `json:"semantic_value_node_id"`
}

type Operation struct {
	ID              string `json:"id"`
	Stage           string `json:"stage"`
	Step            string `json:"step"`
	PriorClaimState string `json:"prior_claim_state"`
	NodeID          string `json:"node_id"`
}

type CapabilityDeclaration struct {
	ValueID         string `json:"value_id"`
	Kind            string `json:"kind"`
	Operation       string `json:"operation"`
	Target          string `json:"target"`
	Policy          string `json:"policy"`
	PriorClaimState string `json:"prior_claim_state"`
	EvidenceClass   string `json:"evidence_class"`
	NodeID          string `json:"node_id"`
}

type CapabilityValue struct {
	ValueID   string `json:"value_id"`
	Kind      string `json:"kind"`
	Operation string `json:"operation"`
	Target    string `json:"target"`
}

type CaseSpec struct {
	ID                          string            `json:"id"`
	Requests                    []CapabilityValue `json:"requests"`
	RequestedRepositoryWrites   int               `json:"requested_repository_writes"`
	RequestedMutationAuthority  bool              `json:"requested_mutation_authority"`
	RequestedPromotionAuthority bool              `json:"requested_promotion_authority"`
	PriorClaimState             string            `json:"prior_claim_state"`
	NodeID                      string            `json:"node_id"`
}

type SourceModel struct {
	SourceDigest      string                  `json:"source_digest"`
	SemanticDigest    string                  `json:"semantic_digest"`
	SemanticCanonical string                  `json:"semantic_canonical"`
	Package           string                  `json:"package"`
	Namespace         string                  `json:"namespace"`
	Policy            Policy                  `json:"policy"`
	Operations        []Operation             `json:"operations"`
	Declarations      []CapabilityDeclaration `json:"declarations"`
	Cases             []CaseSpec              `json:"cases"`
	Reconstructed     bool                    `json:"reconstructed"`
}

type ProviderObservation struct {
	Schema                   string                   `json:"schema"`
	Provider                 string                   `json:"provider"`
	SubjectSHA               string                   `json:"subject_sha"`
	FileReads                []FileReadObservation    `json:"file_reads"`
	LogicalInputs            []LogicalObservation     `json:"logical_inputs"`
	EnvironmentReads         []EnvironmentObservation `json:"environment_reads"`
	NetworkReads             []NetworkObservation     `json:"network_reads"`
	EffectAttempts           []EffectObservation      `json:"effect_attempts"`
	SandboxBefore            SnapshotObservation      `json:"sandbox_before"`
	SandboxAfter             SnapshotObservation      `json:"sandbox_after"`
	ActualRepositoryWrites   int                      `json:"actual_repository_writes"`
	ActualMutationAuthority  bool                     `json:"actual_mutation_authority"`
	ActualPromotionAuthority bool                     `json:"actual_promotion_authority"`
}

type FileReadObservation struct {
	Target        string `json:"target"`
	Path          string `json:"path"`
	ContentDigest string `json:"content_digest"`
	Observed      bool   `json:"observed"`
	EvidenceClass string `json:"evidence_class"`
}

type LogicalObservation struct {
	Target        string `json:"target"`
	Value         string `json:"value"`
	Observed      bool   `json:"observed"`
	EvidenceClass string `json:"evidence_class"`
}

type EnvironmentObservation struct {
	Target        string `json:"target"`
	Observed      bool   `json:"observed"`
	EvidenceClass string `json:"evidence_class"`
}

type NetworkObservation struct {
	Target        string `json:"target"`
	Observed      bool   `json:"observed"`
	EvidenceClass string `json:"evidence_class"`
}

type SnapshotObservation struct {
	Root    string   `json:"root"`
	Entries []string `json:"entries"`
	Digest  string   `json:"digest"`
}

type EffectObservation struct {
	Kind             string `json:"kind"`
	Target           string `json:"target"`
	Requested        bool   `json:"requested"`
	Result           string `json:"result"`
	Reason           string `json:"reason"`
	BoundaryObserved bool   `json:"boundary_observed"`
	BeforeDigest     string `json:"before_digest"`
	AfterDigest      string `json:"after_digest"`
	ActualWrites     int    `json:"actual_writes"`
	ActualMutation   bool   `json:"actual_mutation"`
	ActualPromotion  bool   `json:"actual_promotion"`
}

type Evidence struct {
	ValueID        string `json:"value_id"`
	Observed       string `json:"observed"`
	EvidenceClass  string `json:"evidence_class"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
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
	CurrentEvidenceCapabilities int  `json:"current_evidence_capabilities"`
	CurrentEvidenceDenominator  int  `json:"current_evidence_denominator"`
	RequestedRepositoryWrites   int  `json:"requested_repository_writes"`
	RequestedMutationAuthority  bool `json:"requested_mutation_authority"`
	RequestedPromotionAuthority bool `json:"requested_promotion_authority"`
	RepositoryWrites            int  `json:"repository_writes"`
	MutationAuthority           bool `json:"mutation_authority"`
	PromotionAuthority          bool `json:"promotion_authority"`
	EnforcementObservations     int  `json:"enforcement_observations"`
}

type ClaimTransition struct {
	ClaimID        string `json:"claim_id"`
	PriorState     string `json:"prior_state"`
	NextState      string `json:"next_state"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
}

type Claim struct {
	ID          string `json:"id"`
	PriorState  string `json:"prior_state"`
	Status      string `json:"status"`
	ProofChoice string `json:"proof_choice"`
	Evidence    string `json:"evidence"`
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

type Receipt struct {
	Schema             string                  `json:"schema"`
	MetaOperation      string                  `json:"meta_operation"`
	Producer           string                  `json:"producer"`
	Consumer           string                  `json:"consumer"`
	SubjectSHA         string                  `json:"subject_sha"`
	GoVersion          string                  `json:"go_version"`
	SourceDigest       string                  `json:"source_digest"`
	SemanticDigest     string                  `json:"semantic_digest"`
	CaseID             string                  `json:"case_id"`
	Stage              string                  `json:"stage"`
	Step               string                  `json:"step"`
	Decision           string                  `json:"decision"`
	Resolution         string                  `json:"resolution"`
	EnforcementEffect  string                  `json:"enforcement_effect"`
	Reason             string                  `json:"reason"`
	Policy             Policy                  `json:"policy"`
	Declarations       []CapabilityDeclaration `json:"declarations"`
	Capabilities       []CapabilityValue       `json:"capabilities"`
	Evidence           []Evidence              `json:"evidence"`
	ProviderDigest     string                  `json:"provider_digest"`
	EffectObservations []EffectObservation     `json:"effect_observations"`
	Unknown            *Unknown                `json:"unknown,omitempty"`
	Authority          Authority               `json:"authority"`
	Claims             []Claim                 `json:"claims"`
	ClaimTransitions   []ClaimTransition       `json:"claim_transitions"`
	Indicators         []Indicator             `json:"indicators"`
	RepositoryWrites   int                     `json:"repository_writes"`
	MutationAuthority  bool                    `json:"mutation_authority"`
	PromotionAuthority bool                    `json:"promotion_authority"`
	ReportDigest       string                  `json:"report_digest"`
}

type CaseResult struct {
	CaseID             string `json:"case_id"`
	ObservedDecision   string `json:"observed_decision"`
	ObservedResolution string `json:"observed_resolution"`
	ClaimStatus        string `json:"claim_status"`
	ReceiptDigest      string `json:"receipt_digest"`
	IndependentJudge   string `json:"independent_judge"`
	IndependentReason  string `json:"independent_reason"`
}

type SuiteSummary struct {
	CasesTotal                    int  `json:"cases_total"`
	CasesPassed                   int  `json:"cases_passed"`
	AllowCases                    int  `json:"allow_cases"`
	DenyCases                     int  `json:"deny_cases"`
	UnknownCases                  int  `json:"unknown_cases"`
	CapabilityRequests            int  `json:"capability_requests"`
	CapabilityAuthorized          int  `json:"capability_authorized"`
	CapabilityDenied              int  `json:"capability_denied"`
	CapabilityUnknown             int  `json:"capability_unknown"`
	CurrentEvidenceCapabilities   int  `json:"current_evidence_capabilities"`
	CurrentEvidenceDenominator    int  `json:"current_evidence_denominator"`
	HistoricalFixtureCapabilities int  `json:"historical_fixture_capabilities"`
	EnforcementObservations       int  `json:"enforcement_observations"`
	BlockedWriteAttempts          int  `json:"blocked_write_attempts"`
	BlockedMutationAttempts       int  `json:"blocked_mutation_attempts"`
	RepositoryWrites              int  `json:"repository_writes"`
	MutationAuthority             bool `json:"mutation_authority"`
	PromotionAuthority            bool `json:"promotion_authority"`
	SourceReconstructionPasses    int  `json:"source_reconstruction_passes"`
	SourceReconstructionTotal     int  `json:"source_reconstruction_total"`
	ProducerImportNumerator       int  `json:"producer_import_numerator"`
	ProducerImportDenominator     int  `json:"producer_import_denominator"`
}

type Suite struct {
	Schema             string       `json:"schema"`
	MetaOperation      string       `json:"meta_operation"`
	SubjectSHA         string       `json:"subject_sha"`
	SourceDigest       string       `json:"source_digest"`
	SemanticDigest     string       `json:"semantic_digest"`
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

type Intervention struct {
	ID                      string `json:"id"`
	Kind                    string `json:"kind"`
	BaseSourceDigest        string `json:"base_source_digest"`
	ChangedSourceDigest     string `json:"changed_source_digest"`
	BaseSemanticDigest      string `json:"base_semantic_digest"`
	ChangedSemanticDigest   string `json:"changed_semantic_digest"`
	BaseDecision            string `json:"base_decision"`
	ChangedDecision         string `json:"changed_decision"`
	BaseClaimStatus         string `json:"base_claim_status"`
	ChangedClaimStatus      string `json:"changed_claim_status"`
	DecisionPreserved       bool   `json:"decision_preserved"`
	SemanticDigestPreserved bool   `json:"semantic_digest_preserved"`
	IndependentJudge        string `json:"independent_judge"`
}

func ParseSource(source []byte) (SourceModel, error) {
	if len(strings.TrimSpace(string(source))) == 0 {
		return SourceModel{}, fmt.Errorf("Gooo source is empty")
	}
	file, diagnostics := syntax.ParseFile("capability-scoped-expansion.gooo", string(source))
	if err := diagnostics.Error(); err != nil {
		return SourceModel{}, err
	}
	if file == nil {
		return SourceModel{}, fmt.Errorf("Gooo syntax AST is unavailable")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return SourceModel{}, fmt.Errorf("lower Gooo semantic IR: %w", err)
	}
	model := SourceModel{
		SourceDigest:      digestBytes(source),
		SemanticDigest:    ir.StableHash(),
		SemanticCanonical: ir.SemanticCanonical(),
		Package:           ir.Package,
		Namespace:         ir.Namespace.String(),
		Reconstructed:     true,
	}
	values := make([]SemanticValue, 0)
	for _, node := range ir.Graph.Nodes() {
		if node.ValueProgram == "" {
			continue
		}
		value, err := parseSemanticValue(node.ValueProgram)
		if err != nil {
			return SourceModel{}, fmt.Errorf("node %s value program: %w", node.ID, err)
		}
		value.NodeID = node.ID.String()
		values = append(values, value)
	}
	if err := model.fromValues(values); err != nil {
		return SourceModel{}, err
	}
	return model, nil
}

func ValidateShape(source []byte) error {
	_, err := ParseSource(source)
	return err
}

func parseSemanticValue(raw string) (SemanticValue, error) {
	parts := strings.Split(raw, "|")
	if len(parts) < 2 || strings.TrimSpace(parts[0]) == "" {
		return SemanticValue{}, fmt.Errorf("semantic value must contain a scheme and fields")
	}
	value := SemanticValue{Scheme: parts[0], Fields: make(map[string]string), Raw: raw}
	for _, part := range parts[1:] {
		key, field, ok := strings.Cut(part, "=")
		if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(field) == "" {
			return SemanticValue{}, fmt.Errorf("malformed semantic value field %q", part)
		}
		if _, exists := value.Fields[key]; exists {
			return SemanticValue{}, fmt.Errorf("duplicate semantic value field %q", key)
		}
		value.Fields[key] = field
	}
	return value, nil
}

func (model *SourceModel) fromValues(values []SemanticValue) error {
	for _, value := range values {
		switch value.Scheme {
		case PolicyScheme:
			if model.Policy.ID != "" {
				return fmt.Errorf("multiple capability policies")
			}
			model.Policy = Policy{ID: value.Fields["id"], DefaultDecision: value.Fields["default"], AuthorizationMode: value.Fields["authorization"], Effects: value.Fields["effects"], PriorClaimState: value.Fields["prior"], SemanticValueNodeID: value.NodeID}
		case OperationScheme:
			model.Operations = append(model.Operations, Operation{ID: value.Fields["id"], Stage: value.Fields["stage"], Step: value.Fields["step"], PriorClaimState: value.Fields["prior"], NodeID: value.NodeID})
		case DeclarationScheme:
			model.Declarations = append(model.Declarations, CapabilityDeclaration{ValueID: value.Fields["id"], Kind: value.Fields["kind"], Operation: value.Fields["operation"], Target: value.Fields["target"], Policy: value.Fields["policy"], PriorClaimState: value.Fields["prior"], EvidenceClass: value.Fields["evidence"], NodeID: value.NodeID})
		}
	}
	for _, value := range values {
		if value.Scheme != CaseScheme {
			continue
		}
		caseSpec, err := parseCaseValue(value, model.Declarations)
		if err != nil {
			return err
		}
		model.Cases = append(model.Cases, caseSpec)
	}
	if model.Policy.ID == "" || model.Policy.DefaultDecision != PolicyDefaultDeny || model.Policy.PriorClaimState != ClaimOpen || model.Policy.Effects != EffectPolicyNone {
		return fmt.Errorf("semantic capability policy is missing or not default-deny")
	}
	if model.Policy.AuthorizationMode != PolicyExactCurrent && model.Policy.AuthorizationMode != PolicyDenyAll {
		return fmt.Errorf("semantic capability policy has unknown authorization mode")
	}
	if len(model.Operations) < 3 || len(model.Declarations) != 4 || len(model.Cases) != FixedCaseTotal {
		return fmt.Errorf("semantic capability denominator is incomplete: operations=%d declarations=%d cases=%d", len(model.Operations), len(model.Declarations), len(model.Cases))
	}
	for _, operation := range model.Operations {
		if operation.ID == "" || operation.Stage == "" || operation.Step == "" || operation.PriorClaimState != ClaimOpen {
			return fmt.Errorf("semantic operation is incomplete: %s", operation.ID)
		}
	}
	for _, declaration := range model.Declarations {
		if declaration.ValueID == "" || declaration.Kind == "" || declaration.Operation == "" || declaration.Target == "" || declaration.Policy != model.Policy.ID || declaration.PriorClaimState != ClaimOpen || declaration.EvidenceClass == "" {
			return fmt.Errorf("semantic capability declaration is incomplete: %s", declaration.ValueID)
		}
	}
	sort.Slice(model.Operations, func(i, j int) bool { return model.Operations[i].ID < model.Operations[j].ID })
	sort.Slice(model.Declarations, func(i, j int) bool { return model.Declarations[i].ValueID < model.Declarations[j].ValueID })
	sort.Slice(model.Cases, func(i, j int) bool { return model.Cases[i].ID < model.Cases[j].ID })
	return nil
}

func parseCaseValue(value SemanticValue, declarations []CapabilityDeclaration) (CaseSpec, error) {
	caseSpec := CaseSpec{ID: value.Fields["id"], PriorClaimState: value.Fields["prior"], NodeID: value.NodeID}
	if caseSpec.ID == "" || caseSpec.PriorClaimState != ClaimOpen {
		return CaseSpec{}, fmt.Errorf("semantic case lacks id or OPEN prior state")
	}
	byID := make(map[string]CapabilityDeclaration, len(declarations))
	for _, declaration := range declarations {
		byID[declaration.ValueID] = declaration
	}
	for _, request := range strings.Split(value.Fields["requests"], ",") {
		if strings.TrimSpace(request) == "" {
			return CaseSpec{}, fmt.Errorf("case %s has empty capability request", caseSpec.ID)
		}
		valueID, target, ok := strings.Cut(request, "@")
		if !ok || valueID == "" || target == "" {
			return CaseSpec{}, fmt.Errorf("case %s request %q is not value@target", caseSpec.ID, request)
		}
		declaration, exists := byID[valueID]
		if !exists {
			return CaseSpec{}, fmt.Errorf("case %s references unknown capability %s", caseSpec.ID, valueID)
		}
		caseSpec.Requests = append(caseSpec.Requests, CapabilityValue{ValueID: valueID, Kind: declaration.Kind, Operation: declaration.Operation, Target: target})
	}
	if len(caseSpec.Requests) == 0 {
		return CaseSpec{}, fmt.Errorf("case %s has no requests", caseSpec.ID)
	}
	if writes := value.Fields["writes"]; writes != "" {
		if _, err := fmt.Sscanf(writes, "%d", &caseSpec.RequestedRepositoryWrites); err != nil {
			return CaseSpec{}, fmt.Errorf("case %s writes is not an integer", caseSpec.ID)
		}
	}
	caseSpec.RequestedMutationAuthority = value.Fields["mutation"] == "true"
	caseSpec.RequestedPromotionAuthority = value.Fields["promotion"] == "true"
	return caseSpec, nil
}

func declarationFor(model SourceModel, valueID string) (CapabilityDeclaration, bool) {
	for _, declaration := range model.Declarations {
		if declaration.ValueID == valueID {
			return declaration, true
		}
	}
	return CapabilityDeclaration{}, false
}

func capabilityMatches(value CapabilityValue, declaration CapabilityDeclaration) bool {
	return value.Kind == declaration.Kind && value.Operation == declaration.Operation && value.Target == declaration.Target
}

func firstOperation(model SourceModel, id string) Operation {
	for _, operation := range model.Operations {
		if operation.ID == id {
			return operation
		}
	}
	if len(model.Operations) > 0 {
		return model.Operations[0]
	}
	return Operation{Stage: "UNKNOWN", Step: "UNKNOWN", PriorClaimState: ClaimOpen}
}

func authorizationOperation(model SourceModel) Operation {
	for _, operation := range model.Operations {
		if operation.ID == "authorize-before-expand" {
			return operation
		}
	}
	return firstOperation(model, "")
}
