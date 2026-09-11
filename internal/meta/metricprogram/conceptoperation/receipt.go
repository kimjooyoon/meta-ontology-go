package conceptoperation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
	interventionverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention/verify"
)

const (
	Schema   = "gooo/metric-meta-concept-operation-binding/v1"
	MetricID = "gooo.metric.meta.concept-operation-binding-bps.v1"

	ExecutionPolicy = "READ_ONLY_CONCEPT_OPERATION_RECEIPT"
	CohortRule      = "UPSTREAM_INTERVENTION_OPERATIONS_PLUS_REGISTERED_TERMINALS"
	SourceAuthority = "HANDWRITTEN_GO_REGISTRY_TO_CANONICAL_SOURCE_TO_PARSED_LOWERED_IR"
)

type OperationBinding struct {
	Operation         string `json:"operation"`
	CarrierOperation  string `json:"carrier_operation"`
	IndicatorID       string `json:"indicator_id"`
	ConceptID         string `json:"concept_id"`
	RegisteredActivity string `json:"registered_activity"`
	RegisteredProofChoice string `json:"registered_proof_choice"`
	Activity          string `json:"activity"`
	ProofChoice       string `json:"proof_choice"`
	Expected          string `json:"expected"`
	Actual            string `json:"actual"`
	Status            string `json:"status"`
	EvidenceDigest    string `json:"evidence_digest"`
	OperationDigest   string `json:"operation_digest"`
}

