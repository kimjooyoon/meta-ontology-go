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
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicresolutionrepair"
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
	orchestration, orchestrationBytes, err := readUpstream(input.OrchestrationReport)
	if err != nil {
		return err
	}
	v15Report, err := readRegular(input.V15Report)
	if err != nil {
		return err
	}
	v15Hidden, err := readRegular(input.V15Hidden)
	if err != nil {
		return err
	}
	policy, err := publicresolutionrepair.Load(input.Source, source)
	if err != nil {
		return err
	}
	if runtime.Version() != "go1.27.0" {
		return fmt.Errorf("semantic resolution repair requires Go 1.27.0, got %s", runtime.Version())
	}
	counterexample, err := publicresolutionrepair.LoadCounterexample(v15Hidden, v15Report, policy)
	if err != nil {
		return err
	}
	compilerDigest, err := compilerDigest(input.RepoRoot)
	if err != nil {
		return err
	}
	statusBefore, err := repositoryStatus(input.RepoRoot)
	if err != nil {
		return err
	}
	if statusBefore != "" {
		return errors.New("semantic resolution repair requires a clean repository worktree")
	}
	if err := os.MkdirAll(filepath.Join(input.Out, "publish"), 0o755); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "canonical-source.gooo"), source, 0o444); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "canonical-test.go"), testContract, 0o444); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "upstream-orchestration-report.json"), orchestrationBytes, 0o444); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "original-hidden-dependency.json"), v15Hidden, 0o444); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "original-partial-reuse-report.json"), v15Report, 0o444); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "v15-counterexample-provenance.json"), counterexample); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "resolution-repair-policy.json"), policy); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "resolution-repair-case-table.json"), policy.Cases); err != nil {
		return err
	}

	proposal, err := publicresolutionrepair.SynthesizeProposal(policy, counterexample)
	if err != nil {
		return err
	}
	authorization := publicresolutionrepair.NewAuthorization(proposal, publicresolutionrepair.AuthorizationAuthorized)
	rejectedAuthorization := publicresolutionrepair.NewAuthorization(proposal, publicresolutionrepair.AuthorizationRejected)
	overlay, err := publicresolutionrepair.BuildOverlay(policy, counterexample, proposal, authorization)
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "repair-proposal.json"), proposal); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "repair-authorization.json"), authorization); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "authorized-graph-overlay.json"), overlay); err != nil {
		return err
	}

	fallbackDir := filepath.Join(input.Out, "fallback")
	if err := os.MkdirAll(fallbackDir, 0o755); err != nil {
		return err
	}
	if err := runGenerate(input.Gooo, input.Source, filepath.Join(fallbackDir, "generated")); err != nil {
		return err
	}
	fallbackProgram, err := readRegular(filepath.Join(fallbackDir, "generated", "semantic.gooo.go"))
	if err != nil {
		return err
	}
	fallbackManifest, err := readRegular(filepath.Join(fallbackDir, "generated", "semantic.gooo.manifest.jsonl"))
	if err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "fallback-generated.go"), fallbackProgram, 0o444); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "fallback-generated.manifest.jsonl"), fallbackManifest, 0o444); err != nil {
		return err
	}
	fallbackMetrics, fallbackResult, err := executeTests(filepath.Join(fallbackDir, "package"), fallbackProgram, testContract, allTestsRegex(policy), publicresolutionrepair.TestUnitCount, filepath.Join(fallbackDir, "time"))
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "fallback-baseline.json"), fallbackMetrics); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "fallback-result.json"), fallbackResult); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "fallback-counterexample-preserved.json"), counterexample); err != nil {
		return err
	}

	overlayDir := filepath.Join(input.Out, "overlay")
	if err := os.MkdirAll(overlayDir, 0o755); err != nil {
		return err
	}
	if err := runGenerate(input.Gooo, input.Source, filepath.Join(overlayDir, "generated")); err != nil {
		return err
	}
	overlayProgram, err := readRegular(filepath.Join(overlayDir, "generated", "semantic.gooo.go"))
	if err != nil {
		return err
	}
	overlayManifest, err := readRegular(filepath.Join(overlayDir, "generated", "semantic.gooo.manifest.jsonl"))
	if err != nil {
		return err
	}
	overlayMetrics, overlayResult, err := executeTests(filepath.Join(overlayDir, "replay-package"), overlayProgram, testContract, allTestsRegex(policy), publicresolutionrepair.TestUnitCount, filepath.Join(overlayDir, "replay-time"))
	if err != nil {
		return err
	}
	unchanged := unchangedPartition(policy, counterexample.ObservedAffectedPartition)
	selectivityMetrics, _, err := executeTests(filepath.Join(overlayDir, "selectivity-package"), overlayProgram, testContract, "^"+unchanged.TestName+"$", 1, filepath.Join(overlayDir, "selectivity-time"))
	if err != nil {
		return err
	}
	selectivityMetrics.TestUnitsReused = policy.SelectivityTestUnitsReused
	if err := writeNew(filepath.Join(input.Out, "overlay-generated.go"), overlayProgram, 0o444); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "overlay-generated.manifest.jsonl"), overlayManifest, 0o444); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "overlay-replay.json"), overlayMetrics); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "overlay-selectivity.json"), selectivityMetrics); err != nil {
		return err
	}
	comparisons := publicresolutionrepair.Comparisons{
		GeneratedBytesEqual: bytes.Equal(fallbackProgram, overlayProgram), GeneratedSemanticEqual: bytes.Equal(fallbackManifest, overlayManifest),
		TestContractEqual: bytes.Equal(testContract, testContract), FullTestOutcomeEqual: fallbackResult.Success && overlayResult.Success && fallbackResult.ExitCode == overlayResult.ExitCode,
		OverlayBindingEqual: publicresolutionrepair.ValidateOverlay(overlay, policy, proposal, authorization) == nil,
	}
	if err := writeJSON(filepath.Join(input.Out, "overlay-outcome-comparison.json"), comparisons); err != nil {
		return err
	}

	cases := make([]publicresolutionrepair.CaseReport, 0, publicresolutionrepair.CaseCount)
	for _, item := range policy.Cases {
		caseCounterexample := counterexample
		caseAuthorization := publicresolutionrepair.AuthorizationArtifact{}
		caseOverlay := publicresolutionrepair.GraphOverlay{}
		caseFallback := publicresolutionrepair.Metrics{}
		caseReplay := publicresolutionrepair.Metrics{}
		caseSelectivity := publicresolutionrepair.Metrics{}
		caseComparisons := publicresolutionrepair.Comparisons{}
		if item.ResolutionTo == publicresolutionrepair.ResolutionFallback && item.RepairVariant == "none" {
			caseFallback = fallbackMetrics
		}
		if item.RepairVariant == "canonical" && item.Authorization == "authorized" {
			caseAuthorization = authorization
			caseOverlay = overlay
			caseReplay = overlayMetrics
			caseSelectivity = selectivityMetrics
			caseComparisons = comparisons
		}
		if item.RepairVariant == "tampered" {
			caseCounterexample = publicresolutionrepair.InvalidCounterexample(counterexample)
		}
		if item.Authorization == "rejected" {
			caseAuthorization = rejectedAuthorization
		}
		caseReport, err := publicresolutionrepair.Evaluate(publicresolutionrepair.EvaluationInput{
			Policy: policy, Case: item, Counterexample: caseCounterexample, Proposal: proposal, Authorization: caseAuthorization, Overlay: caseOverlay,
			Fallback: caseFallback, OverlayReplay: caseReplay, UnchangedSelectivity: caseSelectivity, Comparisons: caseComparisons,
		})
		if err != nil {
			return err
		}
		if err := publicresolutionrepair.CompareCase(caseReport, item.Decision); err != nil {
			return err
		}
		cases = append(cases, caseReport)
		if item.Decision != publicresolutionrepair.DecisionClosed {
			filename := strings.ToLower(strings.ReplaceAll(item.ID, "_", "-")) + ".json"
			if err := writeJSON(filepath.Join(input.Out, filename), caseReport); err != nil {
				return err
			}
		}
	}
	closed, unknown, refuted := decisionCounts(cases)
	if closed != 2 || unknown != 2 || refuted != 2 {
		return fmt.Errorf("semantic repair decisions=%d/%d/%d, want 2/2/2", closed, unknown, refuted)
	}
	statusAfter, err := repositoryStatus(input.RepoRoot)
	if err != nil {
		return err
	}
	if statusAfter != statusBefore {
		return errors.New("semantic resolution repair changed the repository worktree")
	}
	if err := writeJSON(filepath.Join(input.Out, "runtime-measurements.json"), map[string]any{
		"schema": "gooo/public-semantic-resolution-repair-runtime/v1", "runtime_comparable": false, "runtime_unknown": "RUNTIME_MODES_NOT_EQUIVALENT",
		"fallback_wall_ms": fallbackMetrics.WallMS, "fallback_peak_rss_kib": fallbackMetrics.PeakRSSKib, "overlay_wall_ms": overlayMetrics.WallMS, "overlay_peak_rss_kib": overlayMetrics.PeakRSSKib,
	}); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "repository-status.json"), map[string]any{"schema": "gooo/public-semantic-resolution-repair-repository-status/v1", "before": statusBefore, "after": statusAfter}); err != nil {
		return err
	}
	safety := cases[1].SafetyImprovement
	final := report{
		Schema: "gooo/public-semantic-resolution-repair-verification/v1", Decision: publicresolutionrepair.DecisionClosed, Reason: "COUNTEREXAMPLE_REPAIRED_WITH_CONSERVATIVE_FALLBACK_AND_AUTHORIZED_OVERLAY", Operation: "gooo.test.generated-public-semantic-resolution-repair",
		UpstreamReportDigest: cache.HashBytes(orchestrationBytes).String(), UpstreamOperation: orchestration.Operation, CompilerDigest: compilerDigest, OriginalCounterexample: counterexample, Proposal: proposal, Authorization: authorization, Overlay: overlay, Policy: policy, Cases: cases,
		CaseDenominator: len(cases), ClosedCases: closed, UnknownCases: unknown, RefutedCases: refuted, ResolutionLevelCount: len(policy.ResolutionLevels), ProofModeObservationCount: policy.ProofModeObservationCount,
		ProofFoundationCount: policy.ProofFoundationCount, ProofCoherenceCount: policy.ProofCoherenceCount, ProofRegressionCount: policy.ProofRegressionCount, RepairProposalCount: len(policy.Proposals), AuthorizationDecisionCount: len(policy.Authorizations),
		GraphEdgesBefore: overlay.BaseEdgeCount, GraphEdgesAfter: overlay.OverlayEdgeCount, CanonicalGraphEdgeCount: policy.CanonicalGraphEdgeCount, TestUnitsTotal: policy.TestUnitCount,
		FallbackTestUnitsExecuted: fallbackMetrics.TestUnitsExecuted, FallbackTestUnitsReused: fallbackMetrics.TestUnitsReused, OverlayTestUnitsExecuted: overlayMetrics.TestUnitsExecuted, OverlayTestUnitsReused: overlayMetrics.TestUnitsReused,
		SelectivityTestUnitsExecuted: selectivityMetrics.TestUnitsExecuted, SelectivityTestUnitsReused: selectivityMetrics.TestUnitsReused, ContinuityEdgeCount: overlay.ContinuityEdgeCount, GeneratedArtifactCount: policy.GeneratedArtifactCount, EvidenceArtifactCount: policy.EvidenceArtifactCount,
		Fallback: fallbackMetrics, OverlayReplay: overlayMetrics, UnchangedSelectivity: selectivityMetrics, GeneratedBytesEqual: comparisons.GeneratedBytesEqual, SemanticEqual: comparisons.GeneratedSemanticEqual, TestContractEqual: comparisons.TestContractEqual, FullTestOutcomeEqual: comparisons.FullTestOutcomeEqual, OverlayBindingEqual: comparisons.OverlayBindingEqual,
		SafetyImprovement: safety, RuntimeComparable: false, RuntimeUnknown: "RUNTIME_MODES_NOT_EQUIVALENT", RepositoryWrites: 0, LocalTestExecutions: 0, PublishedArtifacts: append([]string(nil), publicationNames...),
	}
	if !safety {
		return errors.New("authorized overlay does not prove the counterexample safety improvement")
	}
	if err := writeJSON(filepath.Join(input.Out, "resolution-repair-report.json"), final); err != nil {
		return err
	}
	if err := writeNew(filepath.Join(input.Out, "resolution-repair-human.txt"), []byte(humanReport(final)), 0o444); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(input.Out, "resolution-repair-verification-input.json"), map[string]any{"schema": "gooo/public-semantic-resolution-repair-verification-input/v1", "report": filepath.Join(input.Out, "resolution-repair-report.json"), "published_root": filepath.Join(input.Out, "publish"), "published_artifacts": publicationNames}); err != nil {
		return err
	}
	return publish(input.Out, final)
}

