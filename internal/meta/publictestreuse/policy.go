package publictestreuse

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const (
	PolicySchema        = "gooo/public-test-reuse-policy/v3"
	PolicyActivity      = "Compile"
	PolicyContract      = "v3"
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

//go:embed policy.gooo
var decisionSource []byte

var requiredBindings = []string{
	"canonical-source", "canonical-semantic", "generated-output", "generated-semantic",
	"generated-manifest", "compiler", "released-tool", "toolchain", "test-command",
	"test-contract", "execution-package-inputs", "successful-result",
}

var canonicalCaseIDs = []string{
	CaseBaselineExecution, CaseAuthorizedReuse, CaseMissingAuthorization,
	CaseStaleEvidence, CaseTamperedReceipt, CasePolicyMismatch,
}

var canonicalJourney = []string{
	"baseline-generate", "baseline-build", "baseline-test", "write-receipt",
	"authorize", "replay-validate", "reused-test", "evidence",
}

type Policy struct {
	Schema               string
	SourceDigest         string
	DecisionSourceDigest string
	SemanticDigest       string
	EvaluatorDigest      string
	Package              string
	Namespace            string
	Activity             string
	Name                 string
	Bindings             []string
	Journey              []string
	ArtifactDenom        int
	filename             string
	source               []byte
}

func CanonicalCaseIDs() []string { return append([]string(nil), canonicalCaseIDs...) }

func CanonicalJourney() []string { return append([]string(nil), canonicalJourney...) }

func RequiredBindings() []string { return append([]string(nil), requiredBindings...) }

func Load(filename string, source []byte) (Policy, error) {
	return loadWithDecisionSource(filename, source, decisionSource)
}

func loadWithDecisionSource(filename string, source, program []byte) (Policy, error) {
	if len(source) == 0 || len(source) > 128<<10 || len(program) == 0 || len(program) > 128<<10 {
		return Policy{}, errors.New("public test reuse policy source requires 1..131072 bytes")
	}
	ir, err := lowerPolicy(filename, source)
	if err != nil {
		return Policy{}, err
	}
	markers, err := policyMarkers(ir)
	if err != nil {
		return Policy{}, err
	}
	if markers["test-reuse-contract"] != PolicyContract {
		return Policy{}, errors.New("public test reuse policy contract version is unsupported")
	}
	_, runtimeEvaluator := valueexecution.RuntimeEvidenceIdentities()
	evaluator, err := sha256IdentityDigest(runtimeEvaluator)
	if err != nil {
		return Policy{}, fmt.Errorf("public test reuse evaluator identity: %w", err)
	}
	policy := Policy{
		Schema:               PolicySchema,
		SourceDigest:         cache.HashBytes(source).String(),
		DecisionSourceDigest: cache.HashBytes(program).String(),
		SemanticDigest:       ir.StableHash(),
		EvaluatorDigest:      evaluator,
		Package:              ir.Package,
		Namespace:            ir.Namespace.String(),
		Activity:             PolicyActivity,
		Name:                 markers["test-reuse-policy"],
		Bindings:             strings.Split(markers["test-reuse-bindings"], ","),
		Journey:              strings.Split(markers["test-reuse-journey"], ">"),
		ArtifactDenom:        parsePositiveInt(markers["test-reuse-artifact-denominator"]),
		filename:             "internal/meta/publictestreuse/policy.gooo",
		source:               append([]byte(nil), program...),
	}
	if err := policy.Validate(); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

// Runtime identities carry an algorithm prefix; receipt bindings retain the
// existing lowercase SHA-256 representation. Convert only that exact contract.
func sha256IdentityDigest(identity string) (string, error) {
	digest, prefixed := strings.CutPrefix(identity, "sha256:")
	if !prefixed || !cache.Digest(digest).Known() {
		return "", errors.New("public test reuse requires a known sha256 identity")
	}
	return digest, nil
}

func knownSHA256Identity(identity string) bool {
	_, err := sha256IdentityDigest(identity)
	return err == nil
}

func lowerPolicy(filename string, source []byte) (semantic.IR, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return semantic.IR{}, errors.New("public test reuse policy source has syntax diagnostics")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return semantic.IR{}, fmt.Errorf("lower public test reuse policy: %w", err)
	}
	return ir, nil
}

func policyMarkers(ir semantic.IR) (map[string]string, error) {
	if ir.Package != "publicdiscoveryexample" || ir.Namespace.String() != "public_discovery_example" {
		return nil, errors.New("public test reuse policy is not bound to the canonical public discovery project")
	}
	var activity semantic.Node
	count := 0
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity {
			continue
		}
		activity = node
		count++
	}
	if count != 1 || activity.Name != PolicyActivity || activity.ValueProgram == "" {
		return nil, errors.New("public test reuse metadata requires exactly one Compile activity")
	}
	return parseMarkers(activity.ValueProgram), nil
}

func (policy Policy) Validate() error {
	if policy.Schema != PolicySchema || policy.Name != PolicyName || policy.Activity != PolicyActivity ||
		policy.Package != "publicdiscoveryexample" || policy.Namespace != "public_discovery_example" ||
		policy.ArtifactDenom != ArtifactDenominator || len(policy.Bindings) != len(requiredBindings) ||
		len(policy.Journey) != len(canonicalJourney) || !cache.Digest(policy.EvaluatorDigest).Known() ||
		!cache.Digest(policy.SourceDigest).Known() || !cache.Digest(policy.DecisionSourceDigest).Known() {
		return errors.New("public test reuse policy identity is invalid or its evaluator is unobserved")
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
	return nil
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

func parsePositiveInt(value string) int {
	var result int
	if _, err := fmt.Sscan(value, &result); err != nil || result < 0 {
		return 0
	}
	return result
}
