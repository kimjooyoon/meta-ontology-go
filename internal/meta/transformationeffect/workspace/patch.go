package workspace

import (
	"encoding/base64"
	"sort"
)

func MakePatch(head string, before, after State) Patch {
	left, right := stateIndex(before), stateIndex(after)
	paths := make([]string, 0, len(left)+len(right))
	seen := map[string]bool{}
	for path := range left {
		seen[path], paths = true, append(paths, path)
	}
	for path := range right {
		if !seen[path] {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	changes := make([]Change, 0)
	for _, path := range paths {
		old, hadOld := left[path]
		fresh, hasFresh := right[path]
		if hadOld && hasFresh && old.Kind == fresh.Kind && old.Mode == fresh.Mode && old.SHA256 == fresh.SHA256 {
			continue
		}
		change := Change{Path: path, Kind: "MODIFY", BeforeSHA256: old.SHA256,
			AfterSHA256: fresh.SHA256, Mode: fresh.Mode}
		if !hadOld {
			change.Kind = "ADD"
		}
		if !hasFresh {
			change.Kind, change.Mode = "DELETE", old.Mode
		} else {
			change.AfterContentBase64 = base64.StdEncoding.EncodeToString(fresh.data)
		}
		changes = append(changes, change)
	}
	return seal(Patch{Schema: Schema, HeadSHA: head, Changes: changes})
}

func stateIndex(state State) map[string]Entry {
	result := make(map[string]Entry, len(state.Entries))
	for _, entry := range state.Entries {
		result[entry.Path] = entry
	}
	return result
}
