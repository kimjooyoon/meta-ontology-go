package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
)

func runBaseline(input executionInput) error {
	identity, err := loadIdentity(input)
	if err != nil {
		return err
	}
	if err := prepareEmptyDirectory(input.OutputDir); err != nil {
		return err
	}
	started := time.Now()
	build, err := runTimed(input.PackageDir, "build", []string{"build", "."}, input.OutputDir)
	if err != nil {
		return err
	}
	test, err := runTimed(input.PackageDir, "test", []string{"test", "-tags", "generated_project", "-count=1", "."}, input.OutputDir)
	if err != nil {
		return err
	}
	report := baseReport(identity, publictestreuse.OriginalOperation)
	report.CaseID = publictestreuse.CaseBaselineExecution
	report.BuildExecutions = 1
	report.TestExecutions = 1
	report.ReceiptMisses = 1
	report.BuildMS = build.WallMS
	report.TestMS = test.WallMS
	report.WallMS = elapsedMS(started)
	report.PeakRSSKib = maxInt64(build.PeakRSSKib, test.PeakRSSKib)
	report.GeneratedBytesEqual = true
	report.SemanticEqual = identity.Binding.CanonicalSemanticDigest == identity.Binding.GeneratedSemanticDigest
	report.RepositoryWrites = 0
	report.LocalTestExecutions = 0
	if !build.Success {
		report.Decision = publictestreuse.DecisionRefuted
		report.Reason = "FAILED_ORIGINAL_BUILD"
	} else if !test.Success {
		report.Decision = publictestreuse.DecisionRefuted
		report.Reason = "FAILED_ORIGINAL_TEST"
	} else if !report.SemanticEqual {
		report.Decision = publictestreuse.DecisionRefuted
		report.Reason = "DIGEST_CONTRADICTION"
	} else {
		report.Decision = publictestreuse.DecisionClosed
		report.Reason = publictestreuse.ReasonBaseline
	}
	if expected, ok := identity.Policy.Decision(publictestreuse.CaseBaselineExecution); !ok || expected != publictestreuse.DecisionClosed {
		return errors.New("canonical policy does not declare the baseline case CLOSED")
	}
	if report.Decision == publictestreuse.DecisionClosed {
		receipt := publictestreuse.Receipt{
			Schema: publictestreuse.ReceiptSchema, Operation: publictestreuse.OriginalOperation,
			Decision: publictestreuse.DecisionClosed, Reason: publictestreuse.ReceiptReason,
			Binding: identity.Binding,
			Original: publictestreuse.OriginalExecution{
				Operation: publictestreuse.OriginalOperation, ResultDigest: test.ResultDigest,
				Successful: true, ExitCode: test.ExitCode, BuildExecutions: 1, TestExecutions: 1,
				BuildMS: build.WallMS, TestMS: test.WallMS, WallMS: report.WallMS, PeakRSSKib: report.PeakRSSKib,
			},
			RepositoryWrites: 0, LocalTestExecutions: 0,
		}
		receipt.Original.InvocationID = invocationDigest(receipt.Binding, receipt.Original)
		receipt.ReceiptID, err = publictestreuse.ReceiptContentDigest(receipt)
		if err != nil {
			return err
		}
		receiptPath := filepath.Join(input.OutputDir, "reuse-receipt.json")
		if err := publictestreuse.WriteReceipt(receiptPath, receipt); err != nil {
			return err
		}
		report.ReceiptDigest = receipt.ReceiptID
		report.OriginalReceiptID = receipt.ReceiptID
		report.OriginalInvocationID = receipt.Original.InvocationID
	}
	return writeReport(input.OutputDir, report, "The baseline command executed the fixed test contract and, only on success, sealed an immutable caller-owned receipt.")
}

