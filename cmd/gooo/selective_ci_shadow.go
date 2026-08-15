package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const selectiveCIShadowUsage = "usage: gooo selective-ci shadow --base-snapshot FILE --head-snapshot FILE --plan-input FILE --evidence-input FILE --lane-input FILE"

const selectiveCIShadowSchemaVersion = "gooo/selective-ci-shadow/v1"

type selectiveCIShadowOptions struct {
	baseSnapshot  string
	headSnapshot  string
	planInput     string
	evidenceInput string
	laneInput     string
}

type shadowInputFiles struct {
	baseSnapshot  []byte
	headSnapshot  []byte
	planInput     []byte
	evidenceInput []byte
	laneInput     []byte
}

type shadowCommandSpec struct {
	ID   string   `json:"id"`
	Argv []string `json:"argv"`
}

type shadowResourceReceipt struct {
	CommandID    string `json:"command_id"`
	CPUWorkUnits uint64 `json:"cpu_work_units"`
	MemoryBytes  uint64 `json:"memory_bytes"`
}

type shadowLaneReceipt struct {
	Decision       string `json:"decision"`
	Reason         string `json:"reason"`
	RegistryDigest string `json:"registry_digest"`
	BaseSHA        string `json:"base_sha"`
	LaneHeadSHA    string `json:"lane_head_sha"`
	LaneID         string `json:"lane_id"`
}

// selectiveCIShadowOutput is intentionally a receipt, not an execution plan.
// It contains argv as data only; no field authorizes or represents execution.
type selectiveCIShadowOutput struct {
	SchemaVersion       string                  `json:"schema_version"`
	Command             string                  `json:"command"`
	Status              string                  `json:"status"`
	Stage               string                  `json:"stage"`
	Component           string                  `json:"component"`
	Reason              string                  `json:"reason"`
	ExecutionAuthorized bool                    `json:"execution_authorized"`
	ShadowOnly          bool                    `json:"shadow_only"`
	BaseSourceDigest    string                  `json:"base_source_digest"`
	HeadSourceDigest    string                  `json:"head_source_digest"`
	BaseSemanticDigest  string                  `json:"base_semantic_digest"`
	HeadSemanticDigest  string                  `json:"head_semantic_digest"`
	RegistryDigest      string                  `json:"registry_digest"`
	PlanDigest          string                  `json:"plan_digest"`
	ProofStatus         string                  `json:"proof_status"`
	ProofCode           string                  `json:"proof_code"`
	ChangedSemanticIDs  []string                `json:"changed_semantic_ids"`
	SelectedCommands    []shadowCommandSpec     `json:"selected_commands"`
	SelectedGuards      []shadowCommandSpec     `json:"selected_guard_commands"`
	SelectedWorkIDs     []string                `json:"selected_work_ids"`
	ResourceReceipts    []shadowResourceReceipt `json:"resource_receipts"`
	Lane                shadowLaneReceipt       `json:"lane"`
	CanonicalDigest     string                  `json:"canonical_digest"`
}

func parseSelectiveCIShadowArguments(args []string) (selectiveCIShadowOptions, error) {
	var options selectiveCIShadowOptions
	seen := map[string]bool{}
	for index := 0; index < len(args); index++ {
		flag := args[index]
		var target *string
		switch flag {
		case "--base-snapshot":
			target = &options.baseSnapshot
		case "--head-snapshot":
			target = &options.headSnapshot
		case "--plan-input":
			target = &options.planInput
		case "--evidence-input":
			target = &options.evidenceInput
		case "--lane-input":
			target = &options.laneInput
		default:
			return selectiveCIShadowOptions{}, errors.New(selectiveCIShadowUsage)
		}
		if seen[flag] || index+1 >= len(args) || args[index+1] == "" || strings.HasPrefix(args[index+1], "-") {
			return selectiveCIShadowOptions{}, errors.New(selectiveCIShadowUsage)
		}
		seen[flag] = true
		*target = args[index+1]
		index++
	}
	if options.baseSnapshot == "" || options.headSnapshot == "" || options.planInput == "" || options.evidenceInput == "" || options.laneInput == "" {
		return selectiveCIShadowOptions{}, errors.New(selectiveCIShadowUsage)
	}
	return options, nil
}

