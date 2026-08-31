package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	consumerSchema  = "gooo/meta-policy-compilation-consumer/v3"
	fixedDenom      = 8
	caseDenom       = 3
	decisionPass    = "PASS"
	decisionFail    = "FAIL_CLOSED"
	decisionUnknown = "UNKNOWN"
)

type rule struct {
	ActivityID    string `json:"activity_id"`
	ActivityName  string `json:"activity_name"`
	Role          string `json:"role"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Stage         string `json:"stage"`
	Step          int    `json:"step"`
	Reason        string `json:"reason"`
	Claim         string `json:"claim"`
}
type decisionRule struct {
	Condition     string   `json:"condition"`
	Decision      string   `json:"decision"`
	Stage         string   `json:"stage"`
	Step          int      `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}
type reduction struct {
	Schema string         `json:"schema"`
	Rules  []decisionRule `json:"rules"`
}
type policy struct {
	Schema         string    `json:"schema"`
	PolicyID       string    `json:"policy_id"`
	Package        string    `json:"package"`
	Namespace      string    `json:"namespace"`
	SourceDigest   string    `json:"source_digest"`
	SemanticDigest string    `json:"semantic_digest"`
	Denominator    int       `json:"fixed_denominator"`
	Rules          []rule    `json:"rules"`
	Reduction      reduction `json:"decision_reduction"`
}
type artifact struct {
	Schema             string `json:"schema"`
	Policy             policy `json:"policy"`
	GeneratedJudgeHash string `json:"generated_judge_digest"`
}
type input struct {
	ID                           string `json:"id"`
	ValidatorExpectation         string `json:"validator_expectation"`
	EvidenceClass                string `json:"evidence_class"`
	Provenance                   string `json:"provenance"`
	ProducerAvailable            bool   `json:"producer_available"`
	ConsumerAvailable            bool   `json:"consumer_available"`
	ObservedSourceDigest         string `json:"observed_source_digest"`
	ObservedArtifactSourceDigest string `json:"observed_artifact_source_digest"`
	ObservedGeneratedJudgeDigest string `json:"observed_generated_judge_digest"`
	ObservedIndependentDigest    string `json:"observed_independent_digest"`
	UpperDecision                string `json:"upper_decision,omitempty"`
}
type result struct {
	CaseID         string   `json:"case_id"`
	Decision       string   `json:"decision"`
	Stage          string   `json:"stage"`
	Step           int      `json:"step"`
	Reason         string   `json:"reason"`
	UnknownClass   string   `json:"unknown_class"`
	NextOperation  string   `json:"next_operation"`
	BlockedBy      []string `json:"blocked_by"`
	PolicyDigest   string   `json:"policy_digest"`
	SemanticDigest string   `json:"semantic_digest"`
	Denominator    int      `json:"fixed_denominator"`
}
type syntheticEvidence struct {
	Class             string `json:"class"`
	CaseID            string `json:"case_id"`
	ObservationDigest string `json:"observation_digest"`
	Provenance        string `json:"provenance"`
}
type currentEvidence struct {
	Class                string `json:"class"`
	Provenance           string `json:"provenance"`
	SourceDigest         string `json:"source_digest"`
	ArtifactSourceDigest string `json:"artifact_source_digest"`
	GeneratedJudgeDigest string `json:"generated_judge_digest"`
	IndependentDigest    string `json:"independent_digest"`
}
type report struct {
	Schema                                string              `json:"schema"`
	RawPolicyParsed                       bool                `json:"raw_policy_parsed"`
	RawCasesParsed                        bool                `json:"raw_cases_parsed"`
	GoooDerivedRuleNumerator              int                 `json:"gooo_derived_rule_numerator"`
	GoooDerivedRuleDenominator            int                 `json:"gooo_derived_rule_denominator"`
	SourceExecutionsNumerator             int                 `json:"source_executions_numerator"`
	SourceExecutionsDenominator           int                 `json:"source_executions_denominator"`
	GeneratedExecutionsNumerator          int                 `json:"generated_executions_numerator"`
	GeneratedExecutionsDenominator        int                 `json:"generated_executions_denominator"`
	IndependentReconstructionsNumerator   int                 `json:"independent_reconstructions_numerator"`
	IndependentReconstructionsDenominator int                 `json:"independent_reconstructions_denominator"`
	ContractDigestMatch                   bool                `json:"contract_digest_match"`
	ImportBoundary                        map[string]string   `json:"import_boundary"`
	SyntheticEvidence                     []syntheticEvidence `json:"synthetic_evidence"`
	CurrentEvidence                       currentEvidence     `json:"current_evidence"`
	SubjectResolution                     string              `json:"subject_resolution"`
	SubjectDecision                       result              `json:"subject_decision"`
}

