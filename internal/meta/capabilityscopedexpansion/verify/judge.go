package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	validSchema         = "gooo/capability-scoped-expansion/v3"
	validProviderSchema = "gooo/capability-scoped-expansion/provider/v3"
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
	fixedIndicators     = 12
)

type ProviderContext struct {
	RepositoryRoot   string
	PinnedFile       string
	LogicalInputPath string
	SandboxRoot      string
}

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
	Graph              rawGraph             `json:"graph"`
	Declarations       []rawDeclaration     `json:"declarations"`
	Capabilities       []rawCapability      `json:"capabilities"`
	Evidence           []rawEvidence        `json:"evidence"`
	ProviderDigest     string               `json:"provider_digest"`
	TokenAttempts      []rawToken           `json:"token_attempts"`
	Execution          rawExecution         `json:"execution"`
	Artifact           rawArtifact          `json:"artifact"`
	Propositions       []rawProposition     `json:"propositions"`
	Unknown            *rawUnknown          `json:"unknown,omitempty"`
	Authority          rawAuthority         `json:"authority"`
	Claims             []rawClaim           `json:"claims"`
	ClaimTransitions   []rawClaimTransition `json:"claim_transitions"`
	Indicators         []rawIndicator       `json:"indicators"`
	RepositoryWrites   int                  `json:"repository_writes"`
	SandboxWrites      int                  `json:"sandbox_writes"`
	MutationAuthority  string               `json:"mutation_authority"`
	PromotionAuthority string               `json:"promotion_authority"`
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

type rawCapability struct {
	ValueID   string `json:"value_id"`
	Kind      string `json:"kind"`
	Operation string `json:"operation"`
	Target    string `json:"target"`
}

type rawEvidence struct {
	ValueID        string `json:"value_id"`
	Observed       string `json:"observed"`
	EvidenceClass  string `json:"evidence_class"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
}

type rawUnknown struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type rawGraph struct {
	Facts        []rawGraphFact `json:"facts"`
	RequiredPath []string       `json:"required_path"`
	PathDigest   string         `json:"path_digest"`
	Complete     bool           `json:"complete"`
}

type rawGraphFact struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

type rawToken struct {
	Kind          string `json:"kind"`
	Operation     string `json:"operation"`
	Target        string `json:"target"`
	Requested     bool   `json:"requested"`
	Decision      string `json:"decision"`
	Issued        bool   `json:"issued"`
	Reason        string `json:"reason"`
	PolicyDigest  string `json:"policy_digest"`
	RequestDigest string `json:"request_digest"`
}

type rawExecution struct {
	Requested              bool   `json:"requested"`
	Decision               string `json:"decision"`
	Result                 string `json:"result"`
	ClaimID                string `json:"claim_id"`
	ClaimState             string `json:"claim_state"`
	Reason                 string `json:"reason"`
	ArtifactPath           string `json:"artifact_path"`
	ArtifactValue          string `json:"artifact_value"`
	ArtifactBytes          int    `json:"artifact_bytes"`
	ArtifactDigest         string `json:"artifact_digest"`
	ArtifactSemanticDigest string `json:"artifact_semantic_digest"`
	ReparsedSemanticDigest string `json:"reparsed_semantic_digest"`
}

type rawArtifact struct {
	Schema                 string `json:"schema"`
	Present                bool   `json:"present"`
	Path                   string `json:"path"`
	Value                  string `json:"value"`
	Bytes                  int    `json:"bytes"`
	ContentDigest          string `json:"content_digest"`
	SemanticDigest         string `json:"semantic_digest"`
	Reparsed               bool   `json:"reparsed"`
	ReparsedSemanticDigest string `json:"reparsed_semantic_digest"`
}

type rawProposition struct {
	ID             string `json:"id"`
	Predicate      string `json:"predicate"`
	Decision       string `json:"decision"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
}

type rawAuthority struct {
	CapabilitiesRequested       int    `json:"capabilities_requested"`
	CapabilitiesDeclared        int    `json:"capabilities_declared"`
	CapabilitiesAuthorized      int    `json:"capabilities_authorized"`
	CapabilitiesDenied          int    `json:"capabilities_denied"`
	CapabilitiesUnknown         int    `json:"capabilities_unknown"`
	CurrentEvidenceCapabilities int    `json:"current_evidence_capabilities"`
	CurrentEvidenceDenominator  int    `json:"current_evidence_denominator"`
	RequestedRepositoryWrites   int    `json:"requested_repository_writes"`
	RequestedMutationAuthority  bool   `json:"requested_mutation_authority"`
	RequestedPromotionAuthority bool   `json:"requested_promotion_authority"`
	RepositoryWrites            int    `json:"repository_writes"`
	SandboxWrites               int    `json:"sandbox_writes"`
	MutationAuthority           string `json:"mutation_authority"`
	PromotionAuthority          string `json:"promotion_authority"`
	EnforcementObservations     int    `json:"enforcement_observations"`
}

type rawClaim struct {
	ID          string `json:"id"`
	PriorState  string `json:"prior_state"`
	Status      string `json:"status"`
	ProofChoice string `json:"proof_choice"`
	Evidence    string `json:"evidence"`
}

