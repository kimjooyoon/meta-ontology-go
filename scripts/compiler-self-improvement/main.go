package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const (
	schema                 = "gooo/compiler-self-improvement/v2"
	programRule            = "lsp.refresh:reuse-exact-document-cache/v1"
	targetPath             = "internal/lsp/features_part03.go"
	fixturePath            = "examples/billing/main.gooo"
	graphSchema            = "gooo-graph/v1"
	primaryMetric          = "lsp_refresh_parse_calls"
	decisionClosed         = "CLOSED"
	decisionUnknown        = "UNKNOWN"
	decisionRefuted        = "REFUTED"
	unknownReason          = "MISSING_EXACT_INTEGER_PAIR"
	unknownNextOperation   = "CAPTURE_EXACT_BEFORE_AFTER"
	envelopeSummaryName    = "envelope-summary.json"
	expectedEnvelopeSchema = "gooo/compiler-operation-envelope/v1"
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
	Schema                         string         `json:"schema"`
	AuthorityProgram               string         `json:"authority_program"`
	Target                         string         `json:"target"`
	Fixture                        string         `json:"fixture"`
	GraphSchema                    string         `json:"graph_schema"`
	GraphHash                      string         `json:"graph_hash"`
	GraphSourceDigest              string         `json:"graph_source_digest"`
	BaselineSourceDigest           string         `json:"baseline_source_digest"`
	CandidateSourceDigest          string         `json:"candidate_source_digest"`
	BaselineTransformationCount    int            `json:"baseline_transformation_count"`
	CandidateTransformationCount   int            `json:"candidate_transformation_count"`
	BaselinePrimaryOperationCount  int            `json:"baseline_primary_operation_count"`
	CandidatePrimaryOperationCount int            `json:"candidate_primary_operation_count"`
	EnvelopeAuthorityDigest        string         `json:"envelope_authority_digest"`
	EnvelopeSummaryDigest          string         `json:"envelope_summary_digest"`
	EnvelopeScenarioDenominator    int            `json:"envelope_scenario_denominator"`
	EnvelopeCounts                 map[string]int `json:"envelope_counts"`
	EnvelopeVerifierPassed         bool           `json:"envelope_verifier_passed"`
	RepositoryWrites               int            `json:"repository_writes"`
	OutputMode                     string         `json:"output_mode"`
}

type envelopeScenario struct {
	ScenarioID    string                     `json:"scenario_id"`
	Generated     string                     `json:"generated_decision"`
	Verified      string                     `json:"verified_decision"`
	Reason        string                     `json:"reason"`
	ReceiptDigest string                     `json:"receipt_digest"`
	Metrics       generation.EnvelopeMetrics `json:"metrics"`
	Artifacts     int                        `json:"artifacts"`
}

type envelopeSummary struct {
	Schema              string             `json:"schema"`
	AuthorityProgram    string             `json:"authority_program"`
	AuthorityDigest     string             `json:"authority_digest"`
	GraphHash           string             `json:"graph_hash"`
	ScenarioDenominator int                `json:"scenario_denominator"`
	Counts              map[string]int     `json:"counts"`
	Scenarios           []envelopeScenario `json:"scenarios"`
	RepositoryWrites    int                `json:"repository_writes"`
	LocalTestExecutions int                `json:"local_test_executions"`
}

type envelopeVerification struct {
	ScenarioDenominator int
	Counts              map[string]int
	Scenarios           []envelopeScenario
}

