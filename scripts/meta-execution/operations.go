package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/operationconformance"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type processResult struct {
	Observation generation.ProcessObservation
	Stdout      []byte
	Stderr      []byte
}

type operationMaterialization struct {
	OperationID    string
	InstanceDigest string
	ContractDigest string
	Executor       generation.ProcessObservation
	Evaluator      generation.ProcessObservation
	Verifier       generation.ProcessObservation
	Indicators     []generation.IndicatorReceipt
	Canonical      []byte
}

type replayProcessObservation struct {
	Command      []string `json:"command"`
	ExitCode     int    `json:"exit_code"`
	StdoutBytes  int    `json:"stdout_bytes"`
	StdoutDigest string `json:"stdout_digest"`
	StderrBytes  int    `json:"stderr_bytes"`
	StderrDigest string `json:"stderr_digest"`
}

type operationReplayEvidence struct {
	Executor  replayProcessObservation `json:"executor"`
	Evaluator replayProcessObservation `json:"evaluator"`
	Verifier  replayProcessObservation `json:"verifier"`
}

func operationReplayEvidenceFrom(executor, evaluator, verifier generation.ProcessObservation) operationReplayEvidence {
	return operationReplayEvidence{
		Executor: replayProcess(executor), Evaluator: replayProcess(evaluator),
		Verifier: replayProcess(verifier),
	}
}

func replayProcess(observation generation.ProcessObservation) replayProcessObservation {
	return replayProcessObservation{
		Command:      append([]string{}, observation.Command...),
		ExitCode:     observation.ExitCode, StdoutBytes: observation.StdoutBytes,
		StdoutDigest: observation.StdoutDigest, StderrBytes: observation.StderrBytes,
		StderrDigest: observation.StderrDigest,
	}
}

type extractorReport struct {
	Schema                string                        `json:"schema"`
	SourceSHA             string                        `json:"source_sha"`
	StagedSubjects        int                           `json:"staged_subjects"`
	Subjects              []extractorSubject            `json:"subjects"`
	Unhandled             []string                      `json:"unhandled"`
	Failures              []extractorFailureRecord      `json:"failures,omitempty"`
	Indicators            []json.RawMessage             `json:"indicators"`
	NamespaceReplacements []namespaceReplacementReceipt `json:"namespace_replacements,omitempty"`
	BackupCleanup         backupCleanupObservation      `json:"backup_cleanup"`
}

type extractorSubject struct {
	Logical      string   `json:"logical"`
	State        string   `json:"state"`
	Before       int      `json:"before_lines"`
	After        int      `json:"after_lines"`
	Files        []string `json:"changed_files"`
	CreatedFiles []string `json:"created_files,omitempty"`
	Consumer     string   `json:"consumer"`
	Operation    string   `json:"meta_operation"`
	Operations   []string `json:"meta_operations,omitempty"`
	Proof        string   `json:"proof_choice"`
}

type extractorPlan struct {
	Schema    string                 `json:"schema"`
	SourceSHA string                 `json:"source_sha"`
	Subjects  []extractorPlanSubject `json:"subjects"`
}

type extractorPlanSubject struct {
	Logical string `json:"logical"`
	Lines   int    `json:"lines"`
	Reason  string `json:"reason"`
}

type extractorDensity struct {
	Schema    string                    `json:"schema"`
	SourceSHA string                    `json:"source_sha"`
	Subjects  []extractorDensitySubject `json:"subjects"`
}

type extractorDensitySubject struct {
	Logical string `json:"logical"`
	Status  string `json:"status"`
}

type extractorFailureRecord struct {
	Logical       string   `json:"logical"`
	BlockerID     string   `json:"blocker_id,omitempty"`
	Decision      string   `json:"decision"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class,omitempty"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
	Diagnostics   []string `json:"diagnostics,omitempty"`
}

type extractorIndicatorRecord struct {
	ID        string `json:"id"`
	Value     int    `json:"value"`
	Limit     int    `json:"limit"`
	Blocking  bool   `json:"blocking"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Proof     string `json:"proof_choice"`
}

const extractFunctionOperationID = "gooo/meta/generation/ExtractFunctionSuffix"
const functionExtractionReportSchema = "gooo.function-extraction.v2"

