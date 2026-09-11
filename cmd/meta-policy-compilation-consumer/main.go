package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
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
	consumerSchema   = "gooo/meta-policy-compilation-consumer/v3"
	policyWireSchema = "gooo/meta-policy-compilation/v3"
	fixedDenom       = 8
	caseDenom        = 3
	decisionPass     = "PASS"
	decisionFail     = "FAIL_CLOSED"
	decisionUnknown  = "UNKNOWN"
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
	Schema         string          `json:"schema"`
	PolicyID       string          `json:"policy_id"`
	Package        string          `json:"package"`
	Namespace      string          `json:"namespace"`
	SourceDigest   string          `json:"source_digest"`
	SemanticDigest string          `json:"semantic_digest"`
	Denominator    int             `json:"fixed_denominator"`
	Rules          []rule          `json:"rules"`
	Reduction      reduction       `json:"decision_reduction"`
	Structure      json.RawMessage `json:"structure"`
}
type artifact struct {
	Schema             string `json:"schema"`
	Policy             policy `json:"policy"`
	GeneratedJudgeHash string `json:"generated_judge_digest"`
}
type generationManifest struct {
	Schema               string   `json:"schema"`
	Profile              string   `json:"profile"`
	SourceFile           string   `json:"source_file"`
	SourceDigest         string   `json:"source_digest"`
	SemanticDigest       string   `json:"semantic_digest"`
	Package              string   `json:"package"`
	Namespace            string   `json:"namespace"`
	PolicyBytesDigest    string   `json:"policy_bytes_digest"`
	ArtifactBytesDigest  string   `json:"artifact_bytes_digest"`
	GeneratedJudgeDigest string   `json:"generated_judge_digest"`
	GeneratedFiles       []string `json:"generated_files"`
	OutputRootClass      string   `json:"output_root_class"`
	ExecutionObserved    bool     `json:"execution_observed"`
	CurrentConformance   string   `json:"current_conformance"`
	RepositoryWrites     int      `json:"repository_writes"`
	MutationAuthority    int      `json:"mutation_authority"`
	PromotionAuthority   int      `json:"promotion_authority"`
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
	CaseID           string   `json:"case_id"`
	Decision         string   `json:"decision"`
	MatchedCondition string   `json:"matched_condition"`
	Stage            string   `json:"stage"`
	Step             int      `json:"step"`
	Reason           string   `json:"reason"`
	UnknownClass     string   `json:"unknown_class"`
	NextOperation    string   `json:"next_operation"`
	BlockedBy        []string `json:"blocked_by"`
	PolicyDigest     string   `json:"policy_digest"`
	SemanticDigest   string   `json:"semantic_digest"`
	Denominator      int      `json:"fixed_denominator"`
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
	profilePackage := flag.String("profile-package", "metapolicycompilation", "expected policy package")
	profileNamespace := flag.String("profile-namespace", "metapolicycompilation", "expected policy namespace")
	manifestPath := flag.String("manifest", "", "public generation manifest")
	flag.Parse()
	if *policyPath == "" || *casesPath == "" || *artifactDir == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: meta-policy-compilation-consumer -policy policy.gooo -cases cases.json -artifact DIR -output report.json")
		os.Exit(2)
	}
	if err := consume(*policyPath, *casesPath, *artifactDir, *outputPath, *manifestPath, *profilePackage, *profileNamespace); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func consume(policyPath, casesPath, artifactDir, outputPath, manifestPath, expectedPackage, expectedNamespace string) error {
	policySource, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("read raw policy: %w", err)
	}
	// This parse/lower call is intentionally local to the consumer. It does
	// not import the producer compiler, generated judge, or producer evaluator.
	compiled, err := parseRawPolicy(policyPath, policySource, expectedPackage, expectedNamespace)
	if err != nil {
		return fmt.Errorf("consumer parse raw Gooo policy: %w", err)
	}
	cases, err := readInputs(casesPath)
	if err != nil {
		return err
	}
	cases = bindInputs(cases, compiled.SourceDigest, compiled.SemanticDigest, "")
	boundary, err := readProducerBoundary(artifactDir, manifestPath)
	if err != nil {
		return err
	}
	cases = bindInputs(cases, compiled.SourceDigest, compiled.SemanticDigest, boundary.artifact.GeneratedJudgeHash)
	if err := validateProducerBoundary(boundary, compiled); err != nil {
		return err
	}
	if err := compareProducerExecutions(cases, compiled, boundary.generated, boundary.independent); err != nil {
		return err
	}
	if !standaloneJudge(boundary.judge) {
		return errors.New("generated judge is not a standalone artifact")
	}
	return writeConsumerReport(outputPath, cases, compiled, boundary.artifact, boundary.generated, boundary.judge)
}