func runSelectiveCI(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "shadow" {
		fmt.Fprintln(stderr, selectiveCIShadowUsage)
		return exitUsage
	}
	options, err := parseSelectiveCIShadowArguments(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	files, missing := readSelectiveCIShadowFiles(options, reader)
	if missing != "" {
		// File availability is a CLI contract error. Do not emit a partial
		// semantic receipt when one of the explicit inputs cannot be read.
		fmt.Fprintf(stderr, "gooo: cli.usage: missing %s input file\n%s\n", missing, selectiveCIShadowUsage)
		return exitUsage
	}
	output := evaluateSelectiveCIShadow(files)
	data, err := encodeSelectiveCIShadowOutput(output)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: selective-ci shadow: output encoding failed: %v\n", err)
		return exitFailure
	}
	if _, err := stdout.Write(data); err != nil {
		return exitFailure
	}
	return exitOK
}

func readSelectiveCIShadowFiles(options selectiveCIShadowOptions, reader SourceReader) (shadowInputFiles, string) {
	read := func(name, filename string) ([]byte, string) {
		data, err := reader.ReadFile(filename)
		if err != nil {
			return nil, name
		}
		return data, ""
	}
	base, missing := read("base_snapshot", options.baseSnapshot)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	head, missing := read("head_snapshot", options.headSnapshot)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	plan, missing := read("plan_input", options.planInput)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	evidence, missing := read("evidence_input", options.evidenceInput)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	lane, missing := read("lane_input", options.laneInput)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	return shadowInputFiles{baseSnapshot: base, headSnapshot: head, planInput: plan, evidenceInput: evidence, laneInput: lane}, ""
}

func newSelectiveCIShadowOutput() selectiveCIShadowOutput {
	return selectiveCIShadowOutput{
		SchemaVersion:       selectiveCIShadowSchemaVersion,
		Command:             "selective-ci shadow",
		ExecutionAuthorized: false,
		ShadowOnly:          true,
		ChangedSemanticIDs:  []string{},
		SelectedCommands:    []shadowCommandSpec{},
		SelectedGuards:      []shadowCommandSpec{},
		SelectedWorkIDs:     []string{},
		ResourceReceipts:    []shadowResourceReceipt{},
	}
}

