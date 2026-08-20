package workfrontier

import (
	"crypto/sha256"
	"encoding/hex"
)

// WorkID returns SHA-256(snapshot || obligation || path || policy) as hex.
func WorkID(snapshotDigest, obligationID, pathID, policyDigest string) string {
	digest := sha256.Sum256([]byte(snapshotDigest + obligationID + pathID + policyDigest))
	return hex.EncodeToString(digest[:])
}

// WorkIDFor computes the identity of a path in an input snapshot.
func WorkIDFor(input Input, path RepairPath) string {
	if path.WorkID != "" {
		return path.WorkID
	}
	return WorkID(input.SnapshotDigest, path.ObligationID, path.stableID(), input.PolicyDigest)
}
func inputShapeKnown(input Input) bool {
	if input.SchemaVersion != SchemaVersion || input.SnapshotDigest == "" ||
		input.PolicyDigest == "" || input.RegistryDigest == "" ||
		input.MinimumSelectedPressures < 2 {
		return false
	}
	if input.fromJSON {
		p := input.present
		return p.schemaVersion && p.snapshotDigest && p.policyDigest && p.registryDigest &&
			p.minimumSelectedPressures && p.capacity && p.pressures && p.states && p.paths &&
			input.Capacity.cpuCoreNSPresent
	}
	return true
}
func buildIndexes(input Input) frontierIndexes {
	indexes := frontierIndexes{pressures: make(map[string]struct{}), states: make(map[string]string)}
	for _, pressure := range input.Pressures {
		id := pressure.stableID()
		if !pressureKnown(pressure) || id == "" {
			indexes.invalid = true
			continue
		}
		if _, exists := indexes.pressures[id]; exists {
			indexes.invalid = true
		}
		indexes.pressures[id] = struct{}{}
	}
	for _, state := range input.States {
		id := state.obligationID()
		if !stateKnown(state) || id == "" {
			indexes.invalid = true
			continue
		}
		if _, exists := indexes.states[id]; exists {
			indexes.invalid = true
		}
		indexes.states[id] = state.Status
	}
	return indexes
}
func pressureKnown(pressure Pressure) bool {
	return pressure.stableID() != "" && (!pressure.fromJSON || pressure.stableIDPresent)
}
func stateKnown(state ObligationState) bool {
	if state.obligationID() == "" || state.Status == "" {
		return false
	}
	return !state.fromJSON || (state.obligationIDPresent && state.statusPresent)
}
