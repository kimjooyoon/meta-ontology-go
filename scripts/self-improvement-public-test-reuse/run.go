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
	"github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
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
	build, err := runTimed(identity.GoCommand, input.PackageDir, "build", []string{"build", "."})
	if err != nil {
		return err
	}
	test, err := runTimed(identity.GoCommand, input.PackageDir, "test", []string{"test", "-tags", "generated_project", "-count=1", "."})
	if err != nil {
		return err
	}
	report := baselineReport(identity, build, test)
	observed := publictestreuse.OriginalObservation{
		BuildPassed: build.Success, TestPassed: test.Success, SemanticMatches: report.SemanticEqual,
	}
	if err := assessOriginalReport(identity.Policy, observed, &report); err != nil {
		return err
	}
	report.WallMS = elapsedMS(started)
	if report.Decision == publictestreuse.DecisionClosed {
		if err := sealOriginalReceipt(input.OutputDir, identity, build, test, &report); err != nil {
			return err
		}
	}
	return writeReport(input.OutputDir, report, "Gooo classified the observed original execution; a successful bound result can produce its immutable receipt.")
}

func baselineReport(identity identity, build, test commandResult) publictestreuse.Report {
	report := baseReport(identity, publictestreuse.OriginalOperation)
	report.BuildExecutions = 1
	report.TestExecutions = 1
	report.ReceiptMisses = 1
	report.BuildMS = build.WallMS
	report.TestMS = test.WallMS
	report.PeakRSSKib = maxInt64(build.PeakRSSKib, test.PeakRSSKib)
	report.PeakRSSSubject = publictestreuse.PeakRSSGoCommands
	report.GeneratedBytesEqual = true
	report.SemanticEqual = identity.Binding.CanonicalSemanticDigest == identity.Binding.GeneratedSemanticDigest
	return report
}

func sealOriginalReceipt(output string, identity identity, build, test commandResult, report *publictestreuse.Report) error {
	if !build.Success || !test.Success || identity.Binding.CanonicalSemanticDigest != identity.Binding.GeneratedSemanticDigest {
		return errors.New("Gooo policy cannot seal an unsuccessful original execution")
	}
	if err := identity.GoCommand.VerifyBytes(); err != nil {
		return err
	}
	receipt := publictestreuse.Receipt{
		Schema: publictestreuse.ReceiptSchema, Operation: publictestreuse.OriginalOperation,
		Decision: publictestreuse.DecisionClosed, Reason: publictestreuse.ReceiptReason,
		Binding: identity.Binding,
		Original: publictestreuse.OriginalExecution{
			Operation: publictestreuse.OriginalOperation, ResultDigest: test.ResultDigest,
			Successful: true, ExitCode: test.ExitCode, BuildExecutions: 1, TestExecutions: 1,
			BuildMS: build.WallMS, TestMS: test.WallMS, WallMS: report.WallMS, PeakRSSKib: report.PeakRSSKib,
		},
	}
	receipt.Original.InvocationID = invocationDigest(receipt.Binding, receipt.Original)
	var err error
	receipt.ReceiptID, err = publictestreuse.ReceiptContentDigest(receipt)
	if err != nil {
		return err
	}
	if err := publictestreuse.WriteReceipt(filepath.Join(output, "reuse-receipt.json"), receipt); err != nil {
		return err
	}
	report.ReceiptDigest = receipt.ReceiptID
	report.OriginalReceiptID = receipt.ReceiptID
	report.OriginalInvocationID = receipt.Original.InvocationID
	return nil
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
	observed, receipt, err := observeReuse(input, identity.Binding)
	if err != nil {
		return err
	}
	report := baseReport(identity, publictestreuse.ReplayOperation)
	assessment, err := assessReplayReport(identity.Policy, observed, &report)
	if err != nil {
		return err
	}
	if input.Authorize {
		report.ReceiptMisses = 1
	}
	if assessment.Reuse {
		if err := authorizeObservedReuse(input, identity, receipt); err != nil {
			return err
		}
		report.ReceiptDigest = receipt.ReceiptID
		report.OriginalReceiptID = receipt.ReceiptID
		report.OriginalInvocationID = receipt.Original.InvocationID
		report.ReusedTestExecutions = 1
		report.ReceiptHits = 1
		report.ReceiptMisses = 0
		report.GeneratedBytesEqual = true
		report.SemanticEqual = true
	}
	report.WallMS = elapsedMS(started)
	report.PeakRSSSubject = publictestreuse.PeakRSSObserver
	report.PeakRSSKib, err = toolchaincli.CurrentPeakRSSKiB()
	if err != nil {
		return fmt.Errorf("observe reuse process peak RSS: %w", err)
	}
	return writeReport(input.OutputDir, report, "Gooo classified the retained observations; any proposed reuse also required the current package and exact successful receipt.")
}

func loadIdentity(input executionInput) (identity, error) {
	return loadIdentityWithInputs(input, readExecutionPackageInputs)
}

