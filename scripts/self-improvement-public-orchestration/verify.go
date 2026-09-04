package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publiccontinuity"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicorchestration"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
)

func runVerification(manifestPath, outputPath, humanPath string) error {
	if manifestPath == "" || outputPath == "" || humanPath == "" {
		return errors.New("evidence-manifest, output, and human-output are required")
	}
	inputData, err := readRegular(manifestPath)
	if err != nil {
		return err
	}
	var input evidenceInput
	decoder := json.NewDecoder(bytes.NewReader(inputData))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return fmt.Errorf("decode orchestration evidence manifest: %w", err)
	}
	if input.Schema != "gooo/public-self-improvement-orchestration-verification-input/v1" {
		return errors.New("orchestration evidence manifest schema is unknown")
	}
	policySource, err := readRegular(input.Policy)
	if err != nil {
		return err
	}
	policy, err := publicorchestration.Load(input.Policy, policySource)
	if err != nil {
		return err
	}
	source, err := readRegular(input.Source)
	if err != nil {
		return err
	}
	if !bytes.Equal(policySource, source) {
		return errors.New("policy and source bytes differ")
	}
	compiledPolicy, err := readPolicyArtifact(input.Policy)
	if err != nil {
		return err
	}
	if compiledPolicy.SourceDigest != policy.SourceDigest || compiledPolicy.SemanticDigest != policy.SemanticDigest || compiledPolicy.EvaluatorDigest != policy.EvaluatorDigest || compiledPolicy.Transitions == nil {
		return errors.New("lowered orchestration policy artifact does not match source")
	}
	if err := verifyPublishedArtifacts(input.PublishedRoot, input.PublishedArtifacts); err != nil {
		return err
	}
	if len(input.PublishedArtifacts)+2 != publicorchestration.ArtifactDenominator {
		return fmt.Errorf("published orchestration inputs = %d plus verification outputs, want %d", len(input.PublishedArtifacts), publicorchestration.ArtifactDenominator)
	}
	candidateData, err := readRegular(input.Candidate)
	if err != nil {
		return err
	}
	candidate, err := publiccontinuity.DecodeCandidate(candidateData)
	if err != nil {
		return err
	}
	candidateDigest := cache.HashBytes(candidateData).String()
	handoff, err := publicorchestration.ReadHandoff(input.Handoff, policy)
	if err != nil {
		return err
	}
	if handoff.CandidateDigest != candidateDigest || handoff.CandidateID != candidate.CandidateID {
		return errors.New("handoff does not bind the exact candidate")
	}
	authorizationData, err := readRegular(input.Authorization)
	if err != nil {
		return err
	}
	authorization, err := decodeDecisionReceipt(authorizationData)
	if err != nil || authorization.Decision != publiccontinuity.DecisionAccept || publiccontinuity.ValidateBinding(authorization.Binding, candidate, candidateDigest) != nil {
		return errors.New("accepted authorization is not explicit or candidate-bound")
	}
	rejectedData, err := readRegular(input.RejectedAuthorization)
	if err != nil {
		return err
	}
	rejected, err := decodeDecisionReceipt(rejectedData)
	if err != nil || rejected.Decision != publiccontinuity.DecisionReject || !rejected.ExplicitHumanDecision || rejected.Binding.CandidateDigest != candidateDigest {
		return errors.New("rejected authorization is not an explicit terminal decision")
	}
	certificate, err := readCertificate(input.Certificate)
	if err != nil {
		return err
	}
	if certificate.Binding.CandidateDigest != candidateDigest || certificate.DecisionReceiptDigest != cache.HashBytes(authorizationData).String() {
		return errors.New("certificate does not bind the accepted authorization")
	}
	generatedProgram, err := readRegular(input.GeneratedProgram)
	if err != nil {
		return err
	}
	generatedManifest, err := readRegular(input.GeneratedManifest)
	if err != nil {
		return err
	}
	if !bytes.Equal(generatedProgram, certificate.GeneratedSource) || !bytes.Equal(generatedManifest, certificate.GeneratedManifest) || cache.HashBytes(generatedProgram).String() != candidate.GeneratedOutputDigest || cache.HashBytes(generatedManifest).String() != candidate.GeneratedManifestDigest {
		return errors.New("ordinary generated output is not exactly the certified proposal output")
	}
	semanticDigest, err := manifestSemanticDigest(generatedManifest)
	if err != nil || semanticDigest != policy.SemanticDigest {
		return errors.New("generated output semantic digest is not equal to lowered source semantic digest")
	}
	projectTest, err := readRegular(input.ProjectTest)
	if err != nil {
		return err
	}
	resume, err := readOrchestrationReport(input.ResumeReport)
	if err != nil {
		return err
	}
	prepare, err := readOrchestrationReport(input.PrepareReport)
	if err != nil {
		return err
	}
	missing, err := readOrchestrationReport(input.MissingAuthorization)
	if err != nil {
		return err
	}
	malformed, err := readOrchestrationReport(input.MalformedContinuation)
	if err != nil {
		return err
	}
	contradictory, err := readOrchestrationReport(input.ContradictoryCandidate)
	if err != nil {
		return err
	}
	mismatched, err := readOrchestrationReport(input.MismatchedAuthorization)
	if err != nil {
		return err
	}
	if err := verifyPrepareReport(policy, prepare, handoff); err != nil {
		return err
	}
	if err := verifyClosedResume(policy, resume, handoff, authorization, certificate); err != nil {
		return err
	}
	if err := verifyUnknown(policy, missing, publicorchestration.CaseMissingAuthorization); err != nil {
		return err
	}
	if err := verifyUnknown(policy, malformed, publicorchestration.CaseMalformedContinuation); err != nil {
		return err
	}
	if err := verifyRefuted(policy, contradictory, publicorchestration.CaseContradictoryCandidate); err != nil {
		return err
	}
	if err := verifyRefuted(policy, mismatched, publicorchestration.CaseMismatchedAuthorization); err != nil {
		return err
	}
	baseline, err := readTestReuseReport(input.BaselineReuse)
	if err != nil {
		return err
	}
	replay, err := readTestReuseReport(input.ReplayReuse)
	if err != nil {
		return err
	}
	receipt, err := publictestreuse.ReadReceipt(input.Receipt)
	if err != nil {
		return err
	}
	if baseline.Decision != publictestreuse.DecisionClosed || baseline.BuildExecutions != 1 || baseline.TestExecutions != 1 || baseline.RepositoryWrites != 0 || baseline.LocalTestExecutions != 0 || replay.Decision != publictestreuse.DecisionClosed || replay.TestExecutions != 0 || replay.ReusedTestExecutions != 1 || replay.ReceiptHits != 1 || replay.RepositoryWrites != 0 || replay.LocalTestExecutions != 0 || baseline.Binding != replay.Binding || receipt.Binding != replay.Binding {
		return errors.New("test receipt reuse did not prove one real baseline test and zero duplicate replay tests")
	}
	projectTestDigest := cache.HashBytes(projectTest).String()
	if baseline.Binding.TestContractDigest != projectTestDigest || replay.Binding.TestContractDigest != projectTestDigest {
		return errors.New("baseline and replay test contracts are not the canonical project test")
	}
	status, err := readRepositoryStatus(input.RepositoryStatus)
	if err != nil {
		return err
	}
	if status.Before != "" || status.After != "" {
		return errors.New("orchestration changed the repository worktree")
	}
	if err := verifyRuntimeEvidence(input.RuntimeMeasurements, policy); err != nil {
		return err
	}
	inputInventoryValue := inputInventory(input.Source, input.ProjectTest)
	generatedInventoryValue := generatedInventory(filepath.Dir(input.GeneratedProgram))
	if inputInventoryValue.RegularFiles != 2 || inputInventoryValue.PhysicalLines != 19 || inputInventoryValue.GoFiles != 1 || inputInventoryValue.GoooFiles != 1 || generatedInventoryValue.Files != 2 || generatedInventoryValue.GoFiles != 1 || generatedInventoryValue.GoBytes == 0 || generatedInventoryValue.GoLines == 0 {
		return fmt.Errorf("input/generated inventory is not exact: input=%+v generated=%+v", inputInventoryValue, generatedInventoryValue)
	}
	before := policy.Before
	after := policy.After
	after.WallMS = resume.After.WallMS
	after.PeakRSSKib = resume.After.PeakRSSKib
	if after.WallMS <= 0 || after.PeakRSSKib <= 0 {
		return errors.New("orchestrated wall and RSS measurements are not positive integers")
	}
	caseResults := []publicorchestration.CaseResult{
		caseResult(policy, publicorchestration.CaseAuthorizedOrchestration, resume.Decision, resume.Reason, resume.Unknown),
		caseResult(policy, publicorchestration.CaseAuthorizedReceiptReuse, replay.Decision, publictestreuse.ReuseReason, nil),
		caseResult(policy, publicorchestration.CaseMissingAuthorization, missing.Decision, missing.Reason, missing.Unknown),
		caseResult(policy, publicorchestration.CaseMalformedContinuation, malformed.Decision, malformed.Reason, malformed.Unknown),
		caseResult(policy, publicorchestration.CaseContradictoryCandidate, contradictory.Decision, contradictory.Reason, contradictory.Unknown),
		caseResult(policy, publicorchestration.CaseMismatchedAuthorization, mismatched.Decision, mismatched.Reason, mismatched.Unknown),
	}
	closed, unknown, refuted := decisionCounts(caseResults)
	finalReport := publicorchestration.Report{
		Schema: publicorchestration.ReportSchema, Decision: publicorchestration.DecisionClosed,
		Reason: "EXACT_META_DEFINED_ORCHESTRATION_WITH_EXPLICIT_BOUNDARY_AND_FAIL_CLOSED_CONTINUATIONS",
		CaseID: publicorchestration.CaseAuthorizedOrchestration, PolicySourceDigest: policy.SourceDigest,
		PolicySemanticDigest: policy.SemanticDigest, PolicyEvaluatorDigest: policy.EvaluatorDigest, Operation: policy.Operation,
		StatePath: append([]string(nil), policy.States...), Boundary: policy.Boundary, Before: before, After: after,
		Input: inputInventoryValue, Generated: generatedInventoryValue,
		Comparisons: publicorchestration.Comparisons{GeneratedBytesEqual: true, GeneratedSemanticEqual: true, TestContractBytesEqual: true, ReceiptBindingEqual: true, ContinuityPreserved: true, SafetyOutcomesPreserved: true},
		Cases:       caseResults, CaseDenominator: len(caseResults), ClosedCases: closed, UnknownCases: unknown, RefutedCases: refuted,
		ArtifactDenominator: publicorchestration.ArtifactDenominator, ArtifactCount: publicorchestration.ArtifactDenominator,
		RepositoryWrites: 0, LocalTestExecutions: 0, RuntimeComparable: false, RuntimeUnknown: runtimeUnknown(),
		HandoffDigest: handoff.HandoffID, AuthorizationDigest: cache.HashBytes(authorizationData).String(), CertificateDigest: certificate.CertificateID, ReceiptDigest: receipt.ReceiptID,
		NoAggregateScore: true,
	}
	if closed != 2 || unknown != 2 || refuted != 2 || len(finalReport.Cases) != 6 || finalReport.Before.PublicCLIInvocations != 15 || finalReport.After.PublicCLIInvocations != 4 || finalReport.Before.ExplicitHumanDecisions != 2 || finalReport.After.ExplicitHumanDecisions < finalReport.Before.ExplicitHumanDecisions || finalReport.Before.SemanticOperations != finalReport.After.SemanticOperations || finalReport.Before.LoweringOperations != finalReport.After.LoweringOperations || finalReport.Before.GenerationOperations != finalReport.After.GenerationOperations || finalReport.Before.TestOperations != finalReport.After.TestOperations || !finalReport.Comparisons.ContinuityPreserved || !finalReport.Comparisons.SafetyOutcomesPreserved {
		return errors.New("orchestration utility improvement contract is not exact")
	}
	if err := writeVerificationOutputs(outputPath, humanPath, finalReport, baseline, replay); err != nil {
		return err
	}
	return nil
}

