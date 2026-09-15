package publictestreuse

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	PolicySchema        = "gooo/public-test-reuse-policy/v1"
	EvaluatorSchema     = "gooo/public-test-reuse-policy-evaluator/v1"
	PolicyActivity      = "Compile"
	PolicyContract      = "v1"
	PolicyName          = "EXPLICIT_OPT_IN_IMMUTABLE_RECEIPT"
	Operation           = "gooo.test.generated-public-reuse"
	ArtifactDenominator = 24

	CaseBaselineExecution      = "BASELINE_TEST_EXECUTION"
	CaseAuthorizedReuse        = "AUTHORIZED_IMMUTABLE_RECEIPT_REUSE"
	CaseMissingAuthorization   = "MISSING_EXPLICIT_AUTHORIZATION"
	CaseStaleEvidence          = "STALE_OR_UNBOUNDED_REUSE_EVIDENCE"
	CaseTamperedReceipt        = "TAMPERED_REUSE_RECEIPT"
	CasePolicyMismatch         = "POLICY_MISMATCH"
	DecisionClosed             = "CLOSED"
	DecisionUnknown            = "UNKNOWN"
	DecisionRefuted            = "REFUTED"
	ReasonBaseline             = "SUCCESSFUL_ORIGINAL_TEST_EXECUTION"
	ReasonReuse                = "AUTHORIZED_IMMUTABLE_RECEIPT_REUSE"
	ReasonMissingAuthorization = "MISSING_EXPLICIT_REUSE_AUTHORIZATION"
	ReasonStale                = "STALE_REUSE_EVIDENCE"
	ReasonTampered             = "MALFORMED_OR_TAMPERED_REUSE_RECEIPT"
	ReasonPolicy               = "REUSE_POLICY_MISMATCH"
)

var requiredBindings = []string{
	"canonical-source", "canonical-semantic", "generated-output", "generated-semantic",
	"generated-manifest", "compiler", "released-tool", "toolchain", "test-command",
	"test-contract", "successful-result",
}

var canonicalCaseIDs = []string{
	CaseBaselineExecution, CaseAuthorizedReuse, CaseMissingAuthorization,
	CaseStaleEvidence, CaseTamperedReceipt, CasePolicyMismatch,
}

var canonicalJourney = []string{
	"baseline-generate", "baseline-build", "baseline-test", "write-receipt",
	"authorize", "replay-validate", "reused-test", "evidence",
}

type Case struct {
	ID       string
	Decision string
}

type Policy struct {
	Schema          string
	SourceDigest    string
	SemanticDigest  string
	EvaluatorDigest string
	Package         string
	Namespace       string
	Activity        string
	Name            string
	Bindings        []string
	Journey         []string
	ArtifactDenom   int
	Cases           []Case
}

func CanonicalCaseIDs() []string { return append([]string(nil), canonicalCaseIDs...) }

func CanonicalJourney() []string { return append([]string(nil), canonicalJourney...) }

func RequiredBindings() []string { return append([]string(nil), requiredBindings...) }

func (policy Policy) Decision(caseID string) (string, bool) {
	for _, item := range policy.Cases {
		if item.ID == caseID {
			return item.Decision, true
		}
	}
	return "", false
}

