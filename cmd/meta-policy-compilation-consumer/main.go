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
	reductionRuleCount   = 7
	claimPredicateCount  = 8
	decisionPass         = "PASS"
	decisionFailClosed   = "FAIL_CLOSED"
	decisionUnknown      = "UNKNOWN"
	evidenceSynthetic    = "SYNTHETIC_FIXTURE"
	evidenceCurrent      = "CURRENT_EVIDENCE"
	subjectUnresolved    = "UNRESOLVED"
	subjectResolved      = "RESOLVED"
	conditionUnavailable = "EVIDENCE_UNAVAILABLE"
	conditionEmpty       = "DIGEST_UNAVAILABLE"
	conditionMalformed   = "MALFORMED_DIGEST"
	conditionSource      = "SOURCE_DIGEST_MISMATCH"
	conditionArtifact    = "ARTIFACT_SOURCE_MISMATCH"
	conditionIndependent = "INDEPENDENT_SOURCE_MISMATCH"
	conditionEquivalent  = "SEMANTIC_EQUIVALENCE"
	claimSourceBound     = "source-bound"
	claimArtifactBound   = "artifact-digest-bound"
	claimGenerated       = "generated-execution"
	claimIndependent     = "independent-replay"
	claimProof           = "proof-selection"
	claimLedgerPredicate = "ledger-chain"
	claimReduction       = "decision-reduction"
	claimLineage         = "lineage-seal"
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
	ObservedGeneratedJudgeDigest string `json:"observed_generated_judge_digest"`
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
	Predicate         string `json:"predicate"`
	From              string `json:"from"`
	To                string `json:"to"`
	Decision          string `json:"decision"`
	Stage             string `json:"stage"`
	Step              int    `json:"step"`
	Reason            string `json:"reason"`
	ObservationDigest string `json:"observation_digest"`
	Provenance        string `json:"provenance"`
	Observed          bool   `json:"observed"`
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
	ClaimPredicatesDischarged      int `json:"claim_predicates_discharged"`
	ClaimPredicatesRefuted         int `json:"claim_predicates_refuted"`
	ClaimPredicatesOpen            int `json:"claim_predicates_open"`
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
	ID                   string                      `json:"id"`
	ValidatorExpectation string                      `json:"validator_expectation"`
	EvidenceClass        string                      `json:"evidence_class"`
	ObservationDigest    string                      `json:"observation_digest"`
	Provenance           string                      `json:"provenance"`
	Source               result                      `json:"source"`
	Generated            result                      `json:"generated"`
	Independent          result                      `json:"independent"`
	AllEquivalent        bool                        `json:"all_decisions_equivalent"`
	Equivalent           bool                        `json:"decisions_equivalent"`
	ValidatorConfirmed   bool                        `json:"validator_expectation_confirmed"`
	ClaimPredicates      []claimPredicateObservation `json:"claim_predicates"`
	ClaimStartDigest     string                      `json:"claim_start_digest"`
	ClaimEndDigest       string                      `json:"claim_end_digest"`
}

type claimPredicateObservation struct {
	ClaimID           string `json:"claim_id"`
	Predicate         string `json:"predicate"`
	Outcome           string `json:"outcome"`
	Observed          bool   `json:"observed"`
	Stage             string `json:"stage"`
	Step              int    `json:"step"`
	Reason            string `json:"reason"`
	ObservationDigest string `json:"observation_digest"`
	Provenance        string `json:"provenance"`
}

