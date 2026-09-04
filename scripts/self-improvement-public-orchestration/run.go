package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publiccontinuity"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicorchestration"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
)

func runPrepare(input orchestrationInput) error {
	if err := validatePrepareInput(input); err != nil {
		return err
	}
	started := time.Now()
	policy, err := loadPolicy(input.Policy, input.Source)
	if err != nil {
		return err
	}
	if err := prepareEmptyDirectory(input.Output); err != nil {
		return err
	}
	if err := writePolicy(filepath.Join(input.Output, "orchestration-policy.json"), policy); err != nil {
		return err
	}
	firstDir := filepath.Join(input.Output, "first-generated")
	secondDir := filepath.Join(input.Output, "second-generated")
	ledgerDir := filepath.Join(input.Output, "ledger")
	if err := os.MkdirAll(filepath.Join(input.Output, "cli"), 0o755); err != nil {
		return err
	}
	state := policy.PreparePath[0]
	handlers := map[string]func() error{
		"OBSERVATION_QUORUM": func() error {
			if err := runPublic(input, filepath.Join(input.Output, "cli", "first.json"), "generate", input.Source, "--out", firstDir, "--observation-ledger", ledgerDir, "--json"); err != nil {
				return err
			}
			if err := runPublic(input, filepath.Join(input.Output, "cli", "second.json"), "generate", input.Source, "--out", secondDir, "--observation-ledger", ledgerDir, "--json"); err != nil {
				return err
			}
			return nil
		},
		"PROPOSAL_EMITTED": func() error {
			report, err := readDiscoveryReport(filepath.Join(ledgerDir, "observation-report.json"))
			if err != nil {
				return err
			}
			if report.Decision != publicorchestration.DecisionClosed || report.CandidatePath == "" {
				return errors.New("orchestration proposal did not reach the exact closed candidate boundary")
			}
			candidateBytes, err := readRegular(report.CandidatePath)
			if err != nil {
				return err
			}
			candidate, err := publiccontinuity.DecodeCandidate(candidateBytes)
			if err != nil {
				return err
			}
			if err := writeNew(filepath.Join(input.Output, "candidate.json"), candidateBytes, 0o444); err != nil {
				return err
			}
			candidateDigest := cache.HashBytes(candidateBytes).String()
			unknown, err := publicorchestration.NewUnknown(policy, publicorchestration.CaseMissingAuthorization)
			if err != nil {
				return err
			}
			handoff := publicorchestration.Handoff{
				Schema: publicorchestration.HandoffSchema, Operation: policy.Operation, Decision: publicorchestration.DecisionUnknown,
				State: policy.Boundary, NextOperation: unknown.NextOperation, Unknown: unknown,
				PolicySourceDigest: policy.SourceDigest, PolicySemanticDigest: policy.SemanticDigest, PolicyEvaluatorDigest: policy.EvaluatorDigest,
				CandidateDigest: candidateDigest, CandidateID: candidate.CandidateID,
				GeneratedOutputDigest: candidate.GeneratedOutputDigest, GeneratedManifestDigest: candidate.GeneratedManifestDigest,
				RequiredArtifacts: policy.HandoffArtifacts, AuthorizationRequired: true, ExecutionAllowed: false,
				RepositoryWrites: 0, LocalTestExecutions: 0,
			}
			handoff.HandoffID, err = publicorchestration.HandoffContentDigest(handoff)
			if err != nil {
				return err
			}
			if err := publicorchestration.WriteHandoff(filepath.Join(input.Output, "handoff.json"), handoff, policy); err != nil {
				return err
			}
			return nil
		},
	}
	events, err := policy.EventsFor(policy.PreparePath)
	if err != nil {
		return err
	}
	for _, event := range events {
		transition, ok := policy.Transition(state, event)
		if !ok {
			return fmt.Errorf("prepare event %q is not allowed from %q", event, state)
		}
		handler, ok := handlers[event]
		if !ok {
			return fmt.Errorf("prepare event %q has no executable handler", event)
		}
		if err := handler(); err != nil {
			return fmt.Errorf("prepare event %q: %w", event, err)
		}
		state = transition.To
	}
	handoff, err := publicorchestration.ReadHandoff(filepath.Join(input.Output, "handoff.json"), policy)
	if err != nil {
		return err
	}
	report := baseReport(policy, publicorchestration.CaseMissingAuthorization, publicorchestration.DecisionUnknown, policy.UnknownFor(publicorchestration.CaseMissingAuthorization), policy.PreparePath)
	report.Reason = policyCaseReason(policy, publicorchestration.CaseMissingAuthorization)
	report.Before = policy.Before
	report.After = policy.After
	report.After.WallMS = elapsedMS(started)
	report.After.PeakRSSKib = currentPeakRSSKib()
	report.HandoffDigest = handoff.HandoffID
	report.Input = inputInventory(input.Source, "")
	report.Generated = generatedInventory(firstDir)
	report.RuntimeComparable = false
	report.RuntimeUnknown = runtimeUnknown()
	return writeProtocolReport(input.Output, report, "The meta-defined public journey stopped at the explicit authorization boundary; no certificate, ordinary consumption, or test reuse was attempted.")
}