func verifyPrepareReport(policy publicorchestration.Policy, report publicorchestration.Report, handoff publicorchestration.Handoff) error {
	want := policy.UnknownFor(publicorchestration.CaseMissingAuthorization)
	if report.Decision != publicorchestration.DecisionUnknown || report.CaseID != publicorchestration.CaseMissingAuthorization || !publicorchestration.SameUnknown(report.Unknown, want) || report.HandoffDigest != handoff.HandoffID || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 {
		return errors.New("prepare did not stop at the explicit authorization boundary")
	}
	return nil
}

func verifyClosedResume(policy publicorchestration.Policy, report publicorchestration.Report, handoff publicorchestration.Handoff, authorization publiccontinuity.DecisionReceipt, certificate publiccontinuity.Certificate) error {
	if report.Decision != publicorchestration.DecisionClosed || report.CaseID != publicorchestration.CaseAuthorizedOrchestration || report.Unknown != nil || !sameStrings(report.StatePath, policy.ResumePath) || report.HandoffDigest != handoff.HandoffID || report.AuthorizationDigest == "" || report.CertificateDigest != certificate.CertificateID || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || authorization.Decision != publiccontinuity.DecisionAccept {
		return errors.New("resume did not prove the exact authorized state path")
	}
	return nil
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

func verifyUnknown(policy publicorchestration.Policy, report publicorchestration.Report, caseID string) error {
	if report.Decision != publicorchestration.DecisionUnknown || report.CaseID != caseID || !publicorchestration.SameUnknown(report.Unknown, policy.UnknownFor(caseID)) || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.After.SemanticOperations != policy.After.SemanticOperations {
		return fmt.Errorf("UNKNOWN case %s is not causal and fail closed", caseID)
	}
	return nil
}

func verifyRefuted(policy publicorchestration.Policy, report publicorchestration.Report, caseID string) error {
	if report.Decision != publicorchestration.DecisionRefuted || report.CaseID != caseID || report.Unknown != nil || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.After.SemanticOperations != policy.After.SemanticOperations {
		return fmt.Errorf("REFUTED case %s is not fail closed", caseID)
	}
	return nil
}

func caseResult(policy publicorchestration.Policy, caseID, observed, reason string, unknown *publicorchestration.UnknownState) publicorchestration.CaseResult {
	expected, _ := policy.Decision(caseID)
	return publicorchestration.CaseResult{ID: caseID, ExpectedDecision: expected, ObservedDecision: observed, Reason: reason, Unknown: unknown}
}

func decisionCounts(cases []publicorchestration.CaseResult) (int, int, int) {
	closed, unknown, refuted := 0, 0, 0
	for _, item := range cases {
		switch item.ObservedDecision {
		case publicorchestration.DecisionClosed:
			closed++
		case publicorchestration.DecisionUnknown:
			unknown++
		case publicorchestration.DecisionRefuted:
			refuted++
		}
	}
	return closed, unknown, refuted
}

func readOrchestrationReport(filename string) (publicorchestration.Report, error) {
	data, err := readRegular(filename)
	if err != nil {
		return publicorchestration.Report{}, err
	}
	var report publicorchestration.Report
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return report, err
	}
	return report, nil
}

