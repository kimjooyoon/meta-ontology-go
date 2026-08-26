package selectiveci

func selectChangedOwners(evidence Evidence, commands map[string]Command, paths map[string]PathEvidence) (changeSelection, Reason) {
	if len(evidence.Changes) == 0 {
		return changeSelection{}, noChangesReason
	}
	selection := changeSelection{
		owners:       make(map[string]struct{}),
		changedPaths: make(map[string]struct{}, len(evidence.Changes)),
	}
	for _, change := range evidence.Changes {
		path, reason := validateChange(change, paths)
		if reason != "" {
			return changeSelection{}, reason
		}
		selection.changedPaths[change.Path] = struct{}{}
		for _, owner := range path.Owners {
			if _, exists := commands[owner]; !exists {
				return changeSelection{}, danglingCmdReason
			}
			selection.owners[owner] = struct{}{}
		}
	}
	return selection, ""
}
func validateChange(change PathChange, paths map[string]PathEvidence) (PathEvidence, Reason) {
	path, exists := paths[change.Path]
	if !exists {
		return PathEvidence{}, missingPathReason
	}
	if path.Stale {
		return PathEvidence{}, staleReason
	}
	if !path.Connected {
		return PathEvidence{}, disconnectedReason
	}
	if path.Ambiguous || len(path.Owners) > 1 && change.Kind != ChangeDelete {
		return PathEvidence{}, ambiguousReason
	}
	if path.Authority != AuthorityAuthoritative {
		return PathEvidence{}, nonAuthorityReason
	}
	if len(path.Owners) == 0 {
		return PathEvidence{}, unknownPathReason
	}
	if !changeMatchesEvidence(change, path) {
		return PathEvidence{}, blobMismatchReason
	}
	return path, ""
}
func expandSelection(selected map[string]struct{}, adjacency map[string][]string) {
	queue := make([]string, 0, len(selected))
	for id := range selected {
		queue = append(queue, id)
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, downstream := range adjacency[id] {
			if _, exists := selected[downstream]; exists {
				continue
			}
			selected[downstream] = struct{}{}
			queue = append(queue, downstream)
		}
	}
}
