package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	validSchema         = "gooo/capability-scoped-expansion/v2"
	validSuiteSchema    = "gooo/capability-scoped-expansion-suite/v2"
	validProviderSchema = "gooo/capability-scoped-expansion/provider/v2"
	validMetaOperation  = "expand-capability-scoped-meta-code"
	validProducer       = "capabilityscopedexpansion.Evaluate"
	validConsumer       = "capabilityscopedexpansion.verify.Judge"
	validGoVersion      = "go1.27.0"
	claimOpen           = "OPEN"
	claimDischarged     = "DISCHARGED"
	claimRefuted        = "REFUTED"
	decisionAllow       = "ALLOW"
	decisionDeny        = "DENY"
	decisionUnknown     = "UNKNOWN"
	resolutionExact     = "EXACT"
	resolutionLower     = "LOWER_RESOLUTION"
	operationScheme     = "capability.operation"
	policyScheme        = "capability.policy"
	declarationScheme   = "capability.declare"
	caseScheme          = "capability.case"
	currentEvidence     = "CURRENT_EVIDENCE"
	historicalFixture   = "HISTORICAL_FIXTURE"
	fixedIndicatorTotal = 12
)

type Verdict struct {
	Status               string `json:"status"`
	Decision             string `json:"decision"`
	Resolution           string `json:"resolution"`
	Reason               string `json:"reason"`
	ReceiptDigest        string `json:"receipt_digest"`
	SourceReconstruction string `json:"source_reconstruction"`
	ProducerImports      int    `json:"producer_imports"`
}

type rawReceipt struct {
	Schema             string               `json:"schema"`
	MetaOperation      string               `json:"meta_operation"`
	Producer           string               `json:"producer"`
	Consumer           string               `json:"consumer"`
	SubjectSHA         string               `json:"subject_sha"`
	GoVersion          string               `json:"go_version"`
	SourceDigest       string               `json:"source_digest"`
	SemanticDigest     string               `json:"semantic_digest"`
	CaseID             string               `json:"case_id"`
	Stage              string               `json:"stage"`
	Step               string               `json:"step"`
	Decision           string               `json:"decision"`
	Resolution         string               `json:"resolution"`
	EnforcementEffect  string               `json:"enforcement_effect"`
	Reason             string               `json:"reason"`
	Policy             rawPolicy            `json:"policy"`
	Declarations       []rawDeclaration     `json:"declarations"`
	Capabilities       []rawCapability      `json:"capabilities"`
	Evidence           []rawEvidence        `json:"evidence"`
	ProviderDigest     string               `json:"provider_digest"`
	EffectObservations []rawEffect          `json:"effect_observations"`
	Unknown            *rawUnknown          `json:"unknown,omitempty"`
	Authority          rawAuthority         `json:"authority"`
	Claims             []rawClaim           `json:"claims"`
	ClaimTransitions   []rawClaimTransition `json:"claim_transitions"`
	Indicators         []rawIndicator       `json:"indicators"`
	RepositoryWrites   int                  `json:"repository_writes"`
	MutationAuthority  bool                 `json:"mutation_authority"`
	PromotionAuthority bool                 `json:"promotion_authority"`
	ReportDigest       string               `json:"report_digest"`
}

type rawPolicy struct {
	ID                string `json:"id"`
	DefaultDecision   string `json:"default_decision"`
	AuthorizationMode string `json:"authorization_mode"`
	Effects           string `json:"effects"`
	PriorClaimState   string `json:"prior_claim_state"`
	NodeID            string `json:"semantic_value_node_id"`
}
type rawDeclaration struct {
	ValueID         string `json:"value_id"`
	Kind            string `json:"kind"`
	Operation       string `json:"operation"`
	Target          string `json:"target"`
	Policy          string `json:"policy"`
	PriorClaimState string `json:"prior_claim_state"`
	EvidenceClass   string `json:"evidence_class"`
	NodeID          string `json:"node_id"`
}
type rawCapability struct{ ValueID, Kind, Operation, Target string }

