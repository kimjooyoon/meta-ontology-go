package transformationeffect

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"sort"
)

type PatchChange struct {
	Path               string `json:"path"`
	Kind               string `json:"kind"`
	BeforeSHA256       string `json:"before_sha256,omitempty"`
	AfterSHA256        string `json:"after_sha256,omitempty"`
	Mode               uint32 `json:"mode"`
	AfterContentBase64 string `json:"after_content_base64,omitempty"`
}

type Patch struct {
	Schema       string        `json:"schema"`
	HeadSHA      string        `json:"head_sha"`
	Changes      []PatchChange `json:"changes"`
	ChangeDigest string        `json:"change_digest"`
	PatchDigest  string        `json:"patch_digest"`
}

func makePatch(head string, before, after treeState) Patch {
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
	changes := make([]PatchChange, 0)
	for _, path := range paths {
		old, hadOld := left[path]
		fresh, hasFresh := right[path]
		if hadOld && hasFresh && old.Kind == fresh.Kind && old.Mode == fresh.Mode && old.SHA256 == fresh.SHA256 {
			continue
		}
		change := PatchChange{Path: path, Kind: "MODIFY", BeforeSHA256: old.SHA256,
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
	return sealPatch(Patch{Schema: patchSchema, HeadSHA: head, Changes: changes})
}

func stateIndex(state treeState) map[string]fileState {
	result := make(map[string]fileState, len(state.Entries))
	for _, entry := range state.Entries {
		result[entry.Path] = entry
	}
	return result
}

func sealPatch(patch Patch) Patch {
	patch.ChangeDigest = hashJSON(patch.Changes)
	patch.PatchDigest = ""
	patch.PatchDigest = hashJSON(patch)
	return patch
}

func validatePatch(patch Patch) error {
	expected := patch
	expected.ChangeDigest, expected.PatchDigest = "", ""
	if patch.Schema != patchSchema || !validSHA(patch.HeadSHA) || !reflect.DeepEqual(sealPatch(expected), patch) {
		return fmt.Errorf("patch envelope is not canonical")
	}
	for _, change := range patch.Changes {
		if change.Kind != "DELETE" {
			payload, err := base64.StdEncoding.DecodeString(change.AfterContentBase64)
			if err != nil || hashBytes(payload) != change.AfterSHA256 {
				return fmt.Errorf("patch content is unbound: %s", change.Path)
			}
		}
	}
	return nil
}
