package compatibilitypolicy

import (
	"errors"
	"fmt"
	"go/format"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	generatedpolicy "github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy/generated"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	EvaluatorSchema       = "gooo/compiler-compatibility-policy-evaluator/v1"
	PolicyActivity        = "PublishOperationReceipt"
	CompatibilityHead     = "compatibility-policy:v1"
	DecisionClosed        = "CLOSED"
	DecisionUnknown       = "UNKNOWN"
	DecisionRefuted       = "REFUTED"
	AxisCount             = 7
	TransitionCount       = 8
	ContinuityEdgeCount   = 7
	EvidenceArtifactCount = 24
	TestContract          = "go1.27.0|go test ./...|go test -race ./..."
	DefaultMode           = "strict-default"
	OptInMode             = "caller-owned-immutable-successor-certificate"

	CaseStrictExactReplay              = "STRICT_EXACT_REPLAY"
	CaseBoundedImplementationSuccessor = "BOUNDED_IMPLEMENTATION_SUCCESSOR_REPLAY"
	CaseMissingSuccessorReplay         = "MISSING_SUCCESSOR_REPLAY"
	CaseUnboundedCompatibilityScope    = "UNBOUNDED_COMPATIBILITY_SCOPE"
	CaseSemanticPolicyOutputMismatch   = "SEMANTIC_POLICY_OUTPUT_MISMATCH"
	CaseTamperedWidenedCertificate     = "TAMPERED_WIDENED_CERTIFICATE"
)

var axisNames = [...]string{
	"SEMANTIC_IDENTITY",
	"COMPILER_IMPLEMENTATION_IDENTITY",
	"GO_TOOLCHAIN_IDENTITY",
	"POLICY_IDENTITY",
	"GENERATED_ARTIFACT_IDENTITY",
	"TEST_CONTRACT_IDENTITY",
	"AUTHORIZATION_IDENTITY",
}

var caseIDs = [...]string{
	CaseStrictExactReplay,
	CaseBoundedImplementationSuccessor,
	CaseMissingSuccessorReplay,
	CaseUnboundedCompatibilityScope,
	CaseSemanticPolicyOutputMismatch,
	CaseTamperedWidenedCertificate,
}

var expectedTransitions = [...]Transition{
	{Mode: "STRICT_DEFAULT", Condition: "ALL_AXES_EQUAL", Decision: DecisionClosed},
	{Mode: "STRICT_DEFAULT", Condition: "ANY_AXIS_MISMATCH", Decision: DecisionRefuted},
	{Mode: "OPT_IN", Condition: "ALL_AXES_EQUAL", Decision: DecisionClosed},
	{Mode: "OPT_IN", Condition: "COMPILER_IMPLEMENTATION_ONLY_DIFFERENCE_CERTIFIED_REPLAY", Decision: DecisionClosed},
	{Mode: "OPT_IN", Condition: "MISSING_REPLAY_OR_UNBOUNDED_SCOPE", Decision: DecisionUnknown},
	{Mode: "OPT_IN", Condition: "AMBIGUOUS_AXIS_OR_MISSING_SUCCESSOR_EVIDENCE", Decision: DecisionUnknown},
	{Mode: "OPT_IN", Condition: "OTHER_AXIS_MISMATCH", Decision: DecisionRefuted},
	{Mode: "OPT_IN", Condition: "TAMPERED_OR_SCOPE_WIDENED", Decision: DecisionRefuted},
}

type Transition struct {
	Mode      string `json:"mode"`
	Condition string `json:"condition"`
	Decision  string `json:"decision"`
}

type DecisionCase struct {
	ID       string `json:"id"`
	Decision string `json:"decision"`
}

type Policy struct {
	SourceDigest      string
	EvaluatorDigest   string
	ActivityCount     int
	Mode              string
	OptIn             string
	AxisCount         int
	Axes              []string
	Transitions       []Transition
	Cases             []DecisionCase
	TestContract      string
	ContinuityEdges   int
	EvidenceArtifacts int
}

