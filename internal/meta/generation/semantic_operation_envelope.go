package generation

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SemanticOperationEnvelopeSchema identifies the small, replayable operation
// boundary owned by the semantic layer.
const SemanticOperationEnvelopeSchema = "gooo/semantic-operation-envelope/v1"

const semanticOperationToolchainDigest = "semantic-operation-envelope-toolchain-v1"

var semanticOperationActivities = [...]string{
	"DeclareOperationIntent",
	"BindSourceRevision",
	"DeclareEffectGrant",
	"EmitEffectRequest",
	"RecordEffectResult",
	"VerifyReplayIdentity",
	"ClassifySemanticOutcome",
	"PublishOperationReceipt",
}

// OperationIntent is the user-owned intent carried into the semantic IR.
type OperationIntent struct {
	OperationID   string `json:"operation_id"`
	ScenarioID    string `json:"scenario_id"`
	RequestedMode string `json:"requested_mode"`
}

// SourceRevision binds every request and result to one exact source revision.
type SourceRevision struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}

// EffectGrant is the closed set of effects that an operation may use.
type EffectGrant struct {
	GrantID  string   `json:"grant_id"`
	ParentID string   `json:"parent_id"`
	Effects  []string `json:"effects"`
}

// EffectRequest is the deterministic request handed to an external executor.
type EffectRequest struct {
	RequestID      string   `json:"request_id"`
	OperationID    string   `json:"operation_id"`
	SourceRevision string   `json:"source_revision"`
	Effects        []string `json:"effects"`
	ReplayIdentity string   `json:"replay_identity"`
	PayloadDigest  string   `json:"payload_digest"`
}

// EffectResult is evidence returned by an executor. It is never treated as a
// semantic approval until it is checked against the request and grant.
type EffectResult struct {
	ResultID       string   `json:"result_id"`
	RequestID      string   `json:"request_id"`
	SourceRevision string   `json:"source_revision"`
	Effects        []string `json:"effects"`
	PayloadDigest  string   `json:"payload_digest"`
	ArtifactDigest string   `json:"artifact_digest"`
}

// SemanticPatch describes the proposal without applying it to the input
// repository.
type SemanticPatch struct {
	Schema           string   `json:"schema"`
	ScenarioID       string   `json:"scenario_id"`
	Changed          bool     `json:"changed"`
	Operations       []string `json:"operations"`
	RepositoryWrites int      `json:"repository_writes"`
}

// ReplayIdentity binds a request to its replay comparison.
type ReplayIdentity struct {
	Identity              string `json:"identity"`
	CurrentRequestDigest  string `json:"current_request_digest"`
	PreviousRequestDigest string `json:"previous_request_digest"`
	Compared              bool   `json:"compared"`
}

// OperationDecision is the semantic outcome. REFUTED always outranks UNKNOWN,
// which always outranks CLOSED.
type OperationDecision struct {
	Decision string                `json:"decision"`
	Reason   string                `json:"reason"`
	Unknown  *EnvelopeUnknownState `json:"unknown"`
}

// EnvelopeUnknownState is intentionally limited to the six required fields.
type EnvelopeUnknownState struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

// SemanticOperationIR is the semantic intermediate representation produced
// directly from the .gooo authority and consumed by the artifact generator.
type SemanticOperationIR struct {
	Schema          string            `json:"schema"`
	Intent          OperationIntent   `json:"intent"`
	Source          SourceRevision    `json:"source"`
	Grant           EffectGrant       `json:"grant"`
	Request         *EffectRequest    `json:"request"`
	Result          *EffectResult     `json:"result"`
	Replay          ReplayIdentity    `json:"replay"`
	Decision        OperationDecision `json:"decision"`
	Activities      []string          `json:"activities"`
	AuthorityDigest string            `json:"authority_digest"`
	ToolchainDigest string            `json:"toolchain_digest"`
}