func evaluateSelectiveCIShadow(files shadowInputFiles) selectiveCIShadowOutput {
	output := newSelectiveCIShadowOutput()

	base, err := analyzersci.DecodeSnapshot(files.baseSnapshot)
	if err != nil {
		return shadowFallback(output, "INPUT", "base_snapshot", shadowDecodeReason(err))
	}
	head, err := analyzersci.DecodeSnapshot(files.headSnapshot)
	if err != nil {
		return shadowFallback(output, "INPUT", "head_snapshot", shadowDecodeReason(err))
	}
	planInput, err := plannersci.DecodeJSON(files.planInput)
	if err != nil {
		return shadowFallback(output, "INPUT", "plan_input", shadowDecodeReason(err))
	}
	proofInput, err := proofsci.DecodeInput(files.evidenceInput)
	if err != nil {
		return shadowFallback(output, "INPUT", "evidence_input", shadowDecodeReason(err))
	}
	laneInput, err := lanesci.DecodeJSON(files.laneInput)
	if err != nil {
		return shadowFallback(output, "INPUT", "lane_input", shadowDecodeReason(err))
	}

	output.BaseSourceDigest = base.Digest
	output.HeadSourceDigest = head.Digest
	output.RegistryDigest = planInput.Registry.Digest
	proof := proofsci.Evaluate(proofInput)
	output.ProofStatus = string(proof.Status)
	output.ProofCode = proof.Code
	lane := lanesci.Classify(laneInput)
	output.Lane = shadowLaneReceipt{
		Decision: string(lane.Decision), Reason: string(lane.Reason),
		RegistryDigest: lane.RegistryDigest, BaseSHA: lane.BaseSHA,
		LaneHeadSHA: lane.LaneHeadSHA, LaneID: lane.LaneID,
	}

	baseManifest, err := plannerManifestFromAnalyzerSnapshot(base)
	if err != nil {
		return shadowFallback(output, "SNAPSHOT_BINDING", "base_manifest", "DERIVATION_FAILED")
	}
	headManifest, err := plannerManifestFromAnalyzerSnapshot(head)
	if err != nil {
		return shadowFallback(output, "SNAPSHOT_BINDING", "head_manifest", "DERIVATION_FAILED")
	}
	output.BaseSemanticDigest = baseManifest.Digest
	output.HeadSemanticDigest = headManifest.Digest
	if !reflect.DeepEqual(planInput.Base, baseManifest) {
		return shadowFallback(output, "SNAPSHOT_BINDING", "base_manifest", "MANIFEST_MISMATCH")
	}
	if !reflect.DeepEqual(planInput.Head, headManifest) {
		return shadowFallback(output, "SNAPSHOT_BINDING", "head_manifest", "MANIFEST_MISMATCH")
	}

	if rawDigest(base.RegistryDigest) != planInput.Registry.Digest {
		return shadowFallback(output, "REGISTRY_BINDING", "base_snapshot", "REGISTRY_DIGEST_MISMATCH")
	}
	if rawDigest(head.RegistryDigest) != planInput.Registry.Digest {
		return shadowFallback(output, "REGISTRY_BINDING", "head_snapshot", "REGISTRY_DIGEST_MISMATCH")
	}

	plan := plannersci.Plan(planInput)
	output.PlanDigest = plan.CanonicalDigest
	if plan.Status != plannersci.StatusSelective {
		return shadowFallback(output, "PLAN", "planner", plan.ReasonCode)
	}

	expectedProofSnapshots := proofsci.SnapshotBinding{
		Base: semantic.SnapshotDigests{Source: rawDigest(base.Digest), Semantic: baseManifest.Digest},
		Head: semantic.SnapshotDigests{Source: rawDigest(head.Digest), Semantic: headManifest.Digest},
	}
	expectedSelected := sortedUnion(plan.SelectedCommandIDs, plan.SelectedGuardCommandIDs)
	if proofInput.RegistryDigest != planInput.Registry.Digest {
		return shadowFallback(output, "PLAN_PROOF_BINDING", "registry_digest", "PROOF_REGISTRY_DIGEST_MISMATCH")
	}
	if proofInput.PlanDigest != plan.CanonicalDigest {
		return shadowFallback(output, "PLAN_PROOF_BINDING", "plan_digest", "PROOF_PLAN_DIGEST_MISMATCH")
	}
	if !reflect.DeepEqual(sortedSemanticIDs(proofInput.ChangedRootIDs), plan.ChangedSemanticIDs) {
		return shadowFallback(output, "PLAN_PROOF_BINDING", "changed_root_ids", "PROOF_CHANGED_ROOT_IDS_MISMATCH")
	}
	if !reflect.DeepEqual(sortedSemanticIDs(proofInput.SelectedCommandIDs), expectedSelected) {
		return shadowFallback(output, "PLAN_PROOF_BINDING", "selected_command_ids", "PROOF_SELECTED_COMMAND_IDS_MISMATCH")
	}
	if proofInput.Snapshots != expectedProofSnapshots {
		return shadowFallback(output, "PLAN_PROOF_BINDING", "snapshots", "PROOF_SNAPSHOT_BINDING_MISMATCH")
	}

	switch proof.Status {
	case proofsci.FailClosed:
		return shadowFallback(output, "PROOF_FAIL_CLOSED", "proof", proof.Code)
	case proofsci.Unknown:
		return shadowFallback(output, "PROOF_UNKNOWN", "proof", proof.Code)
	case proofsci.Verified:
		// Continue to the lane gate.
	default:
		return shadowFallback(output, "PROOF_FAIL_CLOSED", "proof", "INVALID_PROOF_STATUS")
	}

	switch lane.Decision {
	case lanesci.DecisionUnknown:
		return shadowFallback(output, "LANE_UNKNOWN", "lane", string(lane.Reason))
	case lanesci.DecisionIneligible:
		return shadowFallback(output, "LANE_INELIGIBLE", "lane", string(lane.Reason))
	case lanesci.DecisionEligible:
		// Continue to the sealed selective receipt.
	default:
		return shadowFallback(output, "LANE_UNKNOWN", "lane", "INVALID_LANE_DECISION")
	}

	commands, guards, receipts, err := selectedShadowCommands(plan, planInput.Registry)
	if err != nil {
		return shadowFallback(output, "PLAN", "selected_commands", "MISSING_COMMAND_SPEC")
	}
	output.Status = "SHADOW_SELECTIVE"
	output.Stage = "SELECTIVE"
	output.Component = "all"
	output.Reason = "VERIFIED"
	output.ChangedSemanticIDs = append([]string{}, plan.ChangedSemanticIDs...)
	output.SelectedCommands = commands
	output.SelectedGuards = guards
	output.SelectedWorkIDs = append([]string{}, plan.SelectedWorkIDs...)
	output.ResourceReceipts = receipts
	return sealSelectiveCIShadowOutput(output)
}

