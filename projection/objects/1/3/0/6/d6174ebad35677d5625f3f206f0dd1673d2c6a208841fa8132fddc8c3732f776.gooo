package shadow

import (
	"fmt"
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
