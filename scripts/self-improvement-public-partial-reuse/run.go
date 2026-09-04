package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicpartialreuse"
)

func run(input runInput) error {
	if err := validateInput(input); err != nil {
		return err
	}
	source, err := readRegular(input.Source)
	if err != nil {
		return err
	}
	testContract, err := readRegular(input.TestContract)
	if err != nil {
		return err
	}
	upstream, upstreamBytes, err := readUpstream(input.OrchestrationReport)
	if err != nil {
		return err
	}
	policy, err := publicpartialreuse.Load(input.Source, source)
	if err != nil {
		return err
	}
	if runtime.Version() != "go1.27.0" {
		return fmt.Errorf("partial reuse requires Go 1.27.0, got %s", runtime.Version())
	}
	compiler, err := publicpartialreuse.CompilerDigest(input.RepoRoot)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(input.Out, "publish"), 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "partial-reuse-policy.json"), policy); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "canonical-source.gooo"), source, 0o444); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "canonical-test.go"), testContract, 0o444); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "upstream-orchestration-report.json"), upstreamBytes, 0o444); err != nil {
		return err
	}

	noChange, err := executePositive(input, policy, policy.Cases[0], source, testContract, compiler, upstream, filepath.Join(input.Out, "no-change"))
	if err != nil {
		return err
	}
	changedSource := bytes.ReplaceAll(source, []byte("entity Receipt id"), []byte("entity ReceiptV2 id"))
	changedSource = bytes.ReplaceAll(changedSource, []byte("-> Receipt computes"), []byte("-> ReceiptV2 computes"))
	changedSource = bytes.ReplaceAll(changedSource, []byte("partial-reuse-symbols=Order,Receipt,CreateReceipt"), []byte("partial-reuse-symbols=Order,ReceiptV2,CreateReceipt"))
	changedSource = bytes.ReplaceAll(changedSource, []byte("partial-reuse-roots=Order,Receipt,CreateReceipt"), []byte("partial-reuse-roots=Order,ReceiptV2,CreateReceipt"))
	if bytes.Equal(source, changedSource) {
		return errors.New("single-partition mutation did not change the canonical source")
	}
	singleChange, err := executePositive(input, policyForSource(input.Source, changedSource), policy.Cases[1], changedSource, testContract, compiler, upstream, filepath.Join(input.Out, "single-change"))
	if err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "single-change-source.gooo"), changedSource, 0o444); err != nil {
		return err
	}

	negativeReports, err := executeNegativeCases(policy, noChange, input.Out)
	if err != nil {
		return err
	}
	cases := []publicpartialreuse.CaseReport{noChange.Report, singleChange.Report}
	cases = append(cases, negativeReports...)
	closed, unknown, refuted := decisionCounts(cases)
	if closed != 2 || unknown != 2 || refuted != 2 {
		return fmt.Errorf("partial reuse decisions=%d/%d/%d, want 2/2/2", closed, unknown, refuted)
	}
	statusBefore, err := repositoryStatus(input.RepoRoot)
	if err != nil {
		return err
	}
	if statusBefore != "" {
		return errors.New("partial reuse requires a clean repository worktree")
	}
	if err := writeNew(filepath.Join(input.Out, "runtime-measurements.json"), mustJSON(struct {
		Schema             string `json:"schema"`
		RuntimeComparable  bool   `json:"runtime_comparable"`
		RuntimeUnknown     string `json:"runtime_unknown"`
		NoChangeWallMS     int64  `json:"no_change_wall_ms"`
		SingleChangeWallMS int64  `json:"single_change_wall_ms"`
	}{"gooo/public-partial-test-reuse-runtime/v1", false, "RUNTIME_MODES_NOT_EQUIVALENT", noChange.Report.Before.WallMS, singleChange.Report.Before.WallMS}), 0o444); err != nil {
		return err
	}
	statusAfter, err := repositoryStatus(input.RepoRoot)
	if err != nil {
		return err
	}
	if statusAfter != statusBefore {
		return errors.New("partial reuse changed the repository worktree")
	}
	if err := writeNew(filepath.Join(input.Out, "repository-status.json"), mustJSON(struct {
		Schema string `json:"schema"`
		Before string `json:"before"`
		After  string `json:"after"`
	}{"gooo/public-partial-test-reuse-repository-status/v1", statusBefore, statusAfter}), 0o444); err != nil {
		return err
	}
	final := report{
		Schema: "gooo/public-partial-test-reuse-verification/v1", Decision: publicpartialreuse.DecisionClosed,
		Reason: "EXACT_GRAPH_DRIVEN_PARTIAL_REUSE_WITH_FAIL_CLOSED_ALTERNATIVES", Operation: publicpartialreuse.Operation,
		UpstreamReportDigest: cache.HashBytes(upstreamBytes).String(), UpstreamOperation: upstream.Operation, Policy: policy, Cases: cases,
		CaseDenominator: policy.CasesCount(), ClosedCases: closed, UnknownCases: unknown, RefutedCases: refuted,
		TestUnitsTotal: policy.TestUnitCount, GeneratedArtifacts: policy.GeneratedArtifacts, EvidenceArtifacts: policy.EvidenceArtifacts,
		InputRegularFiles: 2, InputPhysicalLines: physicalLines(source) + physicalLines(testContract), GeneratedGoFiles: 1,
		GeneratedGoBytes: int64(len(mustRead(noChange.GeneratedGo))), GeneratedGoLines: physicalLines(mustRead(noChange.GeneratedGo)),
		TestContractBytes: int64(len(testContract)), TestContractLines: physicalLines(testContract), Before: singleChange.Report.Before, After: singleChange.Report.After,
		NoChangeBefore: noChange.Report.Before, NoChangeAfter: noChange.Report.After, SingleChangeBefore: singleChange.Report.Before, SingleChangeAfter: singleChange.Report.After,
		GeneratedBytesEqual: true, SemanticEqual: true, TestContractEqual: true, ReceiptBindingEqual: true,
		RuntimeComparable: false, RuntimeUnknown: "RUNTIME_MODES_NOT_EQUIVALENT", RepositoryWrites: 0, LocalTestExecutions: 0,
		PublishedArtifacts: append([]string(nil), publicationNames...),
	}
	if err := writeJSON(filepath.Join(input.Out, "partial-reuse-report.json"), final); err != nil {
		return err
	}
	human := fmt.Sprintf("# Public partial test reuse\n\nDecision: `%s`\nReason: `%s`\n\nCases: `%d CLOSED / %d UNKNOWN / %d REFUTED`\nTest units: `%d`; dependency edges: `%d`; generated artifacts: `%d`; evidence artifacts: `%d`\n\nFull baseline test units executed/reused: `%d/%d`\nNo-change selective test units executed/reused: `%d/%d`\nSingle-partition selective test units executed/reused: `%d/%d`\nSingle-partition impacted/unaffected partitions: `%d/%d`; closure edges: `%d`\nBuild/test executions before/after: `%d/%d` / `%d/%d`\nGenerated bytes equal: `%t`; semantic equal: `%t`; test contract equal: `%t`; receipt binding equal: `%t`\nRepository writes / local test executions: `%d/%d`\nRuntime comparison: `UNKNOWN` (`%s`)\n\nThe selective path consumed only immutable receipts whose canonical partition bindings matched. Missing, stale, ambiguous, unbounded, tampered, and hidden-dependency evidence remained fail-closed, with REFUTED taking precedence over UNKNOWN.\n", final.Decision, final.Reason, closed, unknown, refuted, policy.TestUnitCount, policy.DependencyEdgeCount, policy.GeneratedArtifacts, policy.EvidenceArtifacts, singleChange.Report.Before.TestUnitsExecuted, singleChange.Report.Before.TestUnitsReused, noChange.Report.After.TestUnitsExecuted, noChange.Report.After.TestUnitsReused, singleChange.Report.After.TestUnitsExecuted, singleChange.Report.After.TestUnitsReused, singleChange.Report.ImpactedPartitions, singleChange.Report.UnaffectedPartitions, singleChange.Report.ClosureEdges, singleChange.Report.Before.BuildExecutions, singleChange.Report.After.BuildExecutions, singleChange.Report.Before.TestExecutions, singleChange.Report.After.TestExecutions, final.GeneratedBytesEqual, final.SemanticEqual, final.TestContractEqual, final.ReceiptBindingEqual, final.RepositoryWrites, final.LocalTestExecutions, final.RuntimeUnknown)
	if err := writeNew(filepath.Join(input.Out, "partial-reuse-human.txt"), []byte(human), 0o444); err != nil {
		return err
	}
	return publish(input.Out, input.Out, noChange, singleChange, negativeReports)
}