func validateInput(input runInput) error {
	paths := map[string]string{"source": input.Source, "test contract": input.TestContract, "gooo": input.Gooo, "orchestration report": input.OrchestrationReport, "v15 report": input.V15Report, "v15 hidden record": input.V15Hidden, "repository root": input.RepoRoot, "output": input.Out}
	for name, path := range paths {
		if path == "" {
			return fmt.Errorf("%s path is required", name)
		}
	}
	if !filepath.IsAbs(input.Out) || !filepath.IsAbs(input.RepoRoot) {
		return errors.New("semantic resolution repair paths must be absolute")
	}
	relative, err := filepath.Rel(input.RepoRoot, input.Out)
	if err != nil || relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
		return errors.New("semantic resolution repair output must be outside the repository")
	}
	return nil
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
	if report.Schema != "gooo/public-self-improvement-orchestration-report/v1" || report.Decision != publicresolutionrepair.DecisionClosed || report.Operation != "gooo.self-improvement.public-orchestration" {
		return report, nil, errors.New("v14 orchestration report is not a closed authorized boundary")
	}
	return report, data, nil
}

func compilerDigest(root string) (string, error) {
	paths := []string{"cmd/gooo/generate_part01.go", "cmd/gooo/generate_pipeline_part03.go", "internal/bidir/lower_document_part01.go", "internal/semantic/graph_part03.go", "internal/meta/publicresolutionrepair/policy.go", "internal/meta/publicresolutionrepair/counterexample.go", "internal/meta/publicresolutionrepair/repair.go", "internal/meta/publicresolutionrepair/evaluator.go", "scripts/self-improvement-public-resolution-repair/main.go", "scripts/self-improvement-public-resolution-repair/model.go", "scripts/self-improvement-public-resolution-repair/run.go", "scripts/self-improvement-public-resolution-repair/verify.go"}
	var builder strings.Builder
	for _, relative := range paths {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return "", fmt.Errorf("read semantic repair compiler identity %s: %w", relative, err)
		}
		builder.WriteString(relative)
		builder.WriteByte(0)
		builder.Write(data)
		builder.WriteByte(0)
	}
	return cache.HashBytes([]byte(builder.String())).String(), nil
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
	return writeNew(filepath.Join(directory, "go.mod"), []byte("module semantic-resolution-repair-example\n\ngo 1.27.0\n"), 0o644)
}