type producerBoundary struct {
	policy        policy
	policyBytes   []byte
	artifact      artifact
	artifactBytes []byte
	generated     []result
	independent   []result
	judge         []byte
	manifest      generationManifest
}

func readProducerBoundary(artifactDir, manifestPath string) (producerBoundary, error) {
	producerPolicy, policyBytes, err := readJSONBytes[policy](filepath.Join(artifactDir, "policy.json"))
	if err != nil {
		return producerBoundary{}, err
	}
	producerArtifact, artifactBytes, err := readJSONBytes[artifact](filepath.Join(artifactDir, "artifact.json"))
	if err != nil {
		return producerBoundary{}, err
	}
	generated, err := readJSON[[]result](filepath.Join(artifactDir, "generated-results.json"))
	if err != nil {
		return producerBoundary{}, err
	}
	independent, err := readJSON[[]result](filepath.Join(artifactDir, "independent-results.json"))
	if err != nil {
		return producerBoundary{}, err
	}
	judge, err := os.ReadFile(filepath.Join(artifactDir, "judge.go"))
	if err != nil {
		return producerBoundary{}, fmt.Errorf("read generated judge: %w", err)
	}
	if manifestPath == "" {
		manifestPath = filepath.Join(artifactDir, "generation-manifest.json")
	}
	manifest, err := readGenerationManifest(manifestPath)
	if err != nil {
		return producerBoundary{}, err
	}
	return producerBoundary{
		policy: producerPolicy, policyBytes: policyBytes,
		artifact: producerArtifact, artifactBytes: artifactBytes,
		generated: generated, independent: independent,
		judge: judge, manifest: manifest,
	}, nil
}

func validateProducerBoundary(boundary producerBoundary, compiled policy) error {
	if boundary.artifact.Policy.SourceDigest == "" || boundary.policy.SourceDigest == "" || boundary.artifact.GeneratedJudgeHash != digestBytes(boundary.judge) {
		return errors.New("producer artifact is not bound to its generated judge")
	}
	if mismatch := policySemanticMismatch(boundary.artifact.Policy, boundary.policy); mismatch != "" {
		return fmt.Errorf("producer policy and compiled artifact policy differ at %s", mismatch)
	}
	if mismatch := policySemanticMismatch(boundary.policy, compiled); mismatch != "" {
		return fmt.Errorf("independent raw policy reconstruction differs from artifact at %s", mismatch)
	}
	if mismatch := policySemanticMismatch(boundary.artifact.Policy, compiled); mismatch != "" {
		return fmt.Errorf("independent raw policy reconstruction differs from artifact at %s", mismatch)
	}
	if err := validateGenerationManifest(boundary.manifest, boundary.policyBytes, boundary.artifactBytes, boundary.judge, compiled, boundary.artifact); err != nil {
		return err
	}
	if len(boundary.generated) != caseDenom || len(boundary.independent) != caseDenom {
		return errors.New("producer execution denominator is not 3")
	}
	return nil
}

func compareProducerExecutions(cases []input, compiled policy, generated, independent []result) error {
	for index, current := range cases {
		want := evaluate(compiled, current)
		if !sameResult(want, generated[index]) || !sameResult(want, independent[index]) {
			return fmt.Errorf("consumer reconstruction differs at case %q", current.ID)
		}
	}
	return nil
}

func standaloneJudge(judge []byte) bool {
	producerPackagePath := "internal/meta/" + "policycompilation"
	return strings.Contains(string(judge), "type result struct") && !strings.Contains(string(judge), producerPackagePath)
}

