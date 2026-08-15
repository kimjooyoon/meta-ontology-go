// Package shadow contains an independent, read-only oracle for the selective
// CI shadow decision. It deliberately does not import analyzer, planner,
// proof, lane, or command packages.
package shadow

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	CorpusSchema        = "gooo/selective-ci-shadow-corpus/v1"
	AnalyzerSchema      = "gooo/selective-ci-shadow-analyzer/v1"
	ManifestSchema      = "gooo/selective-ci-shadow-manifest/v1"
	PlannerSchema       = "gooo/selective-ci-shadow-planner/v1"
	ProofSchema         = "gooo/selective-ci-shadow-proof/v1"
	LaneSchema          = "gooo/selective-ci-shadow-lane/v1"
	ShadowSelective     = "SHADOW_SELECTIVE"
	FullSuiteFallback   = "FULL_SUITE_FALLBACK"
	StageInput          = "INPUT"
	StageSnapshot       = "SNAPSHOT_BINDING"
	StageRegistry       = "REGISTRY_BINDING"
	StagePlan           = "PLAN"
	StagePlanProof      = "PLAN_PROOF_BINDING"
	StageProofFail      = "PROOF_FAIL_CLOSED"
	StageProofUnknown   = "PROOF_UNKNOWN"
	StageLaneUnknown    = "LANE_UNKNOWN"
	StageLaneIneligible = "LANE_INELIGIBLE"
	StageSelective      = "SELECTIVE"
)

// Files are the five explicit JSON inputs consumed by the shadow command.
// Values remain raw strings so malformed and duplicate-key partitions cannot
// be normalized away by the corpus decoder.
type Files struct {
	AnalyzerBase string `json:"analyzer_base"`
	AnalyzerHead string `json:"analyzer_head"`
	Planner      string `json:"planner"`
	Proof        string `json:"proof"`
	Lane         string `json:"lane"`
}

type Case struct {
	Name      string `json:"name"`
	Partition string `json:"partition"`
	Files     Files  `json:"files"`
	Expected  Result `json:"expected"`
}

type Corpus struct {
	Schema string `json:"schema"`
	Cases  []Case `json:"cases"`
}

// Result is the complete observable decision vector. Fallback always carries
// empty selections and execution_authorized=false.
type Result struct {
	Status              string              `json:"status"`
	Stage               string              `json:"stage"`
	Reason              string              `json:"reason"`
	SelectedCommandIDs  []string            `json:"selected_command_ids"`
	SelectedGuardIDs    []string            `json:"selected_guard_command_ids"`
	SelectedWorkIDs     []string            `json:"selected_work_ids"`
	SelectedArgv        map[string][]string `json:"selected_argv"`
	ExecutionAuthorized bool                `json:"execution_authorized"`
	CanonicalDigest     string              `json:"canonical_digest"`
}

type analyzerSnapshot struct {
	Schema            string         `json:"schema"`
	Status            string         `json:"status"`
	FullSuiteFallback bool           `json:"full_suite_fallback"`
	RegistryDigest    string         `json:"registry_digest"`
	Files             []manifestFile `json:"files"`
	Digest            string         `json:"digest"`
}

type manifestFile struct {
	Path        string   `json:"path"`
	BlobDigest  string   `json:"blob_digest"`
	SemanticIDs []string `json:"semantic_ids"`
}

type plannerManifest struct {
	Schema string         `json:"schema"`
	Files  []manifestFile `json:"files"`
	Digest string         `json:"digest"`
}

type command struct {
	ID   string   `json:"id"`
	Argv []string `json:"argv"`
}

type plannerInput struct {
	Schema                  string          `json:"schema"`
	Status                  string          `json:"status"`
	RegistryDigest          string          `json:"registry_digest"`
	BaseManifest            plannerManifest `json:"base_manifest"`
	HeadManifest            plannerManifest `json:"head_manifest"`
	PlanDigest              string          `json:"plan_digest"`
	ChangedRootIDs          []string        `json:"changed_root_ids"`
	SelectedCommandIDs      []string        `json:"selected_command_ids"`
	SelectedGuardCommandIDs []string        `json:"selected_guard_command_ids"`
	SelectedWorkIDs         []string        `json:"selected_work_ids"`
	Commands                []command       `json:"commands"`
	GuardCommands           []command       `json:"guard_commands"`
}

