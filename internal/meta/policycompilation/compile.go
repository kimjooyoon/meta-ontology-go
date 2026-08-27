package policycompilation

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const policyID = "gooo://meta-policy-compilation/policy/v2"

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
	var reduction DecisionReduction
	reductionCount := 0
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity {
			continue
		}
		values, err := parseActivityProgram(node.ValueProgram)
		if err != nil {
			return CompiledPolicy{}, fmt.Errorf("activity %q: %w", node.Name, err)
		}
		if values.Reduction != "" {
			reductionCount++
			if reductionCount > 1 {
				return CompiledPolicy{}, errors.New("decision reduction must be declared exactly once")
			}
			reduction, err = parseDecisionReduction(values.Reduction)
			if err != nil {
				return CompiledPolicy{}, fmt.Errorf("activity %q: %w", node.Name, err)
			}
		}
		rules = append(rules, Rule{
			ActivityID: string(node.ID), ActivityName: node.Name,
			Role: values.Role, MetaOperation: values.MetaOperation,
			ProofChoice: values.ProofChoice, Stage: values.Stage,
			Step: values.Step, Reason: values.Reason, Claim: values.Claim,
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
	if reductionCount != 1 {
		return CompiledPolicy{}, errors.New("decision reduction must be declared exactly once")
	}
	return CompiledPolicy{
		Schema: SchemaVersion, PolicyID: policyID,
		Package: ir.Package, Namespace: ir.Namespace.String(),
		SourceDigest: DigestBytes(source), SemanticDigest: ir.StableHash(),
		Denominator: FixedDenominator, Rules: rules, Reduction: reduction,
	}, nil
}

type activityProgram struct {
	Role          string
	MetaOperation string
	ProofChoice   string
	Stage         string
	Step          int
	Reason        string
	Claim         string
	Reduction     string
}

var policyToken = regexp.MustCompile(`^[A-Za-z0-9_.:/-]+$`)

func parseActivityProgram(value string) (activityProgram, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 8 || parts[0] != "policy-compilation:v2" {
		return activityProgram{}, fmt.Errorf("value program must contain the v2 policy metadata fields")
	}
	values := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, fieldValue, found := strings.Cut(part, "=")
		if !found || key == "" || fieldValue == "" || values[key] != "" {
			return activityProgram{}, fmt.Errorf("invalid policy metadata field %q", part)
		}
		if key != "role" && key != "meta-operation" && key != "proof-choice" && key != "stage" && key != "step" && key != "reason" && key != "claim" && key != "decision-reduction" {
			return activityProgram{}, fmt.Errorf("unknown policy metadata field %q", key)
		}
		values[key] = fieldValue
	}
	step, err := strconv.Atoi(values["step"])
	if err != nil {
		return activityProgram{}, fmt.Errorf("invalid step: %w", err)
	}
	program := activityProgram{
		Role: values["role"], MetaOperation: values["meta-operation"],
		ProofChoice: values["proof-choice"], Stage: values["stage"],
		Step: step, Reason: values["reason"], Claim: values["claim"],
		Reduction: values["decision-reduction"],
	}
	if program.Role == "" || program.MetaOperation == "" || program.ProofChoice == "" || program.Stage == "" || program.Reason == "" || program.Claim == "" || !policyToken.MatchString(program.Role) || !policyToken.MatchString(program.MetaOperation) || !policyToken.MatchString(program.ProofChoice) || !policyToken.MatchString(program.Stage) || !policyToken.MatchString(program.Reason) || !policyToken.MatchString(program.Claim) {
		return activityProgram{}, fmt.Errorf("policy metadata has an empty or unsafe required field")
	}
	if step < 1 || step > FixedDenominator {
		return activityProgram{}, fmt.Errorf("policy step %d is outside the fixed safety range", step)
	}
	return program, nil
}

func parseDecisionReduction(value string) (DecisionReduction, error) {
	parts := strings.Split(value, ";")
	if len(parts) != ReductionRuleCount+2 || parts[0] != ReductionSchema {
		return DecisionReduction{}, fmt.Errorf("decision reduction must contain the v1 schema, denominator, and %d source rules", ReductionRuleCount)
	}
	denominator, err := strconv.Atoi(strings.TrimPrefix(parts[1], "denominator="))
	if err != nil || denominator != ReductionRuleCount {
		return DecisionReduction{}, fmt.Errorf("decision reduction denominator must be %d", ReductionRuleCount)
	}
	reduction := DecisionReduction{Schema: ReductionSchema, Rules: make([]DecisionRule, 0, ReductionRuleCount)}
	seen := make(map[string]bool, ReductionRuleCount)
	for _, encoded := range parts[2:] {
		fields := strings.Split(encoded, ":")
		if len(fields) != 5 {
			return DecisionReduction{}, fmt.Errorf("decision rule %q must have condition, decision, stage, step, reason", encoded)
		}
		step, err := strconv.Atoi(fields[3])
		if err != nil || step < 1 || step > FixedDenominator {
			return DecisionReduction{}, fmt.Errorf("decision rule %q has an unsafe step", encoded)
		}
		if !knownCondition(fields[0]) || !knownDecision(fields[1]) || !policyToken.MatchString(fields[2]) || !policyToken.MatchString(fields[4]) || seen[fields[0]] {
			return DecisionReduction{}, fmt.Errorf("decision rule %q violates the reduction schema", encoded)
		}
		seen[fields[0]] = true
		reduction.Rules = append(reduction.Rules, DecisionRule{Condition: fields[0], Decision: fields[1], Stage: fields[2], Step: step, Reason: fields[4]})
	}
	if !seen[ConditionSemanticEquivalence] {
		return DecisionReduction{}, errors.New("decision reduction needs a semantic-equivalence terminal rule")
	}
	return reduction, nil
}

func knownCondition(value string) bool {
	switch value {
	case ConditionEvidenceUnavailable, ConditionDigestUnavailable, ConditionSourceMismatch, ConditionArtifactMismatch, ConditionIndependentMismatch, ConditionSemanticEquivalence:
		return true
	default:
		return false
	}
}

func knownDecision(value string) bool {
	return value == DecisionPass || value == DecisionFailClosed || value == DecisionUnknown
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
	if policy.Denominator != FixedDenominator || len(policy.Rules) != FixedDenominator || policy.Reduction.Schema != ReductionSchema || len(policy.Reduction.Rules) != ReductionRuleCount {
		return errors.New("compiled artifact does not retain the fixed denominator")
	}
	return nil
}