func runReuse(input executionInput) error {
	identity, err := loadIdentity(input)
	if err != nil {
		return err
	}
	if err := prepareEmptyDirectory(input.OutputDir); err != nil {
		return err
	}
	started := time.Now()
	report := baseReport(identity, publictestreuse.ReplayOperation)
	report.CaseID = publictestreuse.CaseMissingAuthorization
	report.WallMS = elapsedMS(started)
	report.PeakRSSKib = currentPeakRSSKib()
	if expected, ok := identity.Policy.Decision(publictestreuse.CaseAuthorizedReuse); !ok || expected != publictestreuse.DecisionClosed {
		return errors.New("canonical policy does not declare the authorized reuse case CLOSED")
	}
	if !input.Authorize {
		report.Decision = publictestreuse.DecisionUnknown
		report.Reason = publictestreuse.ReasonMissingAuthorization
		report.Unknown = publictestreuse.Unknown(publictestreuse.ReasonMissingAuthorization, publictestreuse.UnknownAuthorization, publictestreuse.UnknownNextAuthorization, "explicit_reuse_authorization")
		return writeReport(input.OutputDir, report, "The receipt was not consulted because reuse was not explicitly authorized.")
	}
	receiptPath, state, err := resolveReceiptPath(input.Receipt)
	if err != nil {
		return err
	}
	if state != "" {
		report.CaseID = publictestreuse.CaseStaleEvidence
		report.Decision = publictestreuse.DecisionUnknown
		report.Reason = state
		report.Unknown = publictestreuse.Unknown(state, publictestreuse.UnknownEvidence, publictestreuse.UnknownNextEvidence, "caller_owned_reuse_receipt")
		report.ReceiptMisses = 1
		return writeReport(input.OutputDir, report, "The authorized reuse request had no single bounded receipt to validate, so no test was skipped.")
	}
	receipt, err := publictestreuse.ReadReceipt(receiptPath)
	if err != nil {
		report.CaseID = publictestreuse.CaseTamperedReceipt
		report.Decision = publictestreuse.DecisionRefuted
		report.Reason = publictestreuse.ReasonTampered
		report.ReceiptMisses = 1
		return writeReport(input.OutputDir, report, "The receipt failed immutable decoding or content validation; reuse was rejected fail-closed.")
	}
	if receipt.Binding.PolicySourceDigest != identity.Binding.PolicySourceDigest ||
		receipt.Binding.PolicySemanticDigest != identity.Binding.PolicySemanticDigest ||
		receipt.Binding.PolicyEvaluatorDigest != identity.Binding.PolicyEvaluatorDigest {
		report.CaseID = publictestreuse.CasePolicyMismatch
		report.Decision = publictestreuse.DecisionRefuted
		report.Reason = publictestreuse.ReasonPolicy
		report.ReceiptMisses = 1
		return writeReport(input.OutputDir, report, "The canonical .gooo policy or its lowered evaluator changed; the receipt cannot authorize reuse.")
	}
	if err := publictestreuse.VerifyReceipt(receipt, identity.Binding); err != nil {
		report.CaseID = publictestreuse.CaseStaleEvidence
		report.Decision = publictestreuse.DecisionUnknown
		report.Reason = publictestreuse.ReasonStale
		report.Unknown = publictestreuse.Unknown(publictestreuse.ReasonStale, publictestreuse.UnknownEvidence, publictestreuse.UnknownNextEvidence, "exact_receipt_binding")
		report.ReceiptMisses = 1
		return writeReport(input.OutputDir, report, "The receipt is valid but stale or incomparable with the current generated program and test contract; no test was skipped.")
	}
	report.CaseID = publictestreuse.CaseAuthorizedReuse
	report.Decision = publictestreuse.DecisionClosed
	report.Reason = publictestreuse.ReuseReason
	report.ReceiptDigest = receipt.ReceiptID
	report.OriginalReceiptID = receipt.ReceiptID
	report.OriginalInvocationID = receipt.Original.InvocationID
	report.ReusedTestExecutions = 1
	report.ReceiptHits = 1
	report.GeneratedBytesEqual = true
	report.SemanticEqual = true
	report.WallMS = elapsedMS(started)
	report.PeakRSSKib = currentPeakRSSKib()
	return writeReport(input.OutputDir, report, "The explicit authorization validated the immutable receipt and reused one prior successful test result without executing a duplicate test.")
}