type snapshotBinding struct {
	Source   string `json:"source"`
	Semantic string `json:"semantic"`
}

type proofInput struct {
	Schema             string         `json:"schema"`
	Status             string         `json:"status"`
	Fallback           string         `json:"fallback"`
	RegistryDigest     string         `json:"registry_digest"`
	PlanDigest         string         `json:"plan_digest"`
	Snapshots          proofSnapshots `json:"snapshots"`
	ChangedRootIDs     []string       `json:"changed_root_ids"`
	SelectedCommandIDs []string       `json:"selected_command_ids"`
	VerifiedCommandIDs []string       `json:"verified_command_ids"`
	ProofDigest        string         `json:"proof_digest"`
}

type proofSnapshots struct {
	Base snapshotBinding `json:"base"`
	Head snapshotBinding `json:"head"`
}

type laneInput struct {
	Schema            string   `json:"schema"`
	Decision          string   `json:"decision"`
	Reason            string   `json:"reason"`
	RegistryDigest    string   `json:"registry_digest"`
	BaseSHA           string   `json:"base_sha"`
	LaneHeadSHA       string   `json:"lane_head_sha"`
	LaneID            string   `json:"lane_id"`
	RegisteredBranch  string   `json:"registered_branch"`
	OwnedPathPrefixes []string `json:"owned_path_prefixes"`
	ChangedPaths      []string `json:"changed_paths"`
	AheadCount        int64    `json:"ahead_count"`
	BehindCount       int64    `json:"behind_count"`
	OpenPRCount       int64    `json:"open_pr_count"`
	ActiveLeaseCount  int64    `json:"active_lease_count"`
	CanonicalDigest   string   `json:"canonical_digest"`
}

type decodedInputs struct {
	base, head analyzerSnapshot
	planner    plannerInput
	proof      proofInput
	lane       laneInput
}

// Evaluate applies the contract's fixed precedence. It performs no process,
// filesystem, network, or argv execution side effect.
func Evaluate(c Case) Result {
	digest := caseDigest(c)
	inputs, err := decodeFiles(c.Files)
	if err != nil {
		return fallback(StageInput, err.Error(), digest)
	}
	if err := validateSnapshots(inputs.base, inputs.head, inputs.planner); err != nil {
		return fallback(StageSnapshot, err.Error(), digest)
	}
	if err := validateRegistry(inputs); err != nil {
		return fallback(StageRegistry, err.Error(), digest)
	}
	if err := validatePlan(inputs.base, inputs.head, inputs.planner); err != nil {
		return fallback(StagePlan, err.Error(), digest)
	}
	if err := validatePlanProofBinding(inputs); err != nil {
		return fallback(StagePlanProof, err.Error(), digest)
	}
	if inputs.proof.Status != "VERIFIED" || inputs.proof.Fallback != "NONE" || inputs.proof.ProofDigest != proofDigest(inputs.proof) {
		if inputs.proof.Status == "UNKNOWN" {
			return fallback(StageProofUnknown, "proof is UNKNOWN", digest)
		}
		return fallback(StageProofFail, "proof is not verified", digest)
	}
	if !validLaneFacts(inputs.lane) || inputs.lane.Schema != LaneSchema || inputs.lane.CanonicalDigest != laneDigest(inputs.lane) || inputs.lane.Decision == "UNKNOWN" {
		return fallback(StageLaneUnknown, "lane is UNKNOWN or stale", digest)
	}
	if inputs.lane.Decision == "INELIGIBLE" {
		return fallback(StageLaneIneligible, "lane is INELIGIBLE", digest)
	}
	if inputs.lane.Decision != "ELIGIBLE" || inputs.lane.Reason != "ELIGIBLE" {
		return fallback(StageLaneUnknown, "lane decision is malformed", digest)
	}

	selected, guards, work, argv := normalizedSelection(inputs.planner)
	return Result{
		Status: ShadowSelective, Stage: StageSelective, Reason: "all bindings verified",
		SelectedCommandIDs: selected, SelectedGuardIDs: guards, SelectedWorkIDs: work,
		SelectedArgv: argv, ExecutionAuthorized: true, CanonicalDigest: digest,
	}
}