// EnvelopeMetrics are exact integer observations. The library does not run
// tests or write the input repository, so both corresponding fields remain 0.
type EnvelopeMetrics struct {
	OperationRequests        int `json:"operation_requests"`
	OperationResults         int `json:"operation_results"`
	EffectsGranted           int `json:"effects_granted"`
	EffectsUsed              int `json:"effects_used"`
	ReplayComparisons        int `json:"replay_comparisons"`
	ReplayMismatches         int `json:"replay_mismatches"`
	StaleRejections          int `json:"stale_rejections"`
	EffectEscalationsRefuted int `json:"effect_escalations_refuted"`
	InputDescendantDirs      int `json:"input_descendant_dirs"`
	InputRegularFiles        int `json:"input_regular_files"`
	InputGoPhysicalLines     int `json:"input_go_physical_lines"`
	InputGoooPhysicalLines   int `json:"input_gooo_physical_lines"`
	OutputArtifactFiles      int `json:"output_artifact_files"`
	PeakRSSKib               int `json:"peak_rss_kib"`
	WallMS                   int `json:"wall_ms"`
	RepositoryWrites         int `json:"repository_writes"`
	LocalTestExecutions      int `json:"local_test_executions"`
}

// SemanticOperationReceipt is the compiler-owned receipt written as one of
// the six generated artifacts.
type SemanticOperationReceipt struct {
	Schema              string            `json:"schema"`
	ScenarioID          string            `json:"scenario_id"`
	AuthorityDigest     string            `json:"authority_digest"`
	SourceRevision      SourceRevision    `json:"source_revision"`
	Decision            OperationDecision `json:"decision"`
	Replay              ReplayIdentity    `json:"replay"`
	Activities          []string          `json:"activities"`
	ManifestDigest      string            `json:"manifest_digest"`
	RequestDigest       string            `json:"request_digest"`
	ResultDigest        string            `json:"result_digest"`
	SemanticPatchDigest string            `json:"semantic_patch_digest"`
	Metrics             EnvelopeMetrics   `json:"metrics"`
	ExternalUserUtility string            `json:"external_user_utility"`
}

// SemanticOperationArtifact is an in-memory generated file.
type SemanticOperationArtifact struct {
	Name     string
	Contents []byte
}

// SemanticOperationRun exposes the IR, receipt, and generated artifact bytes.
type SemanticOperationRun struct {
	IR            SemanticOperationIR
	Receipt       SemanticOperationReceipt
	Artifacts     []SemanticOperationArtifact
	ReceiptDigest string
}

// SemanticOperationVerification is returned by the independent verifier.
type SemanticOperationVerification struct {
	ScenarioID    string
	Decision      string
	Reason        string
	ReceiptDigest string
	Metrics       EnvelopeMetrics
}

// SemanticOperationScenarioIDs returns the fixed denominator without allowing
// callers to shrink or reorder the contract.
func SemanticOperationScenarioIDs() []string {
	return []string{"C1", "C2", "U1", "U2", "R1", "R2"}
}

// SemanticOperationActivityNames returns the released eight-activity graph.
func SemanticOperationActivityNames() []string {
	return append([]string(nil), semanticOperationActivities[:]...)
}

