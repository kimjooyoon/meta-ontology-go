package shadow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func analyzerDigest(snapshot analyzerSnapshot) string {
	value := struct {
		Schema            string         `json:"schema"`
		Status            string         `json:"status"`
		FullSuiteFallback bool           `json:"full_suite_fallback"`
		RegistryDigest    string         `json:"registry_digest"`
		Files             []manifestFile `json:"files"`
	}{snapshot.Schema, snapshot.Status, snapshot.FullSuiteFallback, snapshot.RegistryDigest, normalizeFiles(snapshot.Files)}
	return hashJSON(value)
}

func manifestDigest(manifest plannerManifest) string {
	value := struct {
		Schema string         `json:"schema"`
		Files  []manifestFile `json:"files"`
	}{manifest.Schema, normalizeFiles(manifest.Files)}
	return hashJSON(value)
}

func planDigest(planner plannerInput) string {
	value := struct {
		Schema                  string          `json:"schema"`
		Status                  string          `json:"status"`
		RegistryDigest          string          `json:"registry_digest"`
		BaseManifest            plannerManifest `json:"base_manifest"`
		HeadManifest            plannerManifest `json:"head_manifest"`
		ChangedRootIDs          []string        `json:"changed_root_ids"`
		SelectedCommandIDs      []string        `json:"selected_command_ids"`
		SelectedGuardCommandIDs []string        `json:"selected_guard_command_ids"`
		SelectedWorkIDs         []string        `json:"selected_work_ids"`
		Commands                []command       `json:"commands"`
		GuardCommands           []command       `json:"guard_commands"`
	}{planner.Schema, planner.Status, planner.RegistryDigest, normalizeManifest(planner.BaseManifest), normalizeManifest(planner.HeadManifest), sortedCopy(planner.ChangedRootIDs), sortedCopy(planner.SelectedCommandIDs), sortedCopy(planner.SelectedGuardCommandIDs), sortedCopy(planner.SelectedWorkIDs), normalizeCommands(planner.Commands), normalizeCommands(planner.GuardCommands)}
	return hashJSON(value)
}

func proofDigest(proof proofInput) string {
	value := struct {
		Schema             string         `json:"schema"`
		Status             string         `json:"status"`
		Fallback           string         `json:"fallback"`
		RegistryDigest     string         `json:"registry_digest"`
		PlanDigest         string         `json:"plan_digest"`
		Snapshots          proofSnapshots `json:"snapshots"`
		ChangedRootIDs     []string       `json:"changed_root_ids"`
		SelectedCommandIDs []string       `json:"selected_command_ids"`
		VerifiedCommandIDs []string       `json:"verified_command_ids"`
	}{proof.Schema, proof.Status, proof.Fallback, proof.RegistryDigest, proof.PlanDigest, proof.Snapshots, sortedCopy(proof.ChangedRootIDs), sortedCopy(proof.SelectedCommandIDs), sortedCopy(proof.VerifiedCommandIDs)}
	return hashJSON(value)
}

func laneDigest(lane laneInput) string {
	value := struct {
		Schema            string   `json:"schema"`
		Decision          string   `json:"decision"`
		Reason            string   `json:"reason"`
		RegistryDigest    string   `json:"registry_digest"`
		BaseSHA           string   `json:"base_sha"`
		LaneHeadSHA       string   `json:"lane_head_sha"`
		LaneID            string   `json:"lane_id"`
		RegisteredBranch  string   `json:"registered_branch"`
		OwnedPathPrefixes []string `json:"owned_path_prefixes"`
		ChangedPaths      []string `json:"changed_paths"`
		AheadCount        int64    `json:"ahead_count"`
		BehindCount       int64    `json:"behind_count"`
		OpenPRCount       int64    `json:"open_pr_count"`
		ActiveLeaseCount  int64    `json:"active_lease_count"`
	}{lane.Schema, lane.Decision, lane.Reason, lane.RegistryDigest, lane.BaseSHA, lane.LaneHeadSHA, lane.LaneID, lane.RegisteredBranch, sortedCopy(lane.OwnedPathPrefixes), sortedCopy(lane.ChangedPaths), lane.AheadCount, lane.BehindCount, lane.OpenPRCount, lane.ActiveLeaseCount}
	return hashJSON(value)
}

func caseDigest(c Case) string {
	value := struct {
		Name      string `json:"name"`
		Partition string `json:"partition"`
		Files     Files  `json:"files"`
	}{c.Name, c.Partition, c.Files}
	return hashJSON(value)
}

func hashJSON(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func validNonEmpty(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}