func AxisNames() []string { return append([]string(nil), axisNames[:]...) }

func CaseIDs() []string { return append([]string(nil), caseIDs[:]...) }

func GeneratedEvaluatorDigest() string { return generatedpolicy.EvaluatorDigest }

func GeneratedEvaluatorCaseIDs() []string { return generatedpolicy.CaseIDs() }

func Evaluate(caseID string) (string, bool) { return generatedpolicy.Evaluate(caseID) }

func (policy Policy) Decision(caseID string) (string, bool) { return generatedpolicy.Evaluate(caseID) }

func CompileNamed(filename string, source []byte) (Policy, error) {
	if len(source) == 0 {
		return Policy{}, errors.New("compiler compatibility policy source is empty")
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return Policy{}, fmt.Errorf("parse compiler compatibility policy: %w", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Policy{}, fmt.Errorf("lower compiler compatibility policy: %w", err)
	}
	if ir.Package != "selfimprovementobservation" || ir.Namespace.String() != "self_improvement_observation" {
		return Policy{}, fmt.Errorf("compiler compatibility policy package/namespace is %q/%q", ir.Package, ir.Namespace)
	}
	var policy Policy
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.Name != PolicyActivity {
			continue
		}
		policy.ActivityCount++
		parsed, parseErr := parseMarkers(node.ValueProgram)
		if parseErr != nil {
			return Policy{}, fmt.Errorf("activity %q: %w", PolicyActivity, parseErr)
		}
		policy.Axes = parsed.axes
		policy.Transitions = parsed.transitions
		policy.Cases = parsed.cases
		policy.Mode = parsed.mode
		policy.OptIn = parsed.optIn
		policy.AxisCount = parsed.axisCount
		policy.TestContract = parsed.testContract
		policy.ContinuityEdges = parsed.continuityEdges
		policy.EvidenceArtifacts = parsed.evidenceArtifacts
	}
	if policy.ActivityCount != 1 {
		return Policy{}, fmt.Errorf("compiler compatibility policy activity count = %d, want 1", policy.ActivityCount)
	}
	policy.SourceDigest = cache.HashBytes(source).String()
	policy.EvaluatorDigest = evaluatorDigest(policy.Cases)
	return policy, nil
}

func GenerateNamed(filename string, source []byte) (Policy, []byte, error) {
	policy, err := CompileNamed(filename, source)
	if err != nil {
		return Policy{}, nil, err
	}
	generated, err := render(filename, policy)
	if err != nil {
		return Policy{}, nil, err
	}
	return policy, generated, nil
}

func Load(filename string, source []byte) (Policy, error) {
	policy, err := CompileNamed(filename, source)
	if err != nil {
		return Policy{}, err
	}
	if policy.SourceDigest != generatedpolicy.PolicySourceDigest || policy.EvaluatorDigest != generatedpolicy.EvaluatorDigest {
		return Policy{}, errors.New("compiler compatibility policy and generated evaluator digest differ")
	}
	if policy.AxisCount != AxisCount || !sameStrings(policy.Axes, AxisNames()) || len(policy.Transitions) != TransitionCount || !sameTransitions(policy.Transitions, expectedTransitions[:]) ||
		policy.Mode != DefaultMode || policy.OptIn != OptInMode || policy.TestContract != TestContract ||
		policy.ContinuityEdges != ContinuityEdgeCount || policy.EvidenceArtifacts != EvidenceArtifactCount {
		return Policy{}, errors.New("compiler compatibility policy identity or transition table is not canonical")
	}
	ids := generatedpolicy.CaseIDs()
	if !sameStrings(ids, CaseIDs()) || len(policy.Cases) != len(ids) {
		return Policy{}, errors.New("compiler compatibility policy and generated evaluator cases differ")
	}
	for index, row := range policy.Cases {
		if row.ID != ids[index] {
			return Policy{}, fmt.Errorf("compiler compatibility evaluator case %q is not source-bound", row.ID)
		}
		decision, ok := generatedpolicy.Evaluate(row.ID)
		if !ok || decision != row.Decision {
			return Policy{}, fmt.Errorf("compiler compatibility evaluator decision for %q is not source-bound", row.ID)
		}
	}
	return policy, nil
}

