package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	consumerSchema       = "gooo/meta-policy-compilation-consumer/v1"
	policySchema         = "gooo/meta-policy-compilation/v2"
	artifactSchema       = "gooo/meta-policy-compilation-artifact/v2"
	receiptSchema        = "gooo/meta-policy-compilation-receipt/v2"
	reductionSchema      = "decision-reduction:v1"
	fixedDenominator     = 8
	reductionRuleCount   = 6
	decisionPass         = "PASS"
	decisionFailClosed   = "FAIL_CLOSED"
	decisionUnknown      = "UNKNOWN"
	evidenceSynthetic    = "SYNTHETIC_FIXTURE"
	evidenceCurrent      = "CURRENT_EVIDENCE"
	subjectUnresolved    = "UNRESOLVED"
	subjectResolved      = "RESOLVED"
	conditionUnavailable = "EVIDENCE_UNAVAILABLE"
	conditionEmpty       = "DIGEST_UNAVAILABLE"
	conditionSource      = "SOURCE_DIGEST_MISMATCH"
	conditionArtifact    = "ARTIFACT_SOURCE_MISMATCH"
	conditionIndependent = "INDEPENDENT_SOURCE_MISMATCH"
	conditionEquivalent  = "SEMANTIC_EQUIVALENCE"
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

type reductionRule struct {
	Condition string `json:"condition"`
	Decision  string `json:"decision"`
	Stage     string `json:"stage"`
	Step      int    `json:"step"`
	Reason    string `json:"reason"`
}

type reduction struct {
	Schema string          `json:"schema"`
	Rules  []reductionRule `json:"rules"`
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
	ObservedIndependentDigest    string `json:"observed_independent_digest"`
}

type result struct {
	CaseID         string `json:"case_id"`
	Decision       string `json:"decision"`
	Stage          string `json:"stage"`
	Step           int    `json:"step"`
	Reason         string `json:"reason"`
	PolicyDigest   string `json:"policy_digest"`
	SemanticDigest string `json:"semantic_digest"`
	Denominator    int    `json:"fixed_denominator"`
}

type claimTransition struct {
	Event             int    `json:"event"`
	ClaimID           string `json:"claim_id"`
	From              string `json:"from"`
	To                string `json:"to"`
	Decision          string `json:"decision"`
	Stage             string `json:"stage"`
	Step              int    `json:"step"`
	Reason            string `json:"reason"`
	ObservationDigest string `json:"observation_digest"`
	Provenance        string `json:"provenance"`
	PriorDigest       string `json:"prior_digest"`
	Digest            string `json:"digest"`
}

type claimLedger struct {
	Schema     string            `json:"schema"`
	EventCount int               `json:"event_count"`
	Events     []claimTransition `json:"events"`
	HeadDigest string            `json:"head_digest"`
}

type summary struct {
	CaseCount                      int `json:"case_count"`
	PassCount                      int `json:"pass_count"`
	FailClosedCount                int `json:"fail_closed_count"`
	UnknownCount                   int `json:"unknown_count"`
	GeneratedIndependentEqual      int `json:"generated_independent_equivalent"`
	ValidatorExpectationsConfirmed int `json:"validator_expectations_confirmed"`
	SourceAllEquivalent            int `json:"source_all_equivalent"`
}

type verification struct {
	Decision            string `json:"decision"`
	ConformanceDecision string `json:"conformance_decision"`
	SubjectResolution   string `json:"subject_resolution"`
	IndependentReplayed bool   `json:"independent_replayed"`
	GeneratedReplayed   bool   `json:"generated_replayed"`
	LedgerVerified      bool   `json:"ledger_verified"`
	FixedDenominator    int    `json:"fixed_denominator"`
	CaseDenominator     int    `json:"case_denominator"`
}