func validateGenerationManifest(manifest generationManifest, policyBytes, artifactBytes, judge []byte, compiled policy, producerArtifact artifact) error {
	if manifest.Schema != "gooo/meta-policy-compilation-generation-manifest/v1" || manifest.Profile != "meta-policy-compilation-v3" || manifest.SourceFile == "" || manifest.SourceDigest != compiled.SourceDigest || manifest.SemanticDigest != compiled.SemanticDigest || manifest.Package != compiled.Package || manifest.Namespace != compiled.Namespace || manifest.PolicyBytesDigest != digestBytes(policyBytes) || manifest.ArtifactBytesDigest != digestBytes(artifactBytes) || manifest.GeneratedJudgeDigest != digestBytes(judge) || manifest.OutputRootClass != "CALLER_OWNED_EXTERNAL" || manifest.ExecutionObserved || manifest.CurrentConformance != "UNKNOWN" || manifest.RepositoryWrites != 0 || manifest.MutationAuthority != 0 || manifest.PromotionAuthority != 0 {
		return errors.New("generation manifest is not bound to an unexecuted external profile artifact")
	}
	wantFiles := []string{"policy.json", "artifact.json", "judge.go", "generation-manifest.json"}
	if len(manifest.GeneratedFiles) != len(wantFiles) {
		return errors.New("generation manifest file denominator changed")
	}
	for index, want := range wantFiles {
		if manifest.GeneratedFiles[index] != want {
			return errors.New("generation manifest file set is not canonical")
		}
	}
	if producerArtifact.Policy.SourceDigest != compiled.SourceDigest || producerArtifact.Policy.SemanticDigest != compiled.SemanticDigest || producerArtifact.Policy.Package != compiled.Package || producerArtifact.Policy.Namespace != compiled.Namespace {
		return errors.New("generation manifest policy identity differs from raw source")
	}
	return nil
}

func mustJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}

