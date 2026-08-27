package causalci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
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
	if value.Schema != ObservationSchema || value.Repository == "" || value.BaseSHA == "" || value.HeadSHA == "" || filepath.Ext(value.SourcePath) != ".gooo" {
		return fmt.Errorf("%s: observation identity", ReasonMalformedObservation)
	}
	seen := map[string]struct{}{}
	for _, changed := range value.ChangedFiles {
		if changed.Path == "" || strings.Contains(changed.Path, "\t") || changed.Status == "" {
			return fmt.Errorf("%s: changed-file observation", ReasonMalformedObservation)
		}
		if _, exists := seen[changed.Path]; exists {
			return fmt.Errorf("%s: duplicate changed path %q", ReasonMalformedObservation, changed.Path)
		}
		seen[changed.Path] = struct{}{}
	}
	for _, claim := range value.PriorClaims {
		if claim.ClaimID == "" || claim.SubjectPath == "" || claim.State == "" || claim.Provenance == "" {
			return fmt.Errorf("%s: prior claim observation", ReasonMalformedObservation)
		}
		if claim.State != ClaimOpen && claim.State != ClaimDischarged && claim.State != ClaimRefuted {
			return fmt.Errorf("%s: unsupported prior claim state %q", ReasonMalformedObservation, claim.State)
		}
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
	digest, err := digestJSON(value.StatusLines)
	if err != nil || digest != value.StatusDigest {
		return fmt.Errorf("%s: isolation snapshot digest", ReasonMalformedObservation)
	}
	return nil
}

func repositoryWriteCount(value IsolationObservation) int {
	before := map[string]struct{}{}
	for _, line := range value.Before.StatusLines {
		before[line] = struct{}{}
	}
	after := map[string]struct{}{}
	for _, line := range value.After.StatusLines {
		after[line] = struct{}{}
	}
	count := 0
	for line := range before {
		if _, exists := after[line]; !exists {
			count++
		}
	}
	for line := range after {
		if _, exists := before[line]; !exists {
			count++
		}
	}
	return count
}

func sortedChangedFiles(values []ChangedFileObservation) []ChangedFileObservation {
	result := append([]ChangedFileObservation(nil), values...)
	for i := 1; i < len(result); i++ {
		for j := i; j > 0 && result[j].Path < result[j-1].Path; j-- {
			result[j], result[j-1] = result[j-1], result[j]
		}
	}
	return result
}

func sortedPriorClaims(values []PriorClaimObservation) []PriorClaimObservation {
	result := append([]PriorClaimObservation(nil), values...)
	for i := 1; i < len(result); i++ {
		for j := i; j > 0 && priorClaimLess(result[j], result[j-1]); j-- {
			result[j], result[j-1] = result[j-1], result[j]
		}
	}
	return result
}

func priorClaimLess(left, right PriorClaimObservation) bool {
	if left.SubjectPath != right.SubjectPath {
		return left.SubjectPath < right.SubjectPath
	}
	return left.ClaimID < right.ClaimID
}