func main() {
	policyPath := flag.String("policy", "", "raw Gooo policy source")
	casesPath := flag.String("cases", "", "raw canonical cases")
	artifactDir := flag.String("artifact", "", "producer artifact directory")
	outputPath := flag.String("output", "", "consumer report")
	flag.Parse()
	if *policyPath == "" || *casesPath == "" || *artifactDir == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: meta-policy-compilation-consumer -policy policy.gooo -cases cases.json -artifact DIR -output report.json")
		os.Exit(2)
	}
	if err := consume(*policyPath, *casesPath, *artifactDir, *outputPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func consume(policyPath, casesPath, artifactDir, outputPath string) error {
	policySource, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("read raw policy: %w", err)
	}
	// This parse/lower call is intentionally local to the consumer. It does
	// not import the producer compiler, generated judge, or producer evaluator.
	compiled, err := parseRawPolicy(policyPath, policySource)
	if err != nil {
		return fmt.Errorf("consumer parse raw Gooo policy: %w", err)
	}
	cases, err := readInputs(casesPath)
	if err != nil {
		return err
	}
	cases = bindInputs(cases, compiled.SourceDigest, compiled.SemanticDigest, "")
	producerArtifact, err := readJSON[artifact](filepath.Join(artifactDir, "artifact.json"))
	if err != nil {
		return err
	}
	producerPolicy, err := readJSON[policy](filepath.Join(artifactDir, "policy.json"))
	if err != nil {
		return err
	}
	generated, err := readJSON[[]result](filepath.Join(artifactDir, "generated-results.json"))
	if err != nil {
		return err
	}
	independent, err := readJSON[[]result](filepath.Join(artifactDir, "independent-results.json"))
	if err != nil {
		return err
	}
	judge, err := os.ReadFile(filepath.Join(artifactDir, "judge.go"))
	if err != nil {
		return fmt.Errorf("read generated judge: %w", err)
	}
	cases = bindInputs(cases, compiled.SourceDigest, compiled.SemanticDigest, producerArtifact.GeneratedJudgeHash)
	if producerArtifact.Policy.SourceDigest == "" || producerPolicy.SourceDigest == "" || producerArtifact.GeneratedJudgeHash != digestBytes(judge) {
		return errors.New("producer artifact is not bound to its generated judge")
	}
	if producerArtifact.Policy.SourceDigest != compiled.SourceDigest || producerArtifact.Policy.SemanticDigest != compiled.SemanticDigest || producerArtifact.Policy.Denominator != fixedDenom || len(producerArtifact.Policy.Rules) != fixedDenom {
		return errors.New("independent raw policy reconstruction differs from artifact")
	}
	if len(generated) != caseDenom || len(independent) != caseDenom {
		return errors.New("producer execution denominator is not 3")
	}
	for index, current := range cases {
		want := evaluate(compiled, current)
		if !sameResult(want, generated[index]) || !sameResult(want, independent[index]) {
			return fmt.Errorf("consumer reconstruction differs at case %q", current.ID)
		}
	}
	if !standaloneJudge(judge) {
		return errors.New("generated judge is not a standalone artifact")
	}
	return writeConsumerReport(outputPath, cases, compiled, producerArtifact, generated, judge)
}