type caseReceipt struct {
	ID                   string `json:"id"`
	ValidatorExpectation string `json:"validator_expectation"`
	EvidenceClass        string `json:"evidence_class"`
	ObservationDigest    string `json:"observation_digest"`
	Provenance           string `json:"provenance"`
	Source               result `json:"source"`
	Generated            result `json:"generated"`
	Independent          result `json:"independent"`
	AllEquivalent        bool   `json:"all_decisions_equivalent"`
	Equivalent           bool   `json:"decisions_equivalent"`
	ValidatorConfirmed   bool   `json:"validator_expectation_confirmed"`
	ClaimStartDigest     string `json:"claim_start_digest"`
	ClaimEndDigest       string `json:"claim_end_digest"`
}

type writeSet struct {
	RepositoryBeforeDigest string   `json:"repository_before_digest"`
	RepositoryAfterDigest  string   `json:"repository_after_digest"`
	RepositoryBeforeCount  int      `json:"repository_before_count"`
	RepositoryAfterCount   int      `json:"repository_after_count"`
	RepositoryWriteChanged bool     `json:"repository_write_changed"`
	GeneratedRootClass     string   `json:"generated_root_class"`
	GeneratedFiles         []string `json:"generated_files"`
	MutationAuthority      int      `json:"mutation_authority"`
	PromotionAuthority     int      `json:"promotion_authority"`
}

type receipt struct {
	Schema          string                `json:"schema"`
	Policy          policy                `json:"policy"`
	Producer        json.RawMessage       `json:"producer"`
	Consumer        json.RawMessage       `json:"consumer"`
	MetaOperation   string                `json:"meta_operation"`
	ProofChoice     string                `json:"proof_choice"`
	GeneratedDigest string                `json:"generated_judge_digest"`
	Cases           []caseReceipt         `json:"cases"`
	Summary         summary               `json:"summary"`
	Claims          claimLedger           `json:"claims"`
	Verification    verification          `json:"verification"`
	Evidence        []evidenceObservation `json:"evidence"`
	WriteSet        writeSet              `json:"write_set"`
	ReceiptDigest   string                `json:"receipt_digest"`
}

type evidenceObservation struct {
	Class             string `json:"class"`
	CaseID            string `json:"case_id"`
	ProducerAvailable bool   `json:"producer_available"`
	ConsumerAvailable bool   `json:"consumer_available"`
	SourceDigest      string `json:"source_digest"`
	ArtifactDigest    string `json:"artifact_digest"`
	ObservationDigest string `json:"observation_digest"`
	Provenance        string `json:"provenance"`
}

type importBoundary struct {
	ProducerCompilerImports     string `json:"producer_compiler_imports"`
	GeneratedTemplateImports    string `json:"generated_template_imports"`
	IndependentEvaluatorImports string `json:"independent_evaluator_imports"`
}

type consumerReport struct {
	Schema                                string                `json:"schema"`
	RawPolicyParsed                       bool                  `json:"raw_policy_parsed"`
	RawCasesParsed                        bool                  `json:"raw_cases_parsed"`
	GoooDerivedRuleNumerator              int                   `json:"gooo_derived_rule_numerator"`
	GoooDerivedRuleDenominator            int                   `json:"gooo_derived_rule_denominator"`
	GeneratedExecutionsNumerator          int                   `json:"generated_executions_numerator"`
	GeneratedExecutionsDenominator        int                   `json:"generated_executions_denominator"`
	IndependentReconstructionsNumerator   int                   `json:"independent_reconstructions_numerator"`
	IndependentReconstructionsDenominator int                   `json:"independent_reconstructions_denominator"`
	ImportBoundary                        importBoundary        `json:"import_boundary"`
	SyntheticEvidence                     []evidenceObservation `json:"synthetic_evidence"`
	CurrentEvidence                       evidenceObservation   `json:"current_evidence"`
	SubjectResolution                     string                `json:"subject_resolution"`
	SubjectDecision                       result                `json:"subject_decision"`
	IndependentResults                    []result              `json:"independent_results"`
}

