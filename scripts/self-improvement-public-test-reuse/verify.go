package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
)

func runVerification(manifestPath, outputPath, humanPath string) error {
	if manifestPath == "" || outputPath == "" || humanPath == "" {
		return errors.New("verification manifest, output, and human-output are required")
	}
	inputData, err := readRegular(manifestPath)
	if err != nil {
		return err
	}
	var input verificationInput
	decoder := json.NewDecoder(bytes.NewReader(inputData))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return fmt.Errorf("decode verification manifest: %w", err)
	}
	if input.Schema != "gooo/public-test-reuse-verification-input/v1" {
		return errors.New("verification manifest schema is unknown")
	}
	policySource, err := readRegular(input.Policy)
	if err != nil {
		return err
	}
	policy, err := publictestreuse.Load(input.Policy, policySource)
	if err != nil {
		return err
	}
	identity, err := loadIdentity(executionInput{Policy: input.Policy, Source: input.Source, Program: input.BaselineProgram, Manifest: input.Manifest, TestContract: input.TestContract, PackageDir: filepath.Dir(input.BaselineProgram), OutputDir: filepath.Dir(outputPath)})
	if err != nil {
		return err
	}
	replayProgram, err := readRegular(input.ReplayProgram)
	if err != nil {
		return err
	}
	testContract, err := readRegular(input.TestContract)
	if err != nil {
		return err
	}
	baselineProgram, err := readRegular(input.BaselineProgram)
	if err != nil {
		return err
	}
	if !bytes.Equal(baselineProgram, replayProgram) {
		return errors.New("baseline and replay generated program bytes differ")
	}
	baseline, err := readPublicReport(input.BaselineReport)
	if err != nil {
		return err
	}
	replay, err := readPublicReport(input.ReplayReport)
	if err != nil {
		return err
	}
	missingAuthorization, err := readPublicReport(input.MissingAuthorization)
	if err != nil {
		return err
	}
	staleEvidence, err := readPublicReport(input.StaleEvidence)
	if err != nil {
		return err
	}
	tamperedReceipt, err := readPublicReport(input.TamperedReceipt)
	if err != nil {
		return err
	}
	policyMismatch, err := readPublicReport(input.PolicyMismatch)
	if err != nil {
		return err
	}
	receipt, err := publictestreuse.ReadReceipt(input.BaselineReceipt)
	if err != nil {
		return err
	}
	if err := publictestreuse.VerifyReceipt(receipt, identity.Binding); err != nil {
		return fmt.Errorf("baseline receipt: %w", err)
	}
	if err := verifyBaselineReport(baseline, identity.Binding, receipt); err != nil {
		return err
	}
	if err := verifyReplayReport(replay, identity.Binding, receipt); err != nil {
		return err
	}
	if err := verifyUnknownReport(missingAuthorization, identity.Binding, publictestreuse.CaseMissingAuthorization, publictestreuse.ReasonMissingAuthorization, publictestreuse.Unknown(publictestreuse.ReasonMissingAuthorization, publictestreuse.UnknownAuthorization, publictestreuse.UnknownNextAuthorization, "explicit_reuse_authorization")); err != nil {
		return err
	}
	if err := verifyUnknownReport(staleEvidence, staleEvidence.Binding, publictestreuse.CaseStaleEvidence, publictestreuse.ReasonStale, publictestreuse.Unknown(publictestreuse.ReasonStale, publictestreuse.UnknownEvidence, publictestreuse.UnknownNextEvidence, "exact_receipt_binding")); err != nil {
		return err
	}
	if staleEvidence.Binding.GeneratedOutputDigest == identity.Binding.GeneratedOutputDigest {
		return errors.New("stale evidence case did not change the generated program digest")
	}
	if err := verifyRefutedReport(tamperedReceipt, publictestreuse.CaseTamperedReceipt, publictestreuse.ReasonTampered); err != nil {
		return err
	}
	if err := verifyRefutedReport(policyMismatch, publictestreuse.CasePolicyMismatch, publictestreuse.ReasonPolicy); err != nil {
		return err
	}
	if policyMismatch.Binding.PolicySourceDigest == identity.Binding.PolicySourceDigest {
		return errors.New("policy mismatch case did not change the policy source digest")
	}
	if len(input.PublishedArtifacts)+2 != publictestreuse.ArtifactDenominator {
		return fmt.Errorf("published artifact inputs = %d plus verification outputs, want %d", len(input.PublishedArtifacts), publictestreuse.ArtifactDenominator)
	}
	if err := verifyPublishedArtifacts(input.PublishedRoot, input.PublishedArtifacts); err != nil {
		return err
	}
	closed, unknown, refuted := 0, 0, 0
	cases := make([]verifiedCase, 0, len(publictestreuse.CanonicalCaseIDs()))
	for _, report := range []publictestreuse.Report{baseline, replay, missingAuthorization, staleEvidence, tamperedReceipt, policyMismatch} {
		expected, ok := policy.Decision(report.CaseID)
		if !ok || expected != report.Decision {
			return fmt.Errorf("policy decision for %s = %s/%s", report.CaseID, report.Decision, expected)
		}
		caseReport := verifiedCase{ID: report.CaseID, ExpectedDecision: expected, ObservedDecision: report.Decision, Reason: report.Reason, TestExecutions: report.TestExecutions, ReusedTestExecutions: report.ReusedTestExecutions, ReceiptHits: report.ReceiptHits, ReceiptMisses: report.ReceiptMisses, RepositoryWrites: report.RepositoryWrites, LocalTestExecutions: report.LocalTestExecutions}
		cases = append(cases, caseReport)
		switch report.Decision {
		case publictestreuse.DecisionClosed:
			closed++
		case publictestreuse.DecisionUnknown:
			unknown++
		case publictestreuse.DecisionRefuted:
			refuted++
		}
	}
	if len(cases) != 6 || closed != 2 || unknown != 2 || refuted != 2 {
		return fmt.Errorf("case decisions are %d/%d/%d, want 2/2/2", closed, unknown, refuted)
	}
	if baseline.Binding != replay.Binding || receipt.Binding != replay.Binding {
		return errors.New("baseline, replay, and receipt bindings differ")
	}
	if !bytes.Equal(testContract, identity.TestBytes) {
		return errors.New("test contract bytes changed during verification")
	}
	report := verificationReport{
		Schema: "gooo/public-test-reuse-verification/v1", Decision: publictestreuse.DecisionClosed,
		Reason:             "EXACT_AUTHORIZED_TEST_REUSE_WITH_FAIL_CLOSED_ALTERNATIVES",
		PolicySourceDigest: policy.SourceDigest, PolicySemanticDigest: policy.SemanticDigest, PolicyEvaluatorDigest: policy.EvaluatorDigest,
		Journey: policy.Journey, InputRegularFiles: 2, InputPhysicalLines: physicalLines(identity.ProgramBytes[:0]) + physicalLines(identity.TestBytes) + physicalLines(policySource),
		InputGoFiles: 1, InputGoPhysicalLines: physicalLines(identity.TestBytes), GeneratedFiles: 2, GeneratedGoFiles: 1,
		GeneratedProgramBytes: int64(len(identity.ProgramBytes)), GeneratedProgramLines: physicalLines(identity.ProgramBytes),
		TestContractBytes: int64(len(identity.TestBytes)), TestContractLines: physicalLines(identity.TestBytes),
		Before: snapshot(baseline), After: snapshot(replay),
		Comparisons: comparisonSnapshot{GeneratedProgramBytesEqual: true, GeneratedSemanticEqual: identity.Binding.CanonicalSemanticDigest == identity.Binding.GeneratedSemanticDigest, TestContractBytesEqual: true, ReceiptBindingEqual: true},
		Cases:       cases, CaseDenominator: 6, ClosedCases: closed, UnknownCases: unknown, RefutedCases: refuted,
		ArtifactDenominator: publictestreuse.ArtifactDenominator, ArtifactCount: publictestreuse.ArtifactDenominator,
		RepositoryWrites: 0, LocalTestExecutions: 0, NoAggregateScore: true, Binding: identity.Binding,
	}
	if report.InputPhysicalLines != 19 || !report.Comparisons.GeneratedSemanticEqual || baseline.TestExecutions != 1 || replay.TestExecutions != 0 || replay.ReusedTestExecutions != 1 {
		return errors.New("test reuse utility evidence is not the exact equal-input before/after pair")
	}
	if err := writeVerificationOutputs(outputPath, humanPath, report); err != nil {
		return err
	}
	return nil
}

