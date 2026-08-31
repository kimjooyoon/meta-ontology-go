// Package independent is a deliberately separate raw-source replay kernel.
// It shares only the repository syntax and bidirectional lowering kernel with
// the producer. It does not import policycompilation or consume CompiledPolicy.
package independent

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	fixedDenominator     = 8
	reductionRuleCount   = 7
	policySchema         = "decision-reduction:v1"
	decisionPass         = "PASS"
	decisionFailClosed   = "FAIL_CLOSED"
	decisionUnknown      = "UNKNOWN"
	conditionUnavailable = "EVIDENCE_UNAVAILABLE"
	conditionEmpty       = "DIGEST_UNAVAILABLE"
	conditionMalformed   = "MALFORMED_DIGEST"
	conditionSource      = "SOURCE_DIGEST_MISMATCH"
	conditionArtifact    = "ARTIFACT_SOURCE_MISMATCH"
	conditionIndependent = "INDEPENDENT_SOURCE_MISMATCH"
	conditionEquivalent  = "SEMANTIC_EQUIVALENCE"
)

type Case struct {
	ID                           string
	ProducerAvailable            bool
	ConsumerAvailable            bool
	ObservedSourceDigest         string
	ObservedArtifactSourceDigest string
	ObservedGeneratedJudgeDigest string
	ObservedIndependentDigest    string
}

type Result struct {
	CaseID         string `json:"case_id"`
	Decision       string `json:"decision"`
	Stage          string `json:"stage"`
	Step           int    `json:"step"`
	Reason         string `json:"reason"`
	PolicyDigest   string `json:"policy_digest"`
	SemanticDigest string `json:"semantic_digest"`
	Denominator    int    `json:"fixed_denominator"`
}

type reductionRule struct {
	Condition string
	Decision  string
	Stage     string
	Step      int
	Reason    string
}

type Program struct {
	sourceDigest   string
	semanticDigest string
	denominator    int
	reduction      []reductionRule
}

// Compile parses and lowers raw Gooo again. The returned program is not the
// producer's compiled representation; it owns its parser-to-reduction path.
func Compile(source []byte) (Program, error) {
	file, diagnostics := syntax.ParseFile("policy.gooo", string(source))
	if diagnostics.HasErrors() {
		return Program{}, errors.New(diagnostics.Error().Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Program{}, fmt.Errorf("independent lower policy: %w", err)
	}
	if ir.Package != "metapolicycompilation" || ir.Namespace.String() != "metapolicycompilation" {
		return Program{}, errors.New("independent policy package/namespace mismatch")
	}
	activities := 0
	var reduction []reductionRule
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity {
			continue
		}
		activities++
		values, err := parseActivity(node.ValueProgram)
		if err != nil {
			return Program{}, fmt.Errorf("independent activity %q: %w", node.Name, err)
		}
		if values.reduction != "" {
			if reduction != nil {
				return Program{}, errors.New("independent policy has multiple reductions")
			}
			reduction, err = parseReduction(values.reduction)
			if err != nil {
				return Program{}, err
			}
		}
	}
	if activities != fixedDenominator || len(reduction) != reductionRuleCount {
		return Program{}, fmt.Errorf("independent fixed shape is %d/%d", activities, len(reduction))
	}
	return Program{
		sourceDigest: DigestBytes(source), semanticDigest: "sha256:" + ir.StableHash(),
		denominator: fixedDenominator, reduction: reduction,
	}, nil
}

func (program Program) SourceDigest() string   { return program.sourceDigest }
func (program Program) SemanticDigest() string { return program.semanticDigest }

func (program Program) Evaluate(input Case) Result {
	result := Result{CaseID: input.ID, PolicyDigest: program.sourceDigest, SemanticDigest: program.semanticDigest, Denominator: program.denominator}
	if program.denominator != fixedDenominator || len(program.reduction) != reductionRuleCount {
		return safetyFailure(result, "FIXED_DENOMINATOR_CHANGED")
	}
	for _, rule := range program.reduction {
		if matches(rule.Condition, program.sourceDigest, program.semanticDigest, input) {
			result.Decision, result.Stage, result.Step, result.Reason = rule.Decision, rule.Stage, rule.Step, rule.Reason
			return result
		}
	}
	return safetyFailure(result, "NO_REDUCTION_RULE_MATCHED")
}