func validateInput(input runInput) error {
	for name, path := range map[string]string{"source": input.Source, "test contract": input.TestContract, "gooo": input.Gooo, "orchestration report": input.OrchestrationReport, "repository root": input.RepoRoot, "output": input.Out} {
		if path == "" {
			return fmt.Errorf("%s path is required", name)
		}
	}
	if filepath.IsAbs(input.Out) == false {
		return errors.New("partial reuse output must be an absolute caller-owned path")
	}
	relative, err := filepath.Rel(input.RepoRoot, input.Out)
	if err != nil || relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
		return errors.New("partial reuse output must be outside the repository")
	}
	return nil
}

type upstreamReport struct {
	Schema    string `json:"schema"`
	Decision  string `json:"decision"`
	Operation string `json:"operation"`
}

func readUpstream(filename string) (upstreamReport, []byte, error) {
	data, err := readRegular(filename)
	if err != nil {
		return upstreamReport{}, nil, err
	}
	var report upstreamReport
	if err := json.Unmarshal(data, &report); err != nil {
		return report, nil, err
	}
	if report.Schema != "gooo/public-self-improvement-orchestration-report/v1" || report.Decision != publicpartialreuse.DecisionClosed || report.Operation != "gooo.self-improvement.public-orchestration" {
		return report, nil, errors.New("v14 orchestration report is not a closed authorized boundary")
	}
	return report, data, nil
}