func verifyBaselineReport(report publictestreuse.Report, binding publictestreuse.Binding, receipt publictestreuse.Receipt) error {
	if report.Schema != publictestreuse.ReportSchema || report.Operation != publictestreuse.OriginalOperation ||
		report.Decision != publictestreuse.DecisionClosed || report.Reason != publictestreuse.ReasonBaseline || report.CaseID != publictestreuse.CaseBaselineExecution ||
		report.Binding != binding || report.ReceiptDigest != receipt.ReceiptID || report.TestExecutions != 1 || report.ReusedTestExecutions != 0 ||
		report.ReceiptHits != 0 || report.ReceiptMisses != 1 || report.BuildExecutions != 1 || report.BuildMS <= 0 || report.TestMS <= 0 || report.WallMS <= 0 || report.PeakRSSKib <= 0 ||
		report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.Unknown != nil {
		return errors.New("baseline test reuse report is not one successful execution")
	}
	return nil
}

func verifyReplayReport(report publictestreuse.Report, binding publictestreuse.Binding, receipt publictestreuse.Receipt) error {
	if report.Schema != publictestreuse.ReportSchema || report.Operation != publictestreuse.ReplayOperation ||
		report.Decision != publictestreuse.DecisionClosed || report.Reason != publictestreuse.ReuseReason || report.CaseID != publictestreuse.CaseAuthorizedReuse ||
		report.Binding != binding || report.ReceiptDigest != receipt.ReceiptID || report.OriginalReceiptID != receipt.ReceiptID || report.OriginalInvocationID != receipt.Original.InvocationID ||
		report.TestExecutions != 0 || report.ReusedTestExecutions != 1 || report.ReceiptHits != 1 || report.ReceiptMisses != 0 || report.BuildExecutions != 0 || report.BuildMS != 0 || report.TestMS != 0 || report.WallMS <= 0 || report.PeakRSSKib <= 0 ||
		report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.Unknown != nil {
		return errors.New("authorized replay did not prove zero duplicate test executions")
	}
	return nil
}