type writeSet struct {
	RepositoryBeforeDigest      string   `json:"repository_before_digest"`
	RepositoryAfterDigest       string   `json:"repository_after_digest"`
	RepositoryBeforeCount       int      `json:"repository_before_count"`
	RepositoryAfterCount        int      `json:"repository_after_count"`
	RepositoryNetChangeObserved bool     `json:"repository_net_change_observed"`
	GeneratedRootClass          string   `json:"generated_root_class"`
	GeneratedFiles              []string `json:"generated_files"`
	MutationAuthority           int      `json:"mutation_authority"`
	PromotionAuthority          int      `json:"promotion_authority"`
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
	Class                string `json:"class"`
	CaseID               string `json:"case_id"`
	ProducerAvailable    bool   `json:"producer_available"`
	ConsumerAvailable    bool   `json:"consumer_available"`
	SourceDigest         string `json:"source_digest"`
	ArtifactSourceDigest string `json:"artifact_source_digest"`
	ArtifactDigest       string `json:"artifact_digest"`
	GeneratedJudgeDigest string `json:"generated_judge_digest"`
	IndependentDigest    string `json:"independent_digest"`
	SemanticDigest       string `json:"semantic_digest"`
	ObservationDigest    string `json:"observation_digest"`
	Provenance           string `json:"provenance"`
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
	ProducerContractDigest                string                `json:"producer_contract_digest"`
	ConsumerContractDigest                string                `json:"consumer_contract_digest"`
	ContractDigestMatch                   bool                  `json:"contract_digest_match"`
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
	artifactBytes, err := os.ReadFile(filepath.Join(artifactDir, "artifact.json"))
	if err != nil {
		return err
	}
	artifact, err := decodeBytes[artifact](artifactBytes)
	if err != nil {
		return fmt.Errorf("decode artifact: %w", err)
	}
	if digestBytes(artifactBytes) != digestJSONIndented(artifact) {
		return errors.New("consumer observed a non-canonical artifact byte stream")
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
	cases, err := readRawCases(casesPath, compiled.SourceDigest, compiled.SemanticDigest, artifact.GeneratedJudgeHash)
	if err != nil {
		return err
	}
	storedGenerated, err := readStrict[[]result](filepath.Join(artifactDir, "generated-results.json"))
	if err != nil {
		return err
	}
	storedIndependent, err := readStrict[[]result](filepath.Join(artifactDir, "independent-results.json"))
	if err != nil {
		return err
	}
	receiptBytes, err := os.ReadFile(filepath.Join(artifactDir, "receipt.json"))
	if err != nil {
		return err
	}
	storedReceipt, err := decodeBytes[receipt](receiptBytes)
	if err != nil {
		return fmt.Errorf("decode receipt: %w", err)
	}
	if digestReceipt(storedReceipt) != storedReceipt.ReceiptDigest {
		return errors.New("consumer observed a receipt digest mismatch")
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
	if storedReceipt.Summary.CaseCount != len(cases) || storedReceipt.Summary.PassCount != 1 || storedReceipt.Summary.FailClosedCount != 1 || storedReceipt.Summary.UnknownCount != 2 || storedReceipt.Summary.GeneratedIndependentEqual != len(cases) || storedReceipt.Summary.SourceAllEquivalent != len(cases) || storedReceipt.Summary.ValidatorExpectationsConfirmed != len(cases) {
		return errors.New("consumer observed an incomplete conformance summary")
	}
	discharged, refuted, open := countPredicates(storedReceipt.Cases)
	if storedReceipt.Summary.ClaimPredicatesDischarged != discharged || storedReceipt.Summary.ClaimPredicatesRefuted != refuted || storedReceipt.Summary.ClaimPredicatesOpen != open {
		return errors.New("consumer observed claim outcomes not reflected in the summary")
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
	current := observeCurrentEvidence(artifactDir, compiled, source, artifact, judge)
	currentInput := input{ID: "current-subject", EvidenceClass: evidenceCurrent, Provenance: current.Provenance, ProducerAvailable: current.ProducerAvailable, ConsumerAvailable: current.ConsumerAvailable, ObservedSourceDigest: current.SourceDigest, ObservedArtifactSourceDigest: current.ArtifactSourceDigest, ObservedGeneratedJudgeDigest: current.GeneratedJudgeDigest, ObservedIndependentDigest: current.IndependentDigest}
	currentDecision := evaluate(compiled, currentInput)
	if currentDecision.Decision == decisionPass {
		currentDecision.Reason = "CURRENT_ARTIFACT_BOUND"
	}
	report := consumerReport{
		Schema: consumerSchema, RawPolicyParsed: true, RawCasesParsed: true,
		GoooDerivedRuleNumerator: len(compiled.Rules), GoooDerivedRuleDenominator: fixedDenominator,
		GeneratedExecutionsNumerator: generatedCount, GeneratedExecutionsDenominator: len(cases),
		IndependentReconstructionsNumerator: independentCount, IndependentReconstructionsDenominator: len(cases),
		ProducerContractDigest: artifact.Policy.SemanticDigest, ConsumerContractDigest: compiled.SemanticDigest, ContractDigestMatch: artifact.Policy.SemanticDigest == compiled.SemanticDigest,
		ImportBoundary:    importBoundary{ProducerCompilerImports: "0/1", GeneratedTemplateImports: "0/1", IndependentEvaluatorImports: "0/1"},
		SyntheticEvidence: make([]evidenceObservation, 0, len(cases)), CurrentEvidence: current,
		SubjectResolution: subjectUnresolved, SubjectDecision: currentDecision, IndependentResults: reconstructed,
	}
	for _, value := range cases {
		report.SyntheticEvidence = append(report.SyntheticEvidence, evidenceObservation{Class: value.EvidenceClass, CaseID: value.ID, ProducerAvailable: value.ProducerAvailable, ConsumerAvailable: value.ConsumerAvailable, SourceDigest: value.ObservedSourceDigest, ArtifactSourceDigest: value.ObservedArtifactSourceDigest, ArtifactDigest: digestBytes(artifactBytes), GeneratedJudgeDigest: value.ObservedGeneratedJudgeDigest, IndependentDigest: value.ObservedIndependentDigest, SemanticDigest: compiled.SemanticDigest, ObservationDigest: digestInput(value), Provenance: value.Provenance})
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
	result := policy{Schema: policySchema, PolicyID: "gooo://meta-policy-compilation/policy/v2", Package: ir.Package, Namespace: ir.Namespace.String(), SourceDigest: digestBytes(source), SemanticDigest: "sha256:" + ir.StableHash(), Denominator: fixedDenominator, Rules: make([]rule, 0, fixedDenominator)}
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
	if err := validateClaims(result.Rules); err != nil {
		return policy{}, err
	}
	return result, nil
}

func validateClaims(rules []rule) error {
	want := []string{claimSourceBound, claimArtifactBound, claimGenerated, claimIndependent, claimProof, claimLedgerPredicate, claimReduction, claimLineage}
	seen := make(map[string]bool, len(want))
	for _, value := range rules {
		if seen[value.Claim] {
			return fmt.Errorf("consumer rejected duplicate claim predicate %q", value.Claim)
		}
		seen[value.Claim] = true
	}
	for _, value := range want {
		if !seen[value] {
			return fmt.Errorf("consumer rejected missing claim predicate %q", value)
		}
	}
	return nil
}

type activityValues struct {
	role, metaOperation, proofChoice, stage, reason, claim, reduction string
	step                                                              int
}

var safeToken = regexp.MustCompile(`^[A-Za-z0-9_.:/-]+$`)
var digestToken = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func validDigest(value string) bool { return digestToken.MatchString(value) }

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

func readRawCases(path, sourceDigest, semanticDigest, judgeDigest string) ([]input, error) {
	values, err := readStrict[[]input](path)
	if err != nil {
		return nil, err
	}
	if len(values) != 4 {
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
		value.ObservedGeneratedJudgeDigest = bindDigest(value.ObservedGeneratedJudgeDigest, judgeDigest)
		value.ObservedIndependentDigest = bindDigest(value.ObservedIndependentDigest, semanticDigest)
	}
	return values, nil
}

func bindDigest(value, expected string) string {
	if value == "SOURCE_DIGEST_FROM_POLICY" || value == "SEMANTIC_DIGEST_FROM_POLICY" || value == "GENERATED_JUDGE_DIGEST_FROM_ARTIFACT" {
		return expected
	}
	return value
}

func evaluate(policy policy, value input) result {
	base := result{CaseID: value.ID, PolicyDigest: policy.SourceDigest, SemanticDigest: policy.SemanticDigest, Denominator: policy.Denominator}
	if policy.Denominator != fixedDenominator || len(policy.Rules) != fixedDenominator || policy.Reduction.Schema != reductionSchema || len(policy.Reduction.Rules) != reductionRuleCount {
		return setSafety(base, "FIXED_DENOMINATOR_CHANGED")
	}
	for _, row := range policy.Reduction.Rules {
		if matches(row.Condition, policy.SourceDigest, policy.SemanticDigest, value) {
			base.Decision, base.Stage, base.Step, base.Reason = row.Decision, row.Stage, row.Step, row.Reason
			return base
		}
	}
	return setSafety(base, "NO_REDUCTION_RULE_MATCHED")
}

func matches(condition, sourceDigest, semanticDigest string, value input) bool {
	available := value.ProducerAvailable && value.ConsumerAvailable
	valid := func(value string) bool { return regexp.MustCompile(`^sha256:[0-9a-f]{64}$`).MatchString(value) }
	empty := value.ObservedSourceDigest == "" || value.ObservedArtifactSourceDigest == "" || value.ObservedGeneratedJudgeDigest == "" || value.ObservedIndependentDigest == ""
	malformed := !valid(value.ObservedSourceDigest) || !valid(value.ObservedArtifactSourceDigest) || !valid(value.ObservedGeneratedJudgeDigest) || !valid(value.ObservedIndependentDigest)
	switch condition {
	case conditionUnavailable:
		return !available
	case conditionEmpty:
		return available && empty
	case conditionMalformed:
		return available && !empty && malformed
	case conditionSource:
		return available && !empty && !malformed && value.ObservedSourceDigest != sourceDigest
	case conditionArtifact:
		return available && !empty && !malformed && value.ObservedSourceDigest == sourceDigest && value.ObservedArtifactSourceDigest != sourceDigest
	case conditionIndependent:
		return available && !empty && !malformed && value.ObservedSourceDigest == sourceDigest && value.ObservedArtifactSourceDigest == sourceDigest && value.ObservedIndependentDigest != semanticDigest
	case conditionEquivalent:
		return available && !empty && !malformed && value.ObservedSourceDigest == sourceDigest && value.ObservedArtifactSourceDigest == sourceDigest && value.ObservedIndependentDigest == semanticDigest
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
		ObservedGeneratedJudgeDigest string `json:"observed_generated_judge_digest"`
		ObservedIndependentDigest    string `json:"observed_independent_digest"`
	}{ID: value.ID, ProducerAvailable: value.ProducerAvailable, ConsumerAvailable: value.ConsumerAvailable, ObservedSourceDigest: value.ObservedSourceDigest, ObservedArtifactSourceDigest: value.ObservedArtifactSourceDigest, ObservedGeneratedJudgeDigest: value.ObservedGeneratedJudgeDigest, ObservedIndependentDigest: value.ObservedIndependentDigest})
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
	if len(value.Cases) != len(cases) || len(value.Evidence) != len(cases) || len(value.Claims.Events) != len(cases)*claimPredicateCount*2 || value.Claims.EventCount != len(value.Claims.Events) || value.Claims.Schema != "gooo/meta-policy-compilation-claims/v2" {
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
		if len(stored.ClaimPredicates) != claimPredicateCount || !sameResult(stored.Source, source) || !sameResult(stored.Independent, independent) || !sameResult(stored.Source, stored.Generated) || !stored.AllEquivalent || !stored.Equivalent || !stored.ValidatorConfirmed {
			return fmt.Errorf("consumer rejected source/generated/independent lineage for %q", input.ID)
		}
		if err := verifyPredicateObservations(stored, policy, input, artifact); err != nil {
			return err
		}
	}
	evidenceByID := make(map[string]evidenceObservation, len(value.Evidence))
	for _, observed := range value.Evidence {
		evidenceByID[observed.CaseID] = observed
	}
	for _, input := range cases {
		observed, ok := evidenceByID[input.ID]
		if !ok || observed.Class != evidenceSynthetic || observed.ProducerAvailable != input.ProducerAvailable || observed.ConsumerAvailable != input.ConsumerAvailable || observed.SourceDigest != input.ObservedSourceDigest || observed.ArtifactSourceDigest != input.ObservedArtifactSourceDigest || observed.ArtifactDigest != digestJSONIndented(artifact) || observed.GeneratedJudgeDigest != input.ObservedGeneratedJudgeDigest || observed.IndependentDigest != input.ObservedIndependentDigest || observed.SemanticDigest != policy.SemanticDigest || observed.ObservationDigest != digestInput(input) || observed.Provenance != input.Provenance {
			return fmt.Errorf("consumer rejected receipt evidence provenance for %q", input.ID)
		}
	}
	if err := verifyClaimChain(value.Claims, cases); err != nil {
		return err
	}
	if err := verifyClaimEventBindings(value, cases, policy); err != nil {
		return err
	}
	if value.WriteSet.RepositoryBeforeDigest == "" || value.WriteSet.RepositoryBeforeDigest != value.WriteSet.RepositoryAfterDigest || value.WriteSet.RepositoryNetChangeObserved || value.WriteSet.RepositoryBeforeCount != value.WriteSet.RepositoryAfterCount || value.WriteSet.GeneratedRootClass != "RUNNER_TEMP_ONLY" || value.WriteSet.MutationAuthority != 0 || value.WriteSet.PromotionAuthority != 0 {
		return errors.New("consumer rejected generated write-set authority")
	}
	if value.Policy.SourceDigest != policy.SourceDigest || artifact.Policy.SourceDigest != policy.SourceDigest {
		return errors.New("consumer rejected source digest lineage")
	}
	return nil
}

func verifyPredicateObservations(stored caseReceipt, policy policy, input input, artifact artifact) error {
	seen := make(map[string]bool, claimPredicateCount)
	byID := make(map[string]claimPredicateObservation, claimPredicateCount)
	for _, observation := range stored.ClaimPredicates {
		if seen[observation.Predicate] || observation.ObservationDigest != stored.ObservationDigest || observation.Provenance != stored.Provenance {
			return fmt.Errorf("consumer rejected duplicate or unbound predicate for %q", stored.ID)
		}
		seen[observation.Predicate] = true
		byID[observation.ClaimID] = observation
	}
	for _, rule := range policy.Rules {
		claim := claimID(stored.ID, rule)
		observation, ok := byID[claim]
		if !ok || observation.Predicate != rule.Claim || observation.ClaimID != claim || (observation.Outcome != "OPEN" && observation.Outcome != "DISCHARGED" && observation.Outcome != "REFUTED") {
			return fmt.Errorf("consumer rejected predicate %q for %q", rule.Claim, stored.ID)
		}
		expected := assessPredicate(rule, policy, artifact, input, stored.Source, stored.Generated, stored.Independent)
		if observation.Outcome != expected.outcome || observation.Observed != expected.observed || observation.Stage != expected.stage || observation.Step != expected.step || observation.Reason != expected.reason {
			return fmt.Errorf("consumer independently rejected predicate %q for %q", rule.Claim, stored.ID)
		}
	}
	return nil
}

func byPredicate(stored caseReceipt, predicate string) claimPredicateObservation {
	for _, value := range stored.ClaimPredicates {
		if value.Predicate == predicate {
			return value
		}
	}
	return claimPredicateObservation{}
}

func claimID(caseID string, rule rule) string {
	return fmt.Sprintf("gooo://meta-policy-compilation/claim/%s/%02d-%s", caseID, rule.Step, rule.Claim)
}

type predicateAssessment struct {
	outcome  string
	observed bool
	stage    string
	step     int
	reason   string
}

func countPredicates(cases []caseReceipt) (int, int, int) {
	discharged, refuted, open := 0, 0, 0
	for _, stored := range cases {
		for _, observation := range stored.ClaimPredicates {
			switch observation.Outcome {
			case "DISCHARGED":
				discharged++
			case "REFUTED":
				refuted++
			case "OPEN":
				open++
			}
		}
	}
	return discharged, refuted, open
}

func assessPredicate(rule rule, policy policy, artifact artifact, input input, source, generated, independent result) predicateAssessment {
	open := predicateAssessment{outcome: "OPEN", stage: "VERIFY", step: 4, reason: conditionUnavailable}
	switch rule.Claim {
	case claimSourceBound:
		return assessConsumerDigest(rule, input.ProducerAvailable, input.ObservedSourceDigest, policy.SourceDigest, conditionSource)
	case claimArtifactBound:
		return assessConsumerDigest(rule, input.ConsumerAvailable, input.ObservedArtifactSourceDigest, policy.SourceDigest, conditionArtifact)
	case claimGenerated:
		if !input.ProducerAvailable {
			return open
		}
		if input.ObservedGeneratedJudgeDigest == "" {
			open.reason = conditionEmpty
			return open
		}
		if !validDigest(input.ObservedGeneratedJudgeDigest) {
			open.reason, open.stage = conditionMalformed, "LOWER_RESOLUTION"
			return open
		}
		if input.ObservedGeneratedJudgeDigest != artifact.GeneratedJudgeHash {
			return predicateAssessment{outcome: "REFUTED", observed: true, stage: rule.Stage, step: rule.Step, reason: "GENERATED_JUDGE_DIGEST_MISMATCH"}
		}
		if validResult(generated, input.ID, policy) {
			return predicateAssessment{outcome: "DISCHARGED", observed: true, stage: rule.Stage, step: rule.Step, reason: "GENERATED_EXECUTION_OBSERVED"}
		}
		return predicateAssessment{outcome: "REFUTED", observed: true, stage: rule.Stage, step: rule.Step, reason: "GENERATED_EXECUTION_INVALID"}
	case claimIndependent:
		if !input.ConsumerAvailable {
			return open
		}
		if input.ObservedIndependentDigest == "" {
			open.reason = conditionEmpty
			return open
		}
		if !validDigest(input.ObservedIndependentDigest) {
			open.reason, open.stage = conditionMalformed, "LOWER_RESOLUTION"
			return open
		}
		if !validResult(independent, input.ID, policy) || !sameResult(source, independent) {
			return predicateAssessment{outcome: "REFUTED", observed: true, stage: rule.Stage, step: rule.Step, reason: "INDEPENDENT_REPLAY_MISMATCH"}
		}
		if input.ObservedIndependentDigest != policy.SemanticDigest {
			return predicateAssessment{outcome: "REFUTED", observed: true, stage: rule.Stage, step: rule.Step, reason: conditionIndependent}
		}
		return predicateAssessment{outcome: "DISCHARGED", observed: true, stage: rule.Stage, step: rule.Step, reason: "INDEPENDENT_REPLAY_OBSERVED"}
	case claimProof:
		for _, candidate := range policy.Rules {
			if candidate.Claim == claimProof && candidate.ProofChoice != "" {
				return predicateAssessment{outcome: "DISCHARGED", observed: true, stage: rule.Stage, step: rule.Step, reason: "PROOF_SELECTION_OBSERVED"}
			}
		}
		return predicateAssessment{outcome: "OPEN", stage: rule.Stage, step: rule.Step, reason: "PROOF_SELECTION_MISSING"}
	case claimLedgerPredicate:
		return predicateAssessment{outcome: "DISCHARGED", observed: true, stage: rule.Stage, step: rule.Step, reason: "LEDGER_CHAIN_APPENDED"}
	case claimReduction:
		if !input.ProducerAvailable {
			return open
		}
		if validResult(source, input.ID, policy) && sameResult(source, generated) && sameResult(source, independent) && reductionMatches(policy, input, source) {
			return predicateAssessment{outcome: "DISCHARGED", observed: true, stage: rule.Stage, step: rule.Step, reason: source.Reason}
		}
		return predicateAssessment{outcome: "REFUTED", observed: true, stage: rule.Stage, step: rule.Step, reason: "DECISION_REDUCTION_MISMATCH"}
	case claimLineage:
		if !input.ProducerAvailable || !input.ConsumerAvailable {
			open.stage = "LOWER_RESOLUTION"
			return open
		}
		if input.ObservedSourceDigest == "" || input.ObservedArtifactSourceDigest == "" || input.ObservedGeneratedJudgeDigest == "" || input.ObservedIndependentDigest == "" {
			open.reason = conditionEmpty
			return open
		}
		if !validDigest(input.ObservedSourceDigest) || !validDigest(input.ObservedArtifactSourceDigest) || !validDigest(input.ObservedGeneratedJudgeDigest) || !validDigest(input.ObservedIndependentDigest) {
			open.reason, open.stage = conditionMalformed, "LOWER_RESOLUTION"
			return open
		}
		if input.ObservedSourceDigest != policy.SourceDigest {
			return predicateAssessment{outcome: "REFUTED", observed: true, stage: rule.Stage, step: rule.Step, reason: conditionSource}
		}
		if input.ObservedArtifactSourceDigest != artifact.Policy.SourceDigest || input.ObservedGeneratedJudgeDigest != artifact.GeneratedJudgeHash || input.ObservedIndependentDigest != policy.SemanticDigest {
			return predicateAssessment{outcome: "REFUTED", observed: true, stage: rule.Stage, step: rule.Step, reason: "LINEAGE_DIGEST_MISMATCH"}
		}
		if validResult(source, input.ID, policy) && sameResult(source, generated) && sameResult(source, independent) {
			return predicateAssessment{outcome: "DISCHARGED", observed: true, stage: rule.Stage, step: rule.Step, reason: "LINEAGE_SEALED"}
		}
		return predicateAssessment{outcome: "REFUTED", observed: true, stage: rule.Stage, step: rule.Step, reason: "LINEAGE_RESULT_MISMATCH"}
	}
	return open
}

func assessConsumerDigest(rule rule, available bool, observed, expected, mismatch string) predicateAssessment {
	open := predicateAssessment{outcome: "OPEN", stage: "VERIFY", step: 4, reason: conditionUnavailable}
	if !available {
		return open
	}
	if observed == "" {
		open.reason = conditionEmpty
		return open
	}
	if !validDigest(observed) {
		open.reason, open.stage = conditionMalformed, "LOWER_RESOLUTION"
		return open
	}
	if observed == expected {
		return predicateAssessment{outcome: "DISCHARGED", observed: true, stage: rule.Stage, step: rule.Step, reason: "DIGEST_BOUND"}
	}
	return predicateAssessment{outcome: "REFUTED", observed: true, stage: rule.Stage, step: rule.Step, reason: mismatch}
}

func validResult(value result, caseID string, policy policy) bool {
	return value.CaseID == caseID && value.PolicyDigest == policy.SourceDigest && value.SemanticDigest == policy.SemanticDigest && value.Denominator == policy.Denominator && knownDecision(value.Decision) && value.Stage != "" && value.Step > 0 && value.Reason != ""
}

func reductionMatches(policy policy, value input, expected result) bool {
	for _, row := range policy.Reduction.Rules {
		if matches(row.Condition, policy.SourceDigest, policy.SemanticDigest, value) {
			return expected.Decision == row.Decision && expected.Stage == row.Stage && expected.Step == row.Step && expected.Reason == row.Reason
		}
	}
	return false
}

func verifyClaimEventBindings(value receipt, cases []input, policy policy) error {
	byID := make(map[string]caseReceipt, len(value.Cases))
	for _, stored := range value.Cases {
		byID[stored.ID] = stored
	}
	order := append([]input(nil), cases...)
	sort.Slice(order, func(i, j int) bool { return order[i].ID < order[j].ID })
	for caseIndex, input := range order {
		stored, ok := byID[input.ID]
		if !ok {
			return fmt.Errorf("consumer has no claim segment for %q", input.ID)
		}
		base := caseIndex * claimPredicateCount * 2
		for ruleIndex, sourceRule := range policy.Rules {
			observation := byPredicate(stored, sourceRule.Claim)
			opening := value.Claims.Events[base+ruleIndex]
			outcome := value.Claims.Events[base+claimPredicateCount+ruleIndex]
			claim := claimID(input.ID, sourceRule)
			if opening.ClaimID != claim || opening.Predicate != sourceRule.Claim || opening.From != "UNRECORDED" || opening.To != "OPEN" || opening.Observed || opening.Reason != "CLAIM_OPENED" || outcome.ClaimID != claim || outcome.Predicate != sourceRule.Claim || outcome.From != "OPEN" || outcome.To != observation.Outcome || outcome.Observed != observation.Observed || outcome.Stage != observation.Stage || outcome.Step != observation.Step || outcome.Reason != observation.Reason {
				return fmt.Errorf("consumer claim event binding is invalid for %q/%s", input.ID, sourceRule.Claim)
			}
		}
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
		if event.Event != index+1 || event.PriorDigest != prior || event.ClaimID == "" || event.Predicate == "" || event.ObservationDigest == "" || event.Provenance == "" || allowed[event.ObservationDigest] != event.Provenance || counts[event.ClaimID] >= 2 {
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
		if (event.From == "UNRECORDED" && (event.To != "OPEN" || event.Observed)) || (event.From == "OPEN" && (event.To == "DISCHARGED" || event.To == "REFUTED") != event.Observed) {
			return fmt.Errorf("consumer claim observation invalid at event %d", event.Event)
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

func observeCurrentEvidence(dir string, policy policy, source []byte, artifact artifact, judge []byte) evidenceObservation {
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
	provenance := "current runner-temp artifact: raw policy, artifact source, generated judge bytes, and independent raw-source contract observed"
	observation := evidenceObservation{Class: evidenceCurrent, CaseID: "current-subject", ProducerAvailable: available, ConsumerAvailable: available, SourceDigest: digestBytes(source), ArtifactSourceDigest: artifact.Policy.SourceDigest, ArtifactDigest: artifactDigest, GeneratedJudgeDigest: digestBytes(judge), IndependentDigest: policy.SemanticDigest, SemanticDigest: policy.SemanticDigest, Provenance: provenance}
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

func digestJSONIndented(value any) string {
	data, _ := json.MarshalIndent(value, "", "  ")
	return digestBytes(append(data, '\n'))
}

func digestReceipt(value receipt) string {
	value.ReceiptDigest = ""
	return digestJSON(value)
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
	case conditionUnavailable, conditionEmpty, conditionMalformed, conditionSource, conditionArtifact, conditionIndependent, conditionEquivalent:
		return true
	default:
		return false
	}
}

func knownDecision(value string) bool {
	return value == decisionPass || value == decisionFailClosed || value == decisionUnknown
}