func runResume(input orchestrationInput) error {
	if err := validateResumeInput(input); err != nil {
		return err
	}
	started := time.Now()
	policy, err := loadPolicy(input.Policy, input.Source)
	if err != nil {
		return err
	}
	if err := prepareEmptyDirectory(input.Output); err != nil {
		return err
	}
	handoffData, err := readRegular(input.Handoff)
	if err != nil {
		return writeBoundaryReport(input.Output, policy, publicorchestration.CaseMalformedContinuation, started, nil, "continuation artifact is unavailable")
	}
	handoff, err := publicorchestration.DecodeHandoff(handoffData, policy)
	if err != nil {
		return writeBoundaryReport(input.Output, policy, publicorchestration.CaseMalformedContinuation, started, nil, err.Error())
	}
	if input.Authorization == "" {
		return writeBoundaryReport(input.Output, policy, publicorchestration.CaseMissingAuthorization, started, &handoff, "caller authorization artifact was not supplied")
	}
	candidateData, err := readRegular(input.Candidate)
	if err != nil {
		return writeBoundaryReport(input.Output, policy, publicorchestration.CaseContradictoryCandidate, started, &handoff, err.Error())
	}
	candidate, err := publiccontinuity.DecodeCandidate(candidateData)
	if err != nil {
		return writeBoundaryReport(input.Output, policy, publicorchestration.CaseContradictoryCandidate, started, &handoff, err.Error())
	}
	candidateDigest := cache.HashBytes(candidateData).String()
	if candidateDigest != handoff.CandidateDigest || candidate.CandidateID != handoff.CandidateID {
		return writeBoundaryReport(input.Output, policy, publicorchestration.CaseContradictoryCandidate, started, &handoff, "candidate digest or identity contradicts the durable handoff")
	}
	authorizationData, err := readRegular(input.Authorization)
	if err != nil {
		return writeBoundaryReport(input.Output, policy, publicorchestration.CaseMismatchedAuthorization, started, &handoff, err.Error())
	}
	authorization, err := decodeDecisionReceipt(authorizationData)
	if err != nil || authorization.Decision != publiccontinuity.DecisionAccept || authorization.Binding.CandidateDigest != candidateDigest || publiccontinuity.ValidateBinding(authorization.Binding, candidate, candidateDigest) != nil {
		return writeBoundaryReport(input.Output, policy, publicorchestration.CaseMismatchedAuthorization, started, &handoff, "authorization artifact does not explicitly accept the exact proposal")
	}
	state := policy.ResumePath[0]
	certificatePath := filepath.Join(input.Output, "certificate", "continuity-certificate.json")
	generatedDir := filepath.Join(input.Output, "generated")
	baselineDir := filepath.Join(input.Output, "baseline-package")
	replayDir := filepath.Join(input.Output, "replay-package")
	baselineResult := filepath.Join(input.Output, "baseline-result")
	replayResult := filepath.Join(input.Output, "replay-result")
	handlers := map[string]func() error{
		"EXPLICIT_AUTHORIZATION": func() error { return nil },
		"DURABLE_CERTIFICATE": func() error {
			return runPublic(input, filepath.Join(input.Output, "certificate-cli.json"), "certify-discovery", input.Contract, "--input", input.Source, "--candidate", input.Candidate, "--decision", input.Authorization, "--out", filepath.Dir(certificatePath))
		},
		"ORDINARY_OUTPUT": func() error {
			return runPublic(input, filepath.Join(input.Output, "generate-cli.json"), "generate", input.Source, "--continuity-certificate", certificatePath, "--out", generatedDir, "--json")
		},
		"REAL_PROJECT_TEST": func() error {
			return assembleAndRunBaseline(input, generatedDir, baselineDir, baselineResult)
		},
		"IMMUTABLE_RECEIPT_REUSE": func() error {
			return assembleAndRunReuse(input, generatedDir, replayDir, replayResult, filepath.Join(baselineResult, "reuse-receipt.json"))
		},
	}
	events, err := policy.EventsFor(policy.ResumePath)
	if err != nil {
		return err
	}
	for _, event := range events {
		transition, ok := policy.Transition(state, event)
		if !ok {
			return fmt.Errorf("resume event %q is not allowed from %q", event, state)
		}
		handler, ok := handlers[event]
		if !ok {
			return fmt.Errorf("resume event %q has no executable handler", event)
		}
		if err := handler(); err != nil {
			return fmt.Errorf("resume event %q: %w", event, err)
		}
		state = transition.To
	}
	certificate, err := readCertificate(certificatePath)
	if err != nil {
		return err
	}
	if certificate.Binding.CandidateDigest != candidateDigest || certificate.DecisionReceiptDigest != cache.HashBytes(authorizationData).String() {
		return errors.New("certificate does not bind the caller authorization and proposal")
	}
	generatedProgram, err := readRegular(filepath.Join(generatedDir, "semantic.gooo.go"))
	if err != nil {
		return err
	}
	generatedManifest, err := readRegular(filepath.Join(generatedDir, "semantic.gooo.manifest.jsonl"))
	if err != nil {
		return err
	}
	replayProgram, err := readRegular(filepath.Join(replayDir, "semantic.gooo.go"))
	if err != nil {
		return err
	}
	if !bytes.Equal(generatedProgram, replayProgram) {
		return errors.New("replay package generated program differs from ordinary generated output")
	}
	baselineReport, err := readTestReuseReport(filepath.Join(baselineResult, "reuse-report.json"))
	if err != nil {
		return err
	}
	replayReport, err := readTestReuseReport(filepath.Join(replayResult, "reuse-report.json"))
	if err != nil {
		return err
	}
	receipt, err := publictestreuse.ReadReceipt(filepath.Join(baselineResult, "reuse-receipt.json"))
	if err != nil {
		return err
	}
	if baselineReport.Decision != publicorchestration.DecisionClosed || baselineReport.BuildExecutions != 1 || baselineReport.TestExecutions != 1 || baselineReport.RepositoryWrites != 0 || baselineReport.LocalTestExecutions != 0 ||
		replayReport.Decision != publicorchestration.DecisionClosed || replayReport.TestExecutions != 0 || replayReport.ReusedTestExecutions != 1 || replayReport.ReceiptHits != 1 || replayReport.RepositoryWrites != 0 || replayReport.LocalTestExecutions != 0 ||
		baselineReport.Binding != replayReport.Binding || receipt.Binding != replayReport.Binding {
		return errors.New("real project validation or immutable test reuse did not preserve the exact evidence binding")
	}
	semanticDigest, err := manifestSemanticDigest(generatedManifest)
	if err != nil {
		return err
	}
	if semanticDigest != policy.SemanticDigest {
		return errors.New("ordinary generated manifest semantic digest differs from lowered policy")
	}
	report := baseReport(policy, publicorchestration.CaseAuthorizedOrchestration, publicorchestration.DecisionClosed, nil, append([]string(nil), policy.ResumePath...))
	report.Reason = policyCaseReason(policy, publicorchestration.CaseAuthorizedOrchestration)
	report.Before = policy.Before
	report.After = policy.After
	report.After.WallMS = elapsedMS(started)
	report.After.PeakRSSKib = currentPeakRSSKib()
	report.After.ReusedTestExecutions = 1
	report.Input = inputInventory(input.Source, input.ProjectTest)
	report.Generated = generatedInventory(generatedDir)
	report.Comparisons = publicorchestration.Comparisons{GeneratedBytesEqual: true, GeneratedSemanticEqual: true, TestContractBytesEqual: true, ReceiptBindingEqual: true, ContinuityPreserved: true, SafetyOutcomesPreserved: true}
	report.HandoffDigest = handoff.HandoffID
	report.AuthorizationDigest = cache.HashBytes(authorizationData).String()
	report.CertificateDigest = certificate.CertificateID
	report.ReceiptDigest = receipt.ReceiptID
	report.RuntimeComparable = false
	report.RuntimeUnknown = runtimeUnknown()
	if err := writeProtocolReport(input.Output, report, "The caller supplied explicit authorization, after which the meta-defined handlers certified, generated, validated, and reused one exact immutable test receipt."); err != nil {
		return err
	}
	return nil
}