func fallback(stage, reason, digest string) Result {
	return Result{Status: FullSuiteFallback, Stage: stage, Reason: reason,
		SelectedCommandIDs: []string{}, SelectedGuardIDs: []string{}, SelectedWorkIDs: []string{},
		SelectedArgv: map[string][]string{}, ExecutionAuthorized: false, CanonicalDigest: digest}
}

func decodeFiles(files Files) (decodedInputs, error) {
	var result decodedInputs
	if err := decodeStrict(files.AnalyzerBase, []string{"schema", "status", "full_suite_fallback", "registry_digest", "files", "digest"}, &result.base); err != nil {
		return decodedInputs{}, fmt.Errorf("analyzer base: %w", err)
	}
	if err := decodeStrict(files.AnalyzerHead, []string{"schema", "status", "full_suite_fallback", "registry_digest", "files", "digest"}, &result.head); err != nil {
		return decodedInputs{}, fmt.Errorf("analyzer head: %w", err)
	}
	if err := decodeStrict(files.Planner, []string{"schema", "status", "registry_digest", "base_manifest", "head_manifest", "plan_digest", "changed_root_ids", "selected_command_ids", "selected_guard_command_ids", "selected_work_ids", "commands", "guard_commands"}, &result.planner); err != nil {
		return decodedInputs{}, fmt.Errorf("planner: %w", err)
	}
	if err := decodeStrict(files.Proof, []string{"schema", "status", "fallback", "registry_digest", "plan_digest", "snapshots", "changed_root_ids", "selected_command_ids", "verified_command_ids", "proof_digest"}, &result.proof); err != nil {
		return decodedInputs{}, fmt.Errorf("proof: %w", err)
	}
	if err := decodeStrict(files.Lane, []string{"schema", "decision", "reason", "registry_digest", "base_sha", "lane_head_sha", "lane_id", "registered_branch", "owned_path_prefixes", "changed_paths", "ahead_count", "behind_count", "open_pr_count", "active_lease_count", "canonical_digest"}, &result.lane); err != nil {
		return decodedInputs{}, fmt.Errorf("lane: %w", err)
	}
	return result, nil
}

func decodeStrict(data string, required []string, target any) error {
	trimmed := bytes.TrimSpace([]byte(data))
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("top-level object required")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	if err := scanJSON(decoder); err != nil {
		return fmt.Errorf("strict JSON: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil || fields == nil {
		return fmt.Errorf("top-level object required")
	}
	for _, name := range required {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("missing required field %q", name)
		}
	}
	valueDecoder := json.NewDecoder(bytes.NewReader(trimmed))
	valueDecoder.DisallowUnknownFields()
	if err := valueDecoder.Decode(target); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return requireEOF(valueDecoder)
}

func scanJSON(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate object field %q", name)
			}
			seen[name] = struct{}{}
			if err := scanJSON(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := scanJSON(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return fmt.Errorf("trailing JSON data: %w", err)
	}
	return nil
}

func validateSnapshots(base, head analyzerSnapshot, planner plannerInput) error {
	if base.Schema != AnalyzerSchema || head.Schema != AnalyzerSchema || base.Status != "BOUND" || head.Status != "BOUND" || base.FullSuiteFallback || head.FullSuiteFallback {
		return fmt.Errorf("analyzer snapshots are not BOUND")
	}
	if !validSnapshotFacts(base) || !validSnapshotFacts(head) {
		return fmt.Errorf("analyzer snapshots are incomplete")
	}
	if base.Digest == "" || head.Digest == "" || base.Digest != analyzerDigest(base) || head.Digest != analyzerDigest(head) {
		return fmt.Errorf("analyzer snapshot digest is stale")
	}
	if planner.Schema != PlannerSchema || planner.BaseManifest.Schema != ManifestSchema || planner.HeadManifest.Schema != ManifestSchema {
		return fmt.Errorf("planner manifest schema is not bound")
	}
	if planner.BaseManifest.Files == nil || planner.HeadManifest.Files == nil || !validFiles(planner.BaseManifest.Files) || !validFiles(planner.HeadManifest.Files) {
		return fmt.Errorf("planner manifests are incomplete")
	}
	derivedBase, derivedHead := derivedManifest(base), derivedManifest(head)
	if !manifestEqual(planner.BaseManifest, derivedBase) || !manifestEqual(planner.HeadManifest, derivedHead) {
		return fmt.Errorf("planner manifests do not exactly match analyzer snapshots")
	}
	return nil
}

