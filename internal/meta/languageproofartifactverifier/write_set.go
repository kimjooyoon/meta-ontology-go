package languageproofartifactverifier

import (
	"fmt"
	"reflect"
	"sort"
)

const WriteSetSchema = "gooo/repository-write-set-observation/v1"

func DecodeWriteSet(raw []byte) (WriteSetObservation, error) {
	value, err := decodeStrict[WriteSetObservation](raw)
	if err != nil {
		return WriteSetObservation{}, err
	}
	if value.Digest == "" {
		value.Digest = writeSetDigest(value)
	}
	if err := validateWriteSet(value); err != nil {
		return WriteSetObservation{}, err
	}
	return value, nil
}

func validateWriteSet(observation WriteSetObservation) error {
	if observation.Schema != WriteSetSchema || observation.Version != 1 {
		return fmt.Errorf("repository write-set schema mismatch")
	}
	before := sortedEntries(observation.Before)
	after := sortedEntries(observation.After)
	if !reflect.DeepEqual(before, observation.Before) || !reflect.DeepEqual(after, observation.After) {
		return fmt.Errorf("repository write-set entries are not canonical")
	}
	if observation.BeforeDigest != digestValue(before) || observation.AfterDigest != digestValue(after) {
		return fmt.Errorf("repository write-set snapshot digest mismatch")
	}
	want := changedEntries(before, after)
	if !reflect.DeepEqual(want, observation.Changed) || observation.RepositoryWrites != len(want) || observation.MutationAuthority {
		return fmt.Errorf("repository write-set observation mismatch")
	}
	if observation.ObservedScope != "NET_BEFORE_AFTER_TRACKED_AND_UNTRACKED" || !observation.TransientUnknown || observation.NetUnchanged != (len(want) == 0) || observation.AuthorityBasis == "" {
		return fmt.Errorf("repository write-set observation scope mismatch")
	}
	if observation.Digest != writeSetDigest(observation) {
		return fmt.Errorf("repository write-set digest mismatch")
	}
	return nil
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