func executeSelectedOperations(plan generation.Plan, manifest generation.ExecutionManifest, workspace string) (generation.OperationObservationBundle, error) {
	bundle := generation.OperationObservationBundle{
		Schema:         generation.OperationObservationBundleSchema,
		BaseSHA:        plan.BaseSHA,
		HeadSHA:        plan.HeadSHA,
		PlanDigest:     plan.PlanDigest,
		ManifestDigest: manifest.ManifestDigest,
		Receipts:       []generation.OperationReceipt{},
		Failures:       []generation.ObservationFailure{},
	}
	if plan.Decision != generation.DecisionPlan {
		return generation.SealObservationBundle(bundle), nil
	}
	gitDir, err := gitDirectory(workspace)
	if err != nil {
		for _, action := range generationActions(plan) {
			bundle.Failures = append(bundle.Failures, observationFailure(action, "prepare-workspace", "resolve-git-directory", "GIT_CONTEXT_UNAVAILABLE", "DIRECT_MISSING", "restore-git-context", []string{}, generation.ProcessObservation{}))
		}
		bundle.ObservationTotal = len(plan.Selected)
		return generation.SealObservationBundle(bundle), nil
	}
	metricsPath, err := sourceMetricsPath()
	if err != nil {
		for _, action := range generationActions(plan) {
			bundle.Failures = append(bundle.Failures, observationFailure(action, "observe-metrics", "resolve-source-metrics", "SOURCE_METRICS_UNAVAILABLE", "DIRECT_MISSING", "restore-source-metrics", []string{}, generation.ProcessObservation{}))
		}
		bundle.ObservationTotal = len(plan.Selected)
		return generation.SealObservationBundle(bundle), nil
	}
	for _, action := range generationActions(plan) {
		materialized, runErr := executeAction(workspace, gitDir, metricsPath, plan, action)
		if runErr != nil {
			failure := observationFailure(action, runErr.stage, runErr.step, runErr.reason, runErr.class, runErr.next, runErr.blockedBy, materialized.Executor)
			failure.FailureEvidence = append([]generation.ObservationFailureEvidence{}, runErr.evidence...)
			failure.Counterexample = runErr.counterexample
			failure.DerivedRelations = append([]generation.CounterexampleRelation{}, runErr.derivedRelations...)
			bundle.Failures = append(bundle.Failures, failure)
			continue
		}
		receipt := generation.SealReceipt(plan, action, materialized.Indicators)
		receipt = generation.AttachInstanceEvidence(receipt, generation.OperationInstanceEvidence{
			Schema:                 generation.OperationInstanceEvidenceSchema,
			ActionIndicatorID:      action.IndicatorID,
			Subject:                action.Subject,
			HeadSHA:                plan.HeadSHA,
			OperationID:            materialized.OperationID,
			ContractEvidenceDigest: materialized.ContractDigest,
			InstanceEvidenceDigest: materialized.InstanceDigest,
			ExecutorObservation:    materialized.Executor,
			EvaluatorObservation:   materialized.Evaluator,
			ReplayComparisons:      1,
			ReplayMatch:            true,
			VerifierObservation:    &materialized.Verifier,
		})
		bundle.Receipts = append(bundle.Receipts, receipt)
	}
	bundle.ObservationTotal = len(bundle.Receipts) + len(bundle.Failures)
	bundle.ReplayComparisons = len(bundle.Receipts)
	return generation.SealObservationBundle(bundle), nil
}

func generationActions(plan generation.Plan) []generation.Action {
	actions := append([]generation.Action{}, plan.Selected...)
	sort.Slice(actions, func(left, right int) bool { return actions[left].IndicatorID < actions[right].IndicatorID })
	return actions
}

type operationError struct {
	stage, step, reason, class, next string
	blockedBy                        []string
	evidence                         []generation.ObservationFailureEvidence
	counterexample                   string
	derivedRelations                 []generation.CounterexampleRelation
	canonical                        []byte
}

func newOperationError(stage, step, reason, class, next string) *operationError {
	return &operationError{stage: stage, step: step, reason: reason, class: class, next: next}
}

func (err *operationError) Error() string {
	return strings.Join([]string{err.stage, err.step, err.reason, err.class, err.next}, "/")
}

func sameOperationError(left, right *operationError) bool {
	return left != nil && right != nil && left.stage == right.stage && left.step == right.step &&
		left.reason == right.reason && left.class == right.class && left.next == right.next &&
		strings.Join(left.blockedBy, "\x00") == strings.Join(right.blockedBy, "\x00") &&
		strings.Join(left.evidenceCounterexamples(), "\x00") == strings.Join(right.evidenceCounterexamples(), "\x00")
}

func (err *operationError) evidenceCounterexamples() []string {
	result := make([]string, 0, len(err.evidence))
	for _, evidence := range err.evidence {
		result = append(result, evidence.Counterexample)
	}
	return result
}

func observationFailure(action generation.Action, stage, step, reason, class, next string, blockedBy []string, process generation.ProcessObservation) generation.ObservationFailure {
	if len(process.Command) == 0 {
		process = descriptorObservation([]string{"<workspace>", "not-executed", string(action.Operation)}, nil, nil, -1)
	}
	decision := "UNKNOWN"
	unknownClass := class
	if class == "KNOWN_CONTRADICTION" {
		decision, unknownClass = "REFUTED", ""
	}
	if blockedBy == nil {
		blockedBy = []string{}
	}
	return generation.ObservationFailure{ActionIndicatorID: action.IndicatorID, Decision: decision, Stage: stage, Step: step, Reason: reason, UnknownClass: unknownClass, NextOperation: next, BlockedBy: append([]string{}, blockedBy...), Executor: process}
}

