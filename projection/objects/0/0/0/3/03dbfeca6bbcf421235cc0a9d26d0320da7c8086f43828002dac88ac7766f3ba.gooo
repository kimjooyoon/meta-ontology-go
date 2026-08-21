package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
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