func standaloneJudge(judge []byte) bool {
	producerPackagePath := "internal/meta/" + "policycompilation"
	return strings.Contains(string(judge), "type result struct") && !strings.Contains(string(judge), producerPackagePath)
}

func writeConsumerReport(outputPath string, cases []input, compiled policy, producerArtifact artifact, generated []result, judge []byte) error {
	synthetic := make([]syntheticEvidence, 0, len(cases))
	for _, current := range cases {
		synthetic = append(synthetic, syntheticEvidence{Class: current.EvidenceClass, CaseID: current.ID, ObservationDigest: inputDigest(current), Provenance: current.Provenance})
	}
	subject := cases[0]
	for _, current := range cases {
		if current.ID < subject.ID {
			subject = current
		}
	}
	current := currentEvidence{Class: "CURRENT_EVIDENCE", Provenance: "consumer runner-temp artifact observation", SourceDigest: compiled.SourceDigest, ArtifactSourceDigest: producerArtifact.Policy.SourceDigest, GeneratedJudgeDigest: producerArtifact.GeneratedJudgeHash, IndependentDigest: compiled.SemanticDigest}
	output := report{Schema: consumerSchema, RawPolicyParsed: true, RawCasesParsed: true, GoooDerivedRuleNumerator: len(compiled.Rules), GoooDerivedRuleDenominator: fixedDenom, SourceExecutionsNumerator: caseDenom, SourceExecutionsDenominator: caseDenom, GeneratedExecutionsNumerator: len(generated), GeneratedExecutionsDenominator: caseDenom, IndependentReconstructionsNumerator: caseDenom, IndependentReconstructionsDenominator: caseDenom, ContractDigestMatch: producerArtifact.Policy.SourceDigest == compiled.SourceDigest && producerArtifact.Policy.SemanticDigest == compiled.SemanticDigest && producerArtifact.GeneratedJudgeHash == digestBytes(judge), ImportBoundary: map[string]string{"producer_compiler_imports": "0/1", "generated_template_imports": "0/1", "independent_evaluator_imports": "0/1"}, SyntheticEvidence: synthetic, CurrentEvidence: current, SubjectResolution: "RESOLVED", SubjectDecision: evaluate(compiled, subject)}
	return writeJSON(outputPath, output)
}

func parseRawPolicy(filename string, source []byte) (policy, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return policy{}, errors.New(diagnostics.Error().Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return policy{}, fmt.Errorf("lower raw policy: %w", err)
	}
	if ir.Package != "metapolicycompilation" || ir.Namespace.String() != "metapolicycompilation" {
		return policy{}, errors.New("unexpected policy package or namespace")
	}
	result := policy{Schema: "gooo/meta-policy-compilation/v3", PolicyID: "gooo://meta-policy-compilation/policy/v3", Package: ir.Package, Namespace: ir.Namespace.String(), SourceDigest: digestBytes(source), SemanticDigest: "sha256:" + ir.StableHash(), Denominator: fixedDenom, Rules: make([]rule, 0, fixedDenom)}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity {
			continue
		}
		values, err := parseActivity(node.ValueProgram)
		if err != nil {
			return policy{}, fmt.Errorf("activity %q: %w", node.Name, err)
		}
		result.Rules = append(result.Rules, rule{ActivityID: string(node.ID), ActivityName: node.Name, Role: values["role"], MetaOperation: values["meta-operation"], ProofChoice: values["proof-choice"], Stage: values["stage"], Step: atoi(values["step"]), Reason: values["reason"], Claim: values["claim"]})
		if encoded := values["decision-reduction"]; encoded != "" {
			result.Reduction, err = parseReduction(encoded)
			if err != nil {
				return policy{}, err
			}
		}
	}
	sort.Slice(result.Rules, func(i, j int) bool { return result.Rules[i].Step < result.Rules[j].Step })
	if len(result.Rules) != fixedDenom || result.Reduction.Schema == "" || len(result.Reduction.Rules) != fixedDenom {
		return policy{}, errors.New("raw policy did not produce fixed source contract")
	}
	return result, nil
}