type metrics struct {
	TransformationCount   int64 `json:"transformation_count"`
	PrimaryOperationCount int64 `json:"primary_operation_count"`
	AllocationCount       int64 `json:"allocation_count"`
	AllocationBytes       int64 `json:"allocation_bytes"`
	WallMS                int64 `json:"wall_ms"`
	PeakRSSKiB            int64 `json:"peak_rss_kib"`
	BuildMS               int64 `json:"build_ms"`
	TestMS                int64 `json:"test_ms"`
	ExecutedTests         int64 `json:"executed_tests"`
	ReusedTests           int64 `json:"reused_tests"`
	InputDescendantDirs   int64 `json:"input_descendant_dirs"`
	InputRegularFiles     int64 `json:"input_regular_files"`
	InputGoPhysicalLines  int64 `json:"input_go_physical_lines"`
	InputGoooLines        int64 `json:"input_gooo_physical_lines"`
	OutputArtifactFiles   int64 `json:"output_artifact_files"`
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
	ProfileDigest   string `json:"profile_digest"`
	ContractDigest  string `json:"contract_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	CacheState      string `json:"cache_state"`
	Metric          string `json:"metric"`
}

type observation struct {
	ScenarioID          string             `json:"scenario_id"`
	SourceDigest        string             `json:"source_digest"`
	ProfileDigest       string             `json:"profile_digest"`
	ContractDigest      string             `json:"contract_digest"`
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
	ProfileDigest   string             `json:"profile_digest"`
	ContractDigest  string             `json:"contract_digest"`
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
	EnvelopeSuiteDecision   string           `json:"envelope_suite_decision"`
	Precedence              []string         `json:"precedence"`
	ScenarioDenominator     int              `json:"scenario_denominator"`
	EnvelopeScenarioDenom   int              `json:"envelope_scenario_denominator"`
	Counts                  map[string]int   `json:"counts"`
	EnvelopeCounts          map[string]int   `json:"envelope_counts"`
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
	envelopeDir := flag.String("envelope-dir", "", "caller-owned operation envelope output")
	normal := flag.String("normal", "", "normal scenario observation")
	candidateEvidencePath := flag.String("candidate-evidence", "", "candidate evidence JSON")
	output := flag.String("output", "", "report output")
	flag.Parse()

	var err error
	switch *mode {
	case "candidate":
		err = runCandidate(*program, *graphPath, *baseline, *candidate, *evidence, *envelopeDir)
	case "decide":
		err = runDecision(*program, *graphPath, *normal, *candidateEvidencePath, *envelopeDir, *output)
	default:
		err = fmt.Errorf("compiler self-improvement: -mode must be candidate or decide")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCandidate(programPath, graphPath, baselinePath, candidatePath, evidencePath, envelopeDir string) error {
	if programPath == "" || graphPath == "" || baselinePath == "" || candidatePath == "" || evidencePath == "" || envelopeDir == "" {
		return fmt.Errorf("candidate mode requires -program, -graph, -baseline, -candidate, -evidence, and -envelope-dir")
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

	summary, err := generateEnvelope(program, released, envelopeDir)
	if err != nil {
		return err
	}
	summaryDigest, err := digestFile(filepath.Join(envelopeDir, envelopeSummaryName))
	if err != nil {
		return err
	}
	baselineDigest := digestBytes(baseline)
	candidateDigest := digestBytes([]byte(candidate))
	evidence := candidateEvidence{
		Schema:                         schema,
		AuthorityProgram:               programRule,
		Target:                         targetPath,
		Fixture:                        fixturePath,
		GraphSchema:                    released.SchemaVersion,
		GraphHash:                      released.GraphHash,
		GraphSourceDigest:              released.SourceDigest,
		BaselineSourceDigest:           baselineDigest,
		CandidateSourceDigest:          candidateDigest,
		BaselineTransformationCount:    1,
		CandidateTransformationCount:   1,
		BaselinePrimaryOperationCount:  countUnchangedRefreshParseCalls(string(baseline)),
		CandidatePrimaryOperationCount: countUnchangedRefreshParseCalls(candidate),
		EnvelopeAuthorityDigest:        digestBytes(program),
		EnvelopeSummaryDigest:          summaryDigest,
		EnvelopeScenarioDenominator:    summary.ScenarioDenominator,
		EnvelopeCounts:                 summary.Counts,
		EnvelopeVerifierPassed:         true,
		RepositoryWrites:               0,
		OutputMode:                     "caller-owned-temporary-output",
	}
	if evidence.BaselinePrimaryOperationCount != 1 || evidence.CandidatePrimaryOperationCount != 0 {
		return fmt.Errorf("candidate unchanged-refresh parse-call count is not exactly 1 -> 0")
	}
	if err := writeJSON(evidencePath, evidence); err != nil {
		return err
	}
	fmt.Printf("candidate generated: %s (%d -> %d unchanged-refresh parse calls)\n", targetPath, evidence.BaselinePrimaryOperationCount, evidence.CandidatePrimaryOperationCount)
	return nil
}

func runDecision(programPath, graphPath, normalPath, evidencePath, envelopeDir, outputPath string) error {
	if programPath == "" || graphPath == "" || normalPath == "" || evidencePath == "" || envelopeDir == "" || outputPath == "" {
		return fmt.Errorf("decide mode requires -program, -graph, -normal, -candidate-evidence, -envelope-dir, and -output")
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
	if err := verifyCandidateEvidence(evidence, released); err != nil {
		return err
	}
	verifiedEnvelope, err := verifyEnvelopeSuite(program, released, envelopeDir)
	if err != nil {
		return err
	}
	if verifiedEnvelope.ScenarioDenominator != evidence.EnvelopeScenarioDenominator || !sameCounts(verifiedEnvelope.Counts, evidence.EnvelopeCounts) {
		return fmt.Errorf("envelope evidence summary does not match independent verification")
	}
	summaryDigest, err := digestFile(filepath.Join(envelopeDir, envelopeSummaryName))
	if err != nil {
		return err
	}
	if summaryDigest != evidence.EnvelopeSummaryDigest {
		return fmt.Errorf("envelope summary digest mismatch")
	}
	if normal.ScenarioID != "NORMAL" || normal.PrimaryMetric != primaryMetric {
		return fmt.Errorf("normal observation is not bound to %s", primaryMetric)
	}
	if normal.AuthorityProgram != programRule || normal.GraphHash != released.GraphHash {
		return fmt.Errorf("normal observation is not bound to the authoritative graph")
	}
	if normal.Before.PrimaryOperationCount != int64(evidence.BaselinePrimaryOperationCount) || normal.After == nil || normal.After.PrimaryOperationCount != int64(evidence.CandidatePrimaryOperationCount) {
		return fmt.Errorf("normal observation does not carry the generated primary-operation pair")
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
		EnvelopeSuiteDecision:   resolveEnvelopePrecedence(verifiedEnvelope.Scenarios),
		Precedence:              []string{decisionRefuted, decisionUnknown, decisionClosed},
		ScenarioDenominator:     3,
		EnvelopeScenarioDenom:   verifiedEnvelope.ScenarioDenominator,
		Counts:                  counts,
		EnvelopeCounts:          verifiedEnvelope.Counts,
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
	fmt.Printf("decision: candidate=%s compiler-scenarios=%s/%s/%s envelope=%s\n", report.CandidateDecision, results[0].Decision, results[1].Decision, results[2].Decision, report.EnvelopeSuiteDecision)
	return nil
}

func generateEnvelope(program []byte, released graph, outputDir string) (envelopeSummary, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return envelopeSummary{}, fmt.Errorf("create operation envelope output: %w", err)
	}
	counts := map[string]int{decisionClosed: 0, decisionUnknown: 0, decisionRefuted: 0}
	scenarios := make([]envelopeScenario, 0, len(generation.SemanticOperationScenarioIDs()))
	for _, scenarioID := range generation.SemanticOperationScenarioIDs() {
		scenarioDir := filepath.Join(outputDir, scenarioID)
		if err := os.MkdirAll(scenarioDir, 0o755); err != nil {
			return envelopeSummary{}, fmt.Errorf("create envelope scenario %s: %w", scenarioID, err)
		}
		run, err := generation.GenerateSemanticOperationEnvelope(program, scenarioID, scenarioDir)
		if err != nil {
			return envelopeSummary{}, fmt.Errorf("generate envelope %s: %w", scenarioID, err)
		}
		verified, err := generation.VerifySemanticOperationEnvelope(scenarioDir)
		if err != nil {
			return envelopeSummary{}, fmt.Errorf("verify envelope %s: %w", scenarioID, err)
		}
		if run.IR.Decision.Decision != verified.Decision || run.IR.Decision.Reason != verified.Reason {
			return envelopeSummary{}, fmt.Errorf("envelope %s generated and verified decisions differ", scenarioID)
		}
		counts[verified.Decision]++
		if err := writeJSON(filepath.Join(outputDir, "semantic-ir", scenarioID, "semantic-ir.json"), run.IR); err != nil {
			return envelopeSummary{}, fmt.Errorf("write semantic IR %s: %w", scenarioID, err)
		}
		scenarios = append(scenarios, envelopeScenario{
			ScenarioID:    scenarioID,
			Generated:     run.IR.Decision.Decision,
			Verified:      verified.Decision,
			Reason:        verified.Reason,
			ReceiptDigest: verified.ReceiptDigest,
			Metrics:       verified.Metrics,
			Artifacts:     len(run.Artifacts),
		})
	}
	summary := envelopeSummary{
		Schema:              expectedEnvelopeSchema,
		AuthorityProgram:    "compiler.operation-envelope:v1",
		AuthorityDigest:     digestBytes(program),
		GraphHash:           released.GraphHash,
		ScenarioDenominator: len(scenarios),
		Counts:              counts,
		Scenarios:           scenarios,
		RepositoryWrites:    0,
		LocalTestExecutions: 0,
	}
	if err := writeJSON(filepath.Join(outputDir, envelopeSummaryName), summary); err != nil {
		return envelopeSummary{}, err
	}
	return summary, nil
}

func verifyEnvelopeSuite(program []byte, released graph, outputDir string) (envelopeVerification, error) {
	var summary envelopeSummary
	if err := readJSON(filepath.Join(outputDir, envelopeSummaryName), &summary); err != nil {
		return envelopeVerification{}, fmt.Errorf("read envelope summary: %w", err)
	}
	if summary.Schema != expectedEnvelopeSchema || summary.AuthorityDigest != digestBytes(program) || summary.GraphHash != released.GraphHash || summary.RepositoryWrites != 0 || summary.LocalTestExecutions != 0 {
		return envelopeVerification{}, fmt.Errorf("envelope summary authority mismatch")
	}
	expected := map[string]string{"C1": decisionClosed, "C2": decisionClosed, "U1": decisionUnknown, "U2": decisionUnknown, "R1": decisionRefuted, "R2": decisionRefuted}
	counts := map[string]int{decisionClosed: 0, decisionUnknown: 0, decisionRefuted: 0}
	verifiedScenarios := make([]envelopeScenario, 0, len(generation.SemanticOperationScenarioIDs()))
	for _, scenarioID := range generation.SemanticOperationScenarioIDs() {
		verified, err := generation.VerifySemanticOperationEnvelope(filepath.Join(outputDir, scenarioID))
		if err != nil {
			return envelopeVerification{}, fmt.Errorf("independent envelope verification %s: %w", scenarioID, err)
		}
		if verified.ScenarioID != scenarioID || verified.Decision != expected[scenarioID] {
			return envelopeVerification{}, fmt.Errorf("envelope scenario %s has decision %s", scenarioID, verified.Decision)
		}
		counts[verified.Decision]++
		verifiedScenarios = append(verifiedScenarios, envelopeScenario{ScenarioID: scenarioID, Generated: verified.Decision, Verified: verified.Decision, Reason: verified.Reason, ReceiptDigest: verified.ReceiptDigest, Metrics: verified.Metrics, Artifacts: 6})
	}
	if summary.ScenarioDenominator != len(verifiedScenarios) || !sameCounts(summary.Counts, counts) || len(summary.Scenarios) != len(verifiedScenarios) {
		return envelopeVerification{}, fmt.Errorf("envelope summary does not match independent scenario denominator")
	}
	for index := range summary.Scenarios {
		if summary.Scenarios[index].ScenarioID != verifiedScenarios[index].ScenarioID || summary.Scenarios[index].Generated != verifiedScenarios[index].Generated || summary.Scenarios[index].Verified != verifiedScenarios[index].Verified || summary.Scenarios[index].Reason != verifiedScenarios[index].Reason || summary.Scenarios[index].ReceiptDigest != verifiedScenarios[index].ReceiptDigest || summary.Scenarios[index].Artifacts != 6 {
			return envelopeVerification{}, fmt.Errorf("envelope summary scenario %s mismatch", verifiedScenarios[index].ScenarioID)
		}
	}
	return envelopeVerification{ScenarioDenominator: len(verifiedScenarios), Counts: counts, Scenarios: verifiedScenarios}, nil
}

func verifyCandidateEvidence(evidence candidateEvidence, released graph) error {
	if evidence.Schema != schema || evidence.AuthorityProgram != programRule || evidence.Target != targetPath || evidence.Fixture != fixturePath || evidence.GraphSchema != graphSchema || evidence.GraphHash != released.GraphHash || evidence.GraphSourceDigest != released.SourceDigest || evidence.BaselinePrimaryOperationCount != 1 || evidence.CandidatePrimaryOperationCount != 0 || evidence.BaselineTransformationCount != 1 || evidence.CandidateTransformationCount != 1 || evidence.EnvelopeScenarioDenominator != len(generation.SemanticOperationScenarioIDs()) || !evidence.EnvelopeVerifierPassed || evidence.RepositoryWrites != 0 {
		return fmt.Errorf("candidate evidence authority or metric identity mismatch")
	}
	if !validDigest(evidence.BaselineSourceDigest) || !validDigest(evidence.CandidateSourceDigest) || !validDigest(evidence.EnvelopeAuthorityDigest) || !validDigest(evidence.EnvelopeSummaryDigest) {
		return fmt.Errorf("candidate evidence is missing a valid digest")
	}
	return nil
}

func verifyAuthority(program string) error {
	required := []string{
		`activity DeclareOperationIntent(OperationIntentInput) -> OperationIntent computes "compiler.operation-intent:v1;operation=lsp-document-refresh;mode=read-only"`,
		`activity BindSourceRevision(SourceRevisionInput) -> SourceRevision computes "compiler.source-revision:v1;binding=exact-source-digest;target=internal/lsp/features_part03.go;metric=lsp_refresh_parse_calls"`,
		`activity DeclareEffectGrant(EffectGrantInput) -> EffectGrant computes "compiler.effect-grant:v1;effects=read:source;repository-writes=0"`,
		`activity EmitEffectRequest(EffectRequestInput) -> EffectRequest computes "compiler.effect-request:v1;replay=exact-source-profile-toolchain-contract-digest"`,
		`activity RecordEffectResult(EffectResultInput) -> EffectResult computes "compiler.effect-result:v1;match=request-source-grant;cases=unchanged-reuse|changed-input-invalidation"`,
		`activity VerifyReplayIdentity(ReplayIdentityInput) -> ReplayIdentity computes "compiler.replay-identity:v1;compare=current-request-digest;stale=corrupt-cache-reject"`,
		`activity ClassifySemanticOutcome(SemanticOutcomeInput) -> OperationDecision computes "compiler.decision:v1;precedence=REFUTED>UNKNOWN>CLOSED;conditions=unchanged-reuse|changed-input-invalidation|corrupt-stale-evidence"`,
		`activity PublishOperationReceipt(ReceiptInput) -> OperationReceipt computes "compiler.operation-receipt:v1;artifacts=6;utility=UNKNOWN"`,
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
	required := map[string]bool{}
	for _, name := range []string{"DeclareOperationIntent", "BindSourceRevision", "DeclareEffectGrant", "EmitEffectRequest", "RecordEffectResult", "VerifyReplayIdentity", "ClassifySemanticOutcome", "PublishOperationReceipt"} {
		required[name] = false
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
	oldBlock := "\tserver.mu.RLock()\n\tdocument, exists := server.documents[uri]\n\tif exists {\n\t\tversion, source := document.version, document.text\n\t\tserver.mu.RUnlock()\n\t\tresult, err := server.parse(ctx, uri, source)\n"
	if strings.Count(source, oldBlock) != 1 || countUnchangedRefreshParseCalls(source) != 1 {
		return "", fmt.Errorf("baseline does not contain exactly one unchanged-refresh parse call")
	}
	newBlock := "\tif err := ctx.Err(); err != nil {\n\t\treturn err\n\t}\n\tserver.mu.RLock()\n\tdocument, exists := server.documents[uri]\n\tif exists {\n\t\tversion, source, cachedKey := document.version, document.text, document.cacheKey\n\t\tserver.mu.RUnlock()\n\t\tif cachedKey == server.cacheKey(source) {\n\t\t\treturn nil\n\t\t}\n\t\tresult, err := server.parse(ctx, uri, source)\n"
	result := strings.Replace(source, oldBlock, newBlock, 1)
	oldAssignment := "\t\tcurrent.result = result\n"
	newAssignment := "\t\tcurrent.result = result\n\t\tcurrent.cacheKey = server.cacheKey(source)\n"
	if strings.Count(result, oldAssignment) != 1 {
		return "", fmt.Errorf("baseline does not contain exactly one refreshed document result assignment")
	}
	result = strings.Replace(result, oldAssignment, newAssignment, 1)
	if countUnchangedRefreshParseCalls(result) != 0 || !strings.Contains(result, "if cachedKey == server.cacheKey(source)") {
		return "", fmt.Errorf("generated candidate failed the unchanged-refresh reuse invariant")
	}
	return result, nil
}

func countUnchangedRefreshParseCalls(source string) int {
	if strings.Contains(source, "if cachedKey == server.cacheKey(source)") {
		return 0
	}
	if strings.Count(source, "\t\tresult, err := server.parse(ctx, uri, source)") == 1 {
		return 1
	}
	return -1
}

func evaluate(input observation) (scenarioResult, string, string) {
	result := scenarioResult{ID: input.ScenarioID, SourceDigest: input.SourceDigest, ProfileDigest: input.ProfileDigest, ContractDigest: input.ContractDigest, ToolchainDigest: input.ToolchainDigest, CacheState: input.CacheState, BeforeContext: input.BeforeContext, AfterContext: input.AfterContext, Before: input.Before, After: input.After, SemanticGuard: input.SemanticGuard}
	if input.After == nil || !exactContext(input) || !validMetrics(input.Before) || (input.After != nil && !validMetrics(*input.After)) {
		result.Decision = decisionUnknown
		result.Reason = unknownReason
		result.Unknown = &unknownFields{Stage: "COHERENCE", Step: "COMPARE_EXACT_PAIR", Reason: unknownReason, UnknownClass: "INCOMPLETE_EVIDENCE", NextOperation: unknownNextOperation, BlockedBy: []string{"before_after_pair"}}
		return result, decisionUnknown, result.Reason
	}
	if input.RepositoryWrites != 0 || input.LocalTestExecutions != 0 {
		result.Decision = decisionRefuted
		result.Reason = "REPOSITORY_WRITES_OR_LOCAL_EXECUTION_OBSERVED"
		return result, decisionRefuted, result.Reason
	}
	if !input.SemanticGuard.BehaviorEqual || !input.SemanticGuard.DeterministicEqual || !input.SemanticGuard.MemoryGuardrail || !input.SemanticGuard.TimeGuardrail || !input.SemanticGuard.CIPassed || !input.CIPassed {
		result.Decision = decisionRefuted
		result.Reason = "SEMANTIC_OR_GUARDRAIL_NON_REGRESSION_FAILED"
		return result, decisionRefuted, result.Reason
	}
	if input.After.PrimaryOperationCount >= input.Before.PrimaryOperationCount {
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
	return input.PrimaryMetric == primaryMetric && before.ScenarioID == "NORMAL" && after.ScenarioID == "NORMAL" && before.Metric == primaryMetric && after.Metric == primaryMetric && validDigest(before.SourceDigest) && before.SourceDigest == after.SourceDigest && before.SourceDigest == input.SourceDigest && validDigest(before.ProfileDigest) && before.ProfileDigest == after.ProfileDigest && before.ProfileDigest == input.ProfileDigest && validDigest(before.ContractDigest) && before.ContractDigest == after.ContractDigest && before.ContractDigest == input.ContractDigest && validDigest(before.ToolchainDigest) && before.ToolchainDigest == after.ToolchainDigest && before.ToolchainDigest == input.ToolchainDigest && before.CacheState != "" && before.CacheState == after.CacheState && before.CacheState == input.CacheState
}

func validMetrics(value metrics) bool {
	return value.TransformationCount >= 0 && value.PrimaryOperationCount >= 0 && value.AllocationCount >= 0 && value.AllocationBytes >= 0 && value.WallMS > 0 && value.PeakRSSKiB > 0 && value.BuildMS > 0 && value.TestMS > 0 && value.ExecutedTests >= 0 && value.ReusedTests >= 0 && value.InputDescendantDirs >= 0 && value.InputRegularFiles > 0 && value.InputGoPhysicalLines > 0 && value.InputGoooLines > 0 && value.OutputArtifactFiles == 6
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

func resolveEnvelopePrecedence(results []envelopeScenario) string {
	for _, decision := range []string{decisionRefuted, decisionUnknown, decisionClosed} {
		for _, result := range results {
			if result.Verified == decision {
				return decision
			}
		}
	}
	return decisionUnknown
}

func sameCounts(left, right map[string]int) bool {
	for _, decision := range []string{decisionClosed, decisionUnknown, decisionRefuted} {
		if left[decision] != right[decision] {
			return false
		}
	}
	return len(left) == len(right)
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func validDigest(value string) bool {
	value = strings.TrimPrefix(value, "sha256:")
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
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
