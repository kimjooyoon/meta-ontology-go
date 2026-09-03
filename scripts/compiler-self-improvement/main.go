package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	schema               = "gooo/compiler-self-improvement/v1"
	programRule          = "generator.normalizeIR:fold-copy-and-canonicalize/v1"
	targetPath           = "internal/generator/normalize_part01.go"
	fixturePath          = "examples/billing/main.gooo"
	graphSchema          = "gooo-graph/v1"
	primaryMetric        = "normalization_collection_copy_passes"
	decisionClosed       = "CLOSED"
	decisionUnknown      = "UNKNOWN"
	decisionRefuted      = "REFUTED"
	unknownReason        = "MISSING_EXACT_INTEGER_PAIR"
	unknownNextOperation = "CAPTURE_EXACT_BEFORE_AFTER"
)

type graph struct {
	SchemaVersion string      `json:"schema_version"`
	GraphHash     string      `json:"graph_hash"`
	SourceDigest  string      `json:"source_digest"`
	Nodes         []graphNode `json:"nodes"`
}

type graphNode struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type candidateEvidence struct {
	Schema                        string `json:"schema"`
	AuthorityProgram              string `json:"authority_program"`
	Target                        string `json:"target"`
	Fixture                       string `json:"fixture"`
	GraphSchema                   string `json:"graph_schema"`
	GraphHash                     string `json:"graph_hash"`
	GraphSourceDigest             string `json:"graph_source_digest"`
	BaselineSourceDigest          string `json:"baseline_source_digest"`
	CandidateSourceDigest         string `json:"candidate_source_digest"`
	BaselineCollectionCopyPasses  int    `json:"baseline_collection_copy_passes"`
	CandidateCollectionCopyPasses int    `json:"candidate_collection_copy_passes"`
	RepositoryWrites              int    `json:"repository_writes"`
	OutputMode                    string `json:"output_mode"`
}

type metrics struct {
	NormalizationCollectionCopyPasses int64 `json:"normalization_collection_copy_passes"`
	WallMS                            int64 `json:"wall_ms"`
	PeakRSSKiB                        int64 `json:"peak_rss_kib"`
	BuildMS                           int64 `json:"build_ms"`
	TestMS                            int64 `json:"test_ms"`
	ExecutedTests                     int64 `json:"executed_tests"`
	ReusedTests                       int64 `json:"reused_tests"`
}

type semanticGuard struct {
	BehaviorEqual      bool `json:"behavior_equal"`
	DeterministicEqual bool `json:"deterministic_equal"`
	MemoryGuardrail    bool `json:"memory_guardrail"`
	TimeGuardrail      bool `json:"time_guardrail"`
	CIPassed           bool `json:"ci_passed"`
}