func plannerManifestFromAnalyzerSnapshot(snapshot analyzersci.Snapshot) (plannersci.SnapshotManifest, error) {
	files := make([]plannersci.SnapshotFile, 0, len(snapshot.Sources))
	for _, source := range snapshot.Sources {
		ids := make([]string, 0, len(source.Bindings))
		seen := map[string]struct{}{}
		for _, binding := range source.Bindings {
			if binding.Status != analyzersci.StatusBound || binding.ID == "" {
				return plannersci.SnapshotManifest{}, errors.New("source binding is not BOUND")
			}
			if _, exists := seen[binding.ID]; exists {
				return plannersci.SnapshotManifest{}, errors.New("duplicate source binding")
			}
			seen[binding.ID] = struct{}{}
			ids = append(ids, binding.ID)
		}
		sort.Strings(ids)
		files = append(files, plannersci.SnapshotFile{Path: source.Path, BlobDigest: rawDigest(source.BlobDigest), SemanticIDs: ids})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	manifest := plannersci.SnapshotManifest{SchemaVersion: plannersci.ManifestSchemaVersion, Files: files}
	manifest.Digest = manifest.ComputedDigest()
	if err := manifest.Validate(); err != nil {
		return plannersci.SnapshotManifest{}, err
	}
	return manifest, nil
}

// rawDigest bridges the analyzer's labelled SHA-256 spelling to the raw-hex
// spelling used by the planner, proof, and lane contracts. It does not hash or
// otherwise alter the digest identity.
func rawDigest(value string) string {
	return strings.TrimPrefix(value, "sha256:")
}

func selectedShadowCommands(plan plannersci.PlanResult, registry plannersci.Registry) ([]shadowCommandSpec, []shadowCommandSpec, []shadowResourceReceipt, error) {
	commands := map[string]plannersci.Command{}
	for _, command := range registry.Commands {
		commands[command.ID] = command
	}
	guards := map[string]plannersci.Command{}
	for _, command := range registry.GlobalGuardCommands {
		guards[command.ID] = command
	}
	makeSpecs := func(ids []string, source map[string]plannersci.Command) ([]shadowCommandSpec, []shadowResourceReceipt, error) {
		specs := make([]shadowCommandSpec, 0, len(ids))
		receipts := make([]shadowResourceReceipt, 0, len(ids))
		for _, id := range ids {
			command, ok := source[id]
			if !ok {
				return nil, nil, errors.New("selected command is not registered")
			}
			specs = append(specs, shadowCommandSpec{ID: command.ID, Argv: append([]string{}, command.Argv...)})
			receipts = append(receipts, shadowResourceReceipt{CommandID: command.ID, CPUWorkUnits: command.CPUWorkUnits, MemoryBytes: command.MemoryBytes})
		}
		sort.Slice(specs, func(i, j int) bool { return specs[i].ID < specs[j].ID })
		sort.Slice(receipts, func(i, j int) bool { return receipts[i].CommandID < receipts[j].CommandID })
		return specs, receipts, nil
	}
	commandSpecs, commandReceipts, err := makeSpecs(plan.SelectedCommandIDs, commands)
	if err != nil {
		return nil, nil, nil, err
	}
	guardSpecs, guardReceipts, err := makeSpecs(plan.SelectedGuardCommandIDs, guards)
	if err != nil {
		return nil, nil, nil, err
	}
	receipts := append(commandReceipts, guardReceipts...)
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].CommandID < receipts[j].CommandID })
	return commandSpecs, guardSpecs, receipts, nil
}