func parseActivity(value string) (map[string]string, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 8 || parts[0] != "policy-compilation:v3" {
		return nil, errors.New("unsupported policy value program")
	}
	values := make(map[string]string)
	for _, part := range parts[1:] {
		key, field, ok := strings.Cut(part, "=")
		if !ok || key == "" || field == "" || values[key] != "" {
			return nil, fmt.Errorf("invalid policy metadata %q", part)
		}
		values[key] = field
	}
	for _, key := range []string{"role", "meta-operation", "proof-choice", "stage", "step", "reason", "claim"} {
		if values[key] == "" {
			return nil, fmt.Errorf("missing policy field %q", key)
		}
	}
	return values, nil
}

func parseReduction(value string) (reduction, error) {
	parts := strings.Split(value, ";")
	if len(parts) != 10 || parts[0] != "decision-reduction:v2" || parts[1] != "denominator=8" {
		return reduction{}, errors.New("invalid source decision reduction")
	}
	result := reduction{Schema: parts[0], Rules: make([]decisionRule, 0, fixedDenom)}
	for _, encoded := range parts[2:] {
		fields := strings.Split(encoded, ":")
		if len(fields) != 8 {
			return reduction{}, fmt.Errorf("invalid source reduction row %q", encoded)
		}
		row := decisionRule{Condition: fields[0], Decision: fields[1], Stage: fields[2], Step: atoi(fields[3]), Reason: fields[4]}
		if fields[5] != "NONE" {
			row.UnknownClass = fields[5]
		}
		if fields[6] != "NONE" {
			row.NextOperation = fields[6]
		}
		if fields[7] != "NONE" {
			row.BlockedBy = strings.Split(fields[7], ",")
		}
		result.Rules = append(result.Rules, row)
	}
	return result, nil
}

func evaluate(policy policy, value input) result {
	output := result{CaseID: value.ID, PolicyDigest: policy.SourceDigest, SemanticDigest: policy.SemanticDigest, Denominator: policy.Denominator, BlockedBy: []string{}}
	valid := func(item string) bool { return regexp.MustCompile(`^sha256:[0-9a-f]{64}$`).MatchString(item) }
	sourceOK, artifactOK, independentOK, judgeOK := valid(value.ObservedSourceDigest), valid(value.ObservedArtifactSourceDigest), valid(value.ObservedIndependentDigest), valid(value.ObservedGeneratedJudgeDigest)
	complete := sourceOK && artifactOK && independentOK && judgeOK
	empty := value.ObservedSourceDigest == "" || value.ObservedArtifactSourceDigest == "" || value.ObservedGeneratedJudgeDigest == "" || value.ObservedIndependentDigest == ""
	ready := value.ProducerAvailable && value.ConsumerAvailable
	match := func(condition string) bool {
		switch condition {
		case "UNRECOGNIZED_TOP_LEVEL_DECISION":
			return value.UpperDecision != "" && value.UpperDecision != decisionPass && value.UpperDecision != decisionFail && value.UpperDecision != decisionUnknown
		case "SOURCE_DIGEST_MISMATCH":
			return sourceOK && value.ObservedSourceDigest != policy.SourceDigest
		case "ARTIFACT_SOURCE_MISMATCH":
			return sourceOK && artifactOK && value.ObservedSourceDigest == policy.SourceDigest && value.ObservedArtifactSourceDigest != policy.SourceDigest
		case "INDEPENDENT_SOURCE_MISMATCH":
			return sourceOK && artifactOK && independentOK && value.ObservedSourceDigest == policy.SourceDigest && value.ObservedArtifactSourceDigest == policy.SourceDigest && value.ObservedIndependentDigest != policy.SemanticDigest
		case "EVIDENCE_UNAVAILABLE":
			return !ready && !sourceOK && !artifactOK && !independentOK && !judgeOK
		case "DIGEST_UNAVAILABLE":
			return ready && empty
		case "MALFORMED_DIGEST":
			return ready && !empty && !complete
		case "SEMANTIC_EQUIVALENCE":
			return ready && complete && value.ObservedSourceDigest == policy.SourceDigest && value.ObservedArtifactSourceDigest == policy.SourceDigest && value.ObservedIndependentDigest == policy.SemanticDigest
		default:
			return false
		}
	}
	for _, row := range policy.Reduction.Rules {
		if match(row.Condition) {
			output.Decision, output.Stage, output.Step, output.Reason = row.Decision, row.Stage, row.Step, row.Reason
			output.UnknownClass, output.NextOperation, output.BlockedBy = row.UnknownClass, row.NextOperation, append([]string(nil), row.BlockedBy...)
			if output.Decision != decisionUnknown {
				output.UnknownClass, output.NextOperation, output.BlockedBy = "", "", []string{}
			}
			return output
		}
	}
	output.Decision, output.Stage, output.Reason = decisionFail, "COMPILE", "NO_REDUCTION_RULE_MATCHED"
	return output
}