func verifyUnknownReport(report publictestreuse.Report, binding publictestreuse.Binding, caseID, reason string, unknown *publictestreuse.UnknownState) error {
	if report.Schema != publictestreuse.ReportSchema || report.Operation != publictestreuse.ReplayOperation || report.Decision != publictestreuse.DecisionUnknown || report.CaseID != caseID || report.Reason != reason || report.Binding != binding || !publictestreuse.SameUnknown(report.Unknown, unknown) || report.TestExecutions != 0 || report.ReusedTestExecutions != 0 || report.BuildExecutions != 0 || report.BuildMS != 0 || report.TestMS != 0 || report.ReceiptHits != 0 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 {
		return fmt.Errorf("UNKNOWN case %s is not causal and fail-closed", caseID)
	}
	return nil
}

func verifyRefutedReport(report publictestreuse.Report, caseID, reason string) error {
	if report.Schema != publictestreuse.ReportSchema || report.Operation != publictestreuse.ReplayOperation || report.Decision != publictestreuse.DecisionRefuted || report.CaseID != caseID || report.Reason != reason || report.Unknown != nil || report.TestExecutions != 0 || report.ReusedTestExecutions != 0 || report.BuildExecutions != 0 || report.BuildMS != 0 || report.TestMS != 0 || report.ReceiptHits != 0 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 {
		return fmt.Errorf("REFUTED case %s is not fail-closed", caseID)
	}
	return nil
}