func loadIdentity(input executionInput) (identity, error) {
	for name, path := range map[string]string{
		"policy": input.Policy, "source": input.Source, "program": input.Program,
		"manifest": input.Manifest, "test contract": input.TestContract, "package directory": input.PackageDir,
	} {
		if path == "" {
			return identity{}, fmt.Errorf("%s path is required", name)
		}
	}
	policySource, err := readRegular(input.Policy)
	if err != nil {
		return identity{}, fmt.Errorf("read policy: %w", err)
	}
	source, err := readRegular(input.Source)
	if err != nil {
		return identity{}, fmt.Errorf("read canonical source: %w", err)
	}
	if !bytes.Equal(policySource, source) {
		return identity{}, errors.New("policy and canonical source must be the same frozen .gooo file")
	}
	policy, err := publictestreuse.Load(input.Policy, policySource)
	if err != nil {
		return identity{}, err
	}
	program, err := readRegular(input.Program)
	if err != nil {
		return identity{}, fmt.Errorf("read generated program: %w", err)
	}
	manifest, err := readRegular(input.Manifest)
	if err != nil {
		return identity{}, fmt.Errorf("read generated manifest: %w", err)
	}
	testContract, err := readRegular(input.TestContract)
	if err != nil {
		return identity{}, fmt.Errorf("read test contract: %w", err)
	}
	if err := validatePackageInputs(input.PackageDir, input.Program, program, input.TestContract, testContract); err != nil {
		return identity{}, err
	}
	generatedSemantic, err := manifestSemanticDigest(manifest)
	if err != nil {
		return identity{}, err
	}
	compilerDigest, err := publictestreuse.CompilerDigest(os.ReadFile)
	if err != nil {
		return identity{}, err
	}
	toolchainVersion := publictestreuse.CurrentToolchainVersion()
	if toolchainVersion != "go1.27.0" {
		return identity{}, fmt.Errorf("public test reuse requires Go 1.27.0, got %s", toolchainVersion)
	}
	commandDigest := publictestreuse.TestCommandDigest(testCommand, cache.HashBytes(testContract).String())
	return identity{
		Policy: policy, ProgramBytes: program, ManifestBytes: manifest, TestBytes: testContract,
		Binding: publictestreuse.Binding{
			PolicySourceDigest: policy.SourceDigest, PolicySemanticDigest: policy.SemanticDigest, PolicyEvaluatorDigest: policy.EvaluatorDigest,
			CanonicalSourceDigest: cache.HashBytes(source).String(), CanonicalSemanticDigest: policy.SemanticDigest,
			GeneratedOutputDigest: cache.HashBytes(program).String(), GeneratedSemanticDigest: generatedSemantic,
			GeneratedManifestDigest: cache.HashBytes(manifest).String(), CompilerDigest: compilerDigest,
			ReleasedToolDigest: publictestreuse.ReleasedToolDigest(compilerDigest), ToolchainDigest: publictestreuse.ToolchainDigest(),
			ToolchainVersion: toolchainVersion, TestCommand: testCommand, TestCommandDigest: commandDigest,
			TestContractDigest: cache.HashBytes(testContract).String(),
		},
	}, nil
}

func validatePackageInputs(packageDir, programPath string, program []byte, testPath string, testContract []byte) error {
	info, err := os.Stat(packageDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("generated package directory is unavailable: %s", packageDir)
	}
	if !bytes.Equal(program, mustRead(filepath.Join(packageDir, filepath.Base(programPath)))) {
		return errors.New("generated program differs from the package file used by the test command")
	}
	if !bytes.Equal(testContract, mustRead(filepath.Join(packageDir, filepath.Base(testPath)))) {
		return errors.New("test contract differs from the package file used by the test command")
	}
	if _, err := os.Stat(filepath.Join(packageDir, "go.mod")); err != nil {
		return fmt.Errorf("generated package go.mod is unavailable: %w", err)
	}
	return nil
}