func readInputs(path string) ([]input, error) {
	values, err := readJSON[[]input](path)
	if err != nil {
		return nil, err
	}
	if len(values) != caseDenom {
		return nil, fmt.Errorf("canonical case denominator changed: got %d want %d", len(values), caseDenom)
	}
	seen := map[string]bool{}
	for _, value := range values {
		if value.ID == "" || seen[value.ID] || value.EvidenceClass != "SYNTHETIC_FIXTURE" || value.Provenance == "" {
			return nil, fmt.Errorf("invalid canonical case %q", value.ID)
		}
		if value.ValidatorExpectation != decisionPass && value.ValidatorExpectation != decisionFail && value.ValidatorExpectation != decisionUnknown {
			return nil, fmt.Errorf("unsupported validator expectation in %q", value.ID)
		}
		seen[value.ID] = true
	}
	return values, nil
}

func bindInputs(values []input, sourceDigest, semanticDigest, judgeDigest string) []input {
	result := append([]input(nil), values...)
	for index := range result {
		if result[index].ObservedSourceDigest == "SOURCE_DIGEST_FROM_POLICY" {
			result[index].ObservedSourceDigest = sourceDigest
		}
		if result[index].ObservedArtifactSourceDigest == "SOURCE_DIGEST_FROM_POLICY" {
			result[index].ObservedArtifactSourceDigest = sourceDigest
		}
		if result[index].ObservedGeneratedJudgeDigest == "GENERATED_JUDGE_DIGEST_FROM_ARTIFACT" && judgeDigest != "" {
			result[index].ObservedGeneratedJudgeDigest = judgeDigest
		}
		if result[index].ObservedIndependentDigest == "SEMANTIC_DIGEST_FROM_POLICY" {
			result[index].ObservedIndependentDigest = semanticDigest
		}
	}
	return result
}

func inputDigest(value input) string { return digestJSON(value) }
func digestJSON(value any) string    { data, _ := json.Marshal(value); return digestBytes(data) }
func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}
func atoi(value string) int { var number int; fmt.Sscanf(value, "%d", &number); return number }

func sameResult(left, right result) bool {
	return left.CaseID == right.CaseID && left.Decision == right.Decision && left.Stage == right.Stage && left.Step == right.Step && left.Reason == right.Reason && left.UnknownClass == right.UnknownClass && left.NextOperation == right.NextOperation && sameStrings(left.BlockedBy, right.BlockedBy) && left.PolicyDigest == right.PolicyDigest && left.SemanticDigest == right.SemanticDigest && left.Denominator == right.Denominator
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
func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o640)
}
func readJSON[T any](path string) (T, error) {
	data, err := os.ReadFile(path)
	var zero T
	if err != nil {
		return zero, fmt.Errorf("read %s: %w", path, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var value T
	if err := decoder.Decode(&value); err != nil {
		return zero, fmt.Errorf("decode %s: %w", path, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return zero, fmt.Errorf("decode %s: trailing JSON", path)
	}
	return value, nil
}