func policySemanticMismatch(observed, expected policy) string {
	if observed.Schema != expected.Schema {
		return fmt.Sprintf("schema: expected wire schema %q observed %q", expected.Schema, observed.Schema)
	}
	if observed.PolicyID != expected.PolicyID {
		return fmt.Sprintf("policy_id: expected %q observed %q", expected.PolicyID, observed.PolicyID)
	}
	if observed.Package != expected.Package || observed.Namespace != expected.Namespace {
		return fmt.Sprintf("identity: expected package=%q namespace=%q observed package=%q namespace=%q", expected.Package, expected.Namespace, observed.Package, observed.Namespace)
	}
	if observed.SourceDigest != expected.SourceDigest {
		return fmt.Sprintf("source_digest: expected %q observed %q", expected.SourceDigest, observed.SourceDigest)
	}
	if observed.SemanticDigest != expected.SemanticDigest {
		return fmt.Sprintf("semantic_digest: expected %q observed %q", expected.SemanticDigest, observed.SemanticDigest)
	}
	if observed.Denominator != expected.Denominator {
		return fmt.Sprintf("fixed_denominator: expected %d observed %d", expected.Denominator, observed.Denominator)
	}
	if !bytes.Equal(mustJSON(observed.Rules), mustJSON(expected.Rules)) {
		return "rules"
	}
	if !bytes.Equal(mustJSON(observed.Reduction), mustJSON(expected.Reduction)) {
		return "decision_reduction"
	}
	return ""
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

func parseRawPolicy(filename string, source []byte, expectedPackage, expectedNamespace string) (policy, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return policy{}, errors.New(diagnostics.Error().Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return policy{}, fmt.Errorf("lower raw policy: %w", err)
	}
	if ir.Package != expectedPackage || ir.Namespace.String() != expectedNamespace {
		return policy{}, errors.New("unexpected policy package or namespace")
	}
	if len(ir.Policies) > 1 {
		return policy{}, errors.New("raw policy contains multiple first-class policies")
	}
	if len(ir.Policies) == 1 {
		return parseFirstClassPolicy(ir, source)
	}
	result := policy{Schema: policyWireSchema, PolicyID: "gooo://meta-policy-compilation/policy/v3", Package: ir.Package, Namespace: ir.Namespace.String(), SourceDigest: digestBytes(source), SemanticDigest: "sha256:" + ir.StableHash(), Denominator: fixedDenom, Rules: make([]rule, 0, fixedDenom)}
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

func parseFirstClassPolicy(ir semantic.IR, source []byte) (policy, error) {
	declaration := ir.Policies[0]
	if len(declaration.Cases) != fixedDenom || len(declaration.Transitions) != fixedDenom {
		return policy{}, errors.New("first-class policy denominator changed")
	}
	result := policy{Schema: policyWireSchema, PolicyID: string(declaration.ID), Package: ir.Package, Namespace: ir.Namespace.String(), SourceDigest: digestBytes(source), SemanticDigest: "sha256:" + ir.StableHash(), Denominator: fixedDenom, Rules: make([]rule, 0, fixedDenom), Reduction: reduction{Schema: "decision-reduction:v2", Rules: make([]decisionRule, 0, fixedDenom)}}
	for _, current := range declaration.Cases {
		resolution := current.Resolution
		result.Rules = append(result.Rules, rule{ActivityID: string(declaration.ID) + "/case/" + strings.ToLower(current.Name), ActivityName: current.Name, Role: resolution.Role, MetaOperation: resolution.MetaOperation, ProofChoice: resolution.ProofChoice, Stage: resolution.Stage, Step: resolution.Step, Reason: resolution.Reason, Claim: resolution.Claim})
		result.Reduction.Rules = append(result.Reduction.Rules, decisionRule{Condition: current.Name, Decision: resolution.Decision, Stage: resolution.DecisionStage, Step: resolution.DecisionStep, Reason: resolution.DecisionReason, UnknownClass: resolution.UnknownClass, NextOperation: resolution.NextOperation, BlockedBy: append([]string(nil), resolution.BlockedBy...)})
	}
	sort.Slice(result.Rules, func(i, j int) bool { return result.Rules[i].Step < result.Rules[j].Step })
	if len(result.Rules) != fixedDenom || len(result.Reduction.Rules) != fixedDenom {
		return policy{}, errors.New("first-class policy did not produce fixed source contract")
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
			output.Decision, output.MatchedCondition, output.Stage, output.Step, output.Reason = row.Decision, row.Condition, row.Stage, row.Step, row.Reason
			output.UnknownClass, output.NextOperation, output.BlockedBy = row.UnknownClass, row.NextOperation, append([]string(nil), row.BlockedBy...)
			if output.Decision != decisionUnknown {
				output.UnknownClass, output.NextOperation, output.BlockedBy = "", "", []string{}
			}
			return output
		}
	}
	output.Decision, output.MatchedCondition, output.Stage, output.Reason = decisionFail, "", "COMPILE", "NO_REDUCTION_RULE_MATCHED"
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
	return left.CaseID == right.CaseID && left.Decision == right.Decision && left.MatchedCondition == right.MatchedCondition && left.Stage == right.Stage && left.Step == right.Step && left.Reason == right.Reason && left.UnknownClass == right.UnknownClass && left.NextOperation == right.NextOperation && sameStrings(left.BlockedBy, right.BlockedBy) && left.PolicyDigest == right.PolicyDigest && left.SemanticDigest == right.SemanticDigest && left.Denominator == right.Denominator
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
	var value T
	if err := decodeStrictDocument(data, &value); err != nil {
		return zero, fmt.Errorf("decode %s: %w", path, err)
	}
	return value, nil
}

func readJSONBytes[T any](path string) (T, []byte, error) {
	data, err := os.ReadFile(path)
	var zero T
	if err != nil {
		return zero, nil, fmt.Errorf("read %s: %w", path, err)
	}
	var value T
	if err := decodeStrictDocument(data, &value); err != nil {
		return zero, nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return value, data, nil
}

func readGenerationManifest(path string) (generationManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return generationManifest{}, fmt.Errorf("read generation manifest: %w", err)
	}
	var manifest generationManifest
	if err := decodeRequiredDocument(data, &manifest, []string{
		"schema", "profile", "source_file", "source_digest", "semantic_digest", "package", "namespace",
		"policy_bytes_digest", "artifact_bytes_digest", "generated_judge_digest", "generated_files",
		"output_root_class", "execution_observed", "current_conformance", "repository_writes",
		"mutation_authority", "promotion_authority",
	}); err != nil {
		return generationManifest{}, fmt.Errorf("decode generation manifest: %w", err)
	}
	return manifest, nil
}

func decodeStrictDocument(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("json document contains trailing data")
	}
	return nil
}

func decodeRequiredDocument(data []byte, target any, required []string) error {
	var fields map[string]json.RawMessage
	if err := decodeStrictDocument(data, &fields); err != nil {
		return err
	}
	for _, field := range required {
		value, ok := fields[field]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("required JSON field %q is missing", field)
		}
	}
	return decodeStrictDocument(data, target)
}