func safetyFailure(result Result, reason string) Result {
	result.Decision, result.Stage, result.Step, result.Reason = decisionFailClosed, "COMPILE", 3, reason
	return result
}

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func matches(condition, sourceDigest, semanticDigest string, input Case) bool {
	available := input.ProducerAvailable && input.ConsumerAvailable
	valid := func(value string) bool { return digestPattern.MatchString(value) }
	empty := input.ObservedSourceDigest == "" || input.ObservedArtifactSourceDigest == "" || input.ObservedGeneratedJudgeDigest == "" || input.ObservedIndependentDigest == ""
	malformed := !valid(input.ObservedSourceDigest) || !valid(input.ObservedArtifactSourceDigest) || !valid(input.ObservedGeneratedJudgeDigest) || !valid(input.ObservedIndependentDigest)
	switch condition {
	case conditionUnavailable:
		return !available
	case conditionEmpty:
		return available && empty
	case conditionMalformed:
		return available && !empty && malformed
	case conditionSource:
		return available && !empty && !malformed && input.ObservedSourceDigest != sourceDigest
	case conditionArtifact:
		return available && !empty && !malformed && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest != sourceDigest
	case conditionIndependent:
		return available && !empty && !malformed && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest != semanticDigest
	case conditionEquivalent:
		return available && !empty && !malformed && input.ObservedSourceDigest == sourceDigest && input.ObservedArtifactSourceDigest == sourceDigest && input.ObservedIndependentDigest == semanticDigest
	default:
		return false
	}
}

type activityValues struct{ reduction string }

var safeToken = regexp.MustCompile(`^[A-Za-z0-9_.:/-]+$`)

func parseActivity(value string) (activityValues, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 8 || parts[0] != "policy-compilation:v2" {
		return activityValues{}, errors.New("independent activity schema mismatch")
	}
	values := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, field, ok := strings.Cut(part, "=")
		if !ok || key == "" || field == "" || values[key] != "" || !map[string]bool{"role": true, "meta-operation": true, "proof-choice": true, "stage": true, "step": true, "reason": true, "claim": true, "decision-reduction": true}[key] {
			return activityValues{}, fmt.Errorf("independent activity field %q is invalid", part)
		}
		values[key] = field
	}
	step, err := strconv.Atoi(values["step"])
	if err != nil || step < 1 || step > fixedDenominator {
		return activityValues{}, errors.New("independent activity step is unsafe")
	}
	for _, key := range []string{"role", "meta-operation", "proof-choice", "stage", "reason", "claim"} {
		if values[key] == "" || !safeToken.MatchString(values[key]) {
			return activityValues{}, fmt.Errorf("independent activity value %q is unsafe", key)
		}
	}
	return activityValues{reduction: values["decision-reduction"]}, nil
}

func parseReduction(value string) ([]reductionRule, error) {
	parts := strings.Split(value, ";")
	if len(parts) != reductionRuleCount+2 || parts[0] != policySchema || strings.TrimPrefix(parts[1], "denominator=") != strconv.Itoa(reductionRuleCount) {
		return nil, errors.New("independent reduction schema mismatch")
	}
	result := make([]reductionRule, 0, reductionRuleCount)
	seen := make(map[string]bool, reductionRuleCount)
	for _, encoded := range parts[2:] {
		fields := strings.Split(encoded, ":")
		if len(fields) != 5 {
			return nil, fmt.Errorf("independent decision rule %q is malformed", encoded)
		}
		step, err := strconv.Atoi(fields[3])
		if err != nil || step < 1 || step > fixedDenominator || !knownCondition(fields[0]) || !knownDecision(fields[1]) || !safeToken.MatchString(fields[2]) || !safeToken.MatchString(fields[4]) || seen[fields[0]] {
			return nil, fmt.Errorf("independent decision rule %q is unsafe", encoded)
		}
		seen[fields[0]] = true
		result = append(result, reductionRule{Condition: fields[0], Decision: fields[1], Stage: fields[2], Step: step, Reason: fields[4]})
	}
	if !seen[conditionEquivalent] {
		return nil, errors.New("independent reduction has no semantic-equivalence terminal")
	}
	return result, nil
}

func knownCondition(value string) bool {
	switch value {
	case conditionUnavailable, conditionEmpty, conditionMalformed, conditionSource, conditionArtifact, conditionIndependent, conditionEquivalent:
		return true
	default:
		return false
	}
}

func knownDecision(value string) bool {
	return value == decisionPass || value == decisionFailClosed || value == decisionUnknown
}

func DigestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}