func readPolicyArtifact(filename string) (publicorchestration.Policy, error) {
	data, err := readRegular(filename)
	if err != nil {
		return publicorchestration.Policy{}, err
	}
	var policy publicorchestration.Policy
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&policy); err != nil {
		return policy, err
	}
	return policy, policy.Validate()
}

func verifyPublishedArtifacts(root string, names []string) error {
	if root == "" || len(names) == 0 {
		return errors.New("published artifact root or list is empty")
	}
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if name == "" || filepath.IsAbs(name) || strings.Contains(name, "..") {
			return fmt.Errorf("published artifact name %q is unsafe", name)
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("published artifact %q is duplicated", name)
		}
		seen[name] = struct{}{}
		info, err := os.Lstat(filepath.Join(root, name))
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("published artifact %q is unavailable", name)
		}
	}
	return nil
}

func verifyRuntimeEvidence(filename string, policy publicorchestration.Policy) error {
	data, err := readRegular(filename)
	if err != nil {
		return err
	}
	var timing timingEvidence
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&timing); err != nil {
		return err
	}
	if timing.Schema != "gooo/public-self-improvement-orchestration-runtime/v1" || timing.OrchestratedWallMS <= 0 || timing.OrchestratedPeakRSSKib <= 0 || timing.ManualWallMS != 0 || timing.ManualPeakRSSKib != 0 || policy.PerformanceRule != "UNKNOWN_WHEN_RUNTIME_MODES_ARE_NOT_EQUIVALENT" {
		return errors.New("runtime evidence is not exact and explicitly incomparable")
	}
	return nil
}