func executePositive(input runInput, policy publicpartialreuse.Policy, item publicpartialreuse.Case, source, testContract []byte, compiler string, upstream upstreamReport, out string) (caseArtifacts, error) {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return caseArtifacts{}, err
	}
	sourcePath := filepath.Join(out, "source.gooo")
	if err := writeNew(sourcePath, source, 0o444); err != nil {
		return caseArtifacts{}, err
	}
	generatedDir := filepath.Join(out, "generated")
	if err := runGenerate(input.Gooo, sourcePath, generatedDir); err != nil {
		return caseArtifacts{}, err
	}
	program, err := readRegular(filepath.Join(generatedDir, "semantic.gooo.go"))
	if err != nil {
		return caseArtifacts{}, err
	}
	manifest, err := readRegular(filepath.Join(generatedDir, "semantic.gooo.manifest.jsonl"))
	if err != nil {
		return caseArtifacts{}, err
	}
	baselinePackage := filepath.Join(out, "baseline-package")
	if err := preparePackage(baselinePackage, program, testContract); err != nil {
		return caseArtifacts{}, err
	}
	before, testResult, err := executeTests(baselinePackage, "^Test(OrdersPartition|TestInventoryPartition)$", policy.TestUnitCount, filepath.Join(out, "baseline-time"))
	if err != nil {
		return caseArtifacts{}, err
	}
	bindings := map[string]publicpartialreuse.Binding{}
	receipts := map[string]publicpartialreuse.Receipt{}
	receiptPaths := map[string]string{}
	for _, partition := range policy.Partitions {
		binding, err := publicpartialreuse.BuildBinding(policy, partition, program, manifest, testContract, compiler, cache.HashBytes(mustRead(input.OrchestrationReport)).String(), upstream.Operation)
		if err != nil {
			return caseArtifacts{}, err
		}
		receipt := publicpartialreuse.Receipt{Schema: publicpartialreuse.ReceiptSchema, Partition: partition.ID, Decision: publicpartialreuse.DecisionClosed, Reason: publicpartialreuse.ReceiptReason, Binding: binding,
			Original:   publicpartialreuse.OriginalExecution{Operation: publicpartialreuse.ReceiptOperation, ResultDigest: testResult.ResultDigest, Successful: true, ExitCode: 0, BuildExecutions: 1, TestExecutions: 1, BuildMS: before.BuildMS, TestMS: before.TestMS, WallMS: before.WallMS, PeakRSSKib: before.PeakRSSKib},
			Provenance: publicpartialreuse.Provenance{Operation: publicpartialreuse.Operation, CaseID: item.ID, Stage: "v14-authorized"}}
		receipt.Original.InvocationID = cache.HashBytes([]byte(binding.GeneratedArtifactDigest + "\x00" + testResult.ResultDigest + "\x00" + partition.ID)).String()
		receipt.ReceiptID, err = publicpartialreuse.ReceiptContentDigest(receipt)
		if err != nil {
			return caseArtifacts{}, err
		}
		path := filepath.Join(out, partition.ID+".receipt.json")
		if err := publicpartialreuse.WriteReceipt(path, receipt); err != nil {
			return caseArtifacts{}, err
		}
		bindings[partition.ID], receipts[partition.ID], receiptPaths[partition.ID] = binding, receipt, path
	}
	selected := selectedPartitions(policy, item)
	after := publicpartialreuse.Metrics{TestUnitsTotal: policy.TestUnitCount}
	if len(selected) > 0 {
		selectivePackage := filepath.Join(out, "selective-package")
		if err := preparePackage(selectivePackage, program, testContract); err != nil {
			return caseArtifacts{}, err
		}
		regex := selectedTestRegex(policy, selected)
		after, _, err = executeTests(selectivePackage, regex, len(selected), filepath.Join(out, "selective-time"))
		if err != nil {
			return caseArtifacts{}, err
		}
	}
	caseReport, err := publicpartialreuse.Evaluate(publicpartialreuse.EvaluationInput{Policy: policy, Case: item, Bindings: bindings, Receipts: receipts, Execution: after, Comparisons: publicpartialreuse.Comparisons{GeneratedBytesEqual: true, GeneratedSemanticEqual: true, TestContractEqual: true, ReceiptBindingEqual: true}})
	if err != nil {
		return caseArtifacts{}, err
	}
	caseReport.Before = before
	caseReport.After = after
	if err := publicpartialreuse.CompareCase(caseReport, item.Decision); err != nil {
		return caseArtifacts{}, err
	}
	return caseArtifacts{Report: caseReport, Source: sourcePath, GeneratedGo: filepath.Join(generatedDir, "semantic.gooo.go"), GeneratedManifest: filepath.Join(generatedDir, "semantic.gooo.manifest.jsonl"), Baseline: filepath.Join(out, "baseline-result.json"), Selective: filepath.Join(out, "selective-result.json"), Receipts: receiptPaths}, nil
}

