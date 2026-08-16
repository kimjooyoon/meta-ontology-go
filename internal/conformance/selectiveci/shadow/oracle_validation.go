package shadow

import (
	"fmt"
	"sort"
	"strings"
)

func validateSnapshots(base, head analyzerSnapshot, planner plannerInput) error {
	if base.Schema != AnalyzerSchema || head.Schema != AnalyzerSchema || base.Status != "BOUND" || head.Status != "BOUND" || base.FullSuiteFallback || head.FullSuiteFallback {
		return fmt.Errorf("analyzer snapshots are not BOUND")
	}
	if !validSnapshotFacts(base) || !validSnapshotFacts(head) {
		return fmt.Errorf("analyzer snapshots are incomplete")
	}
	if base.Digest == "" || head.Digest == "" || base.Digest != analyzerDigest(base) || head.Digest != analyzerDigest(head) {
		return fmt.Errorf("analyzer snapshot digest is stale")
	}
	if planner.Schema != PlannerSchema || planner.BaseManifest.Schema != ManifestSchema || planner.HeadManifest.Schema != ManifestSchema {
		return fmt.Errorf("planner manifest schema is not bound")
	}
	if planner.BaseManifest.Files == nil || planner.HeadManifest.Files == nil || !validFiles(planner.BaseManifest.Files) || !validFiles(planner.HeadManifest.Files) {
		return fmt.Errorf("planner manifests are incomplete")
	}
	derivedBase, derivedHead := derivedManifest(base), derivedManifest(head)
	if !manifestEqual(planner.BaseManifest, derivedBase) || !manifestEqual(planner.HeadManifest, derivedHead) {
		return fmt.Errorf("planner manifests do not exactly match analyzer snapshots")
	}
	return nil
}

func validateRegistry(inputs decodedInputs) error {
	values := []string{inputs.base.RegistryDigest, inputs.head.RegistryDigest, inputs.planner.RegistryDigest, inputs.proof.RegistryDigest, inputs.lane.RegistryDigest}
	if values[0] == "" {
		return fmt.Errorf("registry digest is missing")
	}
	for _, value := range values[1:] {
		if value != values[0] {
			return fmt.Errorf("registry digest binding mismatch")
		}
	}
	return nil
}

func validatePlan(base, head analyzerSnapshot, planner plannerInput) error {
	if planner.Status != "SELECTIVE" || planner.PlanDigest == "" || planner.PlanDigest != planDigest(planner) {
		return fmt.Errorf("planner is not a sealed SELECTIVE result")
	}
	if planner.ChangedRootIDs == nil || planner.SelectedCommandIDs == nil || planner.SelectedGuardCommandIDs == nil || planner.SelectedWorkIDs == nil || planner.Commands == nil || planner.GuardCommands == nil {
		return fmt.Errorf("planner selection is incomplete")
	}
	if !equalStrings(planner.ChangedRootIDs, changedRoots(base, head)) {
		return fmt.Errorf("changed root IDs mismatch")
	}
	if err := validateCommands(planner); err != nil {
		return err
	}
	if len(planner.SelectedCommandIDs) == 0 && len(planner.SelectedGuardCommandIDs) == 0 {
		return fmt.Errorf("planner selected command union is empty")
	}
	return nil
}

func validateCommands(planner plannerInput) error {
	commands := map[string]command{}
	for _, item := range append(append([]command{}, planner.Commands...), planner.GuardCommands...) {
		if item.ID == "" || len(item.Argv) == 0 {
			return fmt.Errorf("planner command is incomplete")
		}
		if _, exists := commands[item.ID]; exists {
			return fmt.Errorf("planner command ID is duplicated")
		}
		commands[item.ID] = item
	}
	guardIDs := map[string]struct{}{}
	for _, item := range planner.GuardCommands {
		guardIDs[item.ID] = struct{}{}
	}
	seen := map[string]struct{}{}
	for _, id := range append(append([]string{}, planner.SelectedCommandIDs...), planner.SelectedGuardCommandIDs...) {
		if id == "" {
			return fmt.Errorf("planner selected command ID is empty")
		}
		if _, exists := commands[id]; !exists {
			return fmt.Errorf("planner selected command ID is dangling")
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("planner selected command union is duplicated")
		}
		seen[id] = struct{}{}
	}
	for _, id := range planner.SelectedCommandIDs {
		if _, guard := guardIDs[id]; guard {
			return fmt.Errorf("planner command is selected as both command and guard")
		}
	}
	return nil
}

func validatePlanProofBinding(inputs decodedInputs) error {
	proof := inputs.proof
	if proof.Schema != ProofSchema {
		return fmt.Errorf("proof schema mismatch")
	}
	if proof.Snapshots.Base.Source != inputs.base.Digest || proof.Snapshots.Base.Semantic != inputs.planner.BaseManifest.Digest || proof.Snapshots.Head.Source != inputs.head.Digest || proof.Snapshots.Head.Semantic != inputs.planner.HeadManifest.Digest {
		return fmt.Errorf("proof snapshot binding mismatch")
	}
	planner := inputs.planner
	selected, guards, _, _ := normalizedSelection(planner)
	union := append(append([]string{}, selected...), guards...)
	sort.Strings(union)
	if proof.PlanDigest != planner.PlanDigest || !equalStrings(proof.ChangedRootIDs, planner.ChangedRootIDs) || !equalStrings(proof.SelectedCommandIDs, union) || !equalStrings(proof.VerifiedCommandIDs, union) {
		return fmt.Errorf("proof plan binding mismatch")
	}
	return nil
}

func validSnapshotFacts(snapshot analyzerSnapshot) bool {
	return snapshot.RegistryDigest != "" && snapshot.Files != nil && validFiles(snapshot.Files)
}

func validFiles(values []manifestFile) bool {
	seenPaths := map[string]struct{}{}
	seenIDs := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value.Path) == "" || strings.TrimSpace(value.BlobDigest) == "" || value.SemanticIDs == nil || len(value.SemanticIDs) == 0 {
			return false
		}
		if _, exists := seenPaths[value.Path]; exists {
			return false
		}
		seenPaths[value.Path] = struct{}{}
		for _, id := range value.SemanticIDs {
			if strings.TrimSpace(id) == "" {
				return false
			}
			if _, exists := seenIDs[id]; exists {
				return false
			}
			seenIDs[id] = struct{}{}
		}
	}
	return true
}

func validLaneFacts(lane laneInput) bool {
	return validNonEmpty(lane.RegistryDigest, lane.BaseSHA, lane.LaneHeadSHA, lane.LaneID, lane.RegisteredBranch) &&
		lane.OwnedPathPrefixes != nil && lane.ChangedPaths != nil && lane.AheadCount >= 0 && lane.BehindCount >= 0 && lane.OpenPRCount >= 0 && lane.ActiveLeaseCount >= 0
}