// Keep both file observations distinct after checking their exact source binding.
func readCanonicalPolicy(input executionInput) (publictestreuse.Policy, []byte, []byte, error) {
	policySource, err := readRegular(input.Policy)
	if err != nil {
		return publictestreuse.Policy{}, nil, nil, fmt.Errorf("read policy: %w", err)
	}
	source, err := readRegular(input.Source)
	if err != nil {
		return publictestreuse.Policy{}, nil, nil, fmt.Errorf("read canonical source: %w", err)
	}
	if !bytes.Equal(policySource, source) {
		return publictestreuse.Policy{}, nil, nil, errors.New("policy and canonical source must be the same frozen .gooo file")
	}
	policy, err := publictestreuse.Load(input.Policy, policySource)
	if err != nil {
		return publictestreuse.Policy{}, nil, nil, err
	}
	return policy, policySource, source, nil
}

func loadIdentityWithInputs(input executionInput,
	readInputs func(string, []byte, []byte) (publictestreuse.ExecutionPackageInputs, error)) (identity, error) {
	for name, path := range map[string]string{
		"policy": input.Policy, "source": input.Source, "program": input.Program,
		"manifest": input.Manifest, "test contract": input.TestContract, "package directory": input.PackageDir,
	} {
		if path == "" {
			return identity{}, fmt.Errorf("%s path is required", name)
		}
	}
	policy, policySource, source, err := readCanonicalPolicy(input)
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
	packageInputs, err := readInputs(input.PackageDir, program, testContract)
	if err != nil {
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
	command, err := toolchaincli.CaptureGeneratedTestGo()
	if err != nil {
		return identity{}, fmt.Errorf("observe selected Go command: %w", err)
	}
	commandDigest := publictestreuse.TestCommandDigest(testCommand, cache.HashBytes(testContract).String())
	toolInputs := publictestreuse.ExecutionToolInputs(command.Inputs)
	return identity{
		Policy: policy, PolicyBytes: policySource, ProgramBytes: program,
		ManifestBytes: manifest, TestBytes: testContract, GoCommand: command,
		Binding: publictestreuse.Binding{
			PolicySourceDigest: policy.SourceDigest, PolicySemanticDigest: policy.SemanticDigest, PolicyEvaluatorDigest: policy.EvaluatorDigest,
			CanonicalSourceDigest: cache.HashBytes(source).String(), CanonicalSemanticDigest: policy.SemanticDigest,
			GeneratedOutputDigest: cache.HashBytes(program).String(), GeneratedSemanticDigest: generatedSemantic,
			GeneratedManifestDigest: cache.HashBytes(manifest).String(), CompilerDigest: compilerDigest,
			ReleasedToolDigest: publictestreuse.ReleasedToolDigest(compilerDigest), ToolchainDigest: toolInputs.Digest(),
			ToolchainVersion: toolInputs.Version, ToolInputs: toolInputs, TestCommand: testCommand, TestCommandDigest: commandDigest,
			TestContractDigest: cache.HashBytes(testContract).String(), PackageInputs: packageInputs,
		},
	}, nil
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

func runTimed(tool toolchaincli.GeneratedTestGoCommand, packageDir, label string, args []string) (commandResult, error) {
	if err := tool.VerifyBytes(); err != nil {
		return commandResult{}, err
	}
	started := time.Now()
	command := exec.Command(tool.Inputs.Path, args...)
	command.Dir = packageDir
	command.Env = append([]string(nil), tool.Env...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err != nil && command.ProcessState == nil {
		return commandResult{}, fmt.Errorf("start generated command %s: %w", label, err)
	}
	exitCode := 0
	if err != nil {
		exitCode = 1
		if exitError, ok := errors.AsType[*exec.ExitError](err); ok {
			exitCode = exitError.ExitCode()
		}
	}
	peak, readErr := toolchaincli.PeakRSSKiB(command.ProcessState)
	if readErr != nil {
		return commandResult{}, fmt.Errorf("read %s peak RSS: %w", label, readErr)
	}
	resultDigest := cache.HashBytes([]byte(strings.Join(append([]string{"go", label}, args...), "\x00") + "\x00" + stdout.String() + "\x00" + stderr.String() + "\x00" + strconv.Itoa(exitCode))).String()
	return commandResult{Success: err == nil, ExitCode: exitCode, ResultDigest: resultDigest, WallMS: elapsedMS(started), PeakRSSKib: peak}, nil
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
	human := fmt.Sprintf("# Public generated-test reuse\n\nDecision: `%s`\nReason: `%s`\nCase: `%s`\n\nTest executions: `%d`\nReused test executions: `%d`\nReceipt hits/misses: `%d/%d`\nBuild executions: `%d`\nBuild/test ms: `%d/%d`\nWall ms / peak RSS KiB: `%d/%d`\nPeak RSS subject: `%s`\nGenerated program bytes / lines: `%d/%d`\nTest contract bytes / lines: `%d/%d`\nRepository writes / local test executions: `%d/%d`\nReceipt: `%s`\nOriginal invocation: `%s`\n\n%s\n", report.Decision, report.Reason, report.CaseID, report.TestExecutions, report.ReusedTestExecutions, report.ReceiptHits, report.ReceiptMisses, report.BuildExecutions, report.BuildMS, report.TestMS, report.WallMS, report.PeakRSSKib, report.PeakRSSSubject, report.GeneratedProgramBytes, report.GeneratedProgramLines, report.TestContractBytes, report.TestContractLines, report.RepositoryWrites, report.LocalTestExecutions, report.ReceiptDigest, report.OriginalInvocationID, note)
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
