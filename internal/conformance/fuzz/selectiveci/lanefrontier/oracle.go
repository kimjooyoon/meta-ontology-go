// Package lanefrontier contains an independent reference oracle for lane
// frontier eligibility. It deliberately does not import production selectors.
package lanefrontier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

const SchemaV1 = "gooo/lane-frontier/v1"

type Decision string
type Reason string

const (
	Unknown    Decision = "UNKNOWN"
	Ineligible Decision = "INELIGIBLE"
	Eligible   Decision = "ELIGIBLE"
)

const (
	UnknownSchema  Reason = "UNKNOWN_SCHEMA"
	MissingInput   Reason = "MISSING_INPUT"
	InvalidCount   Reason = "INVALID_COUNT"
	AmbiguousOwner Reason = "AMBIGUOUS_OWNER"
	PathOutOfScope Reason = "PATH_OUT_OF_SCOPE"
	ActiveLease    Reason = "ACTIVE_LEASE"
	ActivePR       Reason = "ACTIVE_PR"
	DivergedBranch Reason = "DIVERGED_BRANCH"
	StaleBranch    Reason = "STALE_BRANCH"
	BranchAhead    Reason = "BRANCH_AHEAD"
	Clean          Reason = "CLEAN"
)

type Input struct {
	Schema            string   `json:"schema"`
	RegistryDigest    string   `json:"registry_digest"`
	BaseSHA           string   `json:"base_sha"`
	LaneHeadSHA       string   `json:"lane_head_sha"`
	LaneStableID      string   `json:"lane_stable_id"`
	RegisteredBranch  string   `json:"registered_branch"`
	OwnedPathPrefixes []string `json:"owned_path_prefixes"`
	ChangedPaths      []string `json:"changed_paths"`
	AheadCount        int64    `json:"ahead_count"`
	BehindCount       int64    `json:"behind_count"`
	OpenPRCount       int64    `json:"open_pr_count"`
	ActiveLeaseCount  int64    `json:"active_lease_count"`
}

type Result struct {
	Decision        Decision `json:"decision"`
	Reason          Reason   `json:"reason"`
	CanonicalDigest string   `json:"canonical_digest"`
}

type Case struct {
	Name            string `json:"name"`
	Input           Input  `json:"input"`
	Expected        Result `json:"expected"`
	CanonicalDigest string `json:"canonical_digest"`
}

type Corpus struct {
	Schema string `json:"schema"`
	Cases  []Case `json:"cases"`
}

func Evaluate(input Input) Result {
	decision, reason := decide(input)
	result := Result{Decision: decision, Reason: reason}
	result.CanonicalDigest = digestResult(input, decision, reason)
	return result
}

func decide(input Input) (Decision, Reason) {
	if input.Schema != SchemaV1 {
		return Unknown, UnknownSchema
	}
	if missingInput(input) {
		return Unknown, MissingInput
	}
	if invalidCount(input) {
		return Unknown, InvalidCount
	}
	prefixes, valid := canonicalPrefixes(input.OwnedPathPrefixes)
	if !valid || ambiguousPrefixes(prefixes) {
		if !valid {
			return Unknown, MissingInput
		}
		return Unknown, AmbiguousOwner
	}
	if !validChangedPaths(input.ChangedPaths) {
		return Unknown, MissingInput
	}
	if !pathsInScope(input.ChangedPaths, prefixes) {
		return Ineligible, PathOutOfScope
	}
	if input.ActiveLeaseCount > 0 {
		return Ineligible, ActiveLease
	}
	if input.OpenPRCount > 0 {
		return Ineligible, ActivePR
	}
	if input.AheadCount > 0 && input.BehindCount > 0 {
		return Ineligible, DivergedBranch
	}
	if input.AheadCount == 0 && input.BehindCount > 0 {
		return Ineligible, StaleBranch
	}
	if input.AheadCount > 0 && input.BehindCount == 0 {
		return Ineligible, BranchAhead
	}
	return Eligible, Clean
}

func missingInput(input Input) bool {
	values := []string{input.RegistryDigest, input.BaseSHA, input.LaneHeadSHA, input.LaneStableID, input.RegisteredBranch}
	for _, value := range values {
		if strings.TrimSpace(value) == "" || value != strings.TrimSpace(value) {
			return true
		}
	}
	return len(input.OwnedPathPrefixes) == 0 || input.ChangedPaths == nil
}

func invalidCount(input Input) bool {
	return input.AheadCount < 0 || input.BehindCount < 0 || input.OpenPRCount < 0 || input.ActiveLeaseCount < 0
}

func canonicalPrefixes(prefixes []string) ([]string, bool) {
	result := make([]string, len(prefixes))
	for i, prefix := range prefixes {
		canonical, ok := canonicalPath(prefix, true)
		if !ok {
			return nil, false
		}
		result[i] = canonical
	}
	sort.Strings(result)
	return result, true
}

func canonicalPath(value string, prefix bool) (string, bool) {
	if prefix {
		value = strings.TrimSuffix(value, "/")
	}
	if value == "" || strings.HasPrefix(value, "/") || strings.ContainsAny(value, "\\\x00") {
		return "", false
	}
	parts := strings.Split(value, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", false
		}
	}
	return strings.Join(parts, "/"), true
}

func ambiguousPrefixes(prefixes []string) bool {
	for i := range prefixes {
		for j := i + 1; j < len(prefixes); j++ {
			if pathContains(prefixes[i], prefixes[j]) || pathContains(prefixes[j], prefixes[i]) {
				return true
			}
		}
	}
	return false
}

func pathsInScope(paths, prefixes []string) bool {
	for _, path := range paths {
		canonical, ok := canonicalPath(path, false)
		if !ok || !ownedPath(canonical, prefixes) {
			return false
		}
	}
	return true
}

func validChangedPaths(paths []string) bool {
	for _, path := range paths {
		if _, ok := canonicalPath(path, false); !ok {
			return false
		}
	}
	return true
}

func ownedPath(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if pathContains(prefix, path) {
			return true
		}
	}
	return false
}

func pathContains(prefix, path string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func digestResult(input Input, decision Decision, reason Reason) string {
	canonical := struct {
		Input    Input    `json:"input"`
		Decision Decision `json:"decision"`
		Reason   Reason   `json:"reason"`
	}{normalizedInput(input), decision, reason}
	body, _ := json.Marshal(canonical)
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func normalizedInput(input Input) Input {
	input.OwnedPathPrefixes, _ = canonicalPrefixes(input.OwnedPathPrefixes)
	input.ChangedPaths = append([]string(nil), input.ChangedPaths...)
	sort.Strings(input.ChangedPaths)
	return input
}