func sortedUnion(left, right []string) []string {
	values := append(append([]string{}, left...), right...)
	sort.Strings(values)
	return uniqueStrings(values)
}

func sortedSemanticIDs(values []semantic.ID) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	sort.Strings(result)
	return result
}

func uniqueStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func shadowFallback(output selectiveCIShadowOutput, stage, component, reason string) selectiveCIShadowOutput {
	output.Status = "FULL_SUITE_FALLBACK"
	output.Stage = stage
	output.Component = component
	output.Reason = reason
	output.ExecutionAuthorized = false
	output.ShadowOnly = true
	output.ChangedSemanticIDs = []string{}
	output.SelectedCommands = []shadowCommandSpec{}
	output.SelectedGuards = []shadowCommandSpec{}
	output.SelectedWorkIDs = []string{}
	output.ResourceReceipts = []shadowResourceReceipt{}
	return sealSelectiveCIShadowOutput(output)
}

func shadowDecodeReason(err error) string {
	var snapshotErr *analyzersci.Error
	if errors.As(err, &snapshotErr) {
		return string(snapshotErr.Code)
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "duplicate"):
		return "DUPLICATE_FIELD"
	case strings.Contains(message, "unknown field") || strings.Contains(message, "unknown"):
		return "UNKNOWN_FIELD"
	case strings.Contains(message, "trailing") || strings.Contains(message, "multiple"):
		return "TRAILING_DATA"
	case strings.Contains(message, "stale") || strings.Contains(message, "mismatch"):
		return "STALE_OR_MISMATCHED"
	default:
		return "MALFORMED"
	}
}

func (output selectiveCIShadowOutput) canonicalJSON() ([]byte, error) {
	copy := normalizeSelectiveCIShadowOutput(output)
	copy.CanonicalDigest = ""
	return json.Marshal(copy)
}

func (output selectiveCIShadowOutput) stableDigest() string {
	data, err := output.canonicalJSON()
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func sealSelectiveCIShadowOutput(output selectiveCIShadowOutput) selectiveCIShadowOutput {
	output = normalizeSelectiveCIShadowOutput(output)
	output.CanonicalDigest = output.stableDigest()
	return output
}

func normalizeSelectiveCIShadowOutput(output selectiveCIShadowOutput) selectiveCIShadowOutput {
	if output.ChangedSemanticIDs == nil {
		output.ChangedSemanticIDs = []string{}
	}
	if output.SelectedCommands == nil {
		output.SelectedCommands = []shadowCommandSpec{}
	}
	if output.SelectedGuards == nil {
		output.SelectedGuards = []shadowCommandSpec{}
	}
	if output.SelectedWorkIDs == nil {
		output.SelectedWorkIDs = []string{}
	}
	if output.ResourceReceipts == nil {
		output.ResourceReceipts = []shadowResourceReceipt{}
	}
	sort.Strings(output.ChangedSemanticIDs)
	output.ChangedSemanticIDs = uniqueStrings(output.ChangedSemanticIDs)
	sort.Slice(output.SelectedCommands, func(i, j int) bool { return output.SelectedCommands[i].ID < output.SelectedCommands[j].ID })
	sort.Slice(output.SelectedGuards, func(i, j int) bool { return output.SelectedGuards[i].ID < output.SelectedGuards[j].ID })
	sort.Strings(output.SelectedWorkIDs)
	output.SelectedWorkIDs = uniqueStrings(output.SelectedWorkIDs)
	sort.Slice(output.ResourceReceipts, func(i, j int) bool {
		return output.ResourceReceipts[i].CommandID < output.ResourceReceipts[j].CommandID
	})
	return output
}

func encodeSelectiveCIShadowOutput(output selectiveCIShadowOutput) ([]byte, error) {
	output = sealSelectiveCIShadowOutput(output)
	data, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