func main() {
	policyPath := flag.String("policy", "", "raw Gooo policy source")
	casesPath := flag.String("cases", "", "raw cases and observations")
	artifactDir := flag.String("artifact", "", "producer artifact directory")
	outputPath := flag.String("output", "", "consumer report path")
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
	source, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("read raw policy: %w", err)
	}
	compiled, err := parseRawPolicy(source)
	if err != nil {
		return err
	}
	cases, err := readRawCases(casesPath, compiled.SourceDigest)
	if err != nil {
		return err
	}
	artifact, err := readStrict[artifact](filepath.Join(artifactDir, "artifact.json"))
	if err != nil {
		return err
	}
	if artifact.Schema != artifactSchema || !samePolicy(artifact.Policy, compiled) {
		return errors.New("consumer observed an artifact that differs from its raw-source reconstruction")
	}
	judge, err := os.ReadFile(filepath.Join(artifactDir, "judge.go"))
	if err != nil {
		return fmt.Errorf("read generated judge: %w", err)
	}
	if digestBytes(judge) != artifact.GeneratedJudgeHash {
		return errors.New("consumer observed a generated judge digest mismatch")
	}
	storedGenerated, err := readStrict[[]result](filepath.Join(artifactDir, "generated-results.json"))
	if err != nil {
		return err
	}
	storedIndependent, err := readStrict[[]result](filepath.Join(artifactDir, "independent-results.json"))
	if err != nil {
		return err
	}
	storedReceipt, err := readStrict[receipt](filepath.Join(artifactDir, "receipt.json"))
	if err != nil {
		return err
	}
	if storedReceipt.Schema != receiptSchema || !samePolicy(storedReceipt.Policy, compiled) || storedReceipt.GeneratedDigest != artifact.GeneratedJudgeHash {
		return errors.New("consumer observed a receipt not bound to the raw source and artifact")
	}
	if storedReceipt.Verification.Decision != "PASS" || storedReceipt.Verification.ConformanceDecision != "PASS" || storedReceipt.Verification.SubjectResolution != subjectUnresolved || !storedReceipt.Verification.IndependentReplayed || !storedReceipt.Verification.GeneratedReplayed || !storedReceipt.Verification.LedgerVerified || storedReceipt.Verification.FixedDenominator != fixedDenominator || storedReceipt.Verification.CaseDenominator != len(cases) {
		return errors.New("consumer observed conflated conformance and subject resolution")
	}
	if len(storedGenerated) != len(cases) || len(storedIndependent) != len(cases) || len(storedReceipt.Cases) != len(cases) {
		return errors.New("consumer observed incomplete result denominators")
	}
	if storedReceipt.Summary.CaseCount != len(cases) || storedReceipt.Summary.PassCount != 1 || storedReceipt.Summary.FailClosedCount != 1 || storedReceipt.Summary.UnknownCount != 1 || storedReceipt.Summary.GeneratedIndependentEqual != len(cases) || storedReceipt.Summary.SourceAllEquivalent != len(cases) || storedReceipt.Summary.ValidatorExpectationsConfirmed != len(cases) {
		return errors.New("consumer observed an incomplete conformance summary")
	}
	generatedByID := make(map[string]result, len(storedGenerated))
	independentByID := make(map[string]result, len(storedIndependent))
	for _, value := range storedGenerated {
		generatedByID[value.CaseID] = value
	}
	for _, value := range storedIndependent {
		independentByID[value.CaseID] = value
	}
	ctx := commandContext()
	reconstructed := make([]result, 0, len(cases))
	generatedCount := 0
	independentCount := 0
	for _, value := range cases {
		generated, err := executeJudge(ctx, filepath.Join(artifactDir, "judge.go"), artifactDir, value)
		if err != nil {
			return err
		}
		stored, ok := generatedByID[value.ID]
		if !ok || !sameResult(generated, stored) {
			return fmt.Errorf("generated execution for %q was not reproduced", value.ID)
		}
		generatedCount++
		independent := evaluate(compiled, value)
		stored, ok = independentByID[value.ID]
		if !ok || !sameResult(independent, stored) {
			return fmt.Errorf("independent reconstruction for %q was not reproduced", value.ID)
		}
		independentCount++
		reconstructed = append(reconstructed, independent)
	}
	if err := verifyReceiptObservations(storedReceipt, cases, compiled, artifact); err != nil {
		return err
	}
	current := observeCurrentEvidence(artifactDir, compiled, source)
	currentInput := input{ID: "current-subject", EvidenceClass: evidenceCurrent, Provenance: current.Provenance, ProducerAvailable: current.ProducerAvailable, ConsumerAvailable: current.ConsumerAvailable, ObservedSourceDigest: current.SourceDigest, ObservedArtifactSourceDigest: artifact.Policy.SourceDigest, ObservedIndependentDigest: compiled.SourceDigest}
	currentDecision := evaluate(compiled, currentInput)
	if currentDecision.Decision == decisionPass {
		currentDecision.Reason = "CURRENT_ARTIFACT_BOUND"
	}
	report := consumerReport{
		Schema: consumerSchema, RawPolicyParsed: true, RawCasesParsed: true,
		GoooDerivedRuleNumerator: len(compiled.Rules), GoooDerivedRuleDenominator: fixedDenominator,
		GeneratedExecutionsNumerator: generatedCount, GeneratedExecutionsDenominator: len(cases),
		IndependentReconstructionsNumerator: independentCount, IndependentReconstructionsDenominator: len(cases),
		ImportBoundary:    importBoundary{ProducerCompilerImports: "0/1", GeneratedTemplateImports: "0/1", IndependentEvaluatorImports: "0/1"},
		SyntheticEvidence: make([]evidenceObservation, 0, len(cases)), CurrentEvidence: current,
		SubjectResolution: subjectUnresolved, SubjectDecision: currentDecision, IndependentResults: reconstructed,
	}
	for _, value := range cases {
		report.SyntheticEvidence = append(report.SyntheticEvidence, evidenceObservation{Class: value.EvidenceClass, CaseID: value.ID, ProducerAvailable: value.ProducerAvailable, ConsumerAvailable: value.ConsumerAvailable, SourceDigest: value.ObservedSourceDigest, ArtifactDigest: digestBytes(mustRead(filepath.Join(artifactDir, "artifact.json"))), ObservationDigest: digestInput(value), Provenance: value.Provenance})
	}
	if currentDecision.Decision == decisionPass {
		report.SubjectResolution = subjectResolved
	}
	return writeStrict(outputPath, report)
}

