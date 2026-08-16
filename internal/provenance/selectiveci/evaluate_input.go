package selectiveci

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func normalizeInput(raw Input) (Input, issueState) {
	input, state := normalizeHeader(raw)
	input.Paths = normalizePaths(raw.Paths, &state)
	input.CommandReceipts = normalizeReceipts(raw.CommandReceipts, input, &state)
	return normalizePathEvidence(raw, input, &state)
}

func normalizeHeader(raw Input) (Input, issueState) {
	input, state := raw, issueState{}
	if raw.Schema == "" {
		state.add(issueUnknown, CodeMissing)
	} else if raw.Schema != SchemaVersion {
		state.add(issueFailClosed, CodeMalformed)
	}
	input.Snapshots.Base = normalizeSnapshot(raw.Snapshots.Base, "base", &state)
	input.Snapshots.Head = normalizeSnapshot(raw.Snapshots.Head, "head", &state)
	input.RegistryDigest = normalizeBoundDigest(raw.RegistryDigest, "registry", &state)
	input.PlanDigest = normalizeBoundDigest(raw.PlanDigest, "plan", &state)
	input.ChangedRootIDs = normalizeIDSet(raw.ChangedRootIDs, "changed root ID", &state)
	input.SelectedCommandIDs = normalizeIDSet(raw.SelectedCommandIDs, "selected command ID", &state)
	input.ObligationIDs = normalizeIDSet(raw.ObligationIDs, "obligation ID", &state)
	input.EvidenceIDs = normalizeIDSet(raw.EvidenceIDs, "evidence ID", &state)
	return input, state
}

func normalizeBoundDigest(value, label string, state *issueState) string {
	if value == "" {
		state.add(issueUnknown, CodeMissing)
		return ""
	}
	value, err := normalizeDigest(value, label+" digest")
	if err != nil {
		state.add(issueFailClosed, CodeDigestMismatch)
		return ""
	}
	return value
}

func normalizeIDSet(values []semantic.ID, label string, state *issueState) []semantic.ID {
	ids, err := normalizeIDs(values, label)
	if err != nil {
		state.add(issueFailClosed, CodeDuplicate)
		return append([]semantic.ID(nil), values...)
	}
	if len(ids) == 0 {
		state.add(issueUnknown, CodeMissing)
	}
	return ids
}

func normalizePaths(raw []Path, state *issueState) []Path {
	paths := make([]Path, 0, len(raw))
	seen := make(map[semantic.ID]struct{}, len(raw))
	for _, rawPath := range raw {
		path, err := normalizePath(rawPath)
		if err != nil {
			state.add(issueFailClosed, CodeMalformed)
			continue
		}
		if _, exists := seen[path.PathID]; exists {
			state.add(issueFailClosed, CodeDuplicate)
		}
		seen[path.PathID] = struct{}{}
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool { return paths[i].PathID < paths[j].PathID })
	if len(paths) == 0 {
		state.add(issueUnknown, CodeMissing)
	}
	return paths
}

func normalizeReceipts(raw []CommandReceipt, input Input, state *issueState) []CommandReceipt {
	receipts := make([]CommandReceipt, 0, len(raw))
	commands, receiptIDs := map[semantic.ID]struct{}{}, map[semantic.ID]struct{}{}
	for _, rawReceipt := range raw {
		receipt := normalizeCommandReceipt(rawReceipt, input, state)
		if _, exists := commands[receipt.CommandID]; exists {
			state.add(issueFailClosed, CodeDuplicate)
		}
		if _, exists := receiptIDs[receipt.ReceiptID]; exists {
			state.add(issueFailClosed, CodeDuplicate)
		}
		commands[receipt.CommandID], receiptIDs[receipt.ReceiptID] = struct{}{}, struct{}{}
		receipts = append(receipts, receipt)
	}
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].CommandID < receipts[j].CommandID })
	if len(receipts) == 0 {
		state.add(issueUnknown, CodeMissing)
	}
	return receipts
}

func normalizePathEvidence(raw Input, input Input, state *issueState) (Input, issueState) {
	normalized, err := raw.InferencePath.Normalized()
	if err != nil {
		state.add(classifyInferencePathError(err), classifyInferencePathCode(err))
		return input, *state
	}
	input.InferencePath = normalized
	for _, edge := range normalized.Edges {
		if edge.Kind == semantic.InferenceObservationCandidate {
			state.add(issueUnknown, CodeCandidate)
		}
	}
	if err := bindEvidenceIDs(input); err != nil {
		state.add(classifyBindingError(err), classifyBindingCode(err))
	}
	if err := bindSnapshots(input); err != nil {
		state.add(classifyBindingError(err), classifyBindingCode(err))
	}
	return input, *state
}

func classifyInferencePathError(err error) issueClass {
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "stable-id-collision") || strings.Contains(message, "duplicate") || strings.Contains(message, "stale-evidence") {
		return issueFailClosed
	}
	if strings.Contains(message, "evidence") || strings.Contains(message, "snapshot") || strings.Contains(message, "candidate") {
		return issueUnknown
	}
	return issueFailClosed
}

func classifyInferencePathCode(err error) string {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "stable-id-collision"), strings.Contains(message, "duplicate"):
		return CodeDuplicate
	case strings.Contains(message, "stale-evidence"):
		return CodeDigestMismatch
	case strings.Contains(message, "snapshot"):
		return CodeStaleSnapshot
	case strings.Contains(message, "evidence"):
		return CodeMissing
	case strings.Contains(message, "candidate"):
		return CodeCandidate
	default:
		return CodeMalformed
	}
}

func classifyBindingError(err error) issueClass {
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "missing") || strings.Contains(message, "snapshot") {
		return issueUnknown
	}
	return issueFailClosed
}

func classifyBindingCode(err error) string {
	if strings.Contains(strings.ToLower(err.Error()), "snapshot") {
		return CodeStaleSnapshot
	}
	return CodeDigestMismatch
}

func bindEvidenceIDs(input Input) error {
	actual := make([]semantic.ID, 0, len(input.InferencePath.Evidence))
	for _, evidence := range input.InferencePath.Evidence {
		actual = append(actual, evidence.ID)
	}
	sort.Slice(actual, func(i, j int) bool { return actual[i] < actual[j] })
	if equalIDs(actual, input.EvidenceIDs) {
		return nil
	}
	if len(actual) > len(input.EvidenceIDs) {
		return fmt.Errorf("missing evidence IDs")
	}
	return fmt.Errorf("orphan evidence IDs")
}

func bindSnapshots(input Input) error {
	for _, edge := range input.InferencePath.Edges {
		if edge.Before != input.Snapshots.Base || edge.After != input.Snapshots.Head {
			return fmt.Errorf("stale snapshot on edge %s", edge.RecordID)
		}
	}
	for _, claim := range input.InferencePath.Claims {
		if claim.Before != input.Snapshots.Base || claim.After != input.Snapshots.Head {
			return fmt.Errorf("stale snapshot on claim %s", claim.RecordID)
		}
	}
	for _, evidence := range input.InferencePath.Evidence {
		if evidence.Before != input.Snapshots.Base || evidence.After != input.Snapshots.Head {
			return fmt.Errorf("stale snapshot on evidence %s", evidence.ID)
		}
	}
	return nil
}
