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

const policyID = "gooo://meta-policy-compilation/policy/v3"

// Compile parses the raw Gooo source and lowers it through the repository's
// semantic IR. All policy values and reduction rows are read from the source;
// Go only checks the fixed safety envelope and typed representation.
func Compile(source []byte) (CompiledPolicy, error) {
	return CompileNamed("policy.gooo", source)
}

func CompileNamed(filename string, source []byte) (CompiledPolicy, error) {
	ir, err := lowerPolicy(filename, source)
	if err != nil {
		return CompiledPolicy{}, fmt.Errorf("lower policy: %w", err)
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
		if rule.Claim == "" {
			return CompiledPolicy{}, fmt.Errorf("policy step %d has no claim predicate", rule.Step)
		}
	}
	if err := validateClaimPredicates(rules); err != nil {
		return CompiledPolicy{}, err
	}
	if reductionCount != 1 {
		return CompiledPolicy{}, errors.New("decision reduction must be declared exactly once")
	}
	return CompiledPolicy{
		Schema: SchemaVersion, PolicyID: policyID,
		Package: ir.Package, Namespace: ir.Namespace.String(),
		SourceDigest: DigestBytes(source), SemanticDigest: SemanticDigest(ir.StableHash()),
		Denominator: FixedDenominator, Rules: rules, Reduction: reduction,
	}, nil
}

func lowerPolicy(filename string, source []byte) (semantic.IR, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return semantic.IR{}, errors.New(diagnostics.Error().Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return semantic.IR{}, err
	}
	if ir.Package != "metapolicycompilation" || ir.Namespace.String() != "metapolicycompilation" {
		return semantic.IR{}, fmt.Errorf("policy package/namespace is %q/%q, want metapolicycompilation", ir.Package, ir.Namespace)
	}
	return ir, nil
}

func validateClaimPredicates(rules []Rule) error {
	seen := make(map[string]bool, len(rules))
	for _, rule := range rules {
		if seen[rule.Claim] {
			return fmt.Errorf("claim predicate %q is duplicated", rule.Claim)
		}
		seen[rule.Claim] = true
	}
	if len(seen) != ClaimPredicateCount {
		return fmt.Errorf("claim predicate denominator changed: got %d want %d", len(seen), ClaimPredicateCount)
	}
	return nil
}

type activityProgram struct {
	Role, MetaOperation, ProofChoice, Stage string
	Step                                    int
	Reason, Claim, Reduction                string
}

var policyToken = regexp.MustCompile(`^[A-Za-z0-9_.:/_-]+$`)

func parseActivityProgram(value string) (activityProgram, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 8 || parts[0] != "policy-compilation:v3" {
		return activityProgram{}, errors.New("value program must contain the v3 policy metadata fields")
	}
	values := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, fieldValue, found := strings.Cut(part, "=")
		if !found || key == "" || fieldValue == "" || values[key] != "" {
			return activityProgram{}, fmt.Errorf("invalid policy metadata field %q", part)
		}
		switch key {
		case "role", "meta-operation", "proof-choice", "stage", "step", "reason", "claim", "decision-reduction":
		default:
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
	for _, value := range []string{program.Role, program.MetaOperation, program.ProofChoice, program.Stage, program.Reason, program.Claim} {
		if value == "" || !policyToken.MatchString(value) {
			return activityProgram{}, errors.New("policy metadata has an empty or unsafe required field")
		}
	}
	if step < 1 || step > FixedDenominator {
		return activityProgram{}, fmt.Errorf("policy step %d is outside the fixed safety range", step)
	}
	return program, nil
}

func parseDecisionReduction(value string) (DecisionReduction, error) {
	parts := strings.Split(value, ";")
	if len(parts) != ReductionRuleCount+2 || parts[0] != ReductionSchema {
		return DecisionReduction{}, fmt.Errorf("decision reduction must contain the v2 schema, denominator, and %d source rules", ReductionRuleCount)
	}
	denominator, err := strconv.Atoi(strings.TrimPrefix(parts[1], "denominator="))
	if err != nil || denominator != ReductionRuleCount {
		return DecisionReduction{}, fmt.Errorf("decision reduction denominator must be %d", ReductionRuleCount)
	}
	reduction := DecisionReduction{Schema: ReductionSchema, Rules: make([]DecisionRule, 0, ReductionRuleCount)}
	seen := make(map[string]bool, ReductionRuleCount)
	for _, encoded := range parts[2:] {
		fields := strings.Split(encoded, ":")
		if len(fields) != 8 {
			return DecisionReduction{}, fmt.Errorf("decision rule %q must have condition, decision, stage, step, reason, unknown class, next operation, and blocked_by", encoded)
		}
		step, err := strconv.Atoi(fields[3])
		if err != nil || step < 1 || step > FixedDenominator {
			return DecisionReduction{}, fmt.Errorf("decision rule %q has an unsafe step", encoded)
		}
		if !knownCondition(fields[0]) || !knownDecision(fields[1]) || !policyToken.MatchString(fields[2]) || !policyToken.MatchString(fields[4]) || seen[fields[0]] {
			return DecisionReduction{}, fmt.Errorf("decision rule %q violates the reduction schema", encoded)
		}
		blocked := []string(nil)
		if fields[7] != "NONE" {
			for item := range strings.SplitSeq(fields[7], ",") {
				if item == "" || !policyToken.MatchString(item) {
					return DecisionReduction{}, fmt.Errorf("decision rule %q has unsafe blocked_by", encoded)
				}
				blocked = append(blocked, item)
			}
		}
		if fields[5] != "NONE" && !policyToken.MatchString(fields[5]) || fields[6] != "NONE" && !policyToken.MatchString(fields[6]) {
			return DecisionReduction{}, fmt.Errorf("decision rule %q has unsafe unknown metadata", encoded)
		}
		if fields[1] == DecisionUnknown && (fields[5] == "NONE" || fields[6] == "NONE" || len(blocked) == 0) {
			return DecisionReduction{}, fmt.Errorf("unknown decision rule %q must preserve unknown metadata", encoded)
		}
		if fields[1] != DecisionUnknown && (fields[5] != "NONE" || fields[6] != "NONE" || len(blocked) != 0) {
			return DecisionReduction{}, fmt.Errorf("known decision rule %q cannot carry unknown metadata", encoded)
		}
		seen[fields[0]] = true
		reduction.Rules = append(reduction.Rules, DecisionRule{
			Condition: fields[0], Decision: fields[1], Stage: fields[2], Step: step,
			Reason: fields[4], UnknownClass: noneValue(fields[5]), NextOperation: noneValue(fields[6]), BlockedBy: blocked,
		})
	}
	if !seen[ConditionSemanticEquivalence] || len(seen) != ReductionRuleCount {
		return DecisionReduction{}, errors.New("decision reduction does not cover the complete source condition set")
	}
	return reduction, nil
}

func noneValue(value string) string {
	if value == "NONE" {
		return ""
	}
	return value
}

func knownCondition(value string) bool {
	switch value {
	case ConditionEvidenceUnavailable, ConditionDigestUnavailable, ConditionMalformedDigest, ConditionSourceMismatch, ConditionArtifactMismatch, ConditionIndependentMismatch, ConditionUnrecognizedTopDecision, ConditionSemanticEquivalence:
		return true
	default:
		return false
	}
}

func knownDecision(value string) bool {
	return value == DecisionPass || value == DecisionFailClosed || value == DecisionUnknown
}