func parseRawPolicy(source []byte) (policy, error) {
	file, diagnostics := syntax.ParseFile("policy.gooo", string(source))
	if diagnostics.HasErrors() {
		return policy{}, errors.New(diagnostics.Error().Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return policy{}, fmt.Errorf("consumer lower raw policy: %w", err)
	}
	if ir.Package != "metapolicycompilation" || ir.Namespace.String() != "metapolicycompilation" {
		return policy{}, errors.New("consumer rejected raw policy package/namespace")
	}
	result := policy{Schema: policySchema, PolicyID: "gooo://meta-policy-compilation/policy/v2", Package: ir.Package, Namespace: ir.Namespace.String(), SourceDigest: digestBytes(source), SemanticDigest: ir.StableHash(), Denominator: fixedDenominator, Rules: make([]rule, 0, fixedDenominator)}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity {
			continue
		}
		values, err := parseActivity(node.ValueProgram)
		if err != nil {
			return policy{}, fmt.Errorf("consumer activity %q: %w", node.Name, err)
		}
		if values.reduction != "" {
			if result.Reduction.Schema != "" {
				return policy{}, errors.New("consumer found multiple decision reductions")
			}
			result.Reduction, err = parseReduction(values.reduction)
			if err != nil {
				return policy{}, err
			}
		}
		result.Rules = append(result.Rules, rule{ActivityID: string(node.ID), ActivityName: node.Name, Role: values.role, MetaOperation: values.metaOperation, ProofChoice: values.proofChoice, Stage: values.stage, Step: values.step, Reason: values.reason, Claim: values.claim})
	}
	if len(result.Rules) != fixedDenominator || result.Reduction.Schema != reductionSchema || len(result.Reduction.Rules) != reductionRuleCount {
		return policy{}, errors.New("consumer rejected raw policy fixed safety shape")
	}
	sort.Slice(result.Rules, func(i, j int) bool { return result.Rules[i].Step < result.Rules[j].Step })
	for index, value := range result.Rules {
		if value.Step != index+1 {
			return policy{}, errors.New("consumer rejected duplicate or missing source rule step")
		}
	}
	return result, nil
}