type measurementContext struct {
	ScenarioID      string `json:"scenario_id"`
	SourceDigest    string `json:"source_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	CacheState      string `json:"cache_state"`
	Metric          string `json:"metric"`
}

type observation struct {
	ScenarioID          string             `json:"scenario_id"`
	SourceDigest        string             `json:"source_digest"`
	ToolchainDigest     string             `json:"toolchain_digest"`
	CacheState          string             `json:"cache_state"`
	PrimaryMetric       string             `json:"primary_metric"`
	BeforeContext       measurementContext `json:"before_context"`
	AfterContext        measurementContext `json:"after_context"`
	Before              metrics            `json:"before"`
	After               *metrics           `json:"after"`
	SemanticGuard       semanticGuard      `json:"semantic_guard"`
	RepositoryWrites    int                `json:"repository_writes"`
	LocalTestExecutions int                `json:"local_test_executions"`
	CIPassed            bool               `json:"ci_passed"`
	AuthorityProgram    string             `json:"authority_program"`
	GraphHash           string             `json:"graph_hash"`
}

type unknownFields struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type scenarioResult struct {
	ID              string             `json:"id"`
	Decision        string             `json:"decision"`
	Reason          string             `json:"reason"`
	SourceDigest    string             `json:"source_digest"`
	ToolchainDigest string             `json:"toolchain_digest"`
	CacheState      string             `json:"cache_state"`
	BeforeContext   measurementContext `json:"before_context"`
	AfterContext    measurementContext `json:"after_context"`
	Before          metrics            `json:"before"`
	After           *metrics           `json:"after"`
	SemanticGuard   semanticGuard      `json:"semantic_guard"`
	Unknown         *unknownFields     `json:"unknown,omitempty"`
}

type decisionReport struct {
	Schema                  string           `json:"schema"`
	AuthorityProgram        string           `json:"authority_program"`
	GraphHash               string           `json:"graph_hash"`
	CandidateEvidenceDigest string           `json:"candidate_evidence_digest"`
	CandidateDecision       string           `json:"candidate_decision"`
	ScenarioSuiteDecision   string           `json:"scenario_suite_decision"`
	Precedence              []string         `json:"precedence"`
	ScenarioDenominator     int              `json:"scenario_denominator"`
	Counts                  map[string]int   `json:"counts"`
	PrimaryMetric           string           `json:"primary_metric"`
	Scenarios               []scenarioResult `json:"scenarios"`
	RepositoryWrites        int              `json:"repository_writes"`
	LocalTestExecutions     int              `json:"local_test_executions"`
}

func main() {
	mode := flag.String("mode", "", "candidate or decide")
	program := flag.String("program", "", "authoritative .gooo program")
	graphPath := flag.String("graph", "", "semantic graph JSON")
	baseline := flag.String("baseline", "", "baseline Go source")
	candidate := flag.String("candidate", "", "caller-owned candidate Go output")
	evidence := flag.String("evidence", "", "candidate evidence output")
	normal := flag.String("normal", "", "normal scenario observation")
	candidateEvidencePath := flag.String("candidate-evidence", "", "candidate evidence JSON")
	output := flag.String("output", "", "report output")
	flag.Parse()

	var err error
	switch *mode {
	case "candidate":
		err = runCandidate(*program, *graphPath, *baseline, *candidate, *evidence)
	case "decide":
		err = runDecision(*program, *graphPath, *normal, *candidateEvidencePath, *output)
	default:
		err = fmt.Errorf("compiler self-improvement: -mode must be candidate or decide")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCandidate(programPath, graphPath, baselinePath, candidatePath, evidencePath string) error {
	if programPath == "" || graphPath == "" || baselinePath == "" || candidatePath == "" || evidencePath == "" {
		return fmt.Errorf("candidate mode requires -program, -graph, -baseline, -candidate, and -evidence")
	}
	program, err := os.ReadFile(programPath)
	if err != nil {
		return fmt.Errorf("read authority program: %w", err)
	}
	if err := verifyAuthority(string(program)); err != nil {
		return err
	}
	graphBytes, err := os.ReadFile(graphPath)
	if err != nil {
		return fmt.Errorf("read semantic graph: %w", err)
	}
	var released graph
	if err := json.Unmarshal(graphBytes, &released); err != nil {
		return fmt.Errorf("decode semantic graph: %w", err)
	}
	if err := verifyGraph(released); err != nil {
		return err
	}
	baseline, err := os.ReadFile(baselinePath)
	if err != nil {
		return fmt.Errorf("read baseline source: %w", err)
	}
	candidate, err := transformBaseline(string(baseline))
	if err != nil {
		return err
	}
	if err := os.WriteFile(candidatePath, []byte(candidate), 0o644); err != nil {
		return fmt.Errorf("write caller-owned candidate: %w", err)
	}

	baselineDigest, err := digestFile(baselinePath)
	if err != nil {
		return err
	}
	candidateDigest := digestBytes([]byte(candidate))
	receipt := candidateEvidence{
		Schema:                        schema,
		AuthorityProgram:              programRule,
		Target:                        targetPath,
		Fixture:                       fixturePath,
		GraphSchema:                   released.SchemaVersion,
		GraphHash:                     released.GraphHash,
		GraphSourceDigest:             released.SourceDigest,
		BaselineSourceDigest:          baselineDigest,
		CandidateSourceDigest:         candidateDigest,
		BaselineCollectionCopyPasses:  2,
		CandidateCollectionCopyPasses: 1,
		RepositoryWrites:              0,
		OutputMode:                    "caller-owned-temporary-output",
	}
	if err := writeJSON(evidencePath, receipt); err != nil {
		return err
	}
	fmt.Printf("candidate generated: %s (%s -> %s)\n", targetPath, baselineDigest, candidateDigest)
	return nil
}

func runDecision(programPath, graphPath, normalPath, evidencePath, outputPath string) error {
	if programPath == "" || graphPath == "" || normalPath == "" || evidencePath == "" || outputPath == "" {
		return fmt.Errorf("decide mode requires -program, -graph, -normal, -candidate-evidence, and -output")
	}
	program, err := os.ReadFile(programPath)
	if err != nil {
		return fmt.Errorf("read authority program: %w", err)
	}
	if err := verifyAuthority(string(program)); err != nil {
		return err
	}
	var released graph
	if err := readJSON(graphPath, &released); err != nil {
		return fmt.Errorf("read semantic graph: %w", err)
	}
	if err := verifyGraph(released); err != nil {
		return err
	}
	var normal observation
	if err := readJSON(normalPath, &normal); err != nil {
		return fmt.Errorf("read normal observation: %w", err)
	}
	var evidence candidateEvidence
	if err := readJSON(evidencePath, &evidence); err != nil {
		return fmt.Errorf("read candidate evidence: %w", err)
	}
	if normal.ScenarioID != "NORMAL" || normal.PrimaryMetric != primaryMetric {
		return fmt.Errorf("normal observation is not bound to %s", primaryMetric)
	}
	if evidence.AuthorityProgram != programRule || evidence.Target != targetPath || evidence.Fixture != fixturePath {
		return fmt.Errorf("candidate evidence is not bound to the authoritative transformation")
	}
	if evidence.Schema != schema || evidence.GraphSchema != graphSchema ||
		evidence.BaselineCollectionCopyPasses != 2 || evidence.CandidateCollectionCopyPasses != 1 ||
		evidence.RepositoryWrites != 0 || evidence.GraphHash != released.GraphHash || evidence.GraphSourceDigest != released.SourceDigest {
		return fmt.Errorf("candidate evidence graph identity mismatch")
	}
	if normal.AuthorityProgram != programRule || normal.GraphHash != released.GraphHash {
		return fmt.Errorf("normal observation is not bound to the authoritative graph")
	}
	evidenceDigest, err := digestFile(evidencePath)
	if err != nil {
		return err
	}

	normalResult, normalDecision, normalReason := evaluate(normal)
	unknown := normal
	unknown.ScenarioID = "UNKNOWN"
	unknown.After = nil
	unknownResult, _, _ := evaluate(unknown)
	refuted := normal
	refuted.ScenarioID = "REFUTED"
	refuted.SemanticGuard.MemoryGuardrail = false
	if refuted.After != nil {
		refutedAfter := *refuted.After
		refutedAfter.PeakRSSKiB = refuted.Before.PeakRSSKiB + 1
		refuted.After = &refutedAfter
	}
	refutedResult, _, _ := evaluate(refuted)

	results := []scenarioResult{normalResult, unknownResult, refutedResult}
	counts := map[string]int{decisionClosed: 0, decisionUnknown: 0, decisionRefuted: 0}
	for _, result := range results {
		counts[result.Decision]++
	}
	report := decisionReport{
		Schema:                  schema,
		AuthorityProgram:        programRule,
		GraphHash:               released.GraphHash,
		CandidateEvidenceDigest: evidenceDigest,
		CandidateDecision:       normalDecision,
		ScenarioSuiteDecision:   resolvePrecedence(results),
		Precedence:              []string{decisionRefuted, decisionUnknown, decisionClosed},
		ScenarioDenominator:     3,
		Counts:                  counts,
		PrimaryMetric:           primaryMetric,
		Scenarios:               results,
		RepositoryWrites:        normal.RepositoryWrites,
		LocalTestExecutions:     normal.LocalTestExecutions,
	}
	if normalReason != "" {
		report.Scenarios[0].Reason = normalReason
	}
	if err := writeJSON(outputPath, report); err != nil {
		return err
	}
	fmt.Printf("decision: candidate=%s scenarios=%s/%s/%s\n", report.CandidateDecision, results[0].Decision, results[1].Decision, results[2].Decision)
	return nil
}

func resolvePrecedence(results []scenarioResult) string {
	for _, decision := range []string{decisionRefuted, decisionUnknown, decisionClosed} {
		for _, result := range results {
			if result.Decision == decision {
				return decision
			}
		}
	}
	return decisionUnknown
}

func verifyAuthority(program string) error {
	required := []string{
		`computes "gooo.semantic.graph:v1"`,
		`computes "` + programRule + `;target=` + targetPath + `;fixture=` + fixturePath + `"`,
		`computes "compiler.measure:baseline:v1"`,
		`computes "compiler.measure:candidate:v1"`,
		`computes "compiler.guard:behavior-determinism-memory-time:v1"`,
		`computes "compiler.decision:REFUTED>UNKNOWN>CLOSED:v1"`,
	}
	for _, snippet := range required {
		if !strings.Contains(program, snippet) {
			return fmt.Errorf("authority program is missing %q", snippet)
		}
	}
	return nil
}

func verifyGraph(released graph) error {
	if released.SchemaVersion != graphSchema || !validDigest(released.GraphHash) || !validDigest(released.SourceDigest) {
		return fmt.Errorf("semantic graph is missing a valid schema or digest")
	}
	required := map[string]bool{
		"ObserveSource": false, "GenerateCandidate": false, "ExecuteBaseline": false,
		"ExecuteCandidate": false, "GuardSemantics": false, "Decide": false,
	}
	for _, node := range released.Nodes {
		if node.Kind == "Activity" {
			if _, ok := required[node.Name]; ok {
				required[node.Name] = true
			}
		}
	}
	for name, found := range required {
		if !found {
			return fmt.Errorf("semantic graph is missing activity %q", name)
		}
	}
	return nil
}

func transformBaseline(source string) (string, error) {
	oldCall := "\tresult := copyIR(input)\n\tcanonicalizeIRCollections(&result)"
	if strings.Count(source, oldCall) != 1 {
		return "", fmt.Errorf("baseline does not contain exactly one redundant normalization pass")
	}
	result := strings.Replace(source, oldCall, "\tresult := copyIR(input)", 1)
	oldFunction := `// canonicalizeIRCollections gives semantically equivalent nil and empty
// collections one wire representation without changing caller-owned input.
// This keeps generated metadata digests independent of how an adapter
// materializes absent optional declarations.
func canonicalizeIRCollections(ir *SemanticIR) {
	ir.Imports = append([]Import{}, ir.Imports...)
	ir.Entities = append([]Entity{}, ir.Entities...)
	for index := range ir.Entities {
		ir.Entities[index].Fields = append([]Field{}, ir.Entities[index].Fields...)
	}
	ir.Activities = append([]Activity{}, ir.Activities...)
	for index := range ir.Activities {
		ir.Activities[index].Inputs = append([]Port{}, ir.Activities[index].Inputs...)
		ir.Activities[index].Outputs = append([]Port{}, ir.Activities[index].Outputs...)
		ir.Activities[index].Slots = append([]Slot{}, ir.Activities[index].Slots...)
	}
}
`
	if strings.Count(result, oldFunction) != 1 {
		return "", fmt.Errorf("baseline canonicalization function is not the expected exact implementation")
	}
	result = strings.Replace(result, oldFunction, "", 1)
	copyHeader := "\tresult := input\n\t// Copying with a non-nil zero-length seed preserves the canonical empty\n\t// collection representation while folding canonicalization into the only\n\t// collection copy pass.\n"
	if strings.Count(result, "\tresult := input\n") != 1 {
		return "", fmt.Errorf("baseline copyIR header is not the expected exact implementation")
	}
	result = strings.Replace(result, "\tresult := input\n", copyHeader, 1)
	for _, replacement := range []struct{ old, new string }{
		{"append([]Import(nil), input.Imports...)", "append([]Import{}, input.Imports...)"},
		{"append([]Entity(nil), input.Entities...)", "append([]Entity{}, input.Entities...)"},
		{"append([]Activity(nil), input.Activities...)", "append([]Activity{}, input.Activities...)"},
		{"append([]Field(nil), input.Entities[index].Fields...)", "append([]Field{}, input.Entities[index].Fields...)"},
		{"append([]Port(nil), input.Activities[index].Inputs...)", "append([]Port{}, input.Activities[index].Inputs...)"},
		{"append([]Port(nil), input.Activities[index].Outputs...)", "append([]Port{}, input.Activities[index].Outputs...)"},
		{"append([]Slot(nil), input.Activities[index].Slots...)", "append([]Slot{}, input.Activities[index].Slots...)"},
	} {
		if strings.Count(result, replacement.old) != 1 {
			return "", fmt.Errorf("baseline copyIR is missing exact collection %q", replacement.old)
		}
		result = strings.Replace(result, replacement.old, replacement.new, 1)
	}
	if strings.Contains(result, "canonicalizeIRCollections") || !strings.Contains(result, "func copyIR(input SemanticIR)") {
		return "", fmt.Errorf("generated candidate failed the single-pass normalization invariant")
	}
	return result, nil
}

func evaluate(input observation) (scenarioResult, string, string) {
	result := scenarioResult{
		ID:              input.ScenarioID,
		SourceDigest:    input.SourceDigest,
		ToolchainDigest: input.ToolchainDigest,
		CacheState:      input.CacheState,
		BeforeContext:   input.BeforeContext,
		AfterContext:    input.AfterContext,
		Before:          input.Before,
		After:           input.After,
		SemanticGuard:   input.SemanticGuard,
	}
	if input.After == nil || !exactContext(input) || !validMetrics(input.Before) || (input.After != nil && !validMetrics(*input.After)) {
		reason := unknownReason
		if input.After != nil {
			reason = "MISSING_OR_MISMATCHED_EXACT_CONTEXT_OR_INTEGER"
		}
		result.Decision = decisionUnknown
		result.Reason = reason
		result.Unknown = &unknownFields{
			Stage:         "COHERENCE",
			Step:          "COMPARE_EXACT_PAIR",
			Reason:        reason,
			UnknownClass:  "INCOMPLETE_EVIDENCE",
			NextOperation: unknownNextOperation,
			BlockedBy:     []string{"before_after_pair"},
		}
		return result, decisionUnknown, result.Reason
	}
	if input.RepositoryWrites != 0 || input.LocalTestExecutions != 0 {
		result.Decision = decisionRefuted
		result.Reason = "REPOSITORY_WRITES_OR_LOCAL_EXECUTION_OBSERVED"
		return result, decisionRefuted, result.Reason
	}
	if !input.SemanticGuard.BehaviorEqual || !input.SemanticGuard.DeterministicEqual ||
		!input.SemanticGuard.MemoryGuardrail || !input.SemanticGuard.TimeGuardrail || !input.SemanticGuard.CIPassed || !input.CIPassed {
		result.Decision = decisionRefuted
		result.Reason = "SEMANTIC_OR_GUARDRAIL_NON_REGRESSION_FAILED"
		return result, decisionRefuted, result.Reason
	}
	if input.After.NormalizationCollectionCopyPasses >= input.Before.NormalizationCollectionCopyPasses {
		result.Decision = decisionRefuted
		result.Reason = "PRIMARY_DETERMINISTIC_COUNT_DID_NOT_IMPROVE"
		return result, decisionRefuted, result.Reason
	}
	result.Decision = decisionClosed
	result.Reason = "PRIMARY_DETERMINISTIC_COUNT_IMPROVED_WITH_NON_REGRESSION"
	return result, decisionClosed, result.Reason
}

func exactContext(input observation) bool {
	before := input.BeforeContext
	after := input.AfterContext
	return input.PrimaryMetric == primaryMetric &&
		before.Metric == primaryMetric && after.Metric == primaryMetric &&
		validDigest(before.SourceDigest) && before.SourceDigest == after.SourceDigest && before.SourceDigest == input.SourceDigest &&
		validDigest(before.ToolchainDigest) && before.ToolchainDigest == after.ToolchainDigest && before.ToolchainDigest == input.ToolchainDigest &&
		before.CacheState != "" && before.CacheState == after.CacheState && before.CacheState == input.CacheState
}

func validMetrics(value metrics) bool {
	return value.NormalizationCollectionCopyPasses >= 0 && value.WallMS > 0 && value.PeakRSSKiB > 0 &&
		value.BuildMS > 0 && value.TestMS > 0 && value.ExecutedTests >= 0 && value.ReusedTests >= 0
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestFile(path string) (string, error) {
	value, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("digest %s: %w", path, err)
	}
	return digestBytes(value), nil
}

func readJSON(path string, target any) error {
	value, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(value, target)
}

func writeJSON(path string, value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