func validateRegistry(inputs decodedInputs) error {
	values := []string{inputs.base.RegistryDigest, inputs.head.RegistryDigest, inputs.planner.RegistryDigest, inputs.proof.RegistryDigest, inputs.lane.RegistryDigest}
	if values[0] == "" {
		return fmt.Errorf("registry digest is missing")
	}
	for _, value := range values[1:] {
		if value != values[0] {
			return fmt.Errorf("registry digest binding mismatch")
		}
	}
	return nil
}

func validatePlan(base, head analyzerSnapshot, planner plannerInput) error {
	if planner.Status != "SELECTIVE" || planner.PlanDigest == "" || planner.PlanDigest != planDigest(planner) {
		return fmt.Errorf("planner is not a sealed SELECTIVE result")
	}
	if planner.ChangedRootIDs == nil || planner.SelectedCommandIDs == nil || planner.SelectedGuardCommandIDs == nil || planner.SelectedWorkIDs == nil || planner.Commands == nil || planner.GuardCommands == nil {
		return fmt.Errorf("planner selection is incomplete")
	}
	if !equalStrings(planner.ChangedRootIDs, changedRoots(base, head)) {
		return fmt.Errorf("changed root IDs mismatch")
	}
	if err := validateCommands(planner); err != nil {
		return err
	}
	if len(planner.SelectedCommandIDs) == 0 && len(planner.SelectedGuardCommandIDs) == 0 {
		return fmt.Errorf("planner selected command union is empty")
	}
	return nil
}

func validateCommands(planner plannerInput) error {
	commands := map[string]command{}
	for _, item := range append(append([]command{}, planner.Commands...), planner.GuardCommands...) {
		if item.ID == "" || len(item.Argv) == 0 {
			return fmt.Errorf("planner command is incomplete")
		}
		if _, exists := commands[item.ID]; exists {
			return fmt.Errorf("planner command ID is duplicated")
		}
		commands[item.ID] = item
	}
	guardIDs := map[string]struct{}{}
	for _, item := range planner.GuardCommands {
		guardIDs[item.ID] = struct{}{}
	}
	seen := map[string]struct{}{}
	for _, id := range append(append([]string{}, planner.SelectedCommandIDs...), planner.SelectedGuardCommandIDs...) {
		if id == "" {
			return fmt.Errorf("planner selected command ID is empty")
		}
		if _, exists := commands[id]; !exists {
			return fmt.Errorf("planner selected command ID is dangling")
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("planner selected command union is duplicated")
		}
		seen[id] = struct{}{}
	}
	for _, id := range planner.SelectedCommandIDs {
		if _, guard := guardIDs[id]; guard {
			return fmt.Errorf("planner command is selected as both command and guard")
		}
	}
	return nil
}

func validatePlanProofBinding(inputs decodedInputs) error {
	proof := inputs.proof
	if proof.Schema != ProofSchema {
		return fmt.Errorf("proof schema mismatch")
	}
	if proof.Snapshots.Base.Source != inputs.base.Digest || proof.Snapshots.Base.Semantic != inputs.planner.BaseManifest.Digest || proof.Snapshots.Head.Source != inputs.head.Digest || proof.Snapshots.Head.Semantic != inputs.planner.HeadManifest.Digest {
		return fmt.Errorf("proof snapshot binding mismatch")
	}
	planner := inputs.planner
	selected, guards, _, _ := normalizedSelection(planner)
	union := append(append([]string{}, selected...), guards...)
	sort.Strings(union)
	if proof.PlanDigest != planner.PlanDigest || !equalStrings(proof.ChangedRootIDs, planner.ChangedRootIDs) || !equalStrings(proof.SelectedCommandIDs, union) || !equalStrings(proof.VerifiedCommandIDs, union) {
		return fmt.Errorf("proof plan binding mismatch")
	}
	return nil
}

func validSnapshotFacts(snapshot analyzerSnapshot) bool {
	return snapshot.RegistryDigest != "" && snapshot.Files != nil && validFiles(snapshot.Files)
}