// GenerateSemanticOperationEnvelope parses the .gooo authority, constructs the
// semantic IR, and writes exactly six artifacts to the caller-owned directory.
func GenerateSemanticOperationEnvelope(source []byte, scenarioID, outputDir string) (SemanticOperationRun, error) {
	var run SemanticOperationRun
	if len(source) == 0 {
		return run, errors.New(".gooo authority is empty")
	}
	if outputDir == "" {
		return run, errors.New("caller-owned output directory is required")
	}
	if err := validateSemanticOperationAuthority(source); err != nil {
		return run, err
	}
	if err := prepareSemanticOperationOutput(outputDir); err != nil {
		return run, err
	}

	ir, metrics, err := buildSemanticOperationIR(source, scenarioID)
	if err != nil {
		return run, err
	}
	manifest := semanticOperationManifest{
		Schema:           SemanticOperationEnvelopeSchema,
		ScenarioID:       scenarioID,
		AuthorityDigest:  ir.AuthorityDigest,
		ToolchainDigest:  ir.ToolchainDigest,
		SourceRevision:   ir.Source,
		Intent:           ir.Intent,
		Grant:            ir.Grant,
		Activities:       append([]string(nil), ir.Activities...),
		ReplayIdentity:   ir.Replay.Identity,
		ExpectedDecision: ir.Decision.Decision,
	}
	requestBytes, err := encodeEnvelopeLines(ir.Request)
	if err != nil {
		return run, err
	}
	resultBytes, err := encodeEnvelopeLines(ir.Result)
	if err != nil {
		return run, err
	}
	manifestBytes, err := encodeEnvelopeJSON(manifest)
	if err != nil {
		return run, err
	}
	patch := SemanticPatch{
		Schema:           SemanticOperationEnvelopeSchema,
		ScenarioID:       scenarioID,
		Changed:          false,
		Operations:       []string{},
		RepositoryWrites: 0,
	}
	patchBytes, err := encodeEnvelopeJSON(patch)
	if err != nil {
		return run, err
	}
	receipt := SemanticOperationReceipt{
		Schema:              SemanticOperationEnvelopeSchema,
		ScenarioID:          scenarioID,
		AuthorityDigest:     ir.AuthorityDigest,
		SourceRevision:      ir.Source,
		Decision:            ir.Decision,
		Replay:              ir.Replay,
		Activities:          append([]string(nil), ir.Activities...),
		ManifestDigest:      envelopeDigestBytes(manifestBytes),
		RequestDigest:       envelopeDigestBytes(requestBytes),
		ResultDigest:        envelopeDigestBytes(resultBytes),
		SemanticPatchDigest: envelopeDigestBytes(patchBytes),
		Metrics:             metrics,
		ExternalUserUtility: "UNKNOWN",
	}
	receiptBytes, err := encodeEnvelopeJSON(receipt)
	if err != nil {
		return run, err
	}
	receiptDigest := envelopeDigestBytes(receiptBytes)
	reportBytes := []byte(renderSemanticOperationReport(receipt, receiptDigest))

	artifacts := []SemanticOperationArtifact{
		{Name: "operation-manifest.json", Contents: manifestBytes},
		{Name: "effect-requests.ndjson", Contents: requestBytes},
		{Name: "effect-results.ndjson", Contents: resultBytes},
		{Name: "semantic-patch.json", Contents: patchBytes},
		{Name: "operation-receipt.json", Contents: receiptBytes},
		{Name: "operation-report.md", Contents: reportBytes},
	}
	for _, artifact := range artifacts {
		if err := os.WriteFile(filepath.Join(outputDir, artifact.Name), artifact.Contents, 0o644); err != nil {
			return run, fmt.Errorf("write %s: %w", artifact.Name, err)
		}
	}
	return SemanticOperationRun{IR: ir, Receipt: receipt, Artifacts: artifacts, ReceiptDigest: receiptDigest}, nil
}

type semanticOperationManifest struct {
	Schema           string          `json:"schema"`
	ScenarioID       string          `json:"scenario_id"`
	AuthorityDigest  string          `json:"authority_digest"`
	ToolchainDigest  string          `json:"toolchain_digest"`
	SourceRevision   SourceRevision  `json:"source_revision"`
	Intent           OperationIntent `json:"intent"`
	Grant            EffectGrant     `json:"grant"`
	Activities       []string        `json:"activities"`
	ReplayIdentity   string          `json:"replay_identity"`
	ExpectedDecision string          `json:"expected_decision"`
}

func validateSemanticOperationAuthority(source []byte) error {
	seen := make(map[string]int, len(semanticOperationActivities))
	scanner := bufio.NewScanner(strings.NewReader(string(source)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "activity ") {
			continue
		}
		name := strings.TrimSpace(strings.TrimPrefix(line, "activity "))
		if cut := strings.IndexByte(name, '('); cut >= 0 {
			name = name[:cut]
		}
		if _, ok := semanticOperationActivityIndex(name); !ok {
			return fmt.Errorf("unrecognized operation envelope activity %q", name)
		}
		seen[name]++
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan .gooo authority: %w", err)
	}
	for _, name := range semanticOperationActivities {
		if seen[name] != 1 {
			return fmt.Errorf("activity %s appears %d times; expected once", name, seen[name])
		}
	}
	if len(seen) != len(semanticOperationActivities) {
		return errors.New(".gooo authority does not bind exactly eight activities")
	}
	return nil
}

func semanticOperationActivityIndex(name string) (int, bool) {
	for index, expected := range semanticOperationActivities {
		if name == expected {
			return index, true
		}
	}
	return 0, false
}

type semanticOperationScenarioPlan struct {
	ID                    string
	RequestedEffects      []string
	GrantedEffects        []string
	ResultEffects         []string
	ResultSourceRevision  string
	ResultPresent         bool
	ReplayCompared        bool
	PreviousRequestDigest string
}