type activityValues struct {
	role, metaOperation, proofChoice, stage, reason, claim, reduction string
	step                                                              int
}

var safeToken = regexp.MustCompile(`^[A-Za-z0-9_.:/-]+$`)

func parseActivity(value string) (activityValues, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 8 || parts[0] != "policy-compilation:v2" {
		return activityValues{}, errors.New("consumer rejected activity schema")
	}
	values := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, field, ok := strings.Cut(part, "=")
		if !ok || field == "" || values[key] != "" || !map[string]bool{"role": true, "meta-operation": true, "proof-choice": true, "stage": true, "step": true, "reason": true, "claim": true, "decision-reduction": true}[key] {
			return activityValues{}, fmt.Errorf("consumer rejected activity field %q", part)
		}
		values[key] = field
	}
	step, err := strconv.Atoi(values["step"])
	if err != nil || step < 1 || step > fixedDenominator {
		return activityValues{}, errors.New("consumer rejected unsafe activity step")
	}
	for _, key := range []string{"role", "meta-operation", "proof-choice", "stage", "reason", "claim"} {
		if values[key] == "" || !safeToken.MatchString(values[key]) {
			return activityValues{}, fmt.Errorf("consumer rejected activity value %q", key)
		}
	}
	return activityValues{role: values["role"], metaOperation: values["meta-operation"], proofChoice: values["proof-choice"], stage: values["stage"], step: step, reason: values["reason"], claim: values["claim"], reduction: values["decision-reduction"]}, nil
}

func parseReduction(value string) (reduction, error) {
	parts := strings.Split(value, ";")
	if len(parts) != reductionRuleCount+2 || parts[0] != reductionSchema || strings.TrimPrefix(parts[1], "denominator=") != strconv.Itoa(reductionRuleCount) {
		return reduction{}, errors.New("consumer rejected decision reduction schema")
	}
	result := reduction{Schema: reductionSchema, Rules: make([]reductionRule, 0, reductionRuleCount)}
	seen := make(map[string]bool, reductionRuleCount)
	for _, encoded := range parts[2:] {
		fields := strings.Split(encoded, ":")
		if len(fields) != 5 {
			return reduction{}, fmt.Errorf("consumer rejected decision rule %q", encoded)
		}
		step, err := strconv.Atoi(fields[3])
		if err != nil || step < 1 || step > fixedDenominator || !knownCondition(fields[0]) || !knownDecision(fields[1]) || !safeToken.MatchString(fields[2]) || !safeToken.MatchString(fields[4]) || seen[fields[0]] {
			return reduction{}, fmt.Errorf("consumer rejected decision rule %q", encoded)
		}
		seen[fields[0]] = true
		result.Rules = append(result.Rules, reductionRule{Condition: fields[0], Decision: fields[1], Stage: fields[2], Step: step, Reason: fields[4]})
	}
	if !seen[conditionEquivalent] {
		return reduction{}, errors.New("consumer rejected reduction without terminal equivalence rule")
	}
	return result, nil
}

func readRawCases(path, sourceDigest string) ([]input, error) {
	values, err := readStrict[[]input](path)
	if err != nil {
		return nil, err
	}
	if len(values) != 3 {
		return nil, fmt.Errorf("consumer rejected synthetic case denominator %d", len(values))
	}
	seen := make(map[string]bool, len(values))
	for index := range values {
		value := &values[index]
		if value.ID == "" || seen[value.ID] || value.EvidenceClass != evidenceSynthetic || value.Provenance == "" {
			return nil, fmt.Errorf("consumer rejected synthetic evidence declaration for %q", value.ID)
		}
		seen[value.ID] = true
		if value.ValidatorExpectation != decisionPass && value.ValidatorExpectation != decisionFailClosed && value.ValidatorExpectation != decisionUnknown {
			return nil, fmt.Errorf("consumer rejected validator expectation for %q", value.ID)
		}
		value.ObservedSourceDigest = bindDigest(value.ObservedSourceDigest, sourceDigest)
		value.ObservedArtifactSourceDigest = bindDigest(value.ObservedArtifactSourceDigest, sourceDigest)
		value.ObservedIndependentDigest = bindDigest(value.ObservedIndependentDigest, sourceDigest)
	}
	return values, nil
}