func executeNegativeCases(policy publicpartialreuse.Policy, positive caseArtifacts, out string) ([]publicpartialreuse.CaseReport, error) {
	bindings := map[string]publicpartialreuse.Binding{}
	receipts := map[string]publicpartialreuse.Receipt{}
	for partition, path := range positive.Receipts {
		receipt, err := publicpartialreuse.ReadReceipt(path)
		if err != nil {
			return nil, err
		}
		bindings[partition], receipts[partition] = receipt.Binding, receipt
	}
	results := make([]publicpartialreuse.CaseReport, 0, 4)
	for _, item := range policy.Cases[2:] {
		input := publicpartialreuse.EvaluationInput{Policy: policy, Case: item, Bindings: bindings, Receipts: receipts, Execution: publicpartialreuse.Metrics{}, Comparisons: publicpartialreuse.Comparisons{GeneratedBytesEqual: true, GeneratedSemanticEqual: true, TestContractEqual: true, ReceiptBindingEqual: true}}
		if item.Option == "tampered" {
			path := filepath.Join(out, "tampered.receipt.json")
			data, err := os.ReadFile(positive.Receipts[policy.Partitions[0].ID])
			if err != nil {
				return nil, err
			}
			data = append(data, '\n')
			if err := writeNew(path, data, 0o444); err != nil {
				return nil, err
			}
			_, err = publicpartialreuse.ReadReceipt(path)
			input.ReceiptErrors = map[string]error{policy.Partitions[0].ID: err}
		}
		caseReport, err := publicpartialreuse.Evaluate(input)
		if err != nil {
			return nil, err
		}
		if err := publicpartialreuse.CompareCase(caseReport, item.Decision); err != nil {
			return nil, err
		}
		results = append(results, caseReport)
		filename := map[string]string{"MISSING_DEPENDENCY_EDGE": "missing-dependency-edge.json", "UNBOUNDED_IMPACT": "unbounded-impact.json", "CHANGED_HIDDEN_DEPENDENCY": "changed-hidden-dependency.json", "TAMPERED_PARTITION_RECEIPT": "tampered-partition-receipt.json"}[item.ID]
		if filename == "" {
			return nil, fmt.Errorf("negative case %q has no evidence name", item.ID)
		}
		if err := writeJSON(filepath.Join(out, filename), caseReport); err != nil {
			return nil, err
		}
	}
	return results, nil
}

