package policycompilation

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const policyID = "gooo://meta-policy-compilation/policy/v1"

// Compile lowers the actual .gooo policy to the semantic IR, then binds its
// declared activity metadata to the experiment's meaning. The output is
// canonical and independent of declaration order.
func Compile(source []byte) (CompiledPolicy, error) {
	file, diagnostics := syntax.ParseFile("policy.gooo", string(source))
	if diagnostics.HasErrors() {
		return CompiledPolicy{}, errors.New(diagnostics.Error().Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return CompiledPolicy{}, fmt.Errorf("lower policy: %w", err)
	}
	if ir.Package != "metapolicycompilation" || ir.Namespace.String() != "metapolicycompilation" {
		return CompiledPolicy{}, fmt.Errorf("policy package/namespace is %q/%q, want metapolicycompilation", ir.Package, ir.Namespace)
	}

	rules := make([]Rule, 0, FixedDenominator)
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity {
			continue
		}
		spec, err := parseRuleSpec(node.ValueProgram)
		if err != nil {
			return CompiledPolicy{}, fmt.Errorf("activity %q: %w", node.Name, err)
		}
		spec.Name = node.Name
		expected, ok := expectedRule(node.Name)
		if !ok {
			return CompiledPolicy{}, fmt.Errorf("activity %q is outside the fixed policy ontology", node.Name)
		}
		if spec != expected {
			return CompiledPolicy{}, fmt.Errorf("activity %q changes declared policy meaning: got %#v want %#v", node.Name, spec, expected)
		}
		rules = append(rules, Rule{
			ActivityID: string(node.ID), ActivityName: node.Name,
			Role: spec.Role, MetaOperation: spec.MetaOperation,
			ProofChoice: spec.ProofChoice, Stage: spec.Stage,
			Step: spec.Step, Reason: spec.Reason, Claim: spec.Claim,
		})
	}
	if len(rules) != FixedDenominator {
		return CompiledPolicy{}, fmt.Errorf("fixed denominator changed: got %d want %d", len(rules), FixedDenominator)
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].Step < rules[j].Step })
	for index, rule := range rules {
		if rule.Step != index+1 {
			return CompiledPolicy{}, fmt.Errorf("policy step %d is not present exactly once", index+1)
		}
	}
	return CompiledPolicy{
		Schema: SchemaVersion, PolicyID: policyID,
		Package: ir.Package, Namespace: ir.Namespace.String(),
		SourceDigest: DigestBytes(source), SemanticDigest: ir.StableHash(),
		Denominator: FixedDenominator, Rules: rules,
	}, nil
}

func parseRuleSpec(value string) (RuleSpec, error) {
	parts := strings.Split(value, "|")
	if len(parts) != 8 || parts[0] != "policy-compilation:v1" {
		return RuleSpec{}, fmt.Errorf("value program must contain the v1 policy metadata fields")
	}
	values := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, value, found := strings.Cut(part, "=")
		if !found || key == "" || value == "" || values[key] != "" {
			return RuleSpec{}, fmt.Errorf("invalid policy metadata field %q", part)
		}
		values[key] = value
	}
	step, err := strconv.Atoi(values["step"])
	if err != nil {
		return RuleSpec{}, fmt.Errorf("invalid step: %w", err)
	}
	spec := RuleSpec{
		Role: values["role"], MetaOperation: values["meta-operation"],
		ProofChoice: values["proof-choice"], Stage: values["stage"],
		Step: step, Reason: values["reason"], Claim: values["claim"],
	}
	if values["role"] == "" || values["meta-operation"] == "" || values["proof-choice"] == "" || values["stage"] == "" || values["reason"] == "" || values["claim"] == "" {
		return RuleSpec{}, fmt.Errorf("policy metadata has an empty required field")
	}
	return spec, nil
}

func VerifyCompiledArtifact(artifact PolicyArtifact, policy CompiledPolicy, judgeHash string) error {
	if artifact.Schema != ArtifactSchema {
		return fmt.Errorf("unsupported artifact schema %q", artifact.Schema)
	}
	left, err := canonicalJSON(artifact.Policy)
	if err != nil {
		return err
	}
	right, err := canonicalJSON(policy)
	if err != nil {
		return err
	}
	if string(left) != string(right) {
		return errors.New("consumer observed a compiled policy different from the producer policy")
	}
	if artifact.GeneratedJudgeHash == "" || artifact.GeneratedJudgeHash != judgeHash {
		return errors.New("generated judge digest is not bound to the artifact")
	}
	if policy.Denominator != FixedDenominator || len(policy.Rules) != FixedDenominator {
		return errors.New("compiled artifact does not retain the fixed denominator")
	}
	return nil
}