func (c *rawCapability) UnmarshalJSON(raw []byte) error {
	var value struct {
		ValueID   string `json:"value_id"`
		Kind      string `json:"kind"`
		Operation string `json:"operation"`
		Target    string `json:"target"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	c.ValueID, c.Kind, c.Operation, c.Target = value.ValueID, value.Kind, value.Operation, value.Target
	return nil
}

type rawEvidence struct{ ValueID, Observed, EvidenceClass, EvidenceDigest, Provenance string }

func (e *rawEvidence) UnmarshalJSON(raw []byte) error {
	var value struct {
		ValueID        string `json:"value_id"`
		Observed       string `json:"observed"`
		EvidenceClass  string `json:"evidence_class"`
		EvidenceDigest string `json:"evidence_digest"`
		Provenance     string `json:"provenance"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	e.ValueID, e.Observed, e.EvidenceClass, e.EvidenceDigest, e.Provenance = value.ValueID, value.Observed, value.EvidenceClass, value.EvidenceDigest, value.Provenance
	return nil
}

type rawUnknown struct{ Stage, Step, Reason string }

func (u *rawUnknown) UnmarshalJSON(raw []byte) error {
	var value struct {
		Stage  string `json:"stage"`
		Step   string `json:"step"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	u.Stage, u.Step, u.Reason = value.Stage, value.Step, value.Reason
	return nil
}

type rawAuthority struct {
	CapabilitiesRequested, CapabilitiesDeclared, CapabilitiesAuthorized, CapabilitiesDenied, CapabilitiesUnknown                  int
	CurrentEvidenceCapabilities, CurrentEvidenceDenominator, RequestedRepositoryWrites, RepositoryWrites, EnforcementObservations int
	RequestedMutationAuthority, RequestedPromotionAuthority, MutationAuthority, PromotionAuthority                                bool
}

func (a *rawAuthority) UnmarshalJSON(raw []byte) error {
	var value struct {
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
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	a.CapabilitiesRequested, a.CapabilitiesDeclared, a.CapabilitiesAuthorized, a.CapabilitiesDenied, a.CapabilitiesUnknown = value.CapabilitiesRequested, value.CapabilitiesDeclared, value.CapabilitiesAuthorized, value.CapabilitiesDenied, value.CapabilitiesUnknown
	a.CurrentEvidenceCapabilities, a.CurrentEvidenceDenominator, a.RequestedRepositoryWrites, a.RepositoryWrites, a.EnforcementObservations = value.CurrentEvidenceCapabilities, value.CurrentEvidenceDenominator, value.RequestedRepositoryWrites, value.RepositoryWrites, value.EnforcementObservations
	a.RequestedMutationAuthority, a.RequestedPromotionAuthority, a.MutationAuthority, a.PromotionAuthority = value.RequestedMutationAuthority, value.RequestedPromotionAuthority, value.MutationAuthority, value.PromotionAuthority
	return nil
}

type rawEffect struct {
	Kind, Target, Result, Reason, BeforeDigest, AfterDigest      string
	Requested, BoundaryObserved, ActualMutation, ActualPromotion bool
	ActualWrites                                                 int
}

func (e *rawEffect) UnmarshalJSON(raw []byte) error {
	var value struct {
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
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	e.Kind, e.Target, e.Result, e.Reason, e.BeforeDigest, e.AfterDigest = value.Kind, value.Target, value.Result, value.Reason, value.BeforeDigest, value.AfterDigest
	e.Requested, e.BoundaryObserved, e.ActualMutation, e.ActualPromotion, e.ActualWrites = value.Requested, value.BoundaryObserved, value.ActualMutation, value.ActualPromotion, value.ActualWrites
	return nil
}

type rawClaim struct{ ID, PriorState, Status, ProofChoice, Evidence string }

func (c *rawClaim) UnmarshalJSON(raw []byte) error {
	var value struct {
		ID          string `json:"id"`
		PriorState  string `json:"prior_state"`
		Status      string `json:"status"`
		ProofChoice string `json:"proof_choice"`
		Evidence    string `json:"evidence"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	c.ID, c.PriorState, c.Status, c.ProofChoice, c.Evidence = value.ID, value.PriorState, value.Status, value.ProofChoice, value.Evidence
	return nil
}

type rawClaimTransition struct{ ClaimID, PriorState, NextState, Stage, Step, Reason, EvidenceDigest, Provenance string }

func (c *rawClaimTransition) UnmarshalJSON(raw []byte) error {
	var value struct {
		ClaimID        string `json:"claim_id"`
		PriorState     string `json:"prior_state"`
		NextState      string `json:"next_state"`
		Stage          string `json:"stage"`
		Step           string `json:"step"`
		Reason         string `json:"reason"`
		EvidenceDigest string `json:"evidence_digest"`
		Provenance     string `json:"provenance"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	c.ClaimID, c.PriorState, c.NextState, c.Stage, c.Step, c.Reason, c.EvidenceDigest, c.Provenance = value.ClaimID, value.PriorState, value.NextState, value.Stage, value.Step, value.Reason, value.EvidenceDigest, value.Provenance
	return nil
}

type rawIndicator struct {
	ID, Class, Status, Producer, Consumer, MetaOperation, ProofChoice string
	Observed, Target                                                  int
}

func (i *rawIndicator) UnmarshalJSON(raw []byte) error {
	var value struct {
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
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	i.ID, i.Class, i.Status, i.Producer, i.Consumer, i.MetaOperation, i.ProofChoice, i.Observed, i.Target = value.ID, value.Class, value.Status, value.Producer, value.Consumer, value.MetaOperation, value.ProofChoice, value.Observed, value.Target
	return nil
}

type providerObservation struct {
	Schema                   string            `json:"schema"`
	Provider                 string            `json:"provider"`
	SubjectSHA               string            `json:"subject_sha"`
	FileReads                []fileRead        `json:"file_reads"`
	LogicalInputs            []logicalInput    `json:"logical_inputs"`
	EnvironmentReads         []environmentRead `json:"environment_reads"`
	NetworkReads             []networkRead     `json:"network_reads"`
	EffectAttempts           []rawEffect       `json:"effect_attempts"`
	SandboxBefore            snapshot          `json:"sandbox_before"`
	SandboxAfter             snapshot          `json:"sandbox_after"`
	ActualRepositoryWrites   int               `json:"actual_repository_writes"`
	ActualMutationAuthority  bool              `json:"actual_mutation_authority"`
	ActualPromotionAuthority bool              `json:"actual_promotion_authority"`
}
type fileRead struct {
	Target, Path, ContentDigest, EvidenceClass string
	Observed                                   bool
}

func (f *fileRead) UnmarshalJSON(raw []byte) error {
	var value struct {
		Target        string `json:"target"`
		Path          string `json:"path"`
		ContentDigest string `json:"content_digest"`
		Observed      bool   `json:"observed"`
		EvidenceClass string `json:"evidence_class"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	f.Target, f.Path, f.ContentDigest, f.EvidenceClass, f.Observed = value.Target, value.Path, value.ContentDigest, value.EvidenceClass, value.Observed
	return nil
}

type logicalInput struct {
	Target, Value, EvidenceClass string
	Observed                     bool
}

func (l *logicalInput) UnmarshalJSON(raw []byte) error {
	var value struct {
		Target        string `json:"target"`
		Value         string `json:"value"`
		Observed      bool   `json:"observed"`
		EvidenceClass string `json:"evidence_class"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	l.Target, l.Value, l.EvidenceClass, l.Observed = value.Target, value.Value, value.EvidenceClass, value.Observed
	return nil
}

type environmentRead struct {
	Target, EvidenceClass string
	Observed              bool
}
type networkRead struct {
	Target, EvidenceClass string
	Observed              bool
}
type snapshot struct {
	Root, Digest string
	Entries      []string
}

func (s *snapshot) UnmarshalJSON(raw []byte) error {
	var value struct {
		Root    string   `json:"root"`
		Entries []string `json:"entries"`
		Digest  string   `json:"digest"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	s.Root, s.Entries, s.Digest = value.Root, value.Entries, value.Digest
	return nil
}

type sourceModel struct {
	SourceDigest, SemanticDigest, Package, Namespace string
	Policy                                           policy
	Operations                                       []operation
	Declarations                                     []declaration
	Cases                                            []caseSpec
}
type policy struct{ ID, DefaultDecision, AuthorizationMode, Effects, Prior, NodeID string }
type operation struct{ ID, Stage, Step, Prior string }
type declaration struct {
	ValueID, Kind, Operation, Target, Policy, Prior, Evidence, NodeID string
}
type capability struct{ ValueID, Kind, Operation, Target string }
type caseSpec struct {
	ID                  string
	Requests            []capability
	Writes              int
	Mutation, Promotion bool
	Prior               string
}

// Judge is an independent consumer: it imports syntax/bidir only, parses and
// lowers the raw Gooo source, reconstructs the policy/cases, and then consumes
// raw provider observations. It does not import the producer package.
func Judge(source, providerRaw, receiptRaw []byte) Verdict {
	model, err := reconstructSource(source)
	if err != nil {
		return invalid("source reconstruction: " + err.Error())
	}
	provider, err := decodeProvider(providerRaw)
	if err != nil {
		return invalid("provider reconstruction: " + err.Error())
	}
	var observed rawReceipt
	if err := json.Unmarshal(receiptRaw, &observed); err != nil {
		return invalid("receipt is not JSON")
	}
	if err := validateReceiptShape(source, model, provider, providerRaw, observed); err != nil {
		return invalid(err.Error())
	}
	item, ok := findCase(model.Cases, observed.CaseID)
	if !ok {
		return invalid("receipt case is not a semantic source case")
	}
	if !reflect.DeepEqual(observed.Capabilities, rawCapabilities(item)) {
		return invalid("receipt capabilities are not reconstructed from source values")
	}
	if !reflect.DeepEqual(observed.Evidence, rawEvidenceFor(model, provider, item)) {
		return invalid("receipt evidence is not reconstructed from provider observations")
	}
	wantDecision, wantResolution, wantReason, wantUnknown := expected(model, provider, item)
	if observed.Decision != wantDecision || observed.Resolution != wantResolution || observed.Reason != wantReason {
		return invalid(fmt.Sprintf("decision reconstruction disagrees: want %s/%s/%s", wantDecision, wantResolution, wantReason))
	}
	wantEffect := "BLOCK"
	if wantDecision == decisionAllow {
		wantEffect = "NONE"
	}
	if observed.EnforcementEffect != wantEffect {
		return invalid("receipt enforcement effect disagrees with reconstructed decision")
	}
	if err := validateUnknown(observed, wantUnknown); err != nil {
		return invalid(err.Error())
	}
	if err := validateAuthority(provider, model, item, observed); err != nil {
		return invalid(err.Error())
	}
	if err := validateClaims(model, provider, item, observed, wantUnknown); err != nil {
		return invalid(err.Error())
	}
	if err := validateIndicators(observed.Indicators); err != nil {
		return invalid(err.Error())
	}
	if observed.ReportDigest == "" || canonicalDigest(receiptRaw) != observed.ReportDigest {
		return invalid("receipt digest does not bind raw evidence")
	}
	return Verdict{Status: "PASS", Decision: observed.Decision, Resolution: observed.Resolution, Reason: observed.Reason, ReceiptDigest: observed.ReportDigest, SourceReconstruction: "PASS", ProducerImports: 0}
}

func reconstructSource(source []byte) (sourceModel, error) {
	file, diagnostics := syntax.ParseFile("capability-scoped-expansion.gooo", string(source))
	if err := diagnostics.Error(); err != nil {
		return sourceModel{}, err
	}
	if file == nil {
		return sourceModel{}, fmt.Errorf("syntax AST unavailable")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, err
	}
	model := sourceModel{SourceDigest: digestBytes(source), SemanticDigest: ir.StableHash(), Package: ir.Package, Namespace: ir.Namespace.String()}
	type value struct {
		scheme string
		fields map[string]string
		node   string
	}
	values := make([]value, 0)
	for _, node := range ir.Graph.Nodes() {
		if node.ValueProgram == "" {
			continue
		}
		parsed, err := parseValue(node.ValueProgram)
		if err != nil {
			return sourceModel{}, err
		}
		values = append(values, value{scheme: parsed.scheme, fields: parsed.fields, node: node.ID.String()})
	}
	for _, value := range values {
		switch value.scheme {
		case policyScheme:
			model.Policy = policy{ID: value.fields["id"], DefaultDecision: value.fields["default"], AuthorizationMode: value.fields["authorization"], Effects: value.fields["effects"], Prior: value.fields["prior"], NodeID: value.node}
		case operationScheme:
			model.Operations = append(model.Operations, operation{ID: value.fields["id"], Stage: value.fields["stage"], Step: value.fields["step"], Prior: value.fields["prior"]})
		case declarationScheme:
			model.Declarations = append(model.Declarations, declaration{ValueID: value.fields["id"], Kind: value.fields["kind"], Operation: value.fields["operation"], Target: value.fields["target"], Policy: value.fields["policy"], Prior: value.fields["prior"], Evidence: value.fields["evidence"], NodeID: value.node})
		}
	}
	for _, value := range values {
		if value.scheme != caseScheme {
			continue
		}
		item, err := parseCase(value.fields, model.Declarations)
		if err != nil {
			return sourceModel{}, err
		}
		model.Cases = append(model.Cases, item)
	}
	if model.Package != "capabilityscopedexpansion" || model.Namespace != "capabilityscopedexpansion" || model.Policy.ID == "" || model.Policy.DefaultDecision != "DENY" || model.Policy.Prior != claimOpen || model.Policy.Effects != "NONE" || len(model.Declarations) != 4 || len(model.Cases) != 8 || len(model.Operations) < 3 {
		return sourceModel{}, fmt.Errorf("semantic source denominator is incomplete")
	}
	return model, nil
}

type parsedValue struct {
	scheme string
	fields map[string]string
}

func parseValue(raw string) (parsedValue, error) {
	parts := strings.Split(raw, "|")
	if len(parts) < 2 {
		return parsedValue{}, fmt.Errorf("semantic value has no fields")
	}
	value := parsedValue{scheme: parts[0], fields: make(map[string]string)}
	for _, part := range parts[1:] {
		key, field, ok := strings.Cut(part, "=")
		if !ok || key == "" || field == "" {
			return parsedValue{}, fmt.Errorf("malformed semantic value field")
		}
		if _, exists := value.fields[key]; exists {
			return parsedValue{}, fmt.Errorf("duplicate semantic value field")
		}
		value.fields[key] = field
	}
	return value, nil
}
func parseCase(fields map[string]string, declarations []declaration) (caseSpec, error) {
	item := caseSpec{ID: fields["id"], Prior: fields["prior"], Mutation: fields["mutation"] == "true", Promotion: fields["promotion"] == "true"}
	byID := make(map[string]declaration, len(declarations))
	for _, declaration := range declarations {
		byID[declaration.ValueID] = declaration
	}
	for _, request := range strings.Split(fields["requests"], ",") {
		valueID, target, ok := strings.Cut(request, "@")
		declaration, exists := byID[valueID]
		if !ok || !exists {
			return caseSpec{}, fmt.Errorf("case request is not a declared value")
		}
		item.Requests = append(item.Requests, capability{ValueID: valueID, Kind: declaration.Kind, Operation: declaration.Operation, Target: target})
	}
	if fields["writes"] != "" {
		if _, err := fmt.Sscanf(fields["writes"], "%d", &item.Writes); err != nil {
			return caseSpec{}, err
		}
	}
	if item.ID == "" || item.Prior != claimOpen || len(item.Requests) == 0 {
		return caseSpec{}, fmt.Errorf("semantic case is incomplete")
	}
	return item, nil
}

func rawCapabilities(item caseSpec) []rawCapability {
	capabilities := make([]rawCapability, 0, len(item.Requests))
	for _, request := range item.Requests {
		capabilities = append(capabilities, rawCapability{ValueID: request.ValueID, Kind: request.Kind, Operation: request.Operation, Target: request.Target})
	}
	return capabilities
}

func rawEvidenceFor(model sourceModel, provider providerObservation, item caseSpec) []rawEvidence {
	evidence := make([]rawEvidence, 0, len(item.Requests))
	for _, request := range item.Requests {
		declaration, ok := findDeclaration(model.Declarations, request.ValueID)
		if !ok || declaration.Kind != request.Kind || declaration.Operation != request.Operation || declaration.Target != request.Target {
			continue
		}
		if declaration.Evidence != currentEvidence {
			continue
		}
		if declaration.Kind == "file" {
			for _, observation := range provider.FileReads {
				if observation.Target == declaration.Target && observation.Observed && observation.EvidenceClass == currentEvidence {
					digest := digestBytes([]byte(request.ValueID + "=" + observation.ContentDigest))
					evidence = append(evidence, rawEvidence{ValueID: request.ValueID, Observed: observation.ContentDigest, EvidenceClass: currentEvidence, EvidenceDigest: digest, Provenance: "provider.file.read"})
				}
			}
		}
		if declaration.Kind == "time" {
			for _, observation := range provider.LogicalInputs {
				if observation.Target == declaration.Target && observation.Observed && observation.EvidenceClass == currentEvidence {
					digest := digestBytes([]byte(request.ValueID + "=" + observation.Value))
					evidence = append(evidence, rawEvidence{ValueID: request.ValueID, Observed: observation.Value, EvidenceClass: currentEvidence, EvidenceDigest: digest, Provenance: "provider.logical.input"})
				}
			}
		}
	}
	return evidence
}

func decodeProvider(raw []byte) (providerObservation, error) {
	var provider providerObservation
	if err := json.Unmarshal(raw, &provider); err != nil {
		return providerObservation{}, err
	}
	if provider.Schema != validProviderSchema || provider.Provider == "" || provider.SubjectSHA == "" || len(provider.FileReads) != 1 || len(provider.LogicalInputs) != 1 || len(provider.EnvironmentReads) != 0 || len(provider.NetworkReads) != 0 {
		return providerObservation{}, fmt.Errorf("provider observation shape is incomplete")
	}
	if !provider.FileReads[0].Observed || provider.FileReads[0].EvidenceClass != currentEvidence || !provider.LogicalInputs[0].Observed || provider.LogicalInputs[0].EvidenceClass != currentEvidence {
		return providerObservation{}, fmt.Errorf("provider current evidence is incomplete")
	}
	if provider.SandboxBefore.Digest != provider.SandboxAfter.Digest || provider.ActualRepositoryWrites != 0 || provider.ActualMutationAuthority || provider.ActualPromotionAuthority {
		return providerObservation{}, fmt.Errorf("provider sandbox before/after is not zero effect")
	}
	if len(provider.EffectAttempts) != 3 {
		return providerObservation{}, fmt.Errorf("provider enforcement denominator is incomplete")
	}
	for _, effect := range provider.EffectAttempts {
		if !effect.Requested || effect.Result != "DENIED" || !effect.BoundaryObserved || effect.BeforeDigest != effect.AfterDigest || effect.ActualWrites != 0 || effect.ActualMutation || effect.ActualPromotion {
			return providerObservation{}, fmt.Errorf("provider did not observe denied effect")
		}
	}
	return provider, nil
}

func expected(model sourceModel, provider providerObservation, item caseSpec) (string, string, string, *rawUnknown) {
	op := authorizeOperation(model.Operations)
	if item.Writes != 0 || item.Mutation || item.Promotion {
		if effectBoundary(provider) {
			return decisionDeny, resolutionExact, "CAPABILITY_ENFORCEMENT_OBSERVED", nil
		}
		return decisionUnknown, resolutionLower, "CAPABILITY_ENFORCEMENT_NOT_IMPLEMENTED", &rawUnknown{Stage: opStage(op), Step: opStep(op), Reason: "CAPABILITY_ENFORCEMENT_NOT_IMPLEMENTED"}
	}
	for _, value := range item.Requests {
		declaration, ok := findDeclaration(model.Declarations, value.ValueID)
		if !ok || declaration.Kind != value.Kind || declaration.Operation != value.Operation || declaration.Target != value.Target {
			return decisionDeny, resolutionExact, "CAPABILITY_NOT_DECLARED", nil
		}
	}
	for _, value := range item.Requests {
		declaration, _ := findDeclaration(model.Declarations, value.ValueID)
		if !currentEvidence(provider, declaration) {
			return decisionUnknown, resolutionLower, "EVIDENCE_UNOBSERVED", &rawUnknown{Stage: opStage(op), Step: "bind-capability-evidence", Reason: "EVIDENCE_UNOBSERVED"}
		}
	}
	if model.Policy.AuthorizationMode != "exact-current" {
		return decisionDeny, resolutionExact, "POLICY_REJECTED", nil
	}
	return decisionAllow, resolutionExact, "CAPABILITY_SCOPE_EXACT", nil
}

func validateReceiptShape(source []byte, model sourceModel, provider providerObservation, providerRaw []byte, observed rawReceipt) error {
	if observed.Schema != validSchema || observed.MetaOperation != validMetaOperation || observed.Producer != validProducer || observed.Consumer != validConsumer || observed.GoVersion != validGoVersion || observed.SourceDigest != digestBytes(source) || observed.SemanticDigest != model.SemanticDigest {
		return fmt.Errorf("receipt semantic identity is not source-bound")
	}
	if observed.SubjectSHA != provider.SubjectSHA || observed.ProviderDigest != digestBytes(providerRaw) || !reflect.DeepEqual(observed.Policy, rawPolicy{ID: model.Policy.ID, DefaultDecision: model.Policy.DefaultDecision, AuthorizationMode: model.Policy.AuthorizationMode, Effects: model.Policy.Effects, PriorClaimState: model.Policy.Prior, NodeID: model.Policy.NodeID}) {
		return fmt.Errorf("receipt policy/provider binding is inconsistent")
	}
	wantDeclarations := make([]rawDeclaration, 0, len(model.Declarations))
	for _, item := range model.Declarations {
		wantDeclarations = append(wantDeclarations, rawDeclaration{ValueID: item.ValueID, Kind: item.Kind, Operation: item.Operation, Target: item.Target, Policy: item.Policy, PriorClaimState: item.Prior, EvidenceClass: item.Evidence, NodeID: item.NodeID})
	}
	if !reflect.DeepEqual(observed.Declarations, wantDeclarations) {
		return fmt.Errorf("receipt declarations are not reconstructed from source IR")
	}
	if observed.Authority.CurrentEvidenceDenominator != len(model.Declarations) || observed.Authority.CurrentEvidenceCapabilities != currentEvidenceCount(model, provider) || observed.Authority.EnforcementObservations != len(provider.EffectAttempts) {
		return fmt.Errorf("receipt evidence denominator is inconsistent")
	}
	op := authorizeOperation(model.Operations)
	if observed.Stage != opStage(op) || observed.Step != opStep(op) {
		return fmt.Errorf("receipt operation is not reconstructed from source IR")
	}
	if observed.RepositoryWrites != provider.ActualRepositoryWrites || observed.MutationAuthority != provider.ActualMutationAuthority || observed.PromotionAuthority != provider.ActualPromotionAuthority || !reflect.DeepEqual(observed.EffectObservations, provider.EffectAttempts) {
		return fmt.Errorf("receipt does not bind before/after enforcement observations")
	}
	if len(observed.Indicators) != fixedIndicatorTotal {
		return fmt.Errorf("fixed indicator denominator mismatch")
	}
	return nil
}

func validateAuthority(provider providerObservation, model sourceModel, item caseSpec, observed rawReceipt) error {
	if observed.Authority.CapabilitiesRequested != len(item.Requests) || observed.Authority.RequestedRepositoryWrites != item.Writes || observed.Authority.RequestedMutationAuthority != item.Mutation || observed.Authority.RequestedPromotionAuthority != item.Promotion {
		return fmt.Errorf("authority request is not source-case bound")
	}
	if observed.Decision == decisionAllow && (observed.Authority.CapabilitiesAuthorized != len(item.Requests) || observed.Authority.CapabilitiesDenied != 0 || observed.Authority.CapabilitiesUnknown != 0) {
		return fmt.Errorf("ALLOW authority counters are inconsistent")
	}
	if observed.Decision == decisionDeny && (observed.Authority.CapabilitiesAuthorized != 0 || observed.Authority.CapabilitiesDenied != len(item.Requests) || observed.Authority.CapabilitiesUnknown != 0) {
		return fmt.Errorf("DENY authority counters are inconsistent")
	}
	if observed.Decision == decisionUnknown && (observed.Authority.CapabilitiesAuthorized != 0 || observed.Authority.CapabilitiesDenied != 0 || observed.Authority.CapabilitiesUnknown != len(item.Requests)) {
		return fmt.Errorf("UNKNOWN authority counters are inconsistent")
	}
	declared := 0
	for _, request := range item.Requests {
		if declaration, ok := findDeclaration(model.Declarations, request.ValueID); ok && declaration.Kind == request.Kind && declaration.Operation == request.Operation && declaration.Target == request.Target {
			declared++
		}
	}
	if observed.Authority.CapabilitiesDeclared != declared || observed.Authority.CurrentEvidenceCapabilities > observed.Authority.CurrentEvidenceDenominator {
		return fmt.Errorf("authority counter is out of bounds")
	}
	_ = provider
	_ = model
	return nil
}

func validateClaims(model sourceModel, provider providerObservation, item caseSpec, observed rawReceipt, unknown *rawUnknown) error {
	if len(observed.Claims) != 3 || len(observed.ClaimTransitions) != len(item.Requests)+3 {
		return fmt.Errorf("claim lifecycle denominator is incomplete")
	}
	wantNext := claimOpen
	if observed.Decision == decisionAllow {
		wantNext = claimDischarged
	}
	if observed.Decision == decisionDeny {
		wantNext = claimRefuted
	}
	seen := make(map[string]bool)
	for _, transition := range observed.ClaimTransitions {
		if transition.ClaimID == "capability-scope-exact" {
			if transition.PriorState != model.Policy.Prior || transition.NextState != wantNext {
				return fmt.Errorf("scope claim transition is invalid")
			}
		}
		if transition.ClaimID == "default-deny" {
			if transition.PriorState != model.Policy.Prior || transition.NextState != claimOpen && observed.Decision == decisionUnknown || transition.NextState == "" {
				return fmt.Errorf("default-deny claim transition is invalid")
			}
		}
		if transition.ClaimID == "effect-ceiling" {
			if transition.PriorState != model.Policy.Prior || transition.NextState != claimDischarged {
				return fmt.Errorf("effect claim transition is not observed")
			}
		}
		if strings.HasPrefix(transition.ClaimID, "capability:") && (transition.PriorState != claimOpen || transition.NextState != wantNext) {
			return fmt.Errorf("capability claim transition is invalid")
		}
		if transition.Stage == "" || transition.Step == "" || transition.Reason == "" || transition.EvidenceDigest == "" || transition.Provenance != "source-ir+provider-observation" && transition.Provenance != "sandbox-before-after" {
			return fmt.Errorf("claim transition lacks provenance")
		}
		seen[transition.ClaimID] = true
	}
	for _, claim := range observed.Claims {
		if claim.ID == "" || claim.PriorState != model.Policy.Prior || claim.Status == "" {
			return fmt.Errorf("claim lacks prior state")
		}
		if !seen[claim.ID] {
			return fmt.Errorf("claim is not backed by append-only transition")
		}
	}
	if unknown != nil && observed.Decision != decisionUnknown {
		return fmt.Errorf("UNKNOWN evidence attached to known decision")
	}
	_ = provider
	return nil
}

func validateIndicators(indicators []rawIndicator) error {
	want := map[string]string{
		"CSE-current-file-evidence": "OUTCOME/OBSERVATION", "CSE-current-logical-evidence": "OUTCOME/OBSERVATION", "CSE-declaration-kind": "DRIVER/FOUNDATION", "CSE-declaration-operation": "DRIVER/FOUNDATION", "CSE-declaration-target": "DRIVER/FOUNDATION", "CSE-environment-network-lower-resolution": "OUTCOME/EPISTEMIC", "CSE-prior-claim-open": "DRIVER/FOUNDATION", "CSE-provider-enforcement-boundary": "GUARDRAIL/ENFORCEMENT", "CSE-receipt-seal": "DRIVER/COHERENCE", "CSE-sandbox-before-after": "GUARDRAIL/OBSERVATION", "CSE-semantic-policy": "DRIVER/FOUNDATION", "CSE-source-reconstruction": "DRIVER/FOUNDATION",
	}
	seen := make(map[string]bool, len(indicators))
	for _, item := range indicators {
		if seen[item.ID] || want[item.ID] != item.Class+"/"+item.ProofChoice || item.Producer != validProducer || item.Consumer != validConsumer || item.MetaOperation != validMetaOperation || item.Target != 1 {
			return fmt.Errorf("indicator provenance or denominator is invalid")
		}
		seen[item.ID] = true
	}
	if len(seen) != len(want) {
		return fmt.Errorf("indicator denominator is incomplete")
	}
	return nil
}

func validateUnknown(observed rawReceipt, want *rawUnknown) error {
	if want == nil {
		if observed.Unknown != nil {
			return fmt.Errorf("known receipt carries UNKNOWN evidence")
		}
		return nil
	}
	if observed.Unknown == nil || observed.Unknown.Stage != want.Stage || observed.Unknown.Step != want.Step || observed.Unknown.Reason != want.Reason {
		return fmt.Errorf("UNKNOWN stage/step/reason is not reconstructed")
	}
	return nil
}

func effectBoundary(provider providerObservation) bool {
	for _, effect := range provider.EffectAttempts {
		if !effect.Requested || effect.Result != "DENIED" || !effect.BoundaryObserved || effect.BeforeDigest != effect.AfterDigest {
			return false
		}
	}
	return len(provider.EffectAttempts) == 3
}
func currentEvidence(provider providerObservation, declaration declaration) bool {
	if declaration.Evidence != currentEvidence {
		return false
	}
	if declaration.Kind == "file" {
		for _, item := range provider.FileReads {
			if item.Target == declaration.Target && item.Observed && item.EvidenceClass == currentEvidence {
				return true
			}
		}
	}
	if declaration.Kind == "time" {
		for _, item := range provider.LogicalInputs {
			if item.Target == declaration.Target && item.Observed && item.EvidenceClass == currentEvidence {
				return true
			}
		}
	}
	return false
}
func currentEvidenceCount(model sourceModel, provider providerObservation) int {
	count := 0
	for _, declaration := range model.Declarations {
		if currentEvidence(provider, declaration) {
			count++
		}
	}
	return count
}
func authorizeOperation(operations []operation) operation {
	for _, item := range operations {
		if item.ID == "authorize-before-expand" {
			return item
		}
	}
	return operations[0]
}
func opStage(item operation) string { return item.Stage }
func opStep(item operation) string  { return item.Step }
func findCase(cases []caseSpec, id string) (caseSpec, bool) {
	for _, item := range cases {
		if item.ID == id {
			return item, true
		}
	}
	return caseSpec{}, false
}
func findDeclaration(declarations []declaration, id string) (declaration, bool) {
	for _, item := range declarations {
		if item.ValueID == id {
			return item, true
		}
	}
	return declaration{}, false
}

func invalid(reason string) Verdict {
	return Verdict{Status: "FAIL", Decision: decisionUnknown, Resolution: resolutionLower, Reason: reason, SourceReconstruction: "UNKNOWN", ProducerImports: 0}
}
func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func canonicalDigest(raw []byte) string {
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	value["report_digest"] = ""
	normalized, _ := json.Marshal(value)
	return digestBytes(normalized)
}
