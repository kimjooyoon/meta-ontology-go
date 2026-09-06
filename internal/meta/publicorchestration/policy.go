package publicorchestration

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	PolicySchema        = "gooo/public-self-improvement-orchestration-policy/v1"
	EvaluatorSchema     = "gooo/public-self-improvement-orchestration-evaluator/v1"
	PolicyContract      = "v1"
	PolicyName          = "EXPLICIT_AUTHORIZED_PUBLIC_SELF_IMPROVEMENT"
	Operation           = "gooo.self-improvement.public-orchestration"
	PolicyActivity      = "Compile"
	ArtifactDenominator = 24

	DecisionClosed  = "CLOSED"
	DecisionUnknown = "UNKNOWN"
	DecisionRefuted = "REFUTED"

	CaseAuthorizedOrchestration = "AUTHORIZED_ORCHESTRATION"
	CaseAuthorizedReceiptReuse  = "AUTHORIZED_TEST_RECEIPT_REUSE"
	CaseMissingAuthorization    = "MISSING_EXPLICIT_AUTHORIZATION"
	CaseMalformedContinuation   = "MALFORMED_CONTINUATION"
	CaseContradictoryCandidate  = "CONTRADICTORY_CANDIDATE"
	CaseMismatchedAuthorization = "MISMATCHED_AUTHORIZATION"

	UnknownClassIncomplete = "INCOMPLETE_EVIDENCE"
	UnknownStageAuthorize  = "AUTHORIZE"
	UnknownStepAuthorize   = "EXPLICIT_AUTHORIZATION"
	UnknownStageResume     = "RESUME"
	UnknownStepResume      = "CONTINUATION_ARTIFACT"

	UnknownReasonAuthorization = "MISSING_EXPLICIT_AUTHORIZATION"
	UnknownReasonContinuation  = "MALFORMED_CONTINUATION_ARTIFACT"
	UnknownNextAuthorization   = "PROVIDE_EXPLICIT_AUTHORIZATION"
	UnknownNextContinuation    = "REPAIR_CONTINUATION_ARTIFACT"
)

var (
	canonicalStates   = []string{"DISCOVER", "PROPOSE", "AUTHORIZE", "CERTIFY", "GENERATE", "VALIDATE", "REUSE", "EVIDENCE"}
	canonicalBindings = []string{
		"canonical-source", "canonical-semantic", "candidate", "authorization", "certificate",
		"generated-output", "generated-semantic", "generated-manifest", "generated-test", "receipt",
		"toolchain", "test-contract",
	}
	canonicalJourney = []string{
		"semantic-discovery", "proposal", "explicit-authorization", "durable-certificate",
		"ordinary-generate", "real-project-validation", "immutable-test-receipt-reuse", "evidence",
	}
	canonicalHandoffArtifacts = []string{"proposal", "authorization", "certificate", "test-receipt"}
	tokenPattern              = regexp.MustCompile(`^[A-Za-z0-9_.:/_-]+$`)
)

type Metrics struct {
	PublicCLIInvocations   int   `json:"public_cli_invocations"`
	ExplicitHumanDecisions int   `json:"explicit_human_decisions"`
	SemanticOperations     int   `json:"semantic_operations"`
	LoweringOperations     int   `json:"lowering_operations"`
	GenerationOperations   int   `json:"generation_operations"`
	TestOperations         int   `json:"test_operations"`
	HandoffArtifacts       int   `json:"handoff_artifacts"`
	ReusedTestExecutions   int   `json:"reused_test_executions"`
	WallMS                 int64 `json:"wall_ms"`
	PeakRSSKib             int64 `json:"peak_rss_kib"`
}

type Transition struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Event string `json:"event"`
}

