package causalci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func decodeObservation(raw []byte) (Observation, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var observation Observation
	if err := decoder.Decode(&observation); err != nil {
		return Observation{}, fmt.Errorf("decode raw observation: %w", err)
	}
	if err := validateObservation(observation); err != nil {
		return Observation{}, err
	}
	return observation, nil
}

func validateObservation(value Observation) error {
	if value.Schema != ObservationSchema || value.Repository == "" || value.BaseSHA == "" || value.HeadSHA == "" || value.ObservedCheckoutSHA == "" || filepath.Ext(value.SourcePath) != ".gooo" || value.HeadPathObjectID == "" || value.SourceBytesDigest == "" {
		return fmt.Errorf("%s: observation identity", ReasonMalformedObservation)
	}
	seenPaths := map[string]struct{}{}
	for _, changed := range value.ChangedFiles {
		if changed.Path == "" || strings.Contains(changed.Path, "\t") || changed.Status == "" {
			return fmt.Errorf("%s: changed-file observation", ReasonMalformedObservation)
		}
		if _, exists := seenPaths[changed.Path]; exists {
			return fmt.Errorf("%s: duplicate changed path %q", ReasonMalformedObservation, changed.Path)
		}
		seenPaths[changed.Path] = struct{}{}
	}
	seenClaims := map[string]struct{}{}
	for _, claim := range value.PriorClaims {
		if claim.TemplateID == "" || claim.InstanceID == "" || claim.SubjectPath == "" || claim.Proposition == "" || claim.State == "" || claim.Provenance == "" {
			return fmt.Errorf("%s: prior claim observation", ReasonMalformedObservation)
		}
		if claim.State != ClaimOpen && claim.State != ClaimDischarged && claim.State != ClaimRefuted {
			return fmt.Errorf("%s: unsupported prior claim state %q", ReasonMalformedObservation, claim.State)
		}
		if claim.InstanceID != ClaimInstanceID(claim.TemplateID, claim.SubjectPath, claim.Proposition) {
			return fmt.Errorf("%s: claim instance is not content addressed", ReasonMalformedObservation)
		}
		if _, exists := seenClaims[claim.InstanceID]; exists {
			return fmt.Errorf("%s: duplicate claim instance %q", ReasonMalformedObservation, claim.InstanceID)
		}
		seenClaims[claim.InstanceID] = struct{}{}
	}
	if err := validateSnapshot(value.Isolation.Before); err != nil {
		return err
	}
	if err := validateSnapshot(value.Isolation.After); err != nil {
		return err
	}
	return nil
}

func validateSnapshot(value RepositorySnapshot) error {
	entries := make([]RepositoryEntry, len(value.Entries))
	copy(entries, value.Entries)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Path != entries[j].Path {
			return entries[i].Path < entries[j].Path
		}
		return !entries[i].Tracked && entries[j].Tracked
	})
	seen := map[string]struct{}{}
	for _, entry := range entries {
		if entry.Path == "" || entry.ContentDigest == "" {
			return fmt.Errorf("%s: isolation snapshot entry", ReasonMalformedObservation)
		}
		if _, exists := seen[entry.Path]; exists {
			return fmt.Errorf("%s: duplicate isolation snapshot path %q", ReasonMalformedObservation, entry.Path)
		}
		seen[entry.Path] = struct{}{}
	}
	digest, err := digestJSON(entries)
	if err != nil || digest != value.SnapshotDigest {
		return fmt.Errorf("%s: isolation snapshot digest", ReasonMalformedObservation)
	}
	return nil
}

func repositoryState(value IsolationObservation) RepositoryStateObservation {
	before := snapshotEntries(value.Before)
	after := snapshotEntries(value.After)
	paths := map[string]struct{}{}
	for path := range before {
		paths[path] = struct{}{}
	}
	for path := range after {
		paths[path] = struct{}{}
	}
	changedPaths, changedContents := 0, 0
	for path := range paths {
		left, leftOK := before[path]
		right, rightOK := after[path]
		if !leftOK || !rightOK {
			changedPaths++
			changedContents++
			continue
		}
		if left.Tracked != right.Tracked {
			changedPaths++
		}
		if left.ContentDigest != right.ContentDigest {
			changedContents++
		}
	}
	state := ObservedStateChanged
	if changedPaths == 0 && changedContents == 0 {
		state = ObservedStateUnchanged
	}
	return RepositoryStateObservation{NetState: state, ChangedPathCount: changedPaths, ChangedContentCount: changedContents, TransientWrites: ObservedUnknown, GlobalMutationAuthority: ObservedUnknown}
}

func snapshotEntries(value RepositorySnapshot) map[string]RepositoryEntry {
	result := make(map[string]RepositoryEntry, len(value.Entries))
	for _, entry := range value.Entries {
		result[entry.Path] = entry
	}
	return result
}

func sortedChangedFiles(values []ChangedFileObservation) []ChangedFileObservation {
	result := append([]ChangedFileObservation(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}

func sortedPriorClaims(values []PriorClaimObservation) []PriorClaimObservation {
	result := append([]PriorClaimObservation(nil), values...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].SubjectPath != result[j].SubjectPath {
			return result[i].SubjectPath < result[j].SubjectPath
		}
		return result[i].InstanceID < result[j].InstanceID
	})
	return result
}

func exactIDs(values []string) bool {
	seen := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func sameIDs(expected, observed []string) bool {
	if len(expected) != len(observed) || !exactIDs(expected) || !exactIDs(observed) {
		return false
	}
	left, right := append([]string(nil), expected...), append([]string(nil), observed...)
	sort.Strings(left)
	sort.Strings(right)
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