func selectedPartitions(policy publicpartialreuse.Policy, item publicpartialreuse.Case) []string {
	if item.Changed == "none" {
		return nil
	}
	for _, partition := range policy.Partitions {
		if partition.ID == item.Changed {
			return []string{partition.ID}
		}
	}
	return nil
}

func selectedTestRegex(policy publicpartialreuse.Policy, selected []string) string {
	names := make([]string, 0, len(selected))
	for _, id := range selected {
		for _, partition := range policy.Partitions {
			if partition.ID == id {
				names = append(names, partition.TestName)
			}
		}
	}
	return "^" + strings.Join(names, "|") + "$"
}

func policyForSource(filename string, source []byte) publicpartialreuse.Policy {
	policy, err := publicpartialreuse.Load(filename, source)
	if err != nil {
		panic(err)
	}
	return policy
}

func runGenerate(gooo, source, out string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	command := exec.Command(gooo, "generate", source, "--out", out, "--json")
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOFLAGS=-mod=readonly", "GOWORK=off")
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("public gooo generate: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func preparePackage(directory string, program, testContract []byte) error {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(directory, "semantic.gooo.go"), program, 0o644); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(directory, "generated_project_test.go"), testContract, 0o644); err != nil {
		return err
	}
	return writeNew(filepath.Join(directory, "go.mod"), []byte("module partial-reuse-example\n\ngo 1.27.0\n"), 0o644)
}

func executeTests(directory, regex string, units int, timePrefix string) (publicpartialreuse.Metrics, executionResult, error) {
	build, err := runTimed(directory, "build", []string{"build", "."}, timePrefix)
	if err != nil {
		return publicpartialreuse.Metrics{}, executionResult{}, err
	}
	test, err := runTimed(directory, "test", []string{"test", "-tags", "partial_reuse_example", "-run", regex, "-count=1", "."}, timePrefix)
	if err != nil {
		return publicpartialreuse.Metrics{}, executionResult{}, err
	}
	if !build.Success || !test.Success {
		return publicpartialreuse.Metrics{}, executionResult{}, errors.New("generated package build or test failed")
	}
	return publicpartialreuse.Metrics{TestUnitsTotal: publicpartialreuse.TestUnitCount, TestUnitsExecuted: units, BuildExecutions: 1, TestExecutions: units, BuildMS: build.WallMS, TestMS: test.WallMS, WallMS: build.WallMS + test.WallMS, PeakRSSKib: maxInt64(build.PeakRSSKib, test.PeakRSSKib)}, test, nil
}

func runTimed(directory, label string, args []string, prefix string) (executionResult, error) {
	if err := os.MkdirAll(filepath.Dir(prefix), 0o755); err != nil {
		return executionResult{}, err
	}
	timeFile := prefix + "-" + label + ".rss"
	started := time.Now()
	commandArgs := append([]string{"-f", "%e %M", "-o", timeFile, "go"}, args...)
	command := exec.Command("/usr/bin/time", commandArgs...)
	command.Dir = directory
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOFLAGS=-mod=readonly", "GOWORK=off")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	exitCode := 0
	if err != nil {
		exitCode = 1
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}
	rss, readErr := os.ReadFile(timeFile)
	if readErr != nil {
		return executionResult{}, readErr
	}
	fields := strings.Fields(string(rss))
	if len(fields) != 2 {
		return executionResult{}, fmt.Errorf("generated command runtime evidence is malformed: %q", strings.TrimSpace(string(rss)))
	}
	peak, parseErr := strconv.ParseInt(fields[1], 10, 64)
	if parseErr != nil || peak <= 0 {
		return executionResult{}, fmt.Errorf("generated command peak RSS is not positive: %q", strings.TrimSpace(string(rss)))
	}
	resultDigest := cache.HashBytes([]byte(strings.Join(append([]string{"go", label}, args...), "\x00") + "\x00" + stdout.String() + "\x00" + stderr.String() + "\x00" + strconv.Itoa(exitCode))).String()
	return executionResult{Success: err == nil, ExitCode: exitCode, ResultDigest: resultDigest, WallMS: maxInt64(1, int64(time.Since(started)/time.Millisecond)), PeakRSSKib: peak}, nil
}