func validFiles(values []manifestFile) bool {
	seenPaths := map[string]struct{}{}
	seenIDs := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value.Path) == "" || strings.TrimSpace(value.BlobDigest) == "" || value.SemanticIDs == nil || len(value.SemanticIDs) == 0 {
			return false
		}
		if _, exists := seenPaths[value.Path]; exists {
			return false
		}
		seenPaths[value.Path] = struct{}{}
		for _, id := range value.SemanticIDs {
			if strings.TrimSpace(id) == "" {
				return false
			}
			if _, exists := seenIDs[id]; exists {
				return false
			}
			seenIDs[id] = struct{}{}
		}
	}
	return true
}

func validLaneFacts(lane laneInput) bool {
	return validNonEmpty(lane.RegistryDigest, lane.BaseSHA, lane.LaneHeadSHA, lane.LaneID, lane.RegisteredBranch) &&
		lane.OwnedPathPrefixes != nil && lane.ChangedPaths != nil && lane.AheadCount >= 0 && lane.BehindCount >= 0 && lane.OpenPRCount >= 0 && lane.ActiveLeaseCount >= 0
}

func normalizedSelection(planner plannerInput) ([]string, []string, []string, map[string][]string) {
	selected := sortedCopy(planner.SelectedCommandIDs)
	guards := sortedCopy(planner.SelectedGuardCommandIDs)
	work := sortedCopy(planner.SelectedWorkIDs)
	commands := map[string][]string{}
	for _, item := range append(append([]command{}, planner.Commands...), planner.GuardCommands...) {
		commands[item.ID] = append([]string(nil), item.Argv...)
	}
	argv := map[string][]string{}
	for _, id := range append(append([]string{}, selected...), guards...) {
		argv[id] = append([]string(nil), commands[id]...)
	}
	return selected, guards, work, argv
}