type parsedMarkers struct {
	mode              string
	optIn             string
	axisCount         int
	axes              []string
	transitions       []Transition
	cases             []DecisionCase
	testContract      string
	continuityEdges   int
	evidenceArtifacts int
}

func parseMarkers(value string) (parsedMarkers, error) {
	var parsed parsedMarkers
	seenHead := false
	for part := range strings.SplitSeq(value, ";") {
		switch {
		case part == CompatibilityHead:
			if seenHead {
				return parsedMarkers{}, errors.New("compiler compatibility policy header is duplicated")
			}
			seenHead = true
		case strings.HasPrefix(part, "compatibility-mode="):
			parsed.mode = strings.TrimPrefix(part, "compatibility-mode=")
		case strings.HasPrefix(part, "compatibility-opt-in="):
			parsed.optIn = strings.TrimPrefix(part, "compatibility-opt-in=")
		case strings.HasPrefix(part, "compatibility-axis-count="):
			if _, err := fmt.Sscanf(strings.TrimPrefix(part, "compatibility-axis-count="), "%d", &parsed.axisCount); err != nil {
				return parsedMarkers{}, fmt.Errorf("identity axis count %q is malformed", part)
			}
		case strings.HasPrefix(part, "compatibility-axis="):
			parsed.axes = append(parsed.axes, strings.TrimPrefix(part, "compatibility-axis="))
		case strings.HasPrefix(part, "compatibility-transition="):
			encoded := strings.TrimPrefix(part, "compatibility-transition=")
			pieces := strings.Split(encoded, "|")
			if len(pieces) != 3 || pieces[0] == "" || pieces[1] == "" || pieces[2] == "" {
				return parsedMarkers{}, fmt.Errorf("transition %q is malformed", encoded)
			}
			parsed.transitions = append(parsed.transitions, Transition{Mode: pieces[0], Condition: pieces[1], Decision: pieces[2]})
		case strings.HasPrefix(part, "compatibility-test-contract="):
			parsed.testContract = strings.TrimPrefix(part, "compatibility-test-contract=")
		case strings.HasPrefix(part, "compatibility-continuity-edges="):
			if _, err := fmt.Sscanf(strings.TrimPrefix(part, "compatibility-continuity-edges="), "%d", &parsed.continuityEdges); err != nil {
				return parsedMarkers{}, fmt.Errorf("continuity edges %q are malformed", part)
			}
		case strings.HasPrefix(part, "compatibility-evidence-artifacts="):
			if _, err := fmt.Sscanf(strings.TrimPrefix(part, "compatibility-evidence-artifacts="), "%d", &parsed.evidenceArtifacts); err != nil {
				return parsedMarkers{}, fmt.Errorf("evidence artifact count %q is malformed", part)
			}
		case strings.HasPrefix(part, "compatibility-case="):
			encoded := strings.TrimPrefix(part, "compatibility-case=")
			id, decision, ok := strings.Cut(encoded, ":")
			if !ok || id == "" || decision == "" || strings.Contains(decision, ":") {
				return parsedMarkers{}, fmt.Errorf("case %q is malformed", encoded)
			}
			parsed.cases = append(parsed.cases, DecisionCase{ID: id, Decision: decision})
		}
	}
	if !seenHead {
		return parsedMarkers{}, errors.New("compiler compatibility policy header is missing")
	}
	if parsed.mode != DefaultMode || parsed.optIn != OptInMode {
		return parsedMarkers{}, errors.New("compiler compatibility policy mode is not strict-default with explicit opt-in")
	}
	if parsed.axisCount != AxisCount || !sameStrings(parsed.axes, AxisNames()) || !sameTransitions(parsed.transitions, expectedTransitions[:]) || !sameStrings(caseIDsFrom(parsed.cases), CaseIDs()) {
		return parsedMarkers{}, errors.New("compiler compatibility policy markers do not match the fixed contract")
	}
	if !sameStrings(decisionsFrom(parsed.cases), []string{DecisionClosed, DecisionClosed, DecisionUnknown, DecisionUnknown, DecisionRefuted, DecisionRefuted}) {
		return parsedMarkers{}, errors.New("compiler compatibility policy case decisions are not 2/2/2")
	}
	return parsed, nil
}