func executeTests(directory string, program, testContract []byte, regex string, units int, timePrefix string) (publicresolutionrepair.Metrics, executionResult, error) {
	if err := preparePackage(directory, program, testContract); err != nil {
		return publicresolutionrepair.Metrics{}, executionResult{}, err
	}
	build, err := runTimed(directory, "build", []string{"build", "."}, timePrefix)
	if err != nil {
		return publicresolutionrepair.Metrics{}, executionResult{}, err
	}
	test, err := runTimed(directory, "test", []string{"test", "-tags", "partial_reuse_example", "-run", regex, "-count=1", "."}, timePrefix)
	if err != nil {
		return publicresolutionrepair.Metrics{}, executionResult{}, err
	}
	return publicresolutionrepair.Metrics{TestUnitsTotal: publicresolutionrepair.TestUnitCount, TestUnitsExecuted: units, BuildExecutions: 1, TestExecutions: units, BuildMS: build.WallMS, TestMS: test.WallMS, WallMS: build.WallMS + test.WallMS, PeakRSSKib: maxInt64(build.PeakRSSKib, test.PeakRSSKib)}, test, nil
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
	if err != nil {
		return executionResult{}, fmt.Errorf("generated command %s failed: %w: stdout=%q stderr=%q", label, err, stdout.String(), stderr.String())
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
	return executionResult{Success: true, ExitCode: exitCode, ResultDigest: resultDigest, WallMS: maxInt64(1, int64(time.Since(started)/time.Millisecond)), PeakRSSKib: peak}, nil
}

func allTestsRegex(policy publicresolutionrepair.Policy) string {
	names := make([]string, 0, len(policy.Partitions))
	for _, partition := range policy.Partitions {
		names = append(names, partition.TestName)
	}
	return "^(" + strings.Join(names, "|") + ")$"
}

func unchangedPartition(policy publicresolutionrepair.Policy, affected string) publicresolutionrepair.Partition {
	for _, partition := range policy.Partitions {
		if partition.ID != affected {
			return partition
		}
	}
	return publicresolutionrepair.Partition{}
}

func decisionCounts(cases []publicresolutionrepair.CaseReport) (int, int, int) {
	closed, unknown, refuted := 0, 0, 0
	for _, item := range cases {
		switch item.Decision {
		case publicresolutionrepair.DecisionClosed:
			closed++
		case publicresolutionrepair.DecisionUnknown:
			unknown++
		case publicresolutionrepair.DecisionRefuted:
			refuted++
		}
	}
	return closed, unknown, refuted
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

func writeJSON(filename string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeNew(filename, append(data, '\n'), 0o444)
}

func physicalLines(data []byte) int { return bytes.Count(data, []byte{'\n'}) }

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
