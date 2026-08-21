package shadow

import (
	"fmt"
	"sort"
)

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