func validatePrepareInput(input orchestrationInput) error {
	for name, value := range map[string]string{"policy": input.Policy, "source": input.Source, "gooo": input.Gooo, "repo-root": input.RepoRoot, "out": input.Output} {
		if value == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}

func validateResumeInput(input orchestrationInput) error {
	for name, value := range map[string]string{"policy": input.Policy, "source": input.Source, "contract": input.Contract, "gooo": input.Gooo, "test-reuse": input.TestReuse, "project-test": input.ProjectTest, "repo-root": input.RepoRoot, "out": input.Output, "handoff": input.Handoff, "candidate": input.Candidate} {
		if value == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}

func loadPolicy(policyPath, sourcePath string) (publicorchestration.Policy, error) {
	policySource, err := readRegular(policyPath)
	if err != nil {
		return publicorchestration.Policy{}, err
	}
	source, err := readRegular(sourcePath)
	if err != nil {
		return publicorchestration.Policy{}, err
	}
	if !bytes.Equal(policySource, source) {
		return publicorchestration.Policy{}, errors.New("policy and canonical source must be the same frozen .gooo file")
	}
	return publicorchestration.Load(policyPath, policySource)
}

func runPublic(input orchestrationInput, stdoutPath string, args ...string) error {
	file, err := os.OpenFile(stdoutPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	command := exec.Command(input.Gooo, args...)
	command.Dir = input.RepoRoot
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOFLAGS=-mod=readonly", "GOWORK=off")
	command.Stdout = file
	stderrPath := strings.TrimSuffix(stdoutPath, filepath.Ext(stdoutPath)) + ".stderr"
	errorFile, err := os.OpenFile(stderrPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	command.Stderr = errorFile
	runErr := command.Run()
	closeErr := errorFile.Close()
	if runErr != nil {
		return fmt.Errorf("public command %q: %w", strings.Join(args, " "), runErr)
	}
	return closeErr
}

func assembleAndRunBaseline(input orchestrationInput, generatedDir, packageDir, resultDir string) error {
	if err := assemblePackage(generatedDir, packageDir, input.ProjectTest, input.RepoRoot); err != nil {
		return err
	}
	return runTestReuse(input, resultDir, packageDir, filepath.Join(packageDir, "semantic.gooo.go"), filepath.Join(generatedDir, "semantic.gooo.manifest.jsonl"), filepath.Join(packageDir, "generated_project_test.go"), "baseline")
}

func assembleAndRunReuse(input orchestrationInput, generatedDir, packageDir, resultDir, receipt string) error {
	if err := assemblePackage(generatedDir, packageDir, input.ProjectTest, input.RepoRoot); err != nil {
		return err
	}
	return runTestReuse(input, resultDir, packageDir, filepath.Join(packageDir, "semantic.gooo.go"), filepath.Join(generatedDir, "semantic.gooo.manifest.jsonl"), filepath.Join(packageDir, "generated_project_test.go"), "reuse", receipt)
}

func assemblePackage(generatedDir, packageDir, projectTest, repoRoot string) error {
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		return err
	}
	for _, item := range []struct{ source, target string }{
		{filepath.Join(generatedDir, "semantic.gooo.go"), filepath.Join(packageDir, "semantic.gooo.go")},
		{projectTest, filepath.Join(packageDir, "generated_project_test.go")},
		{filepath.Join(repoRoot, "go.mod"), filepath.Join(packageDir, "go.mod")},
	} {
		data, err := readRegular(item.source)
		if err != nil {
			return err
		}
		if err := os.WriteFile(item.target, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func runTestReuse(input orchestrationInput, resultDir, packageDir, program, manifest, testContract, mode string, receipt ...string) error {
	args := []string{"-mode", mode, "-policy", input.Source, "-source", input.Source, "-program", program, "-manifest", manifest, "-test-contract", testContract, "-package-dir", packageDir, "-out", resultDir}
	if mode == "reuse" {
		args = append(args, "-receipt", receipt[0], "-authorize-reuse")
	}
	command := exec.Command(input.TestReuse, args...)
	command.Dir = input.RepoRoot
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOFLAGS=-mod=readonly", "GOWORK=off")
	command.Stdout = io.Discard
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("public test-reuse %s: %w", mode, err)
	}
	return nil
}

func baseReport(policy publicorchestration.Policy, caseID, decision string, unknown *publicorchestration.UnknownState, statePath []string) publicorchestration.Report {
	return publicorchestration.Report{Schema: publicorchestration.ReportSchema, Decision: decision, CaseID: caseID, Unknown: unknown,
		PolicySourceDigest: policy.SourceDigest, PolicySemanticDigest: policy.SemanticDigest, PolicyEvaluatorDigest: policy.EvaluatorDigest,
		Operation: policy.Operation, StatePath: statePath, Boundary: policy.Boundary, CaseDenominator: len(policy.Cases), ArtifactDenominator: policy.ArtifactDenom,
		RepositoryWrites: 0, LocalTestExecutions: 0}
}

func writeBoundaryReport(output string, policy publicorchestration.Policy, caseID string, started time.Time, handoff *publicorchestration.Handoff, detail string) error {
	decision, ok := policy.Decision(caseID)
	if !ok {
		return fmt.Errorf("policy does not declare boundary case %q", caseID)
	}
	unknown := policy.UnknownFor(caseID)
	report := baseReport(policy, caseID, decision, unknown, append([]string(nil), policy.ResumePath...))
	report.Reason = policyCaseReason(policy, caseID)
	report.Before = policy.Before
	report.After = policy.After
	report.After.WallMS = elapsedMS(started)
	report.After.PeakRSSKib = currentPeakRSSKib()
	report.Input = publicorchestration.Inventory{}
	report.RuntimeComparable = false
	report.RuntimeUnknown = runtimeUnknown()
	if handoff != nil {
		report.HandoffDigest = handoff.HandoffID
	}
	return writeProtocolReport(output, report, detail)
}

func writeProtocolReport(output string, report publicorchestration.Report, note string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := writeNew(filepath.Join(output, "orchestration-report.json"), append(data, '\n'), 0o444); err != nil {
		return err
	}
	human := fmt.Sprintf("# Public orchestration protocol\n\nDecision: `%s`\nCase: `%s`\nReason: `%s`\n\nState path: `%s`\nPublic CLI invocations before/after: `%d/%d`\nExplicit human decisions before/after: `%d/%d`\nSemantic/lowering/generation/test operations before: `%d/%d/%d/%d`; after: `%d/%d/%d/%d`\nHandoff artifacts before/after: `%d/%d`\nRepository writes / local test executions: `%d/%d`\n\n%s\n", report.Decision, report.CaseID, report.Reason, strings.Join(report.StatePath, " -> "), report.Before.PublicCLIInvocations, report.After.PublicCLIInvocations, report.Before.ExplicitHumanDecisions, report.After.ExplicitHumanDecisions, report.Before.SemanticOperations, report.Before.LoweringOperations, report.Before.GenerationOperations, report.Before.TestOperations, report.After.SemanticOperations, report.After.LoweringOperations, report.After.GenerationOperations, report.After.TestOperations, report.Before.HandoffArtifacts, report.After.HandoffArtifacts, report.RepositoryWrites, report.LocalTestExecutions, note)
	return writeNew(filepath.Join(output, "orchestration-report.md"), []byte(human), 0o444)
}

func writePolicy(filename string, policy publicorchestration.Policy) error {
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return err
	}
	return writeNew(filename, append(data, '\n'), 0o444)
}

func runtimeUnknown() *publicorchestration.UnknownState {
	return &publicorchestration.UnknownState{Stage: "UTILITY", Step: "COMPARE_MANUAL_ORCHESTRATED_RUNTIME", Reason: "RUNTIME_MODES_NOT_EQUIVALENT", UnknownClass: "INCOMPARABLE", NextOperation: "PREDECLARE_EQUIVALENT_RUNTIME_MEASUREMENT_RULE", BlockedBy: []string{"manual-vs-orchestrated-runtime-mode"}}
}

func policyCaseReason(policy publicorchestration.Policy, caseID string) string {
	item, ok := policy.CaseFor(caseID)
	if !ok {
		return "POLICY_CASE_NOT_FOUND"
	}
	return item.Reason
}

func readDiscoveryReport(filename string) (struct {
	Decision      string `json:"decision"`
	CandidatePath string `json:"candidate_path"`
}, error) {
	data, err := readRegular(filename)
	if err != nil {
		return struct {
			Decision      string `json:"decision"`
			CandidatePath string `json:"candidate_path"`
		}{}, err
	}
	var report struct {
		Decision      string `json:"decision"`
		CandidatePath string `json:"candidate_path"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return report, err
	}
	return report, nil
}

func readCertificate(filename string) (publiccontinuity.Certificate, error) {
	data, err := readRegular(filename)
	if err != nil {
		return publiccontinuity.Certificate{}, err
	}
	var certificate publiccontinuity.Certificate
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&certificate); err != nil {
		return publiccontinuity.Certificate{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return publiccontinuity.Certificate{}, errors.New("certificate contains multiple JSON values")
	} else if err != io.EOF {
		return publiccontinuity.Certificate{}, err
	}
	if err := publiccontinuity.ValidateCertificate(certificate); err != nil {
		return publiccontinuity.Certificate{}, err
	}
	return certificate, nil
}

func decodeDecisionReceipt(data []byte) (publiccontinuity.DecisionReceipt, error) {
	var receipt publiccontinuity.DecisionReceipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return receipt, err
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return receipt, errors.New("authorization contains multiple JSON values")
	} else if err != io.EOF {
		return receipt, err
	}
	if err := publiccontinuity.ValidateDecisionReceipt(receipt); err != nil {
		return receipt, err
	}
	return receipt, nil
}

func readTestReuseReport(filename string) (publictestreuse.Report, error) {
	data, err := readRegular(filename)
	if err != nil {
		return publictestreuse.Report{}, err
	}
	var report publictestreuse.Report
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return report, err
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return report, errors.New("test reuse report contains multiple JSON values")
	} else if !errors.Is(err, io.EOF) {
		return report, err
	}
	return report, nil
}

func manifestSemanticDigest(data []byte) (string, error) {
	var digest string
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var value struct {
			SemanticDigest string `json:"semantic_digest"`
		}
		if err := json.Unmarshal([]byte(line), &value); err != nil {
			return "", err
		}
		if !cache.Digest(value.SemanticDigest).Known() || (digest != "" && digest != value.SemanticDigest) {
			return "", errors.New("generated manifest semantic digest is missing or contradictory")
		}
		digest = value.SemanticDigest
	}
	if digest == "" {
		return "", errors.New("generated manifest has no semantic digest")
	}
	return digest, nil
}

func inputInventory(source, testContract string) publicorchestration.Inventory {
	result := publicorchestration.Inventory{}
	for _, filename := range []string{source, testContract} {
		if filename == "" {
			continue
		}
		data, err := readRegular(filename)
		if err != nil {
			continue
		}
		result.RegularFiles++
		result.PhysicalLines += physicalLines(data)
		switch filepath.Ext(filename) {
		case ".go":
			result.GoFiles++
			result.GoBytes += len(data)
			result.GoLines += physicalLines(data)
		case ".gooo":
			result.GoooFiles++
			result.GoooLines += physicalLines(data)
		}
	}
	return result
}

func generatedInventory(directory string) publicorchestration.GeneratedInventory {
	var result publicorchestration.GeneratedInventory
	program := filepath.Join(directory, "semantic.gooo.go")
	manifest := filepath.Join(directory, "semantic.gooo.manifest.jsonl")
	if data, err := readRegular(program); err == nil {
		result.GoFiles = 1
		result.GoBytes = len(data)
		result.GoLines = physicalLines(data)
		result.Files++
	}
	if data, err := readRegular(manifest); err == nil {
		result.ManifestBytes = len(data)
		result.Files++
	}
	return result
}

func readRegular(filename string) ([]byte, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", filename)
	}
	return os.ReadFile(filename)
}

func mustRead(filename string) []byte {
	data, _ := readRegular(filename)
	return data
}

func prepareEmptyDirectory(path string) error {
	if path == "" {
		return errors.New("output directory is required")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return errors.New("caller-owned output directory must be empty")
	}
	return nil
}

func writeNew(filename string, data []byte, mode os.FileMode) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func elapsedMS(started time.Time) int64 {
	value := time.Since(started) / time.Millisecond
	if value <= 0 {
		return 1
	}
	return int64(value)
}

func currentPeakRSSKib() int64 {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 1
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if !strings.HasPrefix(line, "VmHWM:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			value, err := strconv.ParseInt(fields[1], 10, 64)
			if err == nil && value > 0 {
				return value
			}
		}
	}
	return 1
}

func physicalLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	count := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		count++
	}
	return count
}