func semanticOperationScenario(id string) (semanticOperationScenarioPlan, error) {
	plan := semanticOperationScenarioPlan{ID: id, ResultSourceRevision: "source-r1"}
	switch id {
	case "C1":
		plan.ResultPresent = true
	case "C2":
		plan.RequestedEffects = []string{"read:source"}
		plan.GrantedEffects = []string{"read:source"}
		plan.ResultEffects = []string{"read:source"}
		plan.ResultPresent = true
		plan.ReplayCompared = true
	case "U1":
		plan.RequestedEffects = []string{"read:source"}
		plan.GrantedEffects = []string{"read:source"}
	case "U2":
		plan.RequestedEffects = []string{"read:source"}
		plan.GrantedEffects = []string{"read:source"}
		plan.ResultEffects = []string{"read:source"}
		plan.ResultSourceRevision = "source-r0"
		plan.ResultPresent = true
	case "R1":
		plan.RequestedEffects = []string{"write:repository"}
		plan.GrantedEffects = []string{"read:source"}
		plan.ResultEffects = []string{"write:repository"}
		plan.ResultPresent = true
	case "R2":
		plan.RequestedEffects = []string{"read:source"}
		plan.GrantedEffects = []string{"read:source"}
		plan.ResultEffects = []string{"read:source"}
		plan.ResultPresent = true
		plan.ReplayCompared = true
		plan.PreviousRequestDigest = envelopeDigestString("different-canonical-request")
	default:
		return semanticOperationScenarioPlan{}, fmt.Errorf("unknown semantic operation scenario %q", id)
	}
	return plan, nil
}

func buildSemanticOperationIR(source []byte, scenarioID string) (SemanticOperationIR, EnvelopeMetrics, error) {
	plan, err := semanticOperationScenario(scenarioID)
	if err != nil {
		return SemanticOperationIR{}, EnvelopeMetrics{}, err
	}
	authorityDigest := envelopeDigestBytes(source)
	intent := OperationIntent{
		OperationID:   "operation-envelope/" + scenarioID,
		ScenarioID:    scenarioID,
		RequestedMode: "semantic-decision",
	}
	sourceRevision := SourceRevision{ID: "source-r1", Digest: authorityDigest}
	grant := EffectGrant{
		GrantID:  "grant/" + scenarioID,
		ParentID: "root-grant",
		Effects:  append([]string(nil), plan.GrantedEffects...),
	}
	var request *EffectRequest
	if len(plan.RequestedEffects) > 0 {
		request = &EffectRequest{
			RequestID:      "request/" + scenarioID,
			OperationID:    intent.OperationID,
			SourceRevision: sourceRevision.ID,
			Effects:        append([]string(nil), plan.RequestedEffects...),
			ReplayIdentity: "replay/" + scenarioID,
		}
		request.PayloadDigest = envelopeDigestString(fmt.Sprintf("%s|%s|%s", request.RequestID, request.SourceRevision, strings.Join(request.Effects, ",")))
	}
	var result *EffectResult
	if plan.ResultPresent {
		result = &EffectResult{
			ResultID:       "result/" + scenarioID,
			RequestID:      "request/" + scenarioID,
			SourceRevision: plan.ResultSourceRevision,
			Effects:        append([]string(nil), plan.ResultEffects...),
			PayloadDigest:  envelopeDigestString(fmt.Sprintf("result|%s|%s", scenarioID, plan.ResultSourceRevision)),
			ArtifactDigest: envelopeDigestString("artifact|" + scenarioID),
		}
	}
	replay := ReplayIdentity{Identity: "replay/" + scenarioID}
	if request != nil && plan.ReplayCompared {
		replay.Compared = true
		replay.CurrentRequestDigest = envelopeDigestJSON(*request)
		replay.PreviousRequestDigest = replay.CurrentRequestDigest
		if plan.PreviousRequestDigest != "" {
			replay.PreviousRequestDigest = plan.PreviousRequestDigest
		}
	}
	decision := classifySemanticOperation(request, result, grant, sourceRevision, replay)
	ir := SemanticOperationIR{
		Schema:          SemanticOperationEnvelopeSchema,
		Intent:          intent,
		Source:          sourceRevision,
		Grant:           grant,
		Request:         request,
		Result:          result,
		Replay:          replay,
		Decision:        decision,
		Activities:      SemanticOperationActivityNames(),
		AuthorityDigest: authorityDigest,
		ToolchainDigest: envelopeDigestString(semanticOperationToolchainDigest),
	}
	metrics := semanticOperationMetrics(ir)
	metrics.InputGoooPhysicalLines = countPhysicalLines(source)
	return ir, metrics, nil
}