func Load(filename string, source []byte) (Policy, error) {
	if len(source) == 0 {
		return Policy{}, errors.New("public test reuse policy source is empty")
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return Policy{}, errors.New("public test reuse policy source has syntax diagnostics")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Policy{}, fmt.Errorf("lower public test reuse policy: %w", err)
	}
	if ir.Package != "publicdiscoveryexample" || ir.Namespace.String() != "public_discovery_example" {
		return Policy{}, errors.New("public test reuse policy is not bound to the canonical public discovery project")
	}
	var activity semantic.Node
	activityCount := 0
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity && node.Name == PolicyActivity {
			activity = node
			activityCount++
		}
	}
	if activityCount != 1 || activity.ValueProgram == "" {
		return Policy{}, errors.New("public test reuse policy activity is missing or ambiguous")
	}
	markers := parseMarkers(activity.ValueProgram)
	policy := Policy{
		Schema:         PolicySchema,
		SourceDigest:   cache.HashBytes(source).String(),
		SemanticDigest: ir.StableHash(),
		Package:        ir.Package,
		Namespace:      ir.Namespace.String(),
		Activity:       PolicyActivity,
		Name:           markers["test-reuse-policy"],
		Bindings:       strings.Split(markers["test-reuse-bindings"], ","),
		Journey:        strings.Split(markers["test-reuse-journey"], ">"),
		ArtifactDenom:  parsePositiveInt(markers["test-reuse-artifact-denominator"]),
		Cases:          parseCases(markers),
	}
	policy.EvaluatorDigest = evaluatorDigest(policy)
	if err := policy.Validate(); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func (policy Policy) Validate() error {
	if policy.Schema != PolicySchema || policy.Name != PolicyName || policy.Activity != PolicyActivity ||
		policy.Package != "publicdiscoveryexample" || policy.Namespace != "public_discovery_example" ||
		policy.ArtifactDenom != ArtifactDenominator || len(policy.Bindings) != len(requiredBindings) ||
		len(policy.Journey) != len(canonicalJourney) || len(policy.Cases) != len(canonicalCaseIDs) {
		return errors.New("public test reuse policy identity is invalid")
	}
	for index, binding := range requiredBindings {
		if policy.Bindings[index] != binding {
			return fmt.Errorf("public test reuse binding %q is not canonical", policy.Bindings[index])
		}
	}
	for index, step := range canonicalJourney {
		if policy.Journey[index] != step {
			return fmt.Errorf("public test reuse journey step %d is %q, want %q", index+1, policy.Journey[index], step)
		}
	}
	seen := make(map[string]struct{}, len(policy.Cases))
	for index, expectedID := range canonicalCaseIDs {
		item := policy.Cases[index]
		if item.ID != expectedID || (item.Decision != DecisionClosed && item.Decision != DecisionUnknown && item.Decision != DecisionRefuted) {
			return fmt.Errorf("public test reuse case %d is not canonical", index+1)
		}
		if _, exists := seen[item.ID]; exists {
			return fmt.Errorf("public test reuse case %q is duplicated", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
	counts := map[string]int{}
	for _, item := range policy.Cases {
		counts[item.Decision]++
	}
	if counts[DecisionClosed] != 2 || counts[DecisionUnknown] != 2 || counts[DecisionRefuted] != 2 {
		return fmt.Errorf("public test reuse policy decisions are %v, want 2/2/2", counts)
	}
	return nil
}

func EvaluatorDigest(policy Policy) string { return evaluatorDigest(policy) }

func evaluatorDigest(policy Policy) string {
	var builder strings.Builder
	builder.WriteString(EvaluatorSchema)
	builder.WriteByte('\n')
	builder.WriteString(policy.Name)
	builder.WriteByte('\n')
	for _, binding := range policy.Bindings {
		builder.WriteString(binding)
		builder.WriteByte(',')
	}
	builder.WriteByte('\n')
	for _, step := range policy.Journey {
		builder.WriteString(step)
		builder.WriteByte('>')
	}
	builder.WriteByte('\n')
	for _, item := range policy.Cases {
		builder.WriteString(item.ID)
		builder.WriteByte('=')
		builder.WriteString(item.Decision)
		builder.WriteByte('\n')
	}
	return cache.HashBytes([]byte(builder.String())).String()
}

func parseMarkers(value string) map[string]string {
	markers := make(map[string]string)
	for part := range strings.SplitSeq(value, ";") {
		key, item, ok := strings.Cut(part, "=")
		if ok && strings.TrimSpace(key) != "" {
			markers[strings.TrimSpace(key)] = strings.TrimSpace(item)
		}
	}
	return markers
}

func parseCases(markers map[string]string) []Case {
	var cases []Case
	for index := 1; ; index++ {
		value, ok := markers[fmt.Sprintf("test-reuse-case-%d", index)]
		if !ok {
			break
		}
		id, decision, ok := strings.Cut(value, ":")
		if ok {
			cases = append(cases, Case{ID: id, Decision: decision})
		}
	}
	return cases
}

func parsePositiveInt(value string) int {
	var result int
	if _, err := fmt.Sscan(value, &result); err != nil || result < 0 {
		return 0
	}
	return result
}