type Case struct {
	ID            string   `json:"id"`
	Decision      string   `json:"decision"`
	Stage         string   `json:"stage"`
	Step          int      `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type CaseResult struct {
	ID               string        `json:"id"`
	ExpectedDecision string        `json:"expected_decision"`
	ObservedDecision string        `json:"observed_decision"`
	Reason           string        `json:"reason"`
	Unknown          *UnknownState `json:"unknown"`
}

type Policy struct {
	Schema           string       `json:"schema"`
	EvaluatorSchema  string       `json:"evaluator_schema"`
	SourceDigest     string       `json:"source_digest"`
	SemanticDigest   string       `json:"semantic_digest"`
	EvaluatorDigest  string       `json:"evaluator_digest"`
	Package          string       `json:"package"`
	Namespace        string       `json:"namespace"`
	Activity         string       `json:"activity"`
	Name             string       `json:"name"`
	Operation        string       `json:"operation"`
	Bindings         []string     `json:"bindings"`
	States           []string     `json:"states"`
	Transitions      []Transition `json:"transitions"`
	Boundary         string       `json:"authorization_boundary"`
	PreparePath      []string     `json:"prepare_path"`
	ResumePath       []string     `json:"resume_path"`
	Journey          []string     `json:"journey"`
	HandoffArtifacts []string     `json:"handoff_artifacts"`
	Before           Metrics      `json:"before"`
	After            Metrics      `json:"after"`
	Cases            []Case       `json:"cases"`
	ArtifactDenom    int          `json:"artifact_denominator"`
	PerformanceRule  string       `json:"performance_rule"`
}

type UnknownState struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type Inventory struct {
	RegularFiles  int `json:"regular_files"`
	PhysicalLines int `json:"physical_lines"`
	GoFiles       int `json:"go_files"`
	GoBytes       int `json:"go_bytes"`
	GoLines       int `json:"go_lines"`
	GoooFiles     int `json:"gooo_files"`
	GoooLines     int `json:"gooo_lines"`
}

type GeneratedInventory struct {
	Files         int `json:"files"`
	GoFiles       int `json:"go_files"`
	GoBytes       int `json:"go_bytes"`
	GoLines       int `json:"go_lines"`
	ManifestBytes int `json:"manifest_bytes"`
}

type Comparisons struct {
	GeneratedBytesEqual     bool `json:"generated_bytes_equal"`
	GeneratedSemanticEqual  bool `json:"generated_semantic_equal"`
	TestContractBytesEqual  bool `json:"test_contract_bytes_equal"`
	ReceiptBindingEqual     bool `json:"receipt_binding_equal"`
	ContinuityPreserved     bool `json:"continuity_preserved"`
	SafetyOutcomesPreserved bool `json:"safety_outcomes_preserved"`
}

type Report struct {
	Schema                string             `json:"schema"`
	Decision              string             `json:"decision"`
	Reason                string             `json:"reason"`
	CaseID                string             `json:"case_id"`
	Unknown               *UnknownState      `json:"unknown"`
	PolicySourceDigest    string             `json:"policy_source_digest"`
	PolicySemanticDigest  string             `json:"policy_semantic_digest"`
	PolicyEvaluatorDigest string             `json:"policy_evaluator_digest"`
	Operation             string             `json:"operation"`
	StatePath             []string           `json:"state_path"`
	Boundary              string             `json:"authorization_boundary"`
	Before                Metrics            `json:"before"`
	After                 Metrics            `json:"after"`
	Input                 Inventory          `json:"input"`
	Generated             GeneratedInventory `json:"generated"`
	Comparisons           Comparisons        `json:"comparisons"`
	Cases                 []CaseResult       `json:"cases"`
	CaseDenominator       int                `json:"case_denominator"`
	ClosedCases           int                `json:"closed_cases"`
	UnknownCases          int                `json:"unknown_cases"`
	RefutedCases          int                `json:"refuted_cases"`
	ArtifactDenominator   int                `json:"artifact_denominator"`
	ArtifactCount         int                `json:"artifact_count"`
	RepositoryWrites      int                `json:"repository_writes"`
	LocalTestExecutions   int                `json:"local_test_executions"`
	RuntimeComparable     bool               `json:"runtime_comparable"`
	RuntimeUnknown        *UnknownState      `json:"runtime_unknown"`
	HandoffDigest         string             `json:"handoff_digest,omitempty"`
	AuthorizationDigest   string             `json:"authorization_digest,omitempty"`
	CertificateDigest     string             `json:"certificate_digest,omitempty"`
	ReceiptDigest         string             `json:"receipt_digest,omitempty"`
	NoAggregateScore      bool               `json:"no_aggregate_score"`
}

type markerSet map[string][]string

func CanonicalCaseIDs() []string {
	return []string{CaseAuthorizedOrchestration, CaseAuthorizedReceiptReuse, CaseMissingAuthorization, CaseMalformedContinuation, CaseContradictoryCandidate, CaseMismatchedAuthorization}
}

func CanonicalStates() []string { return append([]string(nil), canonicalStates...) }

func RequiredBindings() []string { return append([]string(nil), canonicalBindings...) }

func CanonicalJourney() []string { return append([]string(nil), canonicalJourney...) }

func HandoffArtifactNames() []string { return append([]string(nil), canonicalHandoffArtifacts...) }

func (policy Policy) Decision(caseID string) (string, bool) {
	for _, item := range policy.Cases {
		if item.ID == caseID {
			return item.Decision, true
		}
	}
	return "", false
}

func (policy Policy) CaseFor(caseID string) (Case, bool) {
	for _, item := range policy.Cases {
		if item.ID == caseID {
			return item, true
		}
	}
	return Case{}, false
}

func (policy Policy) UnknownFor(caseID string) *UnknownState {
	item, ok := policy.CaseFor(caseID)
	if !ok || item.Decision != DecisionUnknown {
		return nil
	}
	return &UnknownState{Stage: item.Stage, Step: strconv.Itoa(item.Step), Reason: item.Reason, UnknownClass: item.UnknownClass, NextOperation: item.NextOperation, BlockedBy: append([]string(nil), item.BlockedBy...)}
}

func (policy Policy) Transition(from, event string) (Transition, bool) {
	for _, item := range policy.Transitions {
		if item.From == from && item.Event == event {
			return item, true
		}
	}
	return Transition{}, false
}

func (policy Policy) EventsFor(path []string) ([]string, error) {
	if len(path) < 2 {
		return nil, errors.New("orchestration path must contain at least two states")
	}
	events := make([]string, 0, len(path)-1)
	for index := 0; index < len(path)-1; index++ {
		var found Transition
		matches := 0
		for _, item := range policy.Transitions {
			if item.From == path[index] && item.To == path[index+1] {
				found = item
				matches++
			}
		}
		if matches != 1 {
			return nil, fmt.Errorf("orchestration path edge %s>%s has %d transitions", path[index], path[index+1], matches)
		}
		events = append(events, found.Event)
	}
	return events, nil
}

func Load(filename string, source []byte) (Policy, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return Policy{}, errors.New("public orchestration source has syntax diagnostics")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Policy{}, fmt.Errorf("lower public orchestration source: %w", err)
	}
	if ir.Package != "publicdiscoveryexample" || ir.Namespace.String() != "public_discovery_example" {
		return Policy{}, errors.New("public orchestration is not bound to the canonical public discovery project")
	}
	var activity semantic.Node
	count := 0
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity && node.Name == PolicyActivity {
			activity = node
			count++
		}
	}
	if count != 1 || activity.ValueProgram == "" {
		return Policy{}, errors.New("public orchestration Compile activity is missing or ambiguous")
	}
	markers := parseMarkers(activity.ValueProgram)
	if firstMarker(markers, "orchestration-contract") != PolicyContract {
		return Policy{}, errors.New("public orchestration contract marker is missing")
	}
	policy := Policy{
		Schema: PolicySchema, EvaluatorSchema: EvaluatorSchema,
		SourceDigest: cache.HashBytes(source).String(), SemanticDigest: ir.StableHash(),
		Package: ir.Package, Namespace: ir.Namespace.String(), Activity: PolicyActivity,
		Name: firstMarker(markers, "orchestration-policy"), Operation: firstMarker(markers, "orchestration-operation"),
		Bindings:         splitMarker(firstMarker(markers, "orchestration-bindings"), ","),
		States:           splitMarker(firstMarker(markers, "orchestration-states"), ">"),
		Boundary:         firstMarker(markers, "orchestration-boundary"),
		PreparePath:      splitMarker(firstMarker(markers, "orchestration-prepare-path"), ">"),
		ResumePath:       splitMarker(firstMarker(markers, "orchestration-resume-path"), ">"),
		Journey:          splitMarker(firstMarker(markers, "orchestration-journey"), ">"),
		HandoffArtifacts: splitMarker(firstMarker(markers, "orchestration-handoff-artifacts"), ","),
		Before:           parseMetrics(firstMarker(markers, "orchestration-before")),
		After:            parseMetrics(firstMarker(markers, "orchestration-after")),
		Cases:            parseCases(markers), ArtifactDenom: parseInt(firstMarker(markers, "orchestration-artifact-denominator")),
		PerformanceRule: firstMarker(markers, "orchestration-performance-rule"),
	}
	for index := 1; ; index++ {
		encoded, ok := marker(markers, fmt.Sprintf("orchestration-transition-%d", index))
		if !ok {
			break
		}
		fields := strings.Split(encoded, ">")
		if len(fields) != 3 || !safeToken(fields[0]) || !safeToken(fields[1]) || !safeToken(fields[2]) {
			return Policy{}, fmt.Errorf("malformed orchestration transition %q", encoded)
		}
		policy.Transitions = append(policy.Transitions, Transition{From: fields[0], To: fields[1], Event: fields[2]})
	}
	policy.EvaluatorDigest = evaluatorDigest(policy)
	if err := policy.Validate(); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func (policy Policy) Validate() error {
	if policy.Schema != PolicySchema || policy.EvaluatorSchema != EvaluatorSchema || policy.Name != PolicyName || policy.Operation != Operation ||
		policy.Activity != PolicyActivity || policy.Package != "publicdiscoveryexample" || policy.Namespace != "public_discovery_example" ||
		policy.Boundary != "AUTHORIZE" || policy.ArtifactDenom != ArtifactDenominator || policy.PerformanceRule != "UNKNOWN_WHEN_RUNTIME_MODES_ARE_NOT_EQUIVALENT" {
		return errors.New("public orchestration policy identity is invalid")
	}
	if !sameValues(policy.Bindings, canonicalBindings) || !sameValues(policy.States, canonicalStates) || !sameValues(policy.Journey, canonicalJourney) || !sameValues(policy.HandoffArtifacts, canonicalHandoffArtifacts) {
		return errors.New("public orchestration policy lists are not canonical")
	}
	wantBefore := Metrics{PublicCLIInvocations: 15, ExplicitHumanDecisions: 2, SemanticOperations: 3, LoweringOperations: 3, GenerationOperations: 3, TestOperations: 1, HandoffArtifacts: 5}
	wantAfter := Metrics{PublicCLIInvocations: 4, ExplicitHumanDecisions: 2, SemanticOperations: 3, LoweringOperations: 3, GenerationOperations: 3, TestOperations: 1, HandoffArtifacts: 5, ReusedTestExecutions: 1}
	if policy.Before != wantBefore || policy.After != wantAfter {
		return errors.New("public orchestration utility metrics are not the frozen source contract")
	}
	if len(policy.Transitions) != len(canonicalStates)-1 || len(policy.PreparePath) != 3 || len(policy.ResumePath) != len(canonicalStates)-2 {
		return errors.New("public orchestration transition or path denominator changed")
	}
	if policy.PreparePath[len(policy.PreparePath)-1] != policy.Boundary || policy.ResumePath[0] != policy.Boundary || policy.ResumePath[len(policy.ResumePath)-1] != "EVIDENCE" {
		return errors.New("public orchestration authorization boundary is not on both paths")
	}
	seenTransitions := make(map[string]struct{}, len(policy.Transitions))
	for _, item := range policy.Transitions {
		if !contains(policy.States, item.From) || !contains(policy.States, item.To) || !safeToken(item.Event) {
			return fmt.Errorf("orchestration transition %q is outside the state machine", item.Event)
		}
		key := item.From + "\x00" + item.To + "\x00" + item.Event
		if _, exists := seenTransitions[key]; exists {
			return errors.New("public orchestration transition is duplicated")
		}
		seenTransitions[key] = struct{}{}
	}
	if len(policy.Cases) != len(CanonicalCaseIDs()) {
		return errors.New("public orchestration case denominator changed")
	}
	seenCases := make(map[string]struct{}, len(policy.Cases))
	counts := make(map[string]int)
	for index, want := range CanonicalCaseIDs() {
		item := policy.Cases[index]
		if item.ID != want || !knownDecision(item.Decision) || !safeToken(item.Stage) || item.Step < 1 || item.Reason == "" {
			return fmt.Errorf("public orchestration case %d is malformed", index+1)
		}
		if _, exists := seenCases[item.ID]; exists {
			return fmt.Errorf("public orchestration case %q is duplicated", item.ID)
		}
		seenCases[item.ID] = struct{}{}
		counts[item.Decision]++
		if item.Decision == DecisionUnknown {
			if !safeToken(item.UnknownClass) || !safeToken(item.NextOperation) || len(item.BlockedBy) == 0 {
				return fmt.Errorf("UNKNOWN orchestration case %q omits causal metadata", item.ID)
			}
		} else if item.UnknownClass != "" || item.NextOperation != "" || len(item.BlockedBy) != 0 {
			return fmt.Errorf("known orchestration case %q carries UNKNOWN metadata", item.ID)
		}
	}
	if counts[DecisionClosed] != 2 || counts[DecisionUnknown] != 2 || counts[DecisionRefuted] != 2 {
		return fmt.Errorf("public orchestration decisions are %v, want 2/2/2", counts)
	}
	if policy.Before.PublicCLIInvocations <= policy.After.PublicCLIInvocations || policy.After.ExplicitHumanDecisions < policy.Before.ExplicitHumanDecisions ||
		policy.Before.SemanticOperations != policy.After.SemanticOperations || policy.Before.LoweringOperations != policy.After.LoweringOperations ||
		policy.Before.GenerationOperations != policy.After.GenerationOperations || policy.Before.TestOperations != policy.After.TestOperations ||
		policy.Before.HandoffArtifacts != policy.After.HandoffArtifacts {
		return errors.New("orchestration utility contract does not preserve semantic work and decisions")
	}
	return nil
}

func EvaluatorDigest(policy Policy) string { return evaluatorDigest(policy) }

func evaluatorDigest(policy Policy) string {
	var builder strings.Builder
	builder.WriteString(policy.EvaluatorSchema)
	builder.WriteByte('\n')
	builder.WriteString(policy.Name)
	builder.WriteByte('\n')
	builder.WriteString(policy.Operation)
	builder.WriteByte('\n')
	for _, value := range append(append(append(append([]string{}, policy.Bindings...), policy.States...), policy.Journey...), policy.HandoffArtifacts...) {
		builder.WriteString(value)
		builder.WriteByte(',')
	}
	builder.WriteByte('\n')
	for _, item := range policy.Transitions {
		fmt.Fprintf(&builder, "%s>%s>%s\n", item.From, item.To, item.Event)
	}
	for _, item := range policy.Cases {
		fmt.Fprintf(&builder, "%s=%s=%s=%d=%s=%s=%s=%s\n", item.ID, item.Decision, item.Stage, item.Step, item.Reason, item.UnknownClass, item.NextOperation, strings.Join(item.BlockedBy, ","))
	}
	return cache.HashBytes([]byte(builder.String())).String()
}

func parseMarkers(value string) markerSet {
	markers := make(markerSet)
	for part := range strings.SplitSeq(value, ";") {
		key, item, ok := strings.Cut(part, "=")
		if ok && strings.TrimSpace(key) != "" {
			markers[strings.TrimSpace(key)] = append(markers[strings.TrimSpace(key)], strings.TrimSpace(item))
		}
	}
	return markers
}

func marker(markers markerSet, key string) (string, bool) {
	values := markers[key]
	if len(values) != 1 || values[0] == "" {
		return "", false
	}
	return values[0], true
}

func firstMarker(markers markerSet, key string) string {
	value, _ := marker(markers, key)
	return value
}

func splitMarker(value, separator string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, separator)
}

func parseInt(value string) int {
	result, err := strconv.Atoi(value)
	if err != nil || result < 0 {
		return 0
	}
	return result
}

func parseMetrics(value string) Metrics {
	var result Metrics
	for item := range strings.SplitSeq(value, ",") {
		key, raw, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}
		number := parseInt(raw)
		switch key {
		case "public_cli_invocations":
			result.PublicCLIInvocations = number
		case "explicit_human_decisions":
			result.ExplicitHumanDecisions = number
		case "semantic_operations":
			result.SemanticOperations = number
		case "lowering_operations":
			result.LoweringOperations = number
		case "generation_operations":
			result.GenerationOperations = number
		case "test_operations":
			result.TestOperations = number
		case "handoff_artifacts":
			result.HandoffArtifacts = number
		case "reused_test_executions":
			result.ReusedTestExecutions = number
		}
	}
	return result
}

func parseCases(markers markerSet) []Case {
	var result []Case
	for index := 1; ; index++ {
		value, ok := marker(markers, fmt.Sprintf("orchestration-case-%d", index))
		if !ok {
			break
		}
		fields := strings.Split(value, ":")
		if len(fields) != 8 {
			continue
		}
		item := Case{ID: fields[0], Decision: fields[1], Stage: fields[2], Step: parseInt(fields[3]), Reason: fields[4]}
		if fields[5] != "NONE" {
			item.UnknownClass = fields[5]
		}
		if fields[6] != "NONE" {
			item.NextOperation = fields[6]
		}
		if fields[7] != "NONE" {
			item.BlockedBy = strings.Split(fields[7], ",")
		}
		result = append(result, item)
	}
	return result
}

func sameValues(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func contains(values []string, target string) bool {
	return slices.Contains(values, target)
}

func safeToken(value string) bool { return tokenPattern.MatchString(value) }

func knownDecision(value string) bool {
	return value == DecisionClosed || value == DecisionUnknown || value == DecisionRefuted
}
