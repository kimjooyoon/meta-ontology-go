package languageproofartifact

import (
	"fmt"
	"reflect"
	"sort"
)

const WriteSetSchema = "gooo/repository-write-set-observation/v1"

func normalizeWriteSet(observation WriteSetObservation) (WriteSetObservation, error) {
	if observation.Schema != WriteSetSchema || observation.Version != 1 {
		return WriteSetObservation{}, fmt.Errorf("repository write-set schema mismatch")
	}
	observation.Before = sortedEntries(observation.Before)
	observation.After = sortedEntries(observation.After)
	if observation.BeforeDigest == "" {
		observation.BeforeDigest = digestValue(observation.Before)
	}
	if observation.AfterDigest == "" {
		observation.AfterDigest = digestValue(observation.After)
	}
	if observation.BeforeDigest != digestValue(observation.Before) || observation.AfterDigest != digestValue(observation.After) {
		return WriteSetObservation{}, fmt.Errorf("repository write-set snapshot digest mismatch")
	}
	if observation.Digest != "" && observation.Digest != writeSetDigest(observation) {
		return WriteSetObservation{}, fmt.Errorf("repository write-set digest mismatch")
	}
	want := changedEntries(observation.Before, observation.After)
	if len(observation.Changed) != 0 && !reflect.DeepEqual(observation.Changed, want) {
		return WriteSetObservation{}, fmt.Errorf("repository write-set change projection mismatch")
	}
	observation.Changed = want
	observation.RepositoryWrites = len(want)
	if observation.ObservedScope == "" {
		observation.ObservedScope = "NET_BEFORE_AFTER_TRACKED_AND_UNTRACKED"
	}
	if observation.AuthorityBasis == "" {
		observation.AuthorityBasis = "DECLARATION_ONLY"
	}
	observation.NetUnchanged = len(want) == 0
	observation.TransientUnknown = true
	if observation.MutationAuthority {
		return WriteSetObservation{}, fmt.Errorf("repository write-set mutation authority is not allowed")
	}
	observation.Digest = writeSetDigest(observation)
	return observation, nil
}

func sortedEntries(entries []WriteSetEntry) []WriteSetEntry {
	result := append([]WriteSetEntry(nil), entries...)
	sort.Slice(result, func(left, right int) bool { return result[left].Path < result[right].Path })
	return result
}

func changedEntries(before, after []WriteSetEntry) []WriteSetChange {
	left, right := entryIndex(before), entryIndex(after)
	paths := make([]string, 0, len(left)+len(right))
	seen := map[string]bool{}
	for path := range left {
		seen[path] = true
		paths = append(paths, path)
	}
	for path := range right {
		if !seen[path] {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	result := make([]WriteSetChange, 0, len(paths))
	for _, path := range paths {
		beforeEntry, beforeOK := left[path]
		afterEntry, afterOK := right[path]
		if beforeOK && afterOK && beforeEntry == afterEntry {
			continue
		}
		change := WriteSetChange{Path: path}
		if beforeOK {
			change.BeforeDigest, change.BeforeKind = beforeEntry.Digest, beforeEntry.Kind
		}
		if afterOK {
			change.AfterDigest, change.AfterKind = afterEntry.Digest, afterEntry.Kind
		}
		result = append(result, change)
	}
	return result
}

func entryIndex(entries []WriteSetEntry) map[string]WriteSetEntry {
	result := make(map[string]WriteSetEntry, len(entries))
	for _, entry := range entries {
		result[entry.Path] = entry
	}
	return result
}

func writeSetDigest(observation WriteSetObservation) string {
	observation.Digest = ""
	return digestValue(observation)
}