func executeAction(workspace, gitDir, metricsPath string, plan generation.Plan, action generation.Action) (operationMaterialization, *operationError) {
	if action.Operation == sourcepolicy.OperationSplitGo {
		return executeSplit(workspace, gitDir, metricsPath, plan, action)
	}
	if action.Operation == sourcepolicy.OperationExtractFunction {
		return executeExtract(workspace, gitDir, metricsPath, plan, action)
	}
	return operationMaterialization{}, newOperationError("execute-operation", "select-executor", "UNSUPPORTED_SELECTED_OPERATION", "KNOWN_CONTRADICTION", "report-counterexample")
}

func executeSplit(workspace, gitDir, metricsPath string, plan generation.Plan, action generation.Action) (operationMaterialization, *operationError) {
	first, firstErr := materializeSplit(workspace, gitDir, metricsPath, plan, action)
	if firstErr != nil {
		return first, firstErr
	}
	second, secondErr := materializeSplit(workspace, gitDir, metricsPath, plan, action)
	if secondErr != nil {
		return second, secondErr
	}
	if !bytes.Equal(first.Canonical, second.Canonical) || first.InstanceDigest != second.InstanceDigest {
		return first, newOperationError("replay-operation", "compare-instance-evidence", "REPLAY_EVIDENCE_MISMATCH", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	first.Verifier = second.Verifier
	return first, nil
}

func materializeSplit(workspace, gitDir, metricsPath string, plan generation.Plan, action generation.Action) (operationMaterialization, *operationError) {
	temporary, err := copyWorkspace(workspace)
	if err != nil {
		return operationMaterialization{}, newOperationError("prepare-workspace", "materialize-disposable-workspace", "WORKSPACE_MATERIALIZATION_FAILED", "DIRECT_MISSING", "restore-workspace")
	}
	defer os.RemoveAll(temporary)
	snapshot, snapshotErr := readOnlyGitSnapshot(gitDir, plan.HeadSHA)
	if snapshotErr != nil {
		return operationMaterialization{}, newOperationError("prepare-workspace", "isolate-git-context", "GIT_SNAPSHOT_UNAVAILABLE", "DIRECT_MISSING", "restore-git-context")
	}
	defer os.RemoveAll(filepath.Dir(snapshot))
	environment := childEnvironment(snapshot, temporary)
	command := []string{"go", "run", "./scripts/source-splitter", "-root", "<workspace>", "-metrics", "<source-metrics>", "-sha", plan.HeadSHA, "-subject", action.Subject, "-evidence-json"}
	result, runErr := runProcess(temporary, environment, command, []string{"go", "run", "./scripts/source-splitter", "-root", temporary, "-metrics", metricsPath, "-sha", plan.HeadSHA, "-subject", action.Subject, "-evidence-json"})
	if runErr != nil || result.Observation.ExitCode != 0 {
		return operationMaterialization{Executor: result.Observation}, newOperationError("execute-operation", "run-source-splitter", "EXECUTOR_PROCESS_FAILED", "DIRECT_MISSING", "restore-operation-evidence")
	}
	var evidence operationconformance.SplitGoEvidence
	if err := decodeStrictBytes(result.Stdout, &evidence); err != nil || evidence.ExpectedHeadSHA != plan.HeadSHA || evidence.Source.Path != action.Subject || evidence.OperationID != operationconformance.OperationID {
		return operationMaterialization{Executor: result.Observation}, newOperationError("evaluate-operation", "decode-source-splitter-evidence", "INSTANCE_EVIDENCE_MALFORMED", "MALFORMED_EVIDENCE", "restore-operation-evidence")
	}
	contractPath := filepath.Join(temporary, "examples", "source-splitter-conformance", "contract.json")
	contractRaw, err := os.ReadFile(contractPath)
	if err != nil {
		return operationMaterialization{Executor: result.Observation}, newOperationError("evaluate-operation", "read-operation-contract", "CONTRACT_EVIDENCE_UNAVAILABLE", "DIRECT_MISSING", "restore-operation-contract")
	}
	report := operationconformance.Evaluate(contractRaw, evidence)
	reportRaw, _ := json.Marshal(report)
	evaluator := descriptorObservation([]string{"operationconformance.Evaluate", operationconformance.OperationID}, reportRaw, nil)
	if err := operationconformance.Validate(report, contractRaw); err != nil || report.Decision != operationconformance.DecisionPass || report.Evidence.Source.Path != action.Subject {
		failure := newOperationError("evaluate-operation", "adjudicate-source-splitter", "INSTANCE_CONFORMANCE_FAILED", "KNOWN_CONTRADICTION", "report-counterexample")
		failure.evidence = splitFailureEvidence(report)
		return operationMaterialization{Executor: result.Observation, Evaluator: evaluator}, failure
	}
	verifier := runGoTest(temporary, environment)
	if verifier.Observation.ExitCode != 0 {
		return operationMaterialization{Executor: result.Observation, Evaluator: evaluator, Verifier: verifier.Observation}, newOperationError("verify-operation", "go-test-projected-workspace", "PROJECTED_COMPILE_OR_TEST_FAILED", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	canonicalValue := struct {
		Evidence operationconformance.SplitGoEvidence `json:"evidence"`
		Process  operationReplayEvidence              `json:"process"`
	}{Evidence: evidence, Process: operationReplayEvidenceFrom(result.Observation, evaluator, verifier.Observation)}
	canonical, _ := json.Marshal(canonicalValue)
	instance := digestBytes(canonical)
	indicators, ok := splitIndicatorReceipts(report, action, plan.HeadSHA)
	if !ok {
		return operationMaterialization{Executor: result.Observation, Evaluator: evaluator, Verifier: verifier.Observation}, newOperationError("evaluate-operation", "bind-indicator-observations", "INSTANCE_INDICATOR_MISSING", "DIRECT_MISSING", "restore-operation-evidence")
	}
	return operationMaterialization{OperationID: operationconformance.OperationID, InstanceDigest: instance, ContractDigest: digestBytes(contractRaw), Executor: result.Observation, Evaluator: evaluator, Verifier: verifier.Observation, Indicators: indicators, Canonical: canonical}, nil
}

func executeExtract(workspace, gitDir, metricsPath string, plan generation.Plan, action generation.Action) (operationMaterialization, *operationError) {
	subject, err := sourcepolicy.ParseSourceSubject(action.Subject)
	if err != nil {
		return operationMaterialization{}, newOperationError("observe-plan", "parse-function-subject", "SUBJECT_COORDINATE_MALFORMED", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	first, firstErr := materializeExtract(workspace, gitDir, metricsPath, plan, action, subject)
	if firstErr != nil {
		second, secondErr := materializeExtract(workspace, gitDir, metricsPath, plan, action, subject)
		if len(first.Canonical) == 0 || secondErr == nil || len(second.Canonical) == 0 {
			return first, firstErr
		}
		if !sameOperationError(firstErr, secondErr) || !bytes.Equal(first.Canonical, second.Canonical) ||
			first.InstanceDigest != second.InstanceDigest {
			return first, newOperationError("replay-operation", "compare-process-evidence", "REPLAY_EVIDENCE_MISMATCH", "KNOWN_CONTRADICTION", "report-counterexample")
		}
		return first, firstErr
	}
	second, secondErr := materializeExtract(workspace, gitDir, metricsPath, plan, action, subject)
	if secondErr != nil {
		return second, secondErr
	}
	if !bytes.Equal(first.Canonical, second.Canonical) || first.InstanceDigest != second.InstanceDigest {
		return first, newOperationError("replay-operation", "compare-instance-evidence", "REPLAY_EVIDENCE_MISMATCH", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	return first, nil
}

func materializeExtract(workspace, gitDir, metricsPath string, plan generation.Plan, action generation.Action, subject sourcepolicy.SourceSubject) (operationMaterialization, *operationError) {
	temporary, err := copyWorkspace(workspace)
	if err != nil {
		return operationMaterialization{}, newOperationError("prepare-workspace", "materialize-disposable-workspace", "WORKSPACE_MATERIALIZATION_FAILED", "DIRECT_MISSING", "restore-workspace")
	}
	defer os.RemoveAll(temporary)
	snapshot, snapshotErr := readOnlyGitSnapshot(gitDir, plan.HeadSHA)
	if snapshotErr != nil {
		return operationMaterialization{}, newOperationError("prepare-workspace", "isolate-git-context", "GIT_SNAPSHOT_UNAVAILABLE", "DIRECT_MISSING", "restore-git-context")
	}
	defer os.RemoveAll(filepath.Dir(snapshot))
	environment := childEnvironment(snapshot, temporary)
	sourcePath := filepath.Join(temporary, filepath.FromSlash(subject.Path))
	before, err := os.ReadFile(sourcePath)
	if err != nil {
		return operationMaterialization{}, newOperationError("observe-plan", "read-function-source", "SOURCE_UNAVAILABLE", "DIRECT_MISSING", "restore-source")
	}
	planPath, densityPath, prepErr := writeExtractorInputs(temporary, plan, subject, physicalLineCount(before))
	if prepErr != nil {
		return operationMaterialization{}, prepErr
	}
	command := []string{"go", "run", "./bootstrap/function-extractor", "-root", "<workspace>", "-plan", "meta-execution-function-plan.json", "-density-report", "meta-execution-function-density.json", "-expected-sha", plan.HeadSHA, "-output", "<extraction-report>"}
	reportName := "meta-execution-extraction-report.json"
	actual := []string{"go", "run", "./bootstrap/function-extractor", "-root", temporary, "-plan", planPath, "-density-report", densityPath, "-expected-sha", plan.HeadSHA, "-output", filepath.Join(temporary, reportName)}
	result, runErr := runProcess(temporary, environment, command, actual)
	if runErr != nil || result.Observation.ExitCode != 0 {
		failure := failedExtractionError(temporary, reportName, plan, action, result.Observation)
		materialized := operationMaterialization{Executor: result.Observation, Canonical: failure.canonical}
		if len(failure.canonical) != 0 {
			materialized.InstanceDigest = digestBytes(failure.canonical)
			materialized.Executor.StderrDigest = digestBytes(failure.canonical)
		}
		return materialized, failure
	}
	return evaluateExtractMaterialization(temporary, environment, before, result, plan, action, subject, reportName)
}

func writeExtractorInputs(root string, plan generation.Plan, subject sourcepolicy.SourceSubject, lines int) (string, string, *operationError) {
	planName := "meta-execution-function-plan.json"
	densityName := "meta-execution-function-density.json"
	planPath := filepath.Join(root, planName)
	densityPath := filepath.Join(root, densityName)
	if err := writeJSON(planPath, extractorPlan{Schema: "gooo.logical-split-plan.v1", SourceSHA: plan.HeadSHA, Subjects: []extractorPlanSubject{{Logical: subject.Path, Lines: lines, Reason: "fixed-declaration-capacity"}}}); err != nil {
		return "", "", newOperationError("observe-plan", "write-extraction-plan", "PLAN_OBSERVATION_UNAVAILABLE", "DIRECT_MISSING", "restore-extraction-plan")
	}
	if err := writeJSON(densityPath, extractorDensity{Schema: "gooo.line-density-rewrite.v1", SourceSHA: plan.HeadSHA, Subjects: []extractorDensitySubject{{Logical: subject.Path, Status: "blocked"}}}); err != nil {
		return "", "", newOperationError("observe-plan", "write-density-observation", "DENSITY_OBSERVATION_UNAVAILABLE", "DIRECT_MISSING", "restore-density-observation")
	}
	return planPath, densityPath, nil
}

func evaluateExtractMaterialization(temporary string, environment []string, before []byte, result processResult, plan generation.Plan, action generation.Action, subject sourcepolicy.SourceSubject, reportName string) (operationMaterialization, *operationError) {
	reportRaw, report, err := decodeExtractorReport(filepath.Join(temporary, reportName), plan.HeadSHA)
	if err != nil {
		if !os.IsNotExist(err) {
			return operationMaterialization{Executor: result.Observation}, newOperationError("evaluate-operation", "decode-function-extraction-report", "INSTANCE_EVIDENCE_MALFORMED", "KNOWN_CONTRADICTION", "report-counterexample")
		}
		return operationMaterialization{Executor: result.Observation}, newOperationError("evaluate-operation", "read-function-extraction-report", "INSTANCE_EVIDENCE_UNAVAILABLE", "DIRECT_MISSING", "restore-operation-evidence")
	}
	if failure := adjudicateExtractorReport(report); failure != nil {
		failure.counterexample = extractionCounterexample(action.Subject)
		failure.derivedRelations = extractorDerivedRelations(report, action)
		failure.evidence = extractionFailureEvidence(report, action)
		failure.canonical = canonicalStructuredExtractorFailure(action.Subject, failure, result.Observation)
		materialized := operationMaterialization{Executor: result.Observation, Canonical: failure.canonical}
		if len(failure.canonical) != 0 {
			materialized.InstanceDigest = digestBytes(failure.canonical)
			materialized.Executor.StderrDigest = digestBytes(failure.canonical)
		}
		return materialized, failure
	}
	observed, found := findExtractorSubject(report.Subjects, subject.Path)
	if !found || observed.Operation != string(sourcepolicy.OperationExtractFunction) || !containsString(observed.Operations, string(sourcepolicy.OperationExtractFunction)) || observed.Consumer != "function-extractor" || len(observed.Files) == 0 {
		return operationMaterialization{Executor: result.Observation}, newOperationError("evaluate-operation", "bind-function-extraction-subject", "INSTANCE_SUBJECT_MISSING", "DIRECT_MISSING", "restore-operation-evidence")
	}
	if !singleExtractorSubject(report, observed, subject) {
		return operationMaterialization{Executor: result.Observation}, newOperationError("evaluate-operation", "bind-function-extraction-subject", "INSTANCE_SUBJECT_CARDINALITY_MISMATCH", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	validation, err := validateExtractedFiles(temporary, subject, before, observed, report.NamespaceReplacements, report.BackupCleanup)
	if err != nil {
		reason := extractValidationErrorReason(err)
		class := extractValidationErrorClass(err)
		next := extractValidationNextOperation(err)
		return operationMaterialization{Executor: result.Observation}, newOperationError("evaluate-operation", "validate-function-extraction", reason, class, next)
	}
	evaluatorRaw, _ := json.Marshal(report)
	evaluator := descriptorObservation([]string{"bootstrap/function-extractor:independent-evaluator", subject.Path, subject.Name}, evaluatorRaw, nil)
	verifier := runGoTest(temporary, environment)
	if verifier.Observation.ExitCode != 0 {
		return operationMaterialization{Executor: result.Observation, Evaluator: evaluator, Verifier: verifier.Observation}, newOperationError("verify-operation", "go-test-projected-workspace", "PROJECTED_COMPILE_OR_TEST_FAILED", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	validation.ProjectedTestsPassed = true
	outputs, err := outputBytes(temporary, observed)
	if err != nil {
		return operationMaterialization{Executor: result.Observation, Evaluator: evaluator, Verifier: verifier.Observation}, newOperationError("evaluate-operation", "read-generated-output", "OUTPUT_EVIDENCE_UNAVAILABLE", "DIRECT_MISSING", "restore-operation-evidence")
	}
	canonicalValue := struct {
		Report  json.RawMessage   `json:"report"`
		Before  []byte            `json:"before"`
		Outputs map[string][]byte `json:"outputs"`
		Process operationReplayEvidence `json:"process"`
	}{Report: reportRaw, Before: before, Outputs: outputs,
		Process: operationReplayEvidenceFrom(result.Observation, evaluator, verifier.Observation)}
	canonical, _ := json.Marshal(canonicalValue)
	instance := digestBytes(canonical)
	contract := digestBytes([]byte("gooo/function-extraction/suffix-decomposition-contract/v1"))
	indicators, ok := extractIndicatorReceipts(action, validation, plan.HeadSHA)
	if !ok {
		return operationMaterialization{Executor: result.Observation, Evaluator: evaluator, Verifier: verifier.Observation}, newOperationError("evaluate-operation", "bind-indicator-observations", "INSTANCE_INDICATOR_MISSING", "DIRECT_MISSING", "restore-operation-evidence")
	}
	return operationMaterialization{OperationID: extractFunctionOperationID, InstanceDigest: instance, ContractDigest: contract, Executor: result.Observation, Evaluator: evaluator, Verifier: verifier.Observation, Indicators: indicators, Canonical: canonical}, nil
}

func singleExtractorSubject(report extractorReport, observed extractorSubject, subject sourcepolicy.SourceSubject) bool {
	values, ok := extractorIndicatorValues(report.Indicators)
	return ok && len(report.Subjects) == 1 && len(report.Unhandled) == 0 &&
		values["extraction.observed"] == 1 && observed.Logical == subject.Path
}

func splitIndicatorReceipts(report operationconformance.Report, action generation.Action, headSHA string) ([]generation.IndicatorReceipt, bool) {
	byID := make(map[string]operationconformance.IndicatorObservation, len(report.Indicators))
	for _, indicator := range report.Indicators {
		byID[indicator.ID] = indicator
	}
	transformed := splitTransformedSubject(report)
	result := make([]generation.IndicatorReceipt, 0, len(action.RequiredIndicatorIDs))
	for _, identifier := range action.RequiredIndicatorIDs {
		indicator, ok := byID[identifier]
		if !ok {
			return nil, false
		}
		verdict := generation.IndicatorVerdictUnknown
		actual := indicator.Value
		switch indicator.Decision {
		case operationconformance.DecisionPass:
			verdict = generation.IndicatorVerdictPass
		case operationconformance.DecisionFail:
			verdict = generation.IndicatorVerdictFail
		}
		result = append(result, makeIndicatorReceipt(identifier, action.Subject, headSHA, operationconformance.OperationID, actual, indicator.Target, 0, 0, transformed, verdict, action.ProofChoice))
	}
	return result, true
}

func splitFailureEvidence(report operationconformance.Report) []generation.ObservationFailureEvidence {
	result := make([]generation.ObservationFailureEvidence, 0, len(report.Counterexamples))
	for _, counterexample := range report.Counterexamples {
		result = append(result, generation.ObservationFailureEvidence{
			IndicatorID: counterexample.IndicatorID, Observed: counterexample.Observed,
			Expected: counterexample.Expected, Decision: string(counterexample.Decision),
			Counterexample: counterexample.RuleID,
		})
	}
	return result
}

func extractIndicatorReceipts(action generation.Action, validation extractValidation, headSHA string) ([]generation.IndicatorReceipt, bool) {
	result := make([]generation.IndicatorReceipt, 0, len(action.RequiredIndicatorIDs))
	for _, id := range action.RequiredIndicatorIDs {
		actual, ok := extractIndicatorValue(id, validation)
		if !ok {
			return nil, false
		}
		verdict := generation.IndicatorVerdictFail
		if id == "filesystem.atomic-replacement/v1" && !validation.AtomicReplacement {
			verdict = generation.IndicatorVerdictUnknown
		} else {
			verdict = generation.IndicatorVerdictFail
			if actual == 1 {
				verdict = generation.IndicatorVerdictPass
			}
		}
		result = append(result, makeIndicatorReceipt(id, action.Subject, headSHA, extractFunctionOperationID, actual, 1, validation.BeforeFunctionLines, validation.AfterFunctionLines, validation.TransformedSubject, verdict, action.ProofChoice))
	}
	return result, true
}

func extractIndicatorValue(id string, validation extractValidation) (int, bool) {
	checks := map[string]bool{
		"filesystem.atomic-replacement/v1": validation.AtomicReplacement,
		"go.format.fixed-point/v1":         validation.FormatFixedPoint,
		"go.header.preserved/v1":           validation.HeaderPreserved,
		"go.import.identity/v1":            validation.ImportIdentity,
		"go.package.conformance/v1":        validation.PackageConformance && validation.ProjectedTestsPassed,
	}
	passed, ok := checks[id]
	if !ok {
		return 0, false
	}
	if passed {
		return 1, true
	}
	return 0, true
}

func makeIndicatorReceipt(id, subject, headSHA, operationID string, actual, bound, beforeLines, afterLines int, transformed string, verdict generation.IndicatorVerdict, proof generation.ProofChoice) generation.IndicatorReceipt {
	observation := generation.IndicatorObservation{
		Schema: generation.IndicatorObservationSchema, IndicatorID: id, Subject: subject,
		HeadSHA: headSHA, OperationID: operationID,
		ValueKind: "integer", ActualValue: actual, ExpectedPredicate: "equal",
		ExpectedBound: bound, BeforeFunctionLines: beforeLines, AfterFunctionLines: afterLines,
		TransformedSubject: transformed,
	}
	return generation.IndicatorReceipt{ID: id, Verdict: verdict, EvidenceDigest: digestObservation(observation), ProofChoice: proof, Observation: &observation}
}

func digestObservation(observation generation.IndicatorObservation) string {
	payload, _ := json.Marshal(observation)
	return trimDigestPrefix(digestBytes(payload))
}

func splitTransformedSubject(report operationconformance.Report) string {
	paths := make([]string, 0, len(report.Evidence.Candidates))
	for _, candidate := range report.Evidence.Candidates {
		paths = append(paths, candidate.Path)
	}
	sort.Strings(paths)
	return report.Evidence.Source.Path + "=>" + strings.Join(paths, ",")
}

func findExtractorSubject(subjects []extractorSubject, logical string) (extractorSubject, bool) {
	for _, subject := range subjects {
		if subject.Logical == logical {
			return subject, true
		}
	}
	return extractorSubject{}, false
}

func sortedKeys(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func outputBytes(root string, observed extractorSubject) (map[string][]byte, error) {
	result := make(map[string][]byte, len(observed.Files))
	for _, logical := range observed.Files {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(logical)))
		if err != nil {
			return nil, err
		}
		result[logical] = data
	}
	return result, nil
}

func runGoTest(root string, environment []string) processResult {
	return runProcessResult(root, environment, []string{"go", "test", "./..."}, []string{"go", "test", "./..."})
}

func runProcessResult(root string, environment, descriptor, actual []string) processResult {
	result, _ := runProcess(root, environment, descriptor, actual)
	return result
}

func runProcess(root string, environment, descriptor, actual []string) (processResult, error) {
	command := exec.Command(actual[0], actual[1:]...)
	command.Dir = root
	command.Env = environment
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	exitCode := 0
	if err != nil {
		exitCode = 1
		if exitError, ok := errors.AsType[*exec.ExitError](err); ok {
			exitCode = exitError.ExitCode()
		}
	}
	observation := descriptorObservation(descriptor, stdout.Bytes(), stderr.Bytes(), exitCode)
	observation.StdoutDigest = digestBytes(canonicalProcessBytes(root, stdout.Bytes()))
	observation.StderrDigest = digestBytes(canonicalProcessBytes(root, stderr.Bytes()))
	return processResult{Observation: observation, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
}

func canonicalProcessBytes(root string, data []byte) []byte {
	replacements := canonicalProcessReplacements(root)
	sort.Slice(replacements, func(left, right int) bool {
		return len(replacements[left].path) > len(replacements[right].path)
	})
	canonical := append([]byte{}, data...)
	for _, replacement := range replacements {
		canonical = bytes.ReplaceAll(canonical, []byte(replacement.path), []byte(replacement.token))
	}
	return canonical
}

type canonicalProcessReplacement struct {
	path  string
	token string
}

func canonicalProcessReplacements(root string) []canonicalProcessReplacement {
	replacements := make([]canonicalProcessReplacement, 0, 12)
	addCanonicalProcessPath(&replacements, root, "<workspace>")
	addCanonicalProcessPath(&replacements, os.Getenv("LOGICAL_WORKSPACE"), "<workspace>")
	addCanonicalProcessPath(&replacements, os.Getenv("GITHUB_WORKSPACE"), "<workspace>")
	addCanonicalProcessPath(&replacements, os.Getenv("RUNNER_TEMP"), "<temp-workspace>")
	addCanonicalProcessPath(&replacements, os.TempDir(), "<temp-workspace>")
	return replacements
}

func addCanonicalProcessPath(replacements *[]canonicalProcessReplacement, value, token string) {
	if value == "" {
		return
	}
	absolute, err := filepath.Abs(value)
	if err != nil || absolute == "" {
		return
	}
	pathValue := filepath.Clean(absolute)
	*replacements = append(*replacements, canonicalProcessReplacement{path: pathValue, token: token})
	resolved, err := filepath.EvalSymlinks(pathValue)
	if err == nil && resolved != "" && filepath.Clean(resolved) != pathValue {
		*replacements = append(*replacements, canonicalProcessReplacement{path: filepath.Clean(resolved), token: token})
	}
}

func descriptorObservation(command []string, stdout, stderr []byte, exit ...int) generation.ProcessObservation {
	code := 0
	if len(exit) > 0 {
		code = exit[0]
	}
	return generation.ProcessObservation{Command: append([]string{}, command...), ExitCode: code,
		StdoutBytes: len(stdout), RawStdoutDigest: digestBytes(stdout), StdoutDigest: digestBytes(stdout),
		StderrBytes: len(stderr), RawStderrDigest: digestBytes(stderr), StderrDigest: digestBytes(stderr)}
}

func sourceMetricsPath() (string, error) {
	directory := os.Getenv("METRICS_DIR")
	if directory == "" {
		return "", fmt.Errorf("METRICS_DIR is unavailable")
	}
	path := filepath.Join(directory, "source-metrics.json")
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func gitDirectory(root string) (string, error) {
	if configured := os.Getenv("GIT_DIR"); configured != "" {
		if filepath.IsAbs(configured) {
			return configured, nil
		}
		return filepath.Join(root, configured), nil
	}
	command := exec.Command("git", "-C", root, "rev-parse", "--absolute-git-dir")
	output, err := command.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func childEnvironment(gitDir, workspace string) []string {
	environment := append([]string{}, os.Environ()...)
	environment = replaceEnvironment(environment, "GIT_DIR", gitDir)
	environment = replaceEnvironment(environment, "GIT_WORK_TREE", workspace)
	environment = replaceEnvironment(environment, "GIT_INDEX_FILE", "")
	return environment
}

func replaceEnvironment(environment []string, name, value string) []string {
	prefix := name + "="
	result := make([]string, 0, len(environment)+1)
	found := false
	for _, item := range environment {
		if strings.HasPrefix(item, prefix) {
			if value != "" {
				result = append(result, prefix+value)
			}
			found = true
			continue
		}
		result = append(result, item)
	}
	if !found && value != "" {
		result = append(result, prefix+value)
	}
	return result
}

func readOnlyGitSnapshot(source, expected string) (string, error) {
	parent, err := os.MkdirTemp("", "meta-operation-git-")
	if err != nil {
		return "", err
	}
	target := filepath.Join(parent, "snapshot.git")
	gitEnvironment := replaceEnvironment(replaceEnvironment(replaceEnvironment(os.Environ(), "GIT_DIR", ""), "GIT_WORK_TREE", ""), "GIT_INDEX_FILE", "")
	command := exec.Command("git", "clone", "--bare", "--no-local", source, target)
	command.Env = gitEnvironment
	if err := command.Run(); err != nil {
		_ = os.RemoveAll(parent)
		return "", err
	}
	headCommand := exec.Command("git", "--git-dir", target, "rev-parse", "HEAD")
	headCommand.Env = gitEnvironment
	head, err := headCommand.Output()
	if err != nil || strings.TrimSpace(string(head)) != expected {
		_ = os.RemoveAll(parent)
		return "", fmt.Errorf("git snapshot head does not match expected head")
	}
	if err := makeReadOnlySnapshot(target); err != nil {
		_ = os.RemoveAll(parent)
		return "", err
	}
	return target, nil
}

func makeReadOnlySnapshot(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		mode := info.Mode().Perm()
		if info.IsDir() {
			mode &= 0o555
			if mode == 0 {
				mode = 0o555
			}
		} else {
			mode &= 0o444
			if mode == 0 {
				mode = 0o444
			}
		}
		return os.Chmod(path, mode)
	})
}

func copyWorkspace(source string) (string, error) {
	target, err := os.MkdirTemp("", "meta-operation-workspace-")
	if err != nil {
		return "", err
	}
	if err := copyTree(source, target); err != nil {
		_ = os.RemoveAll(target)
		return "", err
	}
	return target, nil
}

func copyTree(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		if relative == ".git" || strings.HasPrefix(relative, ".git"+string(filepath.Separator)) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		destination := filepath.Join(target, relative)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("workspace symlink is not allowed: %s", relative)
		}
		if entry.IsDir() {
			return os.MkdirAll(destination, info.Mode().Perm())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(destination, data, info.Mode().Perm()); err != nil {
			return err
		}
		return os.Chmod(destination, info.Mode().Perm())
	})
}

func decodeStrictBytes(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	return nil
}

func writeJSON(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(path, append(payload, '\n'))
}

func physicalLineCount(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	count := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		count++
	}
	return count
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func trimDigestPrefix(value string) string {
	return strings.TrimPrefix(value, "sha256:")
}