func bindDigest(value, sourceDigest string) string {
	if value == "SOURCE_DIGEST_FROM_POLICY" {
		return sourceDigest
	}
	return value
}

func evaluate(policy policy, value input) result {
	base := result{CaseID: value.ID, PolicyDigest: policy.SourceDigest, SemanticDigest: policy.SemanticDigest, Denominator: policy.Denominator}
	if policy.Denominator != fixedDenominator || len(policy.Rules) != fixedDenominator || policy.Reduction.Schema != reductionSchema || len(policy.Reduction.Rules) != reductionRuleCount {
		return setSafety(base, "FIXED_DENOMINATOR_CHANGED")
	}
	for _, row := range policy.Reduction.Rules {
		if matches(row.Condition, policy.SourceDigest, value) {
			base.Decision, base.Stage, base.Step, base.Reason = row.Decision, row.Stage, row.Step, row.Reason
			return base
		}
	}
	return setSafety(base, "NO_REDUCTION_RULE_MATCHED")
}

func matches(condition, sourceDigest string, value input) bool {
	available := value.ProducerAvailable && value.ConsumerAvailable
	switch condition {
	case conditionUnavailable:
		return !available
	case conditionEmpty:
		return available && (value.ObservedSourceDigest == "" || value.ObservedArtifactSourceDigest == "" || value.ObservedIndependentDigest == "")
	case conditionSource:
		return available && value.ObservedSourceDigest != "" && value.ObservedArtifactSourceDigest != "" && value.ObservedIndependentDigest != "" && value.ObservedSourceDigest != sourceDigest
	case conditionArtifact:
		return available && value.ObservedSourceDigest == sourceDigest && value.ObservedArtifactSourceDigest != "" && value.ObservedArtifactSourceDigest != sourceDigest
	case conditionIndependent:
		return available && value.ObservedSourceDigest == sourceDigest && value.ObservedArtifactSourceDigest == sourceDigest && value.ObservedIndependentDigest != "" && value.ObservedIndependentDigest != sourceDigest
	case conditionEquivalent:
		return available && value.ObservedSourceDigest == sourceDigest && value.ObservedArtifactSourceDigest == sourceDigest && value.ObservedIndependentDigest == sourceDigest
	default:
		return false
	}
}

func setSafety(value result, reason string) result {
	value.Decision, value.Stage, value.Step, value.Reason = decisionFailClosed, "COMPILE", 3, reason
	return value
}

func executeJudge(ctx commandContextValue, path, dir string, value input) (result, error) {
	payload, err := json.Marshal(struct {
		ID                           string `json:"id"`
		ProducerAvailable            bool   `json:"producer_available"`
		ConsumerAvailable            bool   `json:"consumer_available"`
		ObservedSourceDigest         string `json:"observed_source_digest"`
		ObservedArtifactSourceDigest string `json:"observed_artifact_source_digest"`
		ObservedIndependentDigest    string `json:"observed_independent_digest"`
	}{ID: value.ID, ProducerAvailable: value.ProducerAvailable, ConsumerAvailable: value.ConsumerAvailable, ObservedSourceDigest: value.ObservedSourceDigest, ObservedArtifactSourceDigest: value.ObservedArtifactSourceDigest, ObservedIndependentDigest: value.ObservedIndependentDigest})
	if err != nil {
		return result{}, err
	}
	command := exec.CommandContext(ctx.context, "go", "run", path)
	command.Dir = dir
	command.Stdin = bytes.NewReader(payload)
	command.Env = append(os.Environ(), "GO111MODULE=off", "GOTOOLCHAIN=go1.27.0")
	output, err := command.Output()
	if err != nil {
		return result{}, fmt.Errorf("consumer generated-judge execution: %w", err)
	}
	return decodeBytes[result](output)
}