type repositoryStatus struct {
	Schema string `json:"schema"`
	Before string `json:"before"`
	After  string `json:"after"`
}

func readRepositoryStatus(filename string) (repositoryStatus, error) {
	data, err := readRegular(filename)
	if err != nil {
		return repositoryStatus{}, err
	}
	var status repositoryStatus
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&status); err != nil {
		return status, err
	}
	if status.Schema != "gooo/public-self-improvement-repository-status/v1" {
		return status, errors.New("repository status schema is invalid")
	}
	return status, nil
}

func writeVerificationOutputs(outputPath, humanPath string, report publicorchestration.Report, baseline, replay publictestreuse.Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := writeNew(outputPath, append(data, '\n'), 0o444); err != nil {
		return err
	}
	human := fmt.Sprintf("# Public orchestration dossier\n\nDecision: `%s`\nReason: `%s`\n\nJourney: `%s`\nAuthorization boundary: `%s`\n\n## Utility\n\nPublic CLI invocations before/after: `%d/%d`\nExplicit human decisions before/after: `%d/%d`\nSemantic/lowering/generation/test operations before: `%d/%d/%d/%d`; after: `%d/%d/%d/%d`\nHandoff artifacts before/after: `%d/%d`\nWall ms / peak RSS KiB after: `%d/%d` (runtime comparison remains UNKNOWN)\n\n## Evidence\n\nCases: `%d CLOSED / %d UNKNOWN / %d REFUTED`; artifacts: `%d/%d`\nGenerated bytes equal: `%t`; generated semantic equal: `%t`; test contract equal: `%t`; receipt binding equal: `%t`\nReal project build/test executions: `%d/%d`; replay test executions/reused: `%d/%d`\nRepository writes / local test executions: `%d/%d`\n\nThe prepare invocation stopped at the explicit authorization boundary. Resume consumed the caller-supplied authorization, durable certificate, ordinary generated output, real Go validation, and one exact immutable test receipt. Contradictory or mismatched artifacts were REFUTED; incomplete continuation evidence was UNKNOWN. No aggregate score is emitted.\n", report.Decision, report.Reason, strings.Join(report.StatePath, " -> "), report.Boundary, report.Before.PublicCLIInvocations, report.After.PublicCLIInvocations, report.Before.ExplicitHumanDecisions, report.After.ExplicitHumanDecisions, report.Before.SemanticOperations, report.Before.LoweringOperations, report.Before.GenerationOperations, report.Before.TestOperations, report.After.SemanticOperations, report.After.LoweringOperations, report.After.GenerationOperations, report.After.TestOperations, report.Before.HandoffArtifacts, report.After.HandoffArtifacts, report.After.WallMS, report.After.PeakRSSKib, report.ClosedCases, report.UnknownCases, report.RefutedCases, report.ArtifactCount, report.ArtifactDenominator, report.Comparisons.GeneratedBytesEqual, report.Comparisons.GeneratedSemanticEqual, report.Comparisons.TestContractBytesEqual, report.Comparisons.ReceiptBindingEqual, baseline.BuildExecutions, baseline.TestExecutions, replay.TestExecutions, replay.ReusedTestExecutions, report.RepositoryWrites, report.LocalTestExecutions)
	return writeNew(humanPath, []byte(human), 0o444)
}