func readPublicReport(filename string) (publictestreuse.Report, error) {
	data, err := readRegular(filename)
	if err != nil {
		return publictestreuse.Report{}, err
	}
	var report publictestreuse.Report
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return publictestreuse.Report{}, fmt.Errorf("decode report %s: %w", filename, err)
	}
	return report, nil
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

func snapshot(report publictestreuse.Report) metricSnapshot {
	return metricSnapshot{BuildExecutions: report.BuildExecutions, TestExecutions: report.TestExecutions, ReusedTestExecutions: report.ReusedTestExecutions, ReceiptHits: report.ReceiptHits, ReceiptMisses: report.ReceiptMisses, BuildMS: report.BuildMS, TestMS: report.TestMS, WallMS: report.WallMS, PeakRSSKib: report.PeakRSSKib}
}

func writeVerificationOutputs(outputPath, humanPath string, report verificationReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(humanPath), 0o755); err != nil {
		return err
	}
	if err := writeNew(outputPath, data, 0o444); err != nil {
		return err
	}
	human := fmt.Sprintf("# Public generated-test reuse dossier\n\nDecision: `%s`\nReason: `%s`\n\n## DRIVER\n\nThe existing generated public-discovery program and its tagged Go acceptance test are the frozen equal-input pair. The public journey is `%s`. The receipt binds the canonical source and lowered semantic digest, generated bytes and semantic manifest, compiler and released-tool identities, Go 1.27 toolchain, test command, test contract, and successful original result.\n\n## OUTCOME\n\nCases: `%d CLOSED / %d UNKNOWN / %d REFUTED`\nTest executions before/after: `%d/%d`\nReused test executions before/after: `%d/%d`\nBuild executions before/after: `%d/%d`\nBuild/test ms before: `%d/%d`; after: `%d/%d`\nWall ms / peak RSS KiB before: `%d/%d`; after: `%d/%d`\nGenerated program bytes/lines: `%d/%d`; test contract bytes/lines: `%d/%d`\nReceipt hits/misses before: `%d/%d`; after: `%d/%d`\nGenerated bytes equal: `%t`; generated semantic equal: `%t`; test contract equal: `%t`\n\n## GUARDRAIL\n\nReuse is CLOSED only for explicit authorization plus one exact immutable successful receipt. Missing authorization and stale or unbounded evidence are UNKNOWN with no skipped test. Tampered receipt and policy mismatch are REFUTED and fail closed. No aggregate score is emitted. Repository writes=`%d`; local test executions=`%d`; published artifacts=`%d/%d`.\n\n### Cases\n\n%s", report.Decision, report.Reason, strings.Join(report.Journey, " -> "), report.ClosedCases, report.UnknownCases, report.RefutedCases, report.Before.TestExecutions, report.After.TestExecutions, report.Before.ReusedTestExecutions, report.After.ReusedTestExecutions, report.Before.BuildExecutions, report.After.BuildExecutions, report.Before.BuildMS, report.Before.TestMS, report.After.BuildMS, report.After.TestMS, report.Before.WallMS, report.Before.PeakRSSKib, report.After.WallMS, report.After.PeakRSSKib, report.GeneratedProgramBytes, report.GeneratedProgramLines, report.TestContractBytes, report.TestContractLines, report.Before.ReceiptHits, report.Before.ReceiptMisses, report.After.ReceiptHits, report.After.ReceiptMisses, report.Comparisons.GeneratedProgramBytesEqual, report.Comparisons.GeneratedSemanticEqual, report.Comparisons.TestContractBytesEqual, report.RepositoryWrites, report.LocalTestExecutions, report.ArtifactCount, report.ArtifactDenominator, renderCases(report.Cases))
	return writeNew(humanPath, []byte(human), 0o444)
}

func renderCases(cases []verifiedCase) string {
	var builder strings.Builder
	for _, item := range cases {
		fmt.Fprintf(&builder, "- `%s`: `%s` / `%s`; test executions=%d, reused=%d, receipt hits/misses=%d/%d\n", item.ID, item.ObservedDecision, item.Reason, item.TestExecutions, item.ReusedTestExecutions, item.ReceiptHits, item.ReceiptMisses)
	}
	return builder.String()
}