func changedRoots(base, head analyzerSnapshot) []string {
	left, right := fileIndex(base.Files), fileIndex(head.Files)
	paths := map[string]struct{}{}
	for path := range left {
		paths[path] = struct{}{}
	}
	for path := range right {
		paths[path] = struct{}{}
	}
	ids := map[string]struct{}{}
	for path := range paths {
		before, beforeOK := left[path]
		after, afterOK := right[path]
		if beforeOK && afterOK && before.BlobDigest == after.BlobDigest && equalStrings(before.SemanticIDs, after.SemanticIDs) {
			continue
		}
		if beforeOK {
			for _, id := range before.SemanticIDs {
				ids[id] = struct{}{}
			}
		}
		if afterOK {
			for _, id := range after.SemanticIDs {
				ids[id] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

func derivedManifest(snapshot analyzerSnapshot) plannerManifest {
	files := normalizeFiles(snapshot.Files)
	result := plannerManifest{Schema: ManifestSchema, Files: files}
	result.Digest = manifestDigest(result)
	return result
}

func manifestEqual(left, right plannerManifest) bool {
	return left.Schema == right.Schema && left.Digest == right.Digest && equalFiles(left.Files, right.Files)
}

func analyzerDigest(snapshot analyzerSnapshot) string {
	value := struct {
		Schema            string         `json:"schema"`
		Status            string         `json:"status"`
		FullSuiteFallback bool           `json:"full_suite_fallback"`
		RegistryDigest    string         `json:"registry_digest"`
		Files             []manifestFile `json:"files"`
	}{snapshot.Schema, snapshot.Status, snapshot.FullSuiteFallback, snapshot.RegistryDigest, normalizeFiles(snapshot.Files)}
	return hashJSON(value)
}

func manifestDigest(manifest plannerManifest) string {
	value := struct {
		Schema string         `json:"schema"`
		Files  []manifestFile `json:"files"`
	}{manifest.Schema, normalizeFiles(manifest.Files)}
	return hashJSON(value)
}

func planDigest(planner plannerInput) string {
	value := struct {
		Schema                  string          `json:"schema"`
		Status                  string          `json:"status"`
		RegistryDigest          string          `json:"registry_digest"`
		BaseManifest            plannerManifest `json:"base_manifest"`
		HeadManifest            plannerManifest `json:"head_manifest"`
		ChangedRootIDs          []string        `json:"changed_root_ids"`
		SelectedCommandIDs      []string        `json:"selected_command_ids"`
		SelectedGuardCommandIDs []string        `json:"selected_guard_command_ids"`
		SelectedWorkIDs         []string        `json:"selected_work_ids"`
		Commands                []command       `json:"commands"`
		GuardCommands           []command       `json:"guard_commands"`
	}{planner.Schema, planner.Status, planner.RegistryDigest, normalizeManifest(planner.BaseManifest), normalizeManifest(planner.HeadManifest), sortedCopy(planner.ChangedRootIDs), sortedCopy(planner.SelectedCommandIDs), sortedCopy(planner.SelectedGuardCommandIDs), sortedCopy(planner.SelectedWorkIDs), normalizeCommands(planner.Commands), normalizeCommands(planner.GuardCommands)}
	return hashJSON(value)
}

func proofDigest(proof proofInput) string {
	value := struct {
		Schema             string         `json:"schema"`
		Status             string         `json:"status"`
		Fallback           string         `json:"fallback"`
		RegistryDigest     string         `json:"registry_digest"`
		PlanDigest         string         `json:"plan_digest"`
		Snapshots          proofSnapshots `json:"snapshots"`
		ChangedRootIDs     []string       `json:"changed_root_ids"`
		SelectedCommandIDs []string       `json:"selected_command_ids"`
		VerifiedCommandIDs []string       `json:"verified_command_ids"`
	}{proof.Schema, proof.Status, proof.Fallback, proof.RegistryDigest, proof.PlanDigest, proof.Snapshots, sortedCopy(proof.ChangedRootIDs), sortedCopy(proof.SelectedCommandIDs), sortedCopy(proof.VerifiedCommandIDs)}
	return hashJSON(value)
}

func laneDigest(lane laneInput) string {
	value := struct {
		Schema            string   `json:"schema"`
		Decision          string   `json:"decision"`
		Reason            string   `json:"reason"`
		RegistryDigest    string   `json:"registry_digest"`
		BaseSHA           string   `json:"base_sha"`
		LaneHeadSHA       string   `json:"lane_head_sha"`
		LaneID            string   `json:"lane_id"`
		RegisteredBranch  string   `json:"registered_branch"`
		OwnedPathPrefixes []string `json:"owned_path_prefixes"`
		ChangedPaths      []string `json:"changed_paths"`
		AheadCount        int64    `json:"ahead_count"`
		BehindCount       int64    `json:"behind_count"`
		OpenPRCount       int64    `json:"open_pr_count"`
		ActiveLeaseCount  int64    `json:"active_lease_count"`
	}{lane.Schema, lane.Decision, lane.Reason, lane.RegistryDigest, lane.BaseSHA, lane.LaneHeadSHA, lane.LaneID, lane.RegisteredBranch, sortedCopy(lane.OwnedPathPrefixes), sortedCopy(lane.ChangedPaths), lane.AheadCount, lane.BehindCount, lane.OpenPRCount, lane.ActiveLeaseCount}
	return hashJSON(value)
}

func caseDigest(c Case) string {
	value := struct {
		Name      string `json:"name"`
		Partition string `json:"partition"`
		Files     Files  `json:"files"`
		Result    Result `json:"result"`
	}{c.Name, c.Partition, c.Files, resultWithoutDigest(c.Expected)}
	return hashJSON(value)
}

func resultWithoutDigest(result Result) Result { result.CanonicalDigest = ""; return result }

func hashJSON(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func normalizeManifest(value plannerManifest) plannerManifest {
	value.Files = normalizeFiles(value.Files)
	return value
}

func normalizeFiles(values []manifestFile) []manifestFile {
	result := append([]manifestFile(nil), values...)
	for i := range result {
		result[i].SemanticIDs = sortedCopy(result[i].SemanticIDs)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}

func normalizeCommands(values []command) []command {
	result := append([]command(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	for i := range result {
		result[i].Argv = append([]string(nil), result[i].Argv...)
	}
	return result
}

func fileIndex(values []manifestFile) map[string]manifestFile {
	result := map[string]manifestFile{}
	for _, value := range values {
		result[value.Path] = value
	}
	return result
}

func equalFiles(left, right []manifestFile) bool {
	return stringJSON(normalizeFiles(left)) == stringJSON(normalizeFiles(right))
}

func stringJSON(value any) string { data, _ := json.Marshal(value); return string(data) }

func equalStrings(left, right []string) bool {
	return stringJSON(sortedCopy(left)) == stringJSON(sortedCopy(right))
}

func sortedCopy(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func validNonEmpty(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}