type commandContextValue struct{ context context.Context }

func commandContext() commandContextValue {
	return commandContextValue{context: context.Background()}
}

func verifyReceiptObservations(value receipt, cases []input, policy policy, artifact artifact) error {
	if len(value.Cases) != len(cases) || len(value.Evidence) != len(cases) || len(value.Claims.Events) != len(cases)*fixedDenominator*2 || value.Claims.EventCount != len(value.Claims.Events) || value.Claims.Schema != "gooo/meta-policy-compilation-claims/v2" {
		return errors.New("consumer rejected claim or case denominator")
	}
	byID := make(map[string]caseReceipt, len(value.Cases))
	for _, stored := range value.Cases {
		byID[stored.ID] = stored
	}
	for _, input := range cases {
		stored, ok := byID[input.ID]
		if !ok || stored.ValidatorExpectation != input.ValidatorExpectation || stored.EvidenceClass != input.EvidenceClass || stored.Provenance != input.Provenance || stored.ObservationDigest != digestInput(input) {
			return fmt.Errorf("consumer rejected receipt observation for %q", input.ID)
		}
		source := evaluate(policy, input)
		independent := evaluate(policy, input)
		if !sameResult(stored.Source, source) || !sameResult(stored.Independent, independent) || !sameResult(stored.Source, stored.Generated) || !stored.AllEquivalent || !stored.Equivalent || !stored.ValidatorConfirmed {
			return fmt.Errorf("consumer rejected source/generated/independent lineage for %q", input.ID)
		}
	}
	evidenceByID := make(map[string]evidenceObservation, len(value.Evidence))
	for _, observed := range value.Evidence {
		evidenceByID[observed.CaseID] = observed
	}
	for _, input := range cases {
		observed, ok := evidenceByID[input.ID]
		if !ok || observed.Class != evidenceSynthetic || observed.ProducerAvailable != input.ProducerAvailable || observed.ConsumerAvailable != input.ConsumerAvailable || observed.SourceDigest != input.ObservedSourceDigest || observed.ObservationDigest != digestInput(input) || observed.Provenance != input.Provenance {
			return fmt.Errorf("consumer rejected receipt evidence provenance for %q", input.ID)
		}
	}
	if err := verifyClaimChain(value.Claims, cases); err != nil {
		return err
	}
	if value.WriteSet.RepositoryBeforeDigest == "" || value.WriteSet.RepositoryBeforeDigest != value.WriteSet.RepositoryAfterDigest || value.WriteSet.RepositoryWriteChanged || value.WriteSet.RepositoryBeforeCount != value.WriteSet.RepositoryAfterCount || value.WriteSet.GeneratedRootClass != "RUNNER_TEMP_ONLY" || value.WriteSet.MutationAuthority != 0 || value.WriteSet.PromotionAuthority != 0 {
		return errors.New("consumer rejected generated write-set authority")
	}
	if value.Policy.SourceDigest != policy.SourceDigest || artifact.Policy.SourceDigest != policy.SourceDigest {
		return errors.New("consumer rejected source digest lineage")
	}
	return nil
}

