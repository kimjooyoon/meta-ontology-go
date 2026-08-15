package selectiveci

import "sort"

func Plan(input Input) PlanResult {
	result := PlanResult{SchemaVersion: SchemaVersion, BaseSnapshotDigest: input.Base.Digest, HeadSnapshotDigest: input.Head.Digest}
	if err := input.Validate(); err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	changed, err := changedSemanticIDs(input.Base, input.Head)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	result.ChangedSemanticIDs = changed
	graph, err := buildGraph(input)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	required, err := applicableObligations(graph, changed)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	commands, guards, err := selectedCommands(input.Registry, required)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	frontier, selected, err := commandFrontier(input, commands, guards)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	receiptDigests, pathIDs, err := validateSelectedEvidence(input, selected)
	if err != nil {
		return sealResult(fallback(result, reasonFor(err)))
	}
	result = fillSelection(result, selected, frontier)
	result.ResourceReceiptDigests = receiptDigests
	result.ProvenancePathIDs = pathIDs
	result.Status = StatusSelective
	result.ReasonCode = ReasonNone
	return sealResult(result)
}

func Evaluate(input Input) PlanResult { return Plan(input) }

func PlanJSON(data []byte) PlanResult {
	input, err := DecodeJSON(data)
	if err != nil {
		return sealResult(fallback(PlanResult{SchemaVersion: SchemaVersion}, reasonFor(err)))
	}
	return Plan(input)
}

func PlanJSONWithError(data []byte) (PlanResult, error) {
	input, err := DecodeJSON(data)
	if err != nil {
		return sealResult(fallback(PlanResult{SchemaVersion: SchemaVersion}, reasonFor(err))), err
	}
	return Plan(input), nil
}

func fallback(result PlanResult, reason string) PlanResult {
	result.Status = StatusFullSuiteFallback
	result.ReasonCode = reason
	result.SelectedCommandIDs = nil
	result.SelectedGuardCommandIDs = nil
	result.SelectedWorkIDs = nil
	result.ResourceReceiptDigests = nil
	result.ProvenancePathIDs = nil
	return result
}

func changedSemanticIDs(base, head SnapshotManifest) ([]string, error) {
	baseFiles, headFiles := manifestFiles(base), manifestFiles(head)
	ids := map[string]struct{}{}
	paths := map[string]struct{}{}
	for path := range baseFiles {
		paths[path] = struct{}{}
	}
	for path := range headFiles {
		paths[path] = struct{}{}
	}
	for path := range paths {
		before, beforeOK := baseFiles[path]
		after, afterOK := headFiles[path]
		if beforeOK && afterOK && before.BlobDigest == after.BlobDigest && equalStrings(before.SemanticIDs, after.SemanticIDs) {
			continue
		}
		for _, id := range append(append([]string{}, before.SemanticIDs...), after.SemanticIDs...) {
			ids[id] = struct{}{}
		}
		if len(before.SemanticIDs) == 0 && len(after.SemanticIDs) == 0 {
			return nil, failure(ReasonUnknownPath, "changed path has no stable semantic ID")
		}
	}
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	sort.Strings(result)
	return result, nil
}

func manifestFiles(manifest SnapshotManifest) map[string]SnapshotFile {
	result := make(map[string]SnapshotFile, len(manifest.Files))
	for _, file := range manifest.Files {
		result[file.Path] = file
	}
	return result
}

func equalStrings(left, right []string) bool {
	left, right = sortedCopy(left), sortedCopy(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