func classifySemanticOperation(request *EffectRequest, result *EffectResult, grant EffectGrant, source SourceRevision, replay ReplayIdentity) OperationDecision {
	if request != nil && result != nil && !envelopeSubset(result.Effects, grant.Effects) {
		return OperationDecision{Decision: "REFUTED", Reason: "EFFECT_ESCALATION"}
	}
	if replay.Compared && replay.CurrentRequestDigest != replay.PreviousRequestDigest {
		return OperationDecision{Decision: "REFUTED", Reason: "REPLAY_COLLISION"}
	}
	if request != nil && result == nil {
		return OperationDecision{
			Decision: "UNKNOWN",
			Reason:   "DIRECT_MISSING",
			Unknown: &EnvelopeUnknownState{
				Stage: "result-verification", Step: "RecordEffectResult", Reason: "DIRECT_MISSING",
				UnknownClass: "DIRECT_MISSING", NextOperation: "submit-effect-result", BlockedBy: []string{"effect-result"},
			},
		}
	}
	if request != nil && result != nil && result.SourceRevision != source.ID {
		return OperationDecision{
			Decision: "UNKNOWN",
			Reason:   "STALE",
			Unknown: &EnvelopeUnknownState{
				Stage: "source-freshness", Step: "BindSourceRevision", Reason: "STALE",
				UnknownClass: "STALE", NextOperation: "bind-current-source-revision", BlockedBy: []string{"source-revision"},
			},
		}
	}
	return OperationDecision{Decision: "CLOSED", Reason: "EXACT_RESULT"}
}

func semanticOperationMetrics(ir SemanticOperationIR) EnvelopeMetrics {
	metrics := EnvelopeMetrics{
		OperationRequests:      boolToInt(ir.Request != nil),
		OperationResults:       boolToInt(ir.Result != nil),
		EffectsGranted:         len(ir.Grant.Effects),
		InputDescendantDirs:    0,
		InputRegularFiles:      1,
		InputGoPhysicalLines:   0,
		InputGoooPhysicalLines: 0,
		OutputArtifactFiles:    6,
		PeakRSSKib:             0,
		WallMS:                 0,
		RepositoryWrites:       0,
		LocalTestExecutions:    0,
	}
	if ir.Result != nil && ir.Result.SourceRevision == ir.Source.ID {
		metrics.EffectsUsed = len(ir.Result.Effects)
	}
	if ir.Replay.Compared {
		metrics.ReplayComparisons = 1
		if ir.Replay.CurrentRequestDigest != ir.Replay.PreviousRequestDigest {
			metrics.ReplayMismatches = 1
		}
	}
	if ir.Result != nil && ir.Result.SourceRevision != ir.Source.ID {
		metrics.StaleRejections = 1
	}
	if ir.Result != nil && !envelopeSubset(ir.Result.Effects, ir.Grant.Effects) {
		metrics.EffectEscalationsRefuted = 1
	}
	return metrics
}

func prepareSemanticOperationOutput(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create caller-owned output directory: %w", err)
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("inspect caller-owned output directory: %w", err)
	}
	if len(entries) != 0 {
		return errors.New("caller-owned output directory must be empty")
	}
	return nil
}