func mustRead(filename string) []byte {
	data, _ := readRegular(filename)
	return data
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

func manifestSemanticDigest(data []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var digest string
	lines := 0
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var value struct {
			SemanticDigest string `json:"semantic_digest"`
		}
		if err := json.Unmarshal(line, &value); err != nil {
			return "", fmt.Errorf("decode generated manifest: %w", err)
		}
		lines++
		if value.SemanticDigest == "" {
			return "", errors.New("generated manifest semantic digest is missing")
		}
		if digest != "" && digest != value.SemanticDigest {
			return "", errors.New("generated manifest contains contradictory semantic digests")
		}
		digest = value.SemanticDigest
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if lines == 0 || !cache.Digest(digest).Known() {
		return "", errors.New("generated manifest semantic digest is unknown")
	}
	return digest, nil
}

func runTimed(packageDir, label string, args []string, outputDir string) (commandResult, error) {
	timeFile := filepath.Join(outputDir, label+".time")
	started := time.Now()
	commandArgs := append([]string{"-f", "%M", "-o", timeFile, "go"}, args...)
	command := exec.Command("/usr/bin/time", commandArgs...)
	command.Dir = packageDir
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOFLAGS=-mod=readonly", "GOWORK=off")
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
	peak, readErr := readRSSFile(timeFile)
	if readErr != nil {
		return commandResult{}, fmt.Errorf("read %s peak RSS: %w", label, readErr)
	}
	resultDigest := cache.HashBytes([]byte(strings.Join(append([]string{"go", label}, args...), "\x00") + "\x00" + stdout.String() + "\x00" + stderr.String() + "\x00" + strconv.Itoa(exitCode))).String()
	return commandResult{Success: err == nil, ExitCode: exitCode, ResultDigest: resultDigest, WallMS: elapsedMS(started), PeakRSSKib: peak}, nil
}

func readRSSFile(filename string) (int64, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil || value <= 0 {
		return 0, errors.New("peak RSS is not a positive integer")
	}
	return value, nil
}

func resolveReceiptPath(path string) (string, string, error) {
	if path == "" {
		return "", "MISSING_REUSE_RECEIPT", nil
	}
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "MISSING_REUSE_RECEIPT", nil
		}
		return "", "", err
	}
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", "", err
		}
		regular := make([]string, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			candidate := filepath.Join(path, entry.Name())
			candidateInfo, statErr := os.Lstat(candidate)
			if statErr == nil && candidateInfo.Mode().IsRegular() {
				regular = append(regular, candidate)
			}
		}
		if len(regular) == 0 {
			return "", "MISSING_REUSE_RECEIPT", nil
		}
		if len(regular) != 1 {
			return "", "AMBIGUOUS_REUSE_RECEIPT", nil
		}
		return regular[0], "", nil
	}
	if !info.Mode().IsRegular() {
		return "", "", errors.New("reuse receipt path is not a regular file or directory")
	}
	return path, "", nil
}

func baseReport(identity identity, operation string) publictestreuse.Report {
	return publictestreuse.Report{
		Schema: publictestreuse.ReportSchema, Operation: operation, Binding: identity.Binding,
		GeneratedProgramBytes: int64(len(identity.ProgramBytes)), GeneratedProgramLines: physicalLines(identity.ProgramBytes),
		TestContractBytes: int64(len(identity.TestBytes)), TestContractLines: physicalLines(identity.TestBytes),
		RepositoryWrites: 0, LocalTestExecutions: 0,
	}
}

func writeReport(outputDir string, report publictestreuse.Report, note string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := writeNew(filepath.Join(outputDir, "reuse-report.json"), data, 0o444); err != nil {
		return err
	}
	human := fmt.Sprintf("# Public generated-test reuse\n\nDecision: `%s`\nReason: `%s`\nCase: `%s`\n\nTest executions: `%d`\nReused test executions: `%d`\nReceipt hits/misses: `%d/%d`\nBuild executions: `%d`\nBuild/test ms: `%d/%d`\nWall ms / peak RSS KiB: `%d/%d`\nGenerated program bytes / lines: `%d/%d`\nTest contract bytes / lines: `%d/%d`\nRepository writes / local test executions: `%d/%d`\nReceipt: `%s`\nOriginal invocation: `%s`\n\n%s\n", report.Decision, report.Reason, report.CaseID, report.TestExecutions, report.ReusedTestExecutions, report.ReceiptHits, report.ReceiptMisses, report.BuildExecutions, report.BuildMS, report.TestMS, report.WallMS, report.PeakRSSKib, report.GeneratedProgramBytes, report.GeneratedProgramLines, report.TestContractBytes, report.TestContractLines, report.RepositoryWrites, report.LocalTestExecutions, report.ReceiptDigest, report.OriginalInvocationID, note)
	return writeNew(filepath.Join(outputDir, "reuse-report.md"), []byte(human), 0o444)
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

func prepareEmptyDirectory(path string) error {
	if path == "" {
		return errors.New("report output directory is required")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return errors.New("caller-owned report directory must be empty")
	}
	return nil
}

func invocationDigest(binding publictestreuse.Binding, execution publictestreuse.OriginalExecution) string {
	digest, _ := cache.DigestOf(struct {
		Binding   publictestreuse.Binding
		Execution publictestreuse.OriginalExecution
	}{Binding: binding, Execution: execution})
	return digest.String()
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

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