func verifyClaimChain(ledger claimLedger, cases []input) error {
	prior := ""
	counts := make(map[string]int, len(ledger.Events)/2)
	states := make(map[string]string, len(ledger.Events)/2)
	allowed := make(map[string]string, len(cases))
	for _, value := range cases {
		allowed[digestInput(value)] = value.Provenance
	}
	for index, event := range ledger.Events {
		if event.Event != index+1 || event.PriorDigest != prior || event.ClaimID == "" || event.ObservationDigest == "" || event.Provenance == "" || allowed[event.ObservationDigest] != event.Provenance || counts[event.ClaimID] >= 2 {
			return fmt.Errorf("consumer claim chain broken at event %d", event.Event)
		}
		canonical := event
		canonical.Digest = ""
		if digestJSON(canonical) != event.Digest {
			return fmt.Errorf("consumer claim digest broken at event %d", event.Event)
		}
		if !validTransition(event.From, event.To) {
			return fmt.Errorf("consumer claim transition %s -> %s is invalid", event.From, event.To)
		}
		if current, ok := states[event.ClaimID]; ok && current != event.From {
			return fmt.Errorf("consumer claim %q does not continue", event.ClaimID)
		}
		if _, ok := states[event.ClaimID]; !ok && event.From != "UNRECORDED" {
			return fmt.Errorf("consumer claim %q did not begin at UNRECORDED", event.ClaimID)
		}
		counts[event.ClaimID]++
		states[event.ClaimID] = event.To
		prior = event.Digest
	}
	for claimID, count := range counts {
		if count != 2 || states[claimID] == "UNRECORDED" {
			return fmt.Errorf("consumer claim %q does not have two transitions", claimID)
		}
	}
	if prior != ledger.HeadDigest {
		return errors.New("consumer claim head digest mismatch")
	}
	return nil
}

func validTransition(from, to string) bool {
	return (from == "UNRECORDED" && to == "OPEN") || (from == "OPEN" && (to == "OPEN" || to == "DISCHARGED" || to == "REFUTED"))
}

func observeCurrentEvidence(dir string, policy policy, source []byte) evidenceObservation {
	files := []string{"artifact.json", "generated-results.json", "independent-results.json", "judge.go", "policy.json", "receipt.json"}
	available := true
	for _, name := range files {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			available = false
		}
	}
	artifactBytes, err := os.ReadFile(filepath.Join(dir, "artifact.json"))
	artifactDigest := ""
	if err == nil {
		artifactDigest = digestBytes(artifactBytes)
	}
	provenance := "runner-temp producer artifact observed by raw-source consumer"
	observation := evidenceObservation{Class: evidenceCurrent, CaseID: "current-subject", ProducerAvailable: available, ConsumerAvailable: true, SourceDigest: digestBytes(source), ArtifactDigest: artifactDigest, Provenance: provenance}
	observation.ObservationDigest = digestEvidence(observation)
	return observation
}

func samePolicy(left, right policy) bool {
	return left.Schema == right.Schema && left.PolicyID == right.PolicyID && left.Package == right.Package && left.Namespace == right.Namespace && left.SourceDigest == right.SourceDigest && left.SemanticDigest == right.SemanticDigest && left.Denominator == right.Denominator && equalJSON(left.Rules, right.Rules) && equalJSON(left.Reduction, right.Reduction)
}

func sameResult(left, right result) bool {
	return left == right
}

func digestInput(value input) string { return digestJSON(value) }

func digestEvidence(value evidenceObservation) string {
	value.ObservationDigest = ""
	return digestJSON(value)
}

func digestJSON(value any) string {
	data, _ := json.Marshal(value)
	return digestBytes(data)
}

func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func equalJSON(left, right any) bool {
	leftBytes, _ := json.Marshal(left)
	rightBytes, _ := json.Marshal(right)
	return bytes.Equal(leftBytes, rightBytes)
}

func readStrict[T any](path string) (T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("read %s: %w", path, err)
	}
	return decodeBytes[T](data)
}

func decodeBytes[T any](data []byte) (T, error) {
	var value T
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return value, errors.New("trailing JSON")
	}
	return value, nil
}

func writeStrict(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o640)
}

func mustRead(path string) []byte {
	data, _ := os.ReadFile(path)
	return data
}

func knownCondition(value string) bool {
	switch value {
	case conditionUnavailable, conditionEmpty, conditionSource, conditionArtifact, conditionIndependent, conditionEquivalent:
		return true
	default:
		return false
	}
}

func knownDecision(value string) bool {
	return value == decisionPass || value == decisionFailClosed || value == decisionUnknown
}