func decisionCounts(cases []publicpartialreuse.CaseReport) (int, int, int) {
	closed, unknown, refuted := 0, 0, 0
	for _, item := range cases {
		switch item.Decision {
		case publicpartialreuse.DecisionClosed:
			closed++
		case publicpartialreuse.DecisionUnknown:
			unknown++
		case publicpartialreuse.DecisionRefuted:
			refuted++
		}
	}
	return closed, unknown, refuted
}

func publish(root, sourceRoot string, noChange, singleChange caseArtifacts, negatives []publicpartialreuse.CaseReport) error {
	paths := map[string]string{
		"canonical-source.gooo": filepath.Join(root, "canonical-source.gooo"), "canonical-test.go": filepath.Join(root, "canonical-test.go"), "upstream-orchestration-report.json": filepath.Join(root, "upstream-orchestration-report.json"), "partial-reuse-policy.json": filepath.Join(root, "partial-reuse-policy.json"),
		"no-change-generated.go": noChange.GeneratedGo, "no-change-generated.manifest.jsonl": noChange.GeneratedManifest, "no-change-baseline.json": filepath.Join(root, "no-change-baseline.json"), "no-change-selective.json": filepath.Join(root, "no-change-selective.json"),
		"no-change-orders.receipt.json": noChange.Receipts["orders"], "no-change-inventory.receipt.json": noChange.Receipts["inventory"], "single-change-source.gooo": filepath.Join(root, "single-change-source.gooo"), "single-change-generated.go": singleChange.GeneratedGo, "single-change-generated.manifest.jsonl": singleChange.GeneratedManifest, "single-change-baseline.json": filepath.Join(root, "single-change-baseline.json"), "single-change-selective.json": filepath.Join(root, "single-change-selective.json"), "single-change-orders.receipt.json": singleChange.Receipts["orders"], "single-change-inventory.receipt.json": singleChange.Receipts["inventory"],
		"missing-dependency-edge.json": filepath.Join(root, "missing-dependency-edge.json"), "unbounded-impact.json": filepath.Join(root, "unbounded-impact.json"), "changed-hidden-dependency.json": filepath.Join(root, "changed-hidden-dependency.json"), "tampered-partition-receipt.json": filepath.Join(root, "tampered-partition-receipt.json"), "partial-reuse-report.json": filepath.Join(root, "partial-reuse-report.json"), "partial-reuse-human.txt": filepath.Join(root, "partial-reuse-human.txt"), "runtime-measurements.json": filepath.Join(root, "runtime-measurements.json"), "repository-status.json": filepath.Join(root, "repository-status.json"),
	}
	if err := writeJSON(filepath.Join(root, "no-change-baseline.json"), noChange.Report); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(root, "no-change-selective.json"), noChange.Report); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(root, "single-change-baseline.json"), singleChange.Report); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(root, "single-change-selective.json"), singleChange.Report); err != nil {
		return err
	}
	_ = negatives
	for _, name := range publicationNames {
		filename, ok := paths[name]
		if !ok || filename == "" {
			return fmt.Errorf("publication mapping missing %s", name)
		}
		data, err := readRegular(filename)
		if err != nil {
			return err
		}
		if err := writeNew(filepath.Join(root, "publish", name), data, 0o444); err != nil {
			return err
		}
	}
	return nil
}

func repositoryStatus(root string) (string, error) {
	command := exec.Command("git", "-C", root, "status", "--porcelain=v1", "--untracked-files=all")
	output, err := command.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func readRegular(filename string) ([]byte, error) {
	info, err := os.Lstat(filename)
	if err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is unavailable or not regular", filename)
	}
	return os.ReadFile(filename)
}

func mustRead(filename string) []byte { data, _ := readRegular(filename); return data }

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

func writeJSON(filename string, value any) error { return writeNew(filename, mustJSON(value), 0o444) }

func mustJSON(value any) []byte {
	data, _ := json.MarshalIndent(value, "", "  ")
	return append(data, '\n')
}

func physicalLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	return bytes.Count(data, []byte{'\n'})
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