type UnknownCausal struct {
	UnknownClass  string   `json:"unknown_class"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

// SourceInputs are the immutable producer inputs carried with the receipt.
// The consumer replays these inputs before accepting the receipt digest.
type SourceInputs struct {
	Strategy                   []byte
	StrategyVerification      []byte
	SourceMetrics              []byte
	Intervention               []byte
	InterventionVerification  []byte
	Program                    []byte
	ProgramSource              []byte
	ProgramVerification       []byte
}

type Receipt struct {
	Schema                    string             `json:"schema"`
	Repository                string             `json:"repository"`
	SubjectSHA                string             `json:"subject_sha"`
	ExecutionPolicy           string             `json:"execution_policy"`
	MetricID                  string             `json:"metric_id"`
	CohortRule                string             `json:"cohort_rule"`
	SourceAuthority           string             `json:"source_authority"`
	StrategyDigest            string             `json:"strategy_digest"`
	StrategyVerificationDigest string            `json:"strategy_verification_digest"`
	InterventionDigest        string             `json:"intervention_digest"`
	ProgramDigest             string             `json:"program_digest"`
	ProgramVerificationDigest string             `json:"program_verification_digest"`
	ProgramSourcePath         string             `json:"program_source_path"`
	ProgramSourceDigest       string             `json:"program_source_digest"`
	ProgramSemanticDigest     string             `json:"program_semantic_digest"`
	ProgramRegistryDigest     string             `json:"program_registry_digest"`
	Expected                  []OperationBinding `json:"expected"`
	ExpectedCount             int                `json:"expected_count"`
	ObservedCount             int                `json:"observed_count"`
	BoundCount                int                `json:"bound_count"`
	UnknownCount              int                `json:"unknown_count"`
	CoverageBPS               int                `json:"coverage_bps"`
	Status                    string             `json:"status"`
	FailureClass              string             `json:"failure_class,omitempty"`
	Unknown                   *UnknownCausal     `json:"unknown,omitempty"`
	Producer                  string             `json:"producer"`
	Consumer                  string             `json:"consumer"`
	MetaOperation             string             `json:"meta_operation"`
	RepositoryWorkspaceWrites bool               `json:"repository_workspace_writes"`
	PromotionAuthorized       bool               `json:"promotion_authorized"`
	Digest                    string             `json:"digest"`
}

func Build(strategyPayload, strategyVerificationPayload, interventionPayload,
	programPayload, sourcePayload, programVerificationPayload []byte) (Receipt, error) {
	var plan metricstrategy.Plan
	if err := decodeExact(strategyPayload, &plan); err != nil {
		return Receipt{}, fmt.Errorf("decode strategy: %w", err)
	}
	var strategyReceipt strategyverify.Receipt
	if err := decodeExact(strategyVerificationPayload, &strategyReceipt); err != nil {
		return Receipt{}, fmt.Errorf("decode strategy verification: %w", err)
	}
	var ledger metric.Ledger
	if err := decodeExact(interventionPayload, &ledger); err != nil {
		return Receipt{}, fmt.Errorf("decode intervention: %w", err)
	}
	var program metricprogram.Program
	if err := decodeExact(programPayload, &program); err != nil {
		return Receipt{}, fmt.Errorf("decode program: %w", err)
	}
	var programReceipt programverify.Report
	if err := decodeExact(programVerificationPayload, &programReceipt); err != nil {
		return Receipt{}, fmt.Errorf("decode program verification: %w", err)
	}

	if plan.Repository == "" || plan.SubjectSHA == "" || plan.Input.InterventionDigest != ledger.Digest ||
		ledger.Schema != metric.LedgerSchema || ledger.Repository != plan.Repository || ledger.SubjectSHA != plan.SubjectSHA ||
		ledger.RepositoryWorkspaceWrites || ledger.PromotionAuthorized ||
		strategyReceipt.PlanDigest != plan.Digest || strategyReceipt.InterventionDigest != ledger.Digest ||
		strategyReceipt.Status != "VERIFIED" {
		return Receipt{}, fmt.Errorf("strategy and intervention identity is not bound")
	}
	independent, err := programverify.Verify(strategyPayload, strategyVerificationPayload, programPayload, sourcePayload)
	if err != nil {
		return Receipt{}, fmt.Errorf("independent program verification: %w", err)
	}
	if !artifact.Equal(independent, programReceipt) {
		return Receipt{}, fmt.Errorf("program verification receipt is not independently bound")
	}
	if program.Repository != plan.Repository || program.SubjectSHA != plan.SubjectSHA ||
		program.StrategyDigest != plan.Digest || program.StrategyVerificationDigest != strategyReceipt.Digest ||
		programReceipt.ProgramDigest != program.Digest || programReceipt.SourceDigest != program.SourceDigest ||
		programReceipt.SemanticDigest != program.SemanticDigest || programReceipt.RegistryDigest != program.RegistryDigest {
		return Receipt{}, fmt.Errorf("program identity is not bound")
	}

	cohort, err := reconstructCohort(ledger.Indicators, program.Operations)
	if err != nil {
		return Receipt{}, fmt.Errorf("reconstruct operation cohort: %w", err)
	}
	operations := make(map[string]metricprogram.OperationSpec, len(program.Operations))
	for _, operation := range program.Operations {
		operations[operation.ID] = operation
	}
	bindings := make(map[string][]metricstrategy.Binding)
	observed := 0
	for _, binding := range plan.Bindings {
		if metricstrategy.IsConceptOperationBinding(binding.IndicatorID) {
			observed++
			bindings[binding.IndicatorID] = append(bindings[binding.IndicatorID], binding)
		}
	}

	expected := make([]OperationBinding, 0, len(cohort))
	bound := 0
	knownIndicators := make(map[string]bool, len(cohort))
	for _, spec := range cohort {
		conceptID, mapped := metricstrategy.ConceptIDForOperation(spec.Subject)
		indicatorID := metricstrategy.ConceptOperationIndicatorID(spec.Subject)
		expectedValue := conceptID
		if !mapped {
			expectedValue = "REGISTERED_CONCEPT"
		}
		operation, operationOK := operations[spec.Carrier]
		registered, registeredOK := operations[spec.Subject]
		entry := OperationBinding{
			Operation: spec.Subject, CarrierOperation: spec.Carrier, IndicatorID: indicatorID,
			ConceptID: conceptID, Expected: expectedValue, Actual: "UNKNOWN", Status: "UNSATISFIED",
		}
		if registeredOK {
			entry.RegisteredActivity, entry.RegisteredProofChoice = registered.Activity, registered.ProofChoice
		}
		if operationOK {
			entry.Activity, entry.ProofChoice = operation.Activity, operation.ProofChoice
		}
		knownIndicators[indicatorID] = true
		candidates := bindings[indicatorID]
		if operationOK && mapped && len(candidates) == 1 {
			candidate := candidates[0]
			entry.Actual, entry.Status, entry.EvidenceDigest = candidate.Actual, candidate.Status, candidate.EvidenceDigest
			if candidate.MetaOperation == spec.Carrier && candidate.Family == operation.ProofChoice &&
				candidate.Expected == expectedValue && candidate.Actual == expectedValue &&
				candidate.Status == "SATISFIED" && candidate.EvidenceDigest != "" {
				for _, resolved := range program.Bindings {
					if resolved.IndicatorID == candidate.IndicatorID && resolved.OperationID == spec.Carrier &&
						resolved.Activity == operation.Activity && resolved.ProofChoice == candidate.Family {
						entry.OperationDigest = resolved.OperationDigest
						bound++
						break
					}
				}
			}
		}
		expected = append(expected, entry)
	}
	sort.Slice(expected, func(i, j int) bool { return expected[i].Operation < expected[j].Operation })
	unknown := 0
	for indicatorID, candidates := range bindings {
		if !knownIndicators[indicatorID] {
			unknown += len(candidates)
			continue
		}
		if len(candidates) > 1 {
			unknown += len(candidates) - 1
		}
	}
	expectedCount := len(expected)
	coverage := 0
	if expectedCount > 0 {
		coverage = bound * 10000 / expectedCount
	}
	status := "FAIL_CLOSED"
	if expectedCount > 0 && bound == expectedCount && unknown == 0 {
		status = "VERIFIED"
	}
	failureClass := ""
	var unknownCausal *UnknownCausal
	if status != "VERIFIED" {
		if unknown > 0 || observed < expectedCount {
			failureClass = "INCOMPLETE_EVIDENCE"
			unknownCausal = &UnknownCausal{UnknownClass: "INCOMPLETE_EVIDENCE", Stage: "CONCEPT_OPERATION_BINDING", Step: "RECONSTRUCT_EXACT_COHORT", Reason: "CONCEPT_OPERATION_BINDING_EVIDENCE_INCOMPLETE", NextOperation: "CAPTURE_EXACT_PRODUCER_INPUTS", BlockedBy: []string{"concept_operation_binding_receipt"}}
		} else {
			failureClass = "KNOWN_CONTRADICTION"
		}
	}
	receipt := Receipt{
		Schema: Schema, Repository: plan.Repository, SubjectSHA: plan.SubjectSHA, ExecutionPolicy: ExecutionPolicy,
		MetricID: MetricID, CohortRule: CohortRule, StrategyDigest: plan.Digest,
		SourceAuthority: SourceAuthority,
		StrategyVerificationDigest: strategyReceipt.Digest, InterventionDigest: ledger.Digest,
		ProgramDigest: program.Digest, ProgramVerificationDigest: programReceipt.Digest,
		ProgramSourcePath: program.SourcePath,
		ProgramSourceDigest: program.SourceDigest, ProgramSemanticDigest: program.SemanticDigest,
		ProgramRegistryDigest: program.RegistryDigest, Expected: expected, ExpectedCount: expectedCount,
		ObservedCount: observed, BoundCount: bound, UnknownCount: unknown, CoverageBPS: coverage,
		Status: status, FailureClass: failureClass, Unknown: unknownCausal, Producer: "metricprogram/conceptoperation.Build", Consumer: "language-readiness",
		MetaOperation: "bind-concept-operation-metric", RepositoryWorkspaceWrites: false, PromotionAuthorized: false,
	}
	return seal(receipt)
}

// VerifySource independently authenticates the producer inputs and then
// rebuilds the complete receipt. A receipt's self-hash is not provenance.
func VerifySource(receipt Receipt, inputs SourceInputs, repository fs.FS,
	expectedRepository, expectedSubjectSHA string) error {
	if err := VerifyReceipt(receipt, expectedRepository, expectedSubjectSHA); err != nil {
		return fmt.Errorf("verify concept-operation receipt envelope: %w", err)
	}
	if repository == nil {
		return fmt.Errorf("concept-operation repository source is missing")
	}
	for name, payload := range map[string][]byte{
		"strategy": inputs.Strategy, "strategy verification": inputs.StrategyVerification,
		"source metrics": inputs.SourceMetrics, "intervention": inputs.Intervention,
		"intervention verification": inputs.InterventionVerification, "program": inputs.Program,
		"program source": inputs.ProgramSource, "program verification": inputs.ProgramVerification,
	} {
		if len(payload) == 0 {
			return fmt.Errorf("concept-operation %s source is missing", name)
		}
	}
	var plan metricstrategy.Plan
	if err := decodeExact(inputs.Strategy, &plan); err != nil {
		return fmt.Errorf("decode source strategy: %w", err)
	}
	var strategyReceipt strategyverify.Receipt
	if err := decodeExact(inputs.StrategyVerification, &strategyReceipt); err != nil {
		return fmt.Errorf("decode source strategy verification: %w", err)
	}
	var ledger metric.Ledger
	if err := decodeExact(inputs.Intervention, &ledger); err != nil {
		return fmt.Errorf("decode source intervention: %w", err)
	}
	var interventionReceipt interventionverify.Receipt
	if err := decodeExact(inputs.InterventionVerification, &interventionReceipt); err != nil {
		return fmt.Errorf("decode source intervention verification: %w", err)
	}
	var program metricprogram.Program
	if err := decodeExact(inputs.Program, &program); err != nil {
		return fmt.Errorf("decode source program: %w", err)
	}
	var programReceipt programverify.Report
	if err := decodeExact(inputs.ProgramVerification, &programReceipt); err != nil {
		return fmt.Errorf("decode source program verification: %w", err)
	}
	if plan.Repository != expectedRepository || plan.SubjectSHA != expectedSubjectSHA ||
		ledger.Repository != expectedRepository || ledger.SubjectSHA != expectedSubjectSHA ||
		program.Repository != expectedRepository || program.SubjectSHA != expectedSubjectSHA {
		return fmt.Errorf("concept-operation source subject is not exact")
	}
	temporary, err := os.MkdirTemp("", "gooo-concept-operation-")
	if err != nil {
		return fmt.Errorf("create temporary source directory: %w", err)
	}
	defer os.RemoveAll(temporary)
	paths := map[string][]byte{
		"source-metrics.json":          inputs.SourceMetrics,
		"intervention.json":            inputs.Intervention,
		"intervention-verification.json": inputs.InterventionVerification,
	}
	for name, payload := range paths {
		if err := os.WriteFile(filepath.Join(temporary, name), payload, 0o600); err != nil {
			return fmt.Errorf("stage %s: %w", name, err)
		}
	}
	independentIntervention, err := interventionverify.Replay(filepath.Join(temporary, "source-metrics.json"), ledger)
	if err != nil {
		return fmt.Errorf("independent intervention replay: %w", err)
	}
	if !artifact.Equal(independentIntervention, interventionReceipt) {
		return fmt.Errorf("intervention verification receipt is not independently bound")
	}
	independentStrategy, err := strategyverify.Replay(repository, filepath.Join(temporary, "source-metrics.json"), filepath.Join(temporary, "intervention.json"), filepath.Join(temporary, "intervention-verification.json"), plan)
	if err != nil {
		return fmt.Errorf("independent strategy replay: %w", err)
	}
	if !artifact.Equal(independentStrategy, strategyReceipt) {
		return fmt.Errorf("strategy verification receipt is not independently bound")
	}
	independentProgram, err := programverify.Verify(inputs.Strategy, inputs.StrategyVerification, inputs.Program, inputs.ProgramSource)
	if err != nil {
		return fmt.Errorf("independent program replay: %w", err)
	}
	if !artifact.Equal(independentProgram, programReceipt) {
		return fmt.Errorf("program verification receipt is not independently bound")
	}
	rebuilt, err := Build(inputs.Strategy, inputs.StrategyVerification, inputs.Intervention,
		inputs.Program, inputs.ProgramSource, inputs.ProgramVerification)
	if err != nil {
		return fmt.Errorf("rebuild concept-operation receipt: %w", err)
	}
	if !artifact.Equal(rebuilt, receipt) {
		return fmt.Errorf("concept-operation receipt does not match independently rebuilt evidence")
	}
	return nil
}

type cohortOperation struct {
	Subject  string
	Carrier  string
	Family   string
	Trilemma string
}

func reconstructCohort(indicators []metric.Indicator, operations []metricprogram.OperationSpec) ([]cohortOperation, error) {
	available := make(map[string]metricprogram.OperationSpec, len(operations))
	for _, operation := range operations {
		available[operation.ID] = operation
	}
	bySubject := make(map[string]cohortOperation)
	for _, indicator := range indicators {
		if indicator.MetaOperation == "" {
			return nil, fmt.Errorf("intervention indicator has no meta operation")
		}
		operation, ok := available[indicator.MetaOperation]
		if !ok {
			return nil, fmt.Errorf("intervention operation %q is not in the generated registry", indicator.MetaOperation)
		}
		current := cohortOperation{Subject: indicator.MetaOperation, Carrier: indicator.MetaOperation,
			Family: strings.ToUpper(indicator.Family), Trilemma: indicator.Trilemma}
		if previous, exists := bySubject[current.Subject]; exists && previous != current {
			return nil, fmt.Errorf("intervention operation %q spans proof families", current.Subject)
		}
		if operation.ProofChoice != current.Family {
			return nil, fmt.Errorf("intervention operation %q proof family is not registry-bound", current.Subject)
		}
		bySubject[current.Subject] = current
	}
	bySubject["terminate-at-fixed-point"] = cohortOperation{Subject: "terminate-at-fixed-point", Carrier: "replay-counterfactual", Family: "REGRESSION", Trilemma: "REGRESS"}
	bySubject["preserve-non-promoting-terminal"] = cohortOperation{Subject: "preserve-non-promoting-terminal", Carrier: "preserve-non-promoting-terminal", Family: "REGRESSION", Trilemma: "REGRESS"}
	for _, terminal := range []string{"terminate-at-fixed-point", "preserve-non-promoting-terminal"} {
		if _, ok := available[bySubject[terminal].Carrier]; !ok {
			return nil, fmt.Errorf("registered terminal carrier %q is missing", bySubject[terminal].Carrier)
		}
	}
	keys := make([]string, 0, len(bySubject))
	for subject := range bySubject {
		keys = append(keys, subject)
	}
	sort.Strings(keys)
	result := make([]cohortOperation, 0, len(keys))
	for _, subject := range keys {
		result = append(result, bySubject[subject])
	}
	return result, nil
}

func Verify(payload []byte, expectedRepository, expectedSubjectSHA string) (Receipt, error) {
	var receipt Receipt
	if err := decodeExact(payload, &receipt); err != nil {
		return Receipt{}, fmt.Errorf("decode concept-operation receipt: %w", err)
	}
	if err := VerifyReceipt(receipt, expectedRepository, expectedSubjectSHA); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func VerifyReceipt(receipt Receipt, expectedRepository, expectedSubjectSHA string) error {
	if receipt.Schema != Schema || receipt.ExecutionPolicy != ExecutionPolicy || receipt.MetricID != MetricID ||
		receipt.CohortRule != CohortRule || receipt.SourceAuthority != SourceAuthority || receipt.ProgramSourcePath != metricprogram.ProgramSourceFilename ||
		receipt.RepositoryWorkspaceWrites || receipt.PromotionAuthorized ||
		!artifact.ValidSubject(receipt.Repository, receipt.SubjectSHA) ||
		(expectedRepository != "" && receipt.Repository != expectedRepository) ||
		(expectedSubjectSHA != "" && receipt.SubjectSHA != expectedSubjectSHA) {
		return fmt.Errorf("concept-operation receipt identity is invalid")
	}
	if receipt.Producer != "metricprogram/conceptoperation.Build" || receipt.Consumer != "language-readiness" || receipt.MetaOperation != "bind-concept-operation-metric" {
		return fmt.Errorf("concept-operation receipt producer or consumer is invalid")
	}
	for _, digest := range []string{receipt.StrategyDigest, receipt.StrategyVerificationDigest,
		receipt.InterventionDigest, receipt.ProgramDigest, receipt.ProgramVerificationDigest,
		receipt.ProgramSourceDigest, receipt.ProgramSemanticDigest, receipt.ProgramRegistryDigest} {
		if !validDigest(digest) {
			return fmt.Errorf("concept-operation receipt digest field is invalid")
		}
	}
	if receipt.ExpectedCount != len(receipt.Expected) || receipt.ExpectedCount == 0 ||
		receipt.ObservedCount < 0 || receipt.BoundCount < 0 || receipt.UnknownCount < 0 ||
		receipt.BoundCount > receipt.ExpectedCount || receipt.CoverageBPS != receipt.BoundCount*10000/receipt.ExpectedCount {
		return fmt.Errorf("concept-operation receipt coverage is invalid")
	}
	seen := make(map[string]bool, len(receipt.Expected))
	for _, binding := range receipt.Expected {
		if binding.Operation == "" || binding.CarrierOperation == "" || binding.IndicatorID == "" ||
			binding.IndicatorID != metricstrategy.ConceptOperationIndicatorID(binding.Operation) ||
			seen[binding.Operation] || binding.Status == "" || binding.Expected == "" {
			return fmt.Errorf("concept-operation receipt cohort is invalid")
		}
		seen[binding.Operation] = true
	}
	if receipt.Status != "VERIFIED" && receipt.Status != "FAIL_CLOSED" {
		return fmt.Errorf("concept-operation receipt status is invalid")
	}
	if receipt.Status == "VERIFIED" && (receipt.FailureClass != "" || receipt.Unknown != nil) {
		return fmt.Errorf("verified concept-operation receipt has failure state")
	}
	if receipt.Status == "FAIL_CLOSED" {
		switch receipt.FailureClass {
		case "INCOMPLETE_EVIDENCE":
			if receipt.Unknown == nil || receipt.Unknown.UnknownClass != "INCOMPLETE_EVIDENCE" || receipt.Unknown.Stage != "CONCEPT_OPERATION_BINDING" || receipt.Unknown.Step != "RECONSTRUCT_EXACT_COHORT" || receipt.Unknown.Reason != "CONCEPT_OPERATION_BINDING_EVIDENCE_INCOMPLETE" || receipt.Unknown.NextOperation != "CAPTURE_EXACT_PRODUCER_INPUTS" || len(receipt.Unknown.BlockedBy) != 1 || receipt.Unknown.BlockedBy[0] != "concept_operation_binding_receipt" || (receipt.UnknownCount == 0 && receipt.ObservedCount >= receipt.ExpectedCount) {
				return fmt.Errorf("incomplete concept-operation receipt lacks causal UNKNOWN evidence")
			}
		case "KNOWN_CONTRADICTION":
			if receipt.Unknown != nil || receipt.UnknownCount != 0 || receipt.BoundCount >= receipt.ExpectedCount {
				return fmt.Errorf("known concept-operation contradiction has UNKNOWN evidence")
			}
		default:
			return fmt.Errorf("concept-operation receipt failure class is invalid")
		}
	}
	if receipt.Status == "VERIFIED" && (receipt.BoundCount != receipt.ExpectedCount || receipt.UnknownCount != 0 || receipt.CoverageBPS != 10000) {
		return fmt.Errorf("verified concept-operation receipt is incomplete")
	}
	if receipt.Status == "VERIFIED" {
		for _, binding := range receipt.Expected {
			if binding.RegisteredActivity == "" || binding.RegisteredProofChoice == "" || binding.Activity == "" || binding.ProofChoice == "" || binding.Expected != binding.Actual ||
				binding.Status != "SATISFIED" || !validDigest(binding.EvidenceDigest) || !validDigest(binding.OperationDigest) {
				return fmt.Errorf("verified concept-operation receipt binding is incomplete")
			}
		}
	}
	if receipt.Status == "FAIL_CLOSED" && receipt.CoverageBPS == 10000 && receipt.UnknownCount == 0 {
		return fmt.Errorf("fail-closed concept-operation receipt claims complete coverage")
	}
	digest, err := digestReceipt(receipt)
	if err != nil {
		return err
	}
	if receipt.Digest != digest {
		return fmt.Errorf("concept-operation receipt digest mismatch")
	}
	return nil
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	for _, character := range value[len("sha256:"):] {
		if !(character >= '0' && character <= '9') && !(character >= 'a' && character <= 'f') {
			return false
		}
	}
	return true
}

func seal(receipt Receipt) (Receipt, error) {
	digest, err := digestReceipt(receipt)
	if err != nil {
		return Receipt{}, err
	}
	receipt.Digest = digest
	return receipt, nil
}

func digestReceipt(receipt Receipt) (string, error) {
	receipt.Digest = ""
	return artifact.Digest(receipt)
}

func decodeExact[T any](payload []byte, target *T) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("document has trailing JSON")
	}
	return nil
}