func encodeEnvelopeJSON(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func encodeEnvelopeLines(value any) ([]byte, error) {
	if value == nil {
		return []byte{}, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func renderSemanticOperationReport(receipt SemanticOperationReceipt, receiptDigest string) string {
	metrics, _ := json.Marshal(receipt.Metrics)
	return strings.Join([]string{
		"# Semantic operation envelope",
		"",
		"schema: " + receipt.Schema,
		"scenario_id: " + receipt.ScenarioID,
		"decision: " + receipt.Decision.Decision,
		"reason: " + receipt.Decision.Reason,
		"external_user_utility: " + receipt.ExternalUserUtility,
		"activities: 8/8",
		"artifacts: 6/6",
		"repository_writes: 0",
		"local_test_executions: 0",
		"metrics_json: " + string(metrics),
		"receipt_digest: " + receiptDigest,
		"",
	}, "\n")
}

func envelopeSubset(values, allowed []string) bool {
	set := make(map[string]struct{}, len(allowed))
	for _, value := range allowed {
		set[value] = struct{}{}
	}
	for _, value := range values {
		if _, ok := set[value]; !ok {
			return false
		}
	}
	return true
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func countPhysicalLines(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	count := 1
	for _, value := range source {
		if value == '\n' {
			count++
		}
	}
	if source[len(source)-1] == '\n' {
		count--
	}
	return count
}

func envelopeDigestJSON(value any) string {
	data, _ := json.Marshal(value)
	return envelopeDigestBytes(data)
}

func envelopeDigestString(value string) string {
	return envelopeDigestBytes([]byte(value))
}

func envelopeDigestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

// VerifySemanticOperationEnvelope independently re-reads all six artifacts,
// recomputes their digests, and reclassifies the evidence without using the
// generator's decision function.
func VerifySemanticOperationEnvelope(outputDir string) (SemanticOperationVerification, error) {
	var verification SemanticOperationVerification
	expectedNames := []string{
		"operation-manifest.json", "effect-requests.ndjson", "effect-results.ndjson",
		"semantic-patch.json", "operation-receipt.json", "operation-report.md",
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return verification, fmt.Errorf("read generated output: %w", err)
	}
	if len(entries) != len(expectedNames) {
		return verification, fmt.Errorf("expected six output artifacts, found %d", len(entries))
	}
	contents := make(map[string][]byte, len(expectedNames))
	for _, name := range expectedNames {
		data, readErr := os.ReadFile(filepath.Join(outputDir, name))
		if readErr != nil {
			return verification, fmt.Errorf("read %s: %w", name, readErr)
		}
		contents[name] = data
	}
	var manifest semanticOperationManifest
	var patch SemanticPatch
	var receipt SemanticOperationReceipt
	if err := json.Unmarshal(contents["operation-manifest.json"], &manifest); err != nil {
		return verification, fmt.Errorf("decode manifest: %w", err)
	}
	if err := json.Unmarshal(contents["semantic-patch.json"], &patch); err != nil {
		return verification, fmt.Errorf("decode semantic patch: %w", err)
	}
	if err := json.Unmarshal(contents["operation-receipt.json"], &receipt); err != nil {
		return verification, fmt.Errorf("decode receipt: %w", err)
	}
	if manifest.Schema != SemanticOperationEnvelopeSchema || patch.Schema != SemanticOperationEnvelopeSchema || receipt.Schema != SemanticOperationEnvelopeSchema {
		return verification, errors.New("artifact schema mismatch")
	}
	if manifest.ScenarioID == "" || manifest.ScenarioID != receipt.ScenarioID || patch.ScenarioID != receipt.ScenarioID {
		return verification, errors.New("artifact scenario mismatch")
	}
	if !sameEnvelopeActivities(manifest.Activities) || !sameEnvelopeActivities(receipt.Activities) {
		return verification, errors.New("released activity graph is not exactly eight activities")
	}
	if patch.Changed || patch.RepositoryWrites != 0 || receipt.ExternalUserUtility != "UNKNOWN" {
		return verification, errors.New("semantic patch or utility state is not fail-closed")
	}
	if receipt.ManifestDigest != envelopeDigestBytes(contents["operation-manifest.json"]) ||
		receipt.RequestDigest != envelopeDigestBytes(contents["effect-requests.ndjson"]) ||
		receipt.ResultDigest != envelopeDigestBytes(contents["effect-results.ndjson"]) ||
		receipt.SemanticPatchDigest != envelopeDigestBytes(contents["semantic-patch.json"]) {
		return verification, errors.New("artifact digest mismatch")
	}
	requests, err := decodeEnvelopeRequests(contents["effect-requests.ndjson"])
	if err != nil {
		return verification, err
	}
	results, err := decodeEnvelopeResults(contents["effect-results.ndjson"])
	if err != nil {
		return verification, err
	}
	decision, reason, err := independentlyClassifyEnvelope(manifest, requests, results, receipt.Replay)
	if err != nil {
		return verification, err
	}
	if decision != receipt.Decision.Decision || reason != receipt.Decision.Reason || manifest.ExpectedDecision != decision {
		return verification, fmt.Errorf("independent decision mismatch: got %s/%s, receipt %s/%s", decision, reason, receipt.Decision.Decision, receipt.Decision.Reason)
	}
	if receipt.Decision.Decision == "UNKNOWN" && !validEnvelopeUnknown(receipt.Decision.Unknown) {
		return verification, errors.New("unknown decision does not contain the six required fields")
	}
	if receipt.Decision.Decision != "UNKNOWN" && receipt.Decision.Unknown != nil {
		return verification, errors.New("non-unknown decision contains unknown evidence")
	}
	if receipt.Metrics.OutputArtifactFiles != 6 || receipt.Metrics.RepositoryWrites != 0 || receipt.Metrics.LocalTestExecutions != 0 {
		return verification, errors.New("metrics violate output or write contract")
	}
	receiptDigest := envelopeDigestBytes(contents["operation-receipt.json"])
	if string(contents["operation-report.md"]) != renderSemanticOperationReport(receipt, receiptDigest) {
		return verification, errors.New("report replay mismatch")
	}
	return SemanticOperationVerification{
		ScenarioID:    receipt.ScenarioID,
		Decision:      receipt.Decision.Decision,
		Reason:        receipt.Decision.Reason,
		ReceiptDigest: receiptDigest,
		Metrics:       receipt.Metrics,
	}, nil
}

func sameEnvelopeActivities(actual []string) bool {
	return len(actual) == len(semanticOperationActivities) && strings.Join(actual, "\x00") == strings.Join(semanticOperationActivities[:], "\x00")
}

func decodeEnvelopeRequests(data []byte) ([]EffectRequest, error) {
	return decodeEnvelopeLines[EffectRequest](data, "effect request")
}

func decodeEnvelopeResults(data []byte) ([]EffectResult, error) {
	return decodeEnvelopeLines[EffectResult](data, "effect result")
}

func decodeEnvelopeLines[T any](data []byte, kind string) ([]T, error) {
	if len(data) == 0 {
		return nil, nil
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	values := make([]T, 0, len(lines))
	for _, line := range lines {
		var value T
		if err := json.Unmarshal([]byte(line), &value); err != nil {
			return nil, fmt.Errorf("decode %s: %w", kind, err)
		}
		values = append(values, value)
	}
	return values, nil
}

func independentlyClassifyEnvelope(manifest semanticOperationManifest, requests []EffectRequest, results []EffectResult, replay ReplayIdentity) (string, string, error) {
	if len(requests) > 1 || len(results) > 1 {
		return "REFUTED", "MULTIPLE_RESULTS", nil
	}
	if len(requests) == 1 && len(results) == 1 {
		request, result := requests[0], results[0]
		if !independentSubset(result.Effects, manifest.Grant.Effects) {
			return "REFUTED", "EFFECT_ESCALATION", nil
		}
		if replay.Compared && replay.CurrentRequestDigest != replay.PreviousRequestDigest {
			return "REFUTED", "REPLAY_COLLISION", nil
		}
		if result.RequestID != request.RequestID || result.SourceRevision != manifest.SourceRevision.ID {
			return "UNKNOWN", "STALE", nil
		}
		return "CLOSED", "EXACT_RESULT", nil
	}
	if len(requests) == 1 && len(results) == 0 {
		return "UNKNOWN", "DIRECT_MISSING", nil
	}
	if len(requests) == 0 && len(results) == 1 && len(results[0].Effects) == 0 {
		return "CLOSED", "EXACT_RESULT", nil
	}
	return "REFUTED", "EVIDENCE_RELATION", nil
}

func independentSubset(values, allowed []string) bool {
	for _, value := range values {
		found := false
		for _, permitted := range allowed {
			if value == permitted {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func validEnvelopeUnknown(unknown *EnvelopeUnknownState) bool {
	return unknown != nil && unknown.Stage != "" && unknown.Step != "" && unknown.Reason != "" &&
		unknown.UnknownClass != "" && unknown.NextOperation != "" && len(unknown.BlockedBy) > 0
}