type rawClaimTransition struct {
	ClaimID        string `json:"claim_id"`
	PriorState     string `json:"prior_state"`
	NextState      string `json:"next_state"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
}

type rawIndicator struct {
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

type rawProvider struct {
	Schema              string           `json:"schema"`
	Provider            string           `json:"provider"`
	SubjectSHA          string           `json:"subject_sha"`
	FileReads           []rawFileRead    `json:"file_reads"`
	LogicalInputs       []rawLogical     `json:"logical_inputs"`
	EnvironmentReads    []rawEnvironment `json:"environment_reads"`
	NetworkReads        []rawNetwork     `json:"network_reads"`
	TokenAttempts       []rawToken       `json:"token_attempts"`
	BrokerTokenRequests int              `json:"broker_token_requests"`
	BrokerTokensIssued  int              `json:"broker_tokens_issued"`
	BrokerTokenDenials  int              `json:"broker_token_denials"`
	RepositoryBefore    rawSnapshot      `json:"repository_before"`
	RepositoryAfter     rawSnapshot      `json:"repository_after"`
	SandboxBefore       rawSnapshot      `json:"sandbox_before"`
	SandboxAfter        rawSnapshot      `json:"sandbox_after"`
	RepositoryWrites    int              `json:"repository_writes"`
	SandboxWrites       int              `json:"sandbox_writes"`
	MutationAuthority   string           `json:"mutation_authority"`
	PromotionAuthority  string           `json:"promotion_authority"`
	EffectAPIAccess     string           `json:"effect_api_access"`
}

type rawFileRead struct {
	Target        string `json:"target"`
	Path          string `json:"path"`
	ContentDigest string `json:"content_digest"`
	Observed      bool   `json:"observed"`
	EvidenceClass string `json:"evidence_class"`
}

type rawLogical struct {
	Target        string `json:"target"`
	Path          string `json:"path"`
	Value         string `json:"value"`
	Observed      bool   `json:"observed"`
	EvidenceClass string `json:"evidence_class"`
}

type rawEnvironment struct {
	Target        string `json:"target"`
	Observed      bool   `json:"observed"`
	EvidenceClass string `json:"evidence_class"`
}

type rawNetwork struct {
	Target        string `json:"target"`
	Observed      bool   `json:"observed"`
	EvidenceClass string `json:"evidence_class"`
}

type rawSnapshot struct {
	Scope   string   `json:"scope"`
	Root    string   `json:"root"`
	Entries []string `json:"entries"`
	Digest  string   `json:"digest"`
}

type semanticValue struct {
	scheme string
	fields map[string]string
	nodeID string
}

type sourceModel struct {
	sourceDigest   string
	semanticDigest string
	policy         rawPolicy
	operations     []rawOperation
	declarations   []rawDeclaration
	cases          []rawCase
	graph          rawGraph
}

type rawOperation struct {
	ID              string
	Stage           string
	Step            string
	PriorClaimState string
	NodeID          string
}

type rawCase struct {
	ID                          string
	Requests                    []rawCapability
	RequestedRepositoryWrites   int
	RequestedMutationAuthority  bool
	RequestedPromotionAuthority bool
	PriorClaimState             string
	NodeID                      string
}

func Judge(source, providerRaw, receiptRaw []byte) Verdict {
	return JudgeWithContext(source, providerRaw, receiptRaw, ProviderContext{})
}

// JudgeWithContext is deliberately independent of the producer, broker, and
// expansion engine. It reconstructs source semantics, reobserves provider
// inputs, reparses the output artifact, and checks the receipt from raw JSON.
func JudgeWithContext(source, providerRaw, receiptRaw []byte, context ProviderContext) Verdict {
	model, err := reconstructSource(source)
	if err != nil {
		return fail("source reconstruction: " + err.Error())
	}
	var provider rawProvider
	if err := json.Unmarshal(providerRaw, &provider); err != nil {
		return fail("provider JSON: " + err.Error())
	}
	var receipt rawReceipt
	if err := json.Unmarshal(receiptRaw, &receipt); err != nil {
		return fail("receipt JSON: " + err.Error())
	}
	if err := validateProvider(provider, context); err != nil {
		return fail("provider validation: " + err.Error())
	}
	if receipt.Schema != validSchema || receipt.MetaOperation != validMetaOperation || receipt.Producer != validProducer || receipt.Consumer != validConsumer || receipt.GoVersion != validGoVersion {
		return fail("receipt identity is not v3")
	}
	if receipt.SourceDigest != model.sourceDigest || receipt.SemanticDigest != model.semanticDigest || receipt.ProviderDigest != digestBytes(providerRaw) || receipt.SubjectSHA != provider.SubjectSHA {
		return fail("receipt digest binding does not match raw source/provider")
	}
	item, ok := findCase(model.cases, receipt.CaseID)
	if !ok {
		return fail("receipt case is not reconstructed from source")
	}
	expected, resolution, reason, unknown := decisionFor(model, provider, item)
	if receipt.Decision != expected || receipt.Resolution != resolution || receipt.Reason != reason {
		return fail(fmt.Sprintf("decision mismatch: got %s/%s/%s want %s/%s/%s", receipt.Decision, receipt.Resolution, receipt.Reason, expected, resolution, reason))
	}
	if err := validateReceipt(model, provider, receipt, item, unknown, context); err != nil {
		return fail("receipt validation: " + err.Error())
	}
	if receipt.ReportDigest != digestReceipt(receiptRaw) {
		return fail("receipt seal does not match raw receipt")
	}
	return Verdict{Status: "PASS", Decision: receipt.Decision, Resolution: receipt.Resolution, Reason: "INDEPENDENT_SOURCE_PROVIDER_ARTIFACT_REPLAY", ReceiptDigest: receipt.ReportDigest, SourceReconstruction: "PASS"}
}

func fail(reason string) Verdict {
	return Verdict{Status: "FAIL", Decision: decisionUnknown, Resolution: resolutionLower, Reason: reason, SourceReconstruction: "FAIL"}
}

func reconstructSource(source []byte) (sourceModel, error) {
	file, diagnostics := syntax.ParseFile("capability-scoped-expansion.gooo", string(source))
	if err := diagnostics.Error(); err != nil {
		return sourceModel{}, err
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, err
	}
	values := make([]semanticValue, 0)
	for _, node := range ir.Graph.Nodes() {
		if node.ValueProgram == "" {
			continue
		}
		value, err := parseSemanticValue(node.ValueProgram)
		if err != nil {
			return sourceModel{}, err
		}
		value.nodeID = node.ID.String()
		values = append(values, value)
	}
	model := sourceModel{sourceDigest: digestBytes(source), semanticDigest: ir.StableHash()}
	if err := model.fromValues(values); err != nil {
		return sourceModel{}, err
	}
	model.graph = graphProof(ir, model)
	return model, nil
}

func parseSemanticValue(raw string) (semanticValue, error) {
	parts := strings.Split(raw, "|")
	if len(parts) < 2 || parts[0] == "" {
		return semanticValue{}, fmt.Errorf("malformed semantic value")
	}
	value := semanticValue{scheme: parts[0], fields: make(map[string]string)}
	for _, part := range parts[1:] {
		key, field, ok := strings.Cut(part, "=")
		if !ok || key == "" || field == "" {
			return semanticValue{}, fmt.Errorf("malformed semantic value field %q", part)
		}
		value.fields[key] = field
	}
	return value, nil
}

func (model *sourceModel) fromValues(values []semanticValue) error {
	for _, value := range values {
		switch value.scheme {
		case policyScheme:
			if model.policy.ID != "" {
				return fmt.Errorf("multiple policies")
			}
			model.policy = rawPolicy{ID: value.fields["id"], DefaultDecision: value.fields["default"], AuthorizationMode: value.fields["authorization"], Effects: value.fields["effects"], PriorClaimState: value.fields["prior"], NodeID: value.nodeID}
		case operationScheme:
			model.operations = append(model.operations, rawOperation{ID: value.fields["id"], Stage: value.fields["stage"], Step: value.fields["step"], PriorClaimState: value.fields["prior"], NodeID: value.nodeID})
		case declarationScheme:
			model.declarations = append(model.declarations, rawDeclaration{ValueID: value.fields["id"], Kind: value.fields["kind"], Operation: value.fields["operation"], Target: value.fields["target"], Policy: value.fields["policy"], PriorClaimState: value.fields["prior"], EvidenceClass: value.fields["evidence"], NodeID: value.nodeID})
		}
	}
	byID := make(map[string]rawDeclaration, len(model.declarations))
	for _, declaration := range model.declarations {
		byID[declaration.ValueID] = declaration
	}
	for _, value := range values {
		if value.scheme != caseScheme {
			continue
		}
		item := rawCase{ID: value.fields["id"], PriorClaimState: value.fields["prior"], NodeID: value.nodeID}
		for _, request := range strings.Split(value.fields["requests"], ",") {
			valueID, target, ok := strings.Cut(request, "@")
			declaration, exists := byID[valueID]
			if !ok || !exists {
				return fmt.Errorf("case %s has unknown request %q", item.ID, request)
			}
			item.Requests = append(item.Requests, rawCapability{ValueID: valueID, Kind: declaration.Kind, Operation: declaration.Operation, Target: target})
		}
		if value.fields["writes"] != "" {
			if _, err := fmt.Sscanf(value.fields["writes"], "%d", &item.RequestedRepositoryWrites); err != nil {
				return err
			}
		}
		item.RequestedMutationAuthority = value.fields["mutation"] == "true"
		item.RequestedPromotionAuthority = value.fields["promotion"] == "true"
		model.cases = append(model.cases, item)
	}
	if model.policy.ID == "" || model.policy.DefaultDecision != "DENY" || model.policy.Effects != "NONE" || model.policy.PriorClaimState != claimOpen || len(model.declarations) != 4 || len(model.cases) != 9 {
		return fmt.Errorf("source denominator or default-deny policy is incomplete")
	}
	sort.Slice(model.operations, func(i, j int) bool { return model.operations[i].ID < model.operations[j].ID })
	sort.Slice(model.declarations, func(i, j int) bool { return model.declarations[i].ValueID < model.declarations[j].ValueID })
	sort.Slice(model.cases, func(i, j int) bool { return model.cases[i].ID < model.cases[j].ID })
	return nil
}

func graphProof(ir semantic.IR, model sourceModel) rawGraph {
	facts := make([]rawGraphFact, 0, len(ir.Graph.Facts()))
	for _, fact := range ir.Graph.Facts() {
		facts = append(facts, rawGraphFact{Subject: fact.Subject.String(), Predicate: fact.Predicate.String(), Object: fact.Object.String()})
	}
	policyOutput := generatedEntity(facts, model.policy.NodeID)
	policyInput := usedEntity(facts, model.policy.NodeID)
	authorize := operationNode(model, "authorize-before-expand")
	authorizeOutput := generatedEntity(facts, authorize)
	bind := operationNode(model, "bind-capability-evidence")
	bindOutput := generatedEntity(facts, bind)
	expand := operationNode(model, "expand-with-capability-evidence")
	expandOutput := generatedEntity(facts, expand)
	path := []string{policyInput, model.policy.NodeID, policyOutput, authorize, authorizeOutput, bind, bindOutput, expand, expandOutput}
	complete := true
	for index := 0; index+1 < len(path); index++ {
		if path[index] == "" || path[index+1] == "" {
			complete = false
			continue
		}
		subject, predicate, object := path[index+1], "used", path[index]
		if index%2 == 1 {
			subject, predicate, object = path[index+1], "wasGeneratedBy", path[index]
		}
		if !graphFactExists(facts, subject, predicate, object) {
			complete = false
		}
	}
	canonical := append([]string{"path"}, path...)
	canonical = append(canonical, "facts")
	for _, fact := range facts {
		canonical = append(canonical, fact.Subject+"|"+fact.Predicate+"|"+fact.Object)
	}
	return rawGraph{Facts: facts, RequiredPath: path, PathDigest: digestBytes([]byte(strings.Join(canonical, "\n"))), Complete: complete}
}

func operationNode(model sourceModel, id string) string {
	for _, operation := range model.operations {
		if operation.ID == id {
			return operation.NodeID
		}
	}
	return ""
}

func generatedEntity(facts []rawGraphFact, activity string) string {
	for _, fact := range facts {
		if fact.Predicate == "wasGeneratedBy" && fact.Object == activity {
			return fact.Subject
		}
	}
	return ""
}

func usedEntity(facts []rawGraphFact, activity string) string {
	for _, fact := range facts {
		if fact.Predicate == "used" && fact.Subject == activity {
			return fact.Object
		}
	}
	return ""
}

func graphFactExists(facts []rawGraphFact, subject, predicate, object string) bool {
	for _, fact := range facts {
		if fact.Subject == subject && fact.Predicate == predicate && fact.Object == object {
			return true
		}
	}
	return false
}

func validateProvider(provider rawProvider, context ProviderContext) error {
	if provider.Schema != validProviderSchema || provider.Provider == "" || provider.SubjectSHA == "" || len(provider.FileReads) != 1 || len(provider.LogicalInputs) != 1 || len(provider.EnvironmentReads) != 1 || len(provider.NetworkReads) != 1 {
		return fmt.Errorf("provider identity or evidence slots incomplete")
	}
	file := provider.FileReads[0]
	if file.Target != "pinned-file" || !file.Observed || file.EvidenceClass != currentEvidence || file.Path == "" || file.ContentDigest == "" {
		return fmt.Errorf("current pinned-file observation incomplete")
	}
	logical := provider.LogicalInputs[0]
	if logical.Target != "logical-clock" || logical.Path == "" || logical.Value != "logical-clock:0" || !logical.Observed || logical.EvidenceClass != currentEvidence {
		return fmt.Errorf("current logical observation incomplete")
	}
	if provider.EnvironmentReads[0].Observed || provider.EnvironmentReads[0].EvidenceClass != "UNKNOWN" || provider.NetworkReads[0].Observed || provider.NetworkReads[0].EvidenceClass != historicalFixture {
		return fmt.Errorf("environment/network evidence was promoted without observation")
	}
	if provider.RepositoryBefore.Scope != "repository" || provider.RepositoryAfter.Scope != "repository" || provider.SandboxBefore.Scope != "sandbox" || provider.SandboxAfter.Scope != "sandbox" || provider.RepositoryWrites != 0 || provider.SandboxWrites != 0 || provider.RepositoryBefore.Digest != provider.RepositoryAfter.Digest || provider.SandboxBefore.Digest != provider.SandboxAfter.Digest {
		return fmt.Errorf("repository and sandbox effects are not separately zero by snapshot")
	}
	if provider.MutationAuthority != "NOT_OBSERVED" || provider.PromotionAuthority != "NOT_OBSERVED" || provider.EffectAPIAccess != "NOT_REACHED_WITHOUT_TOKEN" {
		return fmt.Errorf("unmeasured authority was presented as a boolean result")
	}
	if len(provider.TokenAttempts) != 3 || provider.BrokerTokenRequests != 3 || provider.BrokerTokensIssued != 0 || provider.BrokerTokenDenials != 3 {
		return fmt.Errorf("broker issuance denominator is not 3/0/3")
	}
	expectedTokens := []brokerRequestWire{
		{Kind: "file", Operation: "write", Target: "repository", PolicyID: "default-deny"},
		{Kind: "mutation", Operation: "mutate", Target: "sandbox", PolicyID: "default-deny"},
		{Kind: "promotion", Operation: "promote", Target: "repository", PolicyID: "default-deny"},
	}
	for index, token := range provider.TokenAttempts {
		if !token.Requested || token.Decision != decisionDeny || token.Issued || token.PolicyDigest == "" || token.RequestDigest == "" {
			return fmt.Errorf("effect token was not denied by the broker")
		}
		if token.Kind != expectedTokens[index].Kind || token.Operation != expectedTokens[index].Operation || token.Target != expectedTokens[index].Target || token.PolicyDigest != brokerPolicyDigest() || token.RequestDigest != brokerRequestDigest(expectedTokens[index]) {
			return fmt.Errorf("broker request or digest does not match the declared effect boundary")
		}
	}
	if context.PinnedFile != "" {
		pinned, err := filepath.Abs(context.PinnedFile)
		if err != nil {
			return err
		}
		observed, err := filepath.Abs(file.Path)
		if err != nil || pinned != observed {
			return fmt.Errorf("provider pinned path does not match consumer context")
		}
		contents, err := os.ReadFile(pinned)
		if err != nil {
			return fmt.Errorf("consumer pinned-file reobservation: %w", err)
		}
		if digestBytes(contents) != file.ContentDigest {
			return fmt.Errorf("provider pinned bytes failed independent reobservation")
		}
	}
	if context.LogicalInputPath != "" {
		logicalPath, err := filepath.Abs(context.LogicalInputPath)
		if err != nil {
			return err
		}
		observed, err := filepath.Abs(logical.Path)
		if err != nil || logicalPath != observed {
			return fmt.Errorf("provider logical input path does not match consumer context")
		}
		contents, err := os.ReadFile(logicalPath)
		if err != nil {
			return fmt.Errorf("consumer logical input reobservation: %w", err)
		}
		if strings.TrimSpace(string(contents)) != logical.Value {
			return fmt.Errorf("provider logical input failed independent reobservation")
		}
	}
	if context.RepositoryRoot != "" {
		repository, err := snapshotRepository(context.RepositoryRoot)
		if err != nil {
			return err
		}
		if repository.Digest != provider.RepositoryAfter.Digest || !reflect.DeepEqual(repository.Entries, provider.RepositoryAfter.Entries) {
			return fmt.Errorf("provider repository snapshot failed independent replay")
		}
	}
	if context.SandboxRoot != "" {
		sandbox, err := snapshotSandbox(context.SandboxRoot)
		if err != nil {
			return err
		}
		if sandbox.Digest != provider.SandboxAfter.Digest || !reflect.DeepEqual(sandbox.Entries, provider.SandboxAfter.Entries) {
			return fmt.Errorf("provider sandbox snapshot failed independent replay")
		}
	}
	return nil
}

type brokerPolicyWire struct {
	ID                string
	DefaultDecision   string
	AuthorizationMode string
	Effects           string
}

type brokerRequestWire struct {
	Kind      string `json:"kind"`
	Operation string `json:"operation"`
	Target    string `json:"target"`
	PolicyID  string `json:"policy_id"`
}

func brokerPolicyDigest() string {
	raw, _ := json.Marshal(brokerPolicyWire{ID: "default-deny", DefaultDecision: "DENY", AuthorizationMode: "exact-current", Effects: "NONE"})
	return digestBytes(raw)
}

func brokerRequestDigest(request brokerRequestWire) string {
	raw, _ := json.Marshal(request)
	return digestBytes(raw)
}

func validateReceipt(model sourceModel, provider rawProvider, receipt rawReceipt, item rawCase, unknown *rawUnknown, context ProviderContext) error {
	if !reflect.DeepEqual(receipt.Policy, model.policy) || !reflect.DeepEqual(receipt.Graph, model.graph) || !reflect.DeepEqual(receipt.Declarations, model.declarations) || !reflect.DeepEqual(receipt.Capabilities, item.Requests) || !reflect.DeepEqual(receipt.TokenAttempts, provider.TokenAttempts) {
		return fmt.Errorf("receipt did not carry lowered policy, graph, declarations, capabilities, or broker observations")
	}
	expectedEvidence := evidenceFor(provider, model.declarations, item.Requests)
	if !reflect.DeepEqual(receipt.Evidence, expectedEvidence) {
		return fmt.Errorf("evidence is not reconstructed from raw provider observations")
	}
	if receipt.RepositoryWrites != provider.RepositoryWrites || receipt.SandboxWrites != provider.SandboxWrites || receipt.MutationAuthority != provider.MutationAuthority || receipt.PromotionAuthority != provider.PromotionAuthority {
		return fmt.Errorf("receipt effect snapshots are not bound")
	}
	current := currentEvidenceCount(provider, model.declarations)
	declared := declaredCount(model.declarations, item.Requests)
	if receipt.Authority.CapabilitiesRequested != len(item.Requests) || receipt.Authority.CapabilitiesDeclared != declared || receipt.Authority.CurrentEvidenceCapabilities != current || receipt.Authority.CurrentEvidenceDenominator != len(model.declarations) || receipt.Authority.RepositoryWrites != provider.RepositoryWrites || receipt.Authority.SandboxWrites != provider.SandboxWrites || receipt.Authority.MutationAuthority != provider.MutationAuthority || receipt.Authority.PromotionAuthority != provider.PromotionAuthority || receipt.Authority.EnforcementObservations != provider.BrokerTokenDenials {
		return fmt.Errorf("authority counters do not match observations")
	}
	if receipt.Authority.RequestedRepositoryWrites != item.RequestedRepositoryWrites || receipt.Authority.RequestedMutationAuthority != item.RequestedMutationAuthority || receipt.Authority.RequestedPromotionAuthority != item.RequestedPromotionAuthority {
		return fmt.Errorf("effect request flags are not source-derived")
	}
	if receipt.Decision == decisionAllow {
		if receipt.EnforcementEffect != "NONE" || receipt.Authority.CapabilitiesAuthorized != len(item.Requests) {
			return fmt.Errorf("allow authority is inconsistent")
		}
	} else if receipt.EnforcementEffect != "BLOCK" {
		return fmt.Errorf("blocked decision lacks block effect")
	}
	if unknown == nil {
		if receipt.Unknown != nil {
			return fmt.Errorf("exact decision unexpectedly has UNKNOWN payload")
		}
	} else if !reflect.DeepEqual(receipt.Unknown, unknown) {
		return fmt.Errorf("UNKNOWN stage/step/reason is not bound")
	}
	expectedPropositions := propositions(model, provider, item, receipt)
	if !reflect.DeepEqual(receipt.Propositions, expectedPropositions) || !uniquePropositions(receipt.Propositions) {
		return fmt.Errorf("proposition predicates/digests are not unique or source-derived")
	}
	if err := validateArtifact(model, provider, item, receipt, context); err != nil {
		return err
	}
	if err := validateClaims(model, provider, item, receipt); err != nil {
		return err
	}
	if len(receipt.Indicators) != fixedIndicators {
		return fmt.Errorf("indicator repetition is not fixed at %d per receipt", fixedIndicators)
	}
	for _, indicator := range receipt.Indicators {
		if indicator.Producer != validProducer || indicator.Consumer != validConsumer || indicator.MetaOperation != validMetaOperation || indicator.Target != 1 {
			return fmt.Errorf("indicator provenance is incomplete")
		}
	}
	return nil
}

func validateArtifact(model sourceModel, provider rawProvider, item rawCase, receipt rawReceipt, context ProviderContext) error {
	allow := receipt.Decision == decisionAllow
	if !allow {
		if receipt.Artifact.Present || receipt.Artifact.Path != "" || receipt.Artifact.ContentDigest != "" || receipt.Artifact.SemanticDigest != "" || receipt.Execution.ArtifactDigest != "" || receipt.Execution.Result != "NOT_EXECUTED" || receipt.Execution.Requested {
			return fmt.Errorf("blocked decision has an expansion artifact or execution")
		}
		if receipt.Execution.ClaimState != claimRefuted && receipt.Decision == decisionDeny || receipt.Execution.ClaimState != claimOpen && receipt.Decision == decisionUnknown {
			return fmt.Errorf("blocked execution claim state is inconsistent")
		}
		return nil
	}
	artifact := receipt.Artifact
	if !artifact.Present || artifact.Schema != "gooo/capability-scoped-expansion/artifact/v1" || artifact.Path == "" || artifact.Value == "" || artifact.Bytes <= 0 || artifact.ContentDigest == "" || artifact.SemanticDigest == "" || !artifact.Reparsed || artifact.ReparsedSemanticDigest != artifact.SemanticDigest {
		return fmt.Errorf("ALLOW lacks a complete expansion artifact")
	}
	fileDigest, logicalValue := "", ""
	for _, observation := range provider.FileReads {
		if observation.Target == "pinned-file" && observation.Observed {
			fileDigest = observation.ContentDigest
		}
	}
	for _, observation := range provider.LogicalInputs {
		if observation.Target == "logical-clock" && observation.Observed {
			logicalValue = observation.Value
		}
	}
	expectedValue := fmt.Sprintf("capability.expanded|source=%s|graph=%s|file-digest=%s|logical=%s", model.semanticDigest, model.graph.PathDigest, fileDigest, logicalValue)
	if artifact.Value != expectedValue || item.ID != "allow-current-file-time" {
		return fmt.Errorf("expansion artifact value is not bound to current semantic/evidence inputs")
	}
	if context.RepositoryRoot != "" {
		relative, err := filepath.Rel(context.RepositoryRoot, artifact.Path)
		if err != nil || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
			return fmt.Errorf("expansion artifact is inside the repository")
		}
	}
	contents, err := os.ReadFile(artifact.Path)
	if err != nil {
		return fmt.Errorf("read expansion artifact: %w", err)
	}
	if len(contents) != artifact.Bytes || digestBytes(contents) != artifact.ContentDigest || !strings.Contains(string(contents), artifact.Value) {
		return fmt.Errorf("expansion artifact bytes/value/digest are not bound")
	}
	file, diagnostics := syntax.ParseFile(artifact.Path, string(contents))
	if err := diagnostics.Error(); err != nil {
		return fmt.Errorf("reparse expansion artifact: %w", err)
	}
	ir, err := bidir.Lower(file)
	if err != nil || ir.StableHash() != artifact.SemanticDigest {
		return fmt.Errorf("reparsed expansion artifact semantic digest mismatch")
	}
	if receipt.Execution.Result != "EXPANDED" || !receipt.Execution.Requested || receipt.Execution.ClaimState != claimDischarged || receipt.Execution.ArtifactDigest != artifact.ContentDigest || receipt.Execution.ArtifactPath != artifact.Path || receipt.Execution.ArtifactSemanticDigest != artifact.SemanticDigest || receipt.Execution.ReparsedSemanticDigest != artifact.SemanticDigest {
		return fmt.Errorf("execution receipt is not bound to the artifact")
	}
	return nil
}

func validateClaims(model sourceModel, provider rawProvider, item rawCase, receipt rawReceipt) error {
	expectedTransitions := transitions(model, provider, item, receipt)
	if !reflect.DeepEqual(receipt.ClaimTransitions, expectedTransitions) {
		return fmt.Errorf("claim transitions are not append-only proposition transitions")
	}
	expectedClaims := claims(expectedTransitions)
	if !reflect.DeepEqual(receipt.Claims, expectedClaims) {
		return fmt.Errorf("claims do not preserve unique proposition states")
	}
	return nil
}

func evidenceFor(provider rawProvider, declarations []rawDeclaration, requests []rawCapability) []rawEvidence {
	result := make([]rawEvidence, 0)
	for _, request := range requests {
		declaration, ok := declarationByID(declarations, request.ValueID)
		if !ok || declaration.Kind != request.Kind || declaration.Operation != request.Operation || declaration.Target != request.Target || declaration.EvidenceClass != currentEvidence {
			continue
		}
		if declaration.Kind == "file" {
			for _, observation := range provider.FileReads {
				if observation.Target == declaration.Target && observation.Observed && observation.EvidenceClass == currentEvidence {
					result = append(result, rawEvidence{ValueID: declaration.ValueID, Observed: observation.ContentDigest, EvidenceClass: currentEvidence, EvidenceDigest: digestBytes([]byte(declaration.ValueID + "=" + observation.ContentDigest)), Provenance: "provider.file.read"})
				}
			}
		}
		if declaration.Kind == "time" {
			for _, observation := range provider.LogicalInputs {
				if observation.Target == declaration.Target && observation.Observed && observation.EvidenceClass == currentEvidence {
					result = append(result, rawEvidence{ValueID: declaration.ValueID, Observed: observation.Value, EvidenceClass: currentEvidence, EvidenceDigest: digestBytes([]byte(declaration.ValueID + "=" + observation.Value)), Provenance: "provider.logical.input"})
				}
			}
		}
	}
	return result
}

func propositions(model sourceModel, provider rawProvider, item rawCase, receipt rawReceipt) []rawProposition {
	result := make([]rawProposition, 0, len(item.Requests)+3)
	for index, request := range item.Requests {
		declaration, declared := declarationByID(model.declarations, request.ValueID)
		predicate := fmt.Sprintf("capability:%s:%s:%s:%s", request.ValueID, request.Kind, request.Operation, request.Target)
		status, decision, evidenceDigest := claimOpen, decisionUnknown, digestBytes([]byte(predicate))
		if !declared || declaration.Kind != request.Kind || declaration.Operation != request.Operation || declaration.Target != request.Target {
			status, decision = claimRefuted, decisionDeny
		} else if evidence := evidenceFor(provider, model.declarations, []rawCapability{request}); len(evidence) == 1 {
			status, decision, evidenceDigest = claimDischarged, decisionAllow, evidence[0].EvidenceDigest
		}
		result = append(result, rawProposition{ID: fmt.Sprintf("capability:%s:%d", request.ValueID, index), Predicate: predicate, Decision: decision, Status: status, EvidenceDigest: evidenceDigest, Provenance: "source-ir+independent-provider-replay"})
	}
	result = append(result, rawProposition{ID: "authorization:" + item.ID, Predicate: "authorization:" + item.ID + ":" + model.graph.PathDigest, Decision: receipt.Decision, Status: stateForDecision(receipt.Decision), EvidenceDigest: digestBytes([]byte(model.policy.ID + "=" + model.policy.AuthorizationMode)), Provenance: "lowered-graph-path"})
	if kind := requestedEffectKind(item); kind != "" {
		for _, token := range provider.TokenAttempts {
			if token.Kind == kind {
				result = append(result, rawProposition{ID: "effect-token:" + kind, Predicate: "effect-token:" + kind + ":" + token.RequestDigest, Decision: token.Decision, Status: tokenState(token), EvidenceDigest: token.RequestDigest, Provenance: "broker.issuance-receipt"})
				break
			}
		}
	}
	executionDigest := receipt.SemanticDigest
	executionPredicate := "execution:expanded-syntax:" + receipt.SemanticDigest
	if receipt.Decision != decisionAllow {
		executionDigest = digestBytes([]byte(receipt.CaseID + "=" + receipt.Reason))
		executionPredicate = "execution:expanded-syntax:" + receipt.CaseID
	}
	result = append(result, rawProposition{ID: "execution:expanded-syntax", Predicate: executionPredicate, Decision: receipt.Decision, Status: stateForDecision(receipt.Decision), EvidenceDigest: executionDigest, Provenance: map[bool]string{true: "engine.output-reparse", false: "expansion-gate"}[receipt.Decision == decisionAllow]})
	return result
}

func uniquePropositions(items []rawProposition) bool {
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		key := item.ID + "|" + item.Predicate + "|" + item.EvidenceDigest
		if seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}

func transitions(model sourceModel, provider rawProvider, item rawCase, receipt rawReceipt) []rawClaimTransition {
	result := make([]rawClaimTransition, 0, len(receipt.Propositions)+3)
	for _, proposition := range receipt.Propositions {
		result = append(result, rawClaimTransition{ClaimID: proposition.ID, PriorState: claimOpen, NextState: proposition.Status, Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, EvidenceDigest: proposition.EvidenceDigest, Provenance: proposition.Provenance})
	}
	result = append(result,
		rawClaimTransition{ClaimID: "capability-scope-exact", PriorState: model.policy.PriorClaimState, NextState: stateForDecision(receipt.Decision), Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, EvidenceDigest: receipt.ProviderDigest, Provenance: "source-ir+provider-observation"},
		rawClaimTransition{ClaimID: "default-deny", PriorState: model.policy.PriorClaimState, NextState: knownState(receipt.Decision), Stage: receipt.Stage, Step: receipt.Step, Reason: receipt.Reason, EvidenceDigest: receipt.ProviderDigest, Provenance: "source-ir+provider-observation"},
		rawClaimTransition{ClaimID: "effect-ceiling", PriorState: model.policy.PriorClaimState, NextState: effectState(provider), Stage: receipt.Stage, Step: receipt.Step, Reason: "BROKER_TOKEN_DENIALS_OBSERVED", EvidenceDigest: digestBytes([]byte(fmt.Sprintf("%s|%d|%d|%d", provider.Schema, provider.BrokerTokenRequests, provider.BrokerTokensIssued, provider.BrokerTokenDenials))), Provenance: "broker.issuance+repository-sandbox-snapshots"},
	)
	return result
}

func claims(transitions []rawClaimTransition) []rawClaim {
	result := make([]rawClaim, 0, len(transitions))
	for _, transition := range transitions {
		proof, evidence := "PROPOSITION", transition.Provenance
		if transition.ClaimID == "capability-scope-exact" {
			proof, evidence = "COHERENCE", "source-ir+provider-observation"
		}
		if transition.ClaimID == "default-deny" || transition.ClaimID == "effect-ceiling" {
			proof = "REGRESSION"
		}
		result = append(result, rawClaim{ID: transition.ClaimID, PriorState: transition.PriorState, Status: transition.NextState, ProofChoice: proof, Evidence: evidence})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func decisionFor(model sourceModel, provider rawProvider, item rawCase) (string, string, string, *rawUnknown) {
	operation := authorizationOperation(model)
	if !model.graph.Complete {
		return decisionUnknown, resolutionLower, "GRAPH_TOPOLOGY_UNOBSERVED", &rawUnknown{Stage: operation.Stage, Step: "graph-reconstruct", Reason: "GRAPH_TOPOLOGY_UNOBSERVED"}
	}
	for _, request := range item.Requests {
		declaration, ok := declarationByID(model.declarations, request.ValueID)
		if !ok || declaration.Kind != request.Kind || declaration.Operation != request.Operation || declaration.Target != request.Target {
			return decisionDeny, resolutionExact, "CAPABILITY_NOT_DECLARED", nil
		}
	}
	if kind := requestedEffectKind(item); kind != "" {
		if !deniedToken(provider, kind) {
			return decisionUnknown, resolutionLower, "CAPABILITY_ENFORCEMENT_NOT_IMPLEMENTED", &rawUnknown{Stage: operation.Stage, Step: "authorize-before-expand", Reason: "CAPABILITY_ENFORCEMENT_NOT_IMPLEMENTED"}
		}
		return decisionDeny, resolutionExact, "CAPABILITY_TOKEN_DENIED", nil
	}
	for _, request := range item.Requests {
		declaration, _ := declarationByID(model.declarations, request.ValueID)
		if len(evidenceFor(provider, model.declarations, []rawCapability{request})) == 0 {
			return decisionUnknown, resolutionLower, "EVIDENCE_UNOBSERVED", &rawUnknown{Stage: operation.Stage, Step: "bind-capability-evidence", Reason: "EVIDENCE_UNOBSERVED"}
		}
		_ = declaration
	}
	if model.policy.AuthorizationMode != "exact-current" {
		return decisionDeny, resolutionExact, "POLICY_REJECTED", nil
	}
	return decisionAllow, resolutionExact, "CAPABILITY_SCOPE_EXACT", nil
}

func authorizationOperation(model sourceModel) rawOperation {
	for _, operation := range model.operations {
		if operation.ID == "authorize-before-expand" {
			return operation
		}
	}
	return rawOperation{Stage: "UNKNOWN", Step: "UNKNOWN"}
}

func deniedToken(provider rawProvider, kind string) bool {
	for _, token := range provider.TokenAttempts {
		if token.Kind == kind && token.Requested && token.Decision == decisionDeny && !token.Issued {
			return true
		}
	}
	return false
}

func requestedEffectKind(item rawCase) string {
	if item.RequestedRepositoryWrites != 0 {
		return "file"
	}
	if item.RequestedMutationAuthority {
		return "mutation"
	}
	if item.RequestedPromotionAuthority {
		return "promotion"
	}
	return ""
}

func stateForDecision(decision string) string {
	if decision == decisionAllow {
		return claimDischarged
	}
	if decision == decisionDeny {
		return claimRefuted
	}
	return claimOpen
}

func knownState(decision string) string {
	if decision == decisionUnknown {
		return claimOpen
	}
	return claimDischarged
}

func tokenState(token rawToken) string {
	if token.Issued {
		return claimDischarged
	}
	if token.Decision == decisionDeny {
		return claimRefuted
	}
	return claimOpen
}

func effectState(provider rawProvider) string {
	if provider.BrokerTokenRequests == 3 && provider.BrokerTokenDenials == 3 && provider.BrokerTokensIssued == 0 {
		return claimDischarged
	}
	return claimOpen
}

func currentEvidenceCount(provider rawProvider, declarations []rawDeclaration) int {
	count := 0
	for _, declaration := range declarations {
		if len(evidenceFor(provider, declarations, []rawCapability{{ValueID: declaration.ValueID, Kind: declaration.Kind, Operation: declaration.Operation, Target: declaration.Target}})) == 1 {
			count++
		}
	}
	return count
}

func declaredCount(declarations []rawDeclaration, requests []rawCapability) int {
	count := 0
	for _, request := range requests {
		declaration, ok := declarationByID(declarations, request.ValueID)
		if ok && declaration.Kind == request.Kind && declaration.Operation == request.Operation && declaration.Target == request.Target {
			count++
		}
	}
	return count
}

func declarationByID(declarations []rawDeclaration, id string) (rawDeclaration, bool) {
	for _, declaration := range declarations {
		if declaration.ValueID == id {
			return declaration, true
		}
	}
	return rawDeclaration{}, false
}

func findCase(cases []rawCase, id string) (rawCase, bool) {
	for _, item := range cases {
		if item.ID == id {
			return item, true
		}
	}
	return rawCase{}, false
}

func snapshotRepository(root string) (rawSnapshot, error) {
	raw, err := exec.Command("git", "-C", root, "ls-files", "--cached", "--others", "--exclude-standard", "-z").Output()
	if err != nil {
		return rawSnapshot{}, err
	}
	entries := make([]string, 0)
	for _, relative := range strings.Split(string(raw), "\x00") {
		if relative == "" {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return rawSnapshot{}, err
		}
		entries = append(entries, filepath.ToSlash(relative)+"="+digestBytes(contents))
	}
	sort.Strings(entries)
	return snapshot("repository", root, entries), nil
}

func snapshotSandbox(root string) (rawSnapshot, error) {
	files, err := os.ReadDir(root)
	if err != nil {
		return rawSnapshot{}, err
	}
	entries := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(root, file.Name()))
		if err != nil {
			return rawSnapshot{}, err
		}
		entries = append(entries, file.Name()+"="+digestBytes(contents))
	}
	sort.Strings(entries)
	return snapshot("sandbox", root, entries), nil
}

func snapshot(scope, root string, entries []string) rawSnapshot {
	return rawSnapshot{Scope: scope, Root: root, Entries: entries, Digest: digestBytes([]byte(strings.Join(entries, "\n")))}
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestReceipt(raw []byte) string {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	object, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	object["report_digest"] = ""
	normalized, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return digestBytes(normalized)
}