func evaluatorDigest(rows []DecisionCase) string {
	var builder strings.Builder
	builder.WriteString(EvaluatorSchema)
	builder.WriteByte('\n')
	for _, row := range rows {
		builder.WriteString(row.ID)
		builder.WriteByte('=')
		builder.WriteString(row.Decision)
		builder.WriteByte('\n')
	}
	return cache.HashBytes([]byte(builder.String())).String()
}

func render(filename string, policy Policy) ([]byte, error) {
	var builder strings.Builder
	fmt.Fprintln(&builder, "// Code generated by gooo compiler compatibility policy generator; DO NOT EDIT.")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "package compatibilitypolicygenerated")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "const (")
	fmt.Fprintf(&builder, "\tEvaluatorSchema = %q\n", EvaluatorSchema)
	fmt.Fprintf(&builder, "\tPolicySourcePath = %q\n", filename)
	fmt.Fprintf(&builder, "\tPolicySourceDigest = %q\n", policy.SourceDigest)
	fmt.Fprintf(&builder, "\tEvaluatorDigest = %q\n", policy.EvaluatorDigest)
	fmt.Fprintln(&builder, ")")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "type DecisionCase struct {")
	fmt.Fprintln(&builder, "\tID string")
	fmt.Fprintln(&builder, "\tDecision string")
	fmt.Fprintln(&builder, "}")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "var cases = [...]DecisionCase{")
	for _, row := range policy.Cases {
		fmt.Fprintf(&builder, "\t{ID: %q, Decision: %q},\n", row.ID, row.Decision)
	}
	fmt.Fprintln(&builder, "}")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "func CaseIDs() []string {")
	fmt.Fprintln(&builder, "\tresult := make([]string, 0, len(cases))")
	fmt.Fprintln(&builder, "\tfor _, row := range cases {")
	fmt.Fprintln(&builder, "\t\tresult = append(result, row.ID)")
	fmt.Fprintln(&builder, "\t}")
	fmt.Fprintln(&builder, "\treturn result")
	fmt.Fprintln(&builder, "}")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "func Evaluate(caseID string) (string, bool) {")
	fmt.Fprintln(&builder, "\tswitch caseID {")
	for _, decision := range []string{DecisionClosed, DecisionUnknown, DecisionRefuted} {
		var ids []string
		for _, row := range policy.Cases {
			if row.Decision == decision {
				ids = append(ids, row.ID)
			}
		}
		if len(ids) == 0 {
			continue
		}
		fmt.Fprint(&builder, "\tcase ")
		for index, id := range ids {
			if index > 0 {
				fmt.Fprint(&builder, ", ")
			}
			fmt.Fprintf(&builder, "%q", id)
		}
		fmt.Fprintf(&builder, ":\n\t\treturn %q, true\n", decision)
	}
	fmt.Fprintln(&builder, "\tdefault:")
	fmt.Fprintln(&builder, "\t\treturn \"\", false")
	fmt.Fprintln(&builder, "\t}")
	fmt.Fprintln(&builder, "}")
	return format.Source([]byte(builder.String()))
}

func sameStrings(left, right []string) bool {
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

func sameTransitions(left, right []Transition) bool {
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

func caseIDsFrom(rows []DecisionCase) []string {
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ID)
	}
	return result
}

func decisionsFrom(rows []DecisionCase) []string {
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.Decision)
	}
	return result
}
