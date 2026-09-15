package selectiveci

import (
	"strings"
)

func validateCommands(commands, guards []Command) error {
	seen := map[string]struct{}{}
	for _, command := range append(append([]Command(nil), commands...), guards...) {
		if !validStableID(command.ID) {
			return failure(ReasonInvalidInput, "command ID is invalid")
		}
		if _, exists := seen[command.ID]; exists {
			return failure(ReasonDuplicateID, "duplicate command ID")
		}
		seen[command.ID] = struct{}{}
		if len(command.Argv) == 0 || command.Argv[0] == "" {
			return failure(ReasonInvalidArgv, "argv must contain an executable")
		}
		for _, arg := range command.Argv {
			if strings.IndexByte(arg, 0) >= 0 {
				return failure(ReasonInvalidArgv, "argv contains NUL")
			}
		}
		if !validRepoDir(command.WorkingDir) {
			return failure(ReasonUnknownPath, "working directory is not normalized")
		}
		if command.CPUWorkUnits == 0 || command.MemoryBytes == 0 {
			return failure(ReasonInvalidInput, "command resource ceiling is zero")
		}
	}
	return nil
}
func validateBindings(bindings []ObligationBinding) error {
	seen := map[string]struct{}{}
	for _, binding := range bindings {
		if !validStableID(binding.ID) || !validStableID(binding.Subject) {
			return failure(ReasonMissingBinding, "obligation binding identity is invalid")
		}
		if _, exists := seen[binding.ID]; exists {
			return failure(ReasonDuplicateID, "duplicate obligation ID")
		}
		seen[binding.ID] = struct{}{}
		if len(binding.CommandIDs) == 0 {
			return failure(ReasonMissingCommand, "obligation has no commands")
		}
		if len(binding.CommandIDs) != len(sortedUnique(binding.CommandIDs)) {
			return failure(ReasonDuplicateID, "duplicate command reference")
		}
		if err := validateIDs(binding.CommandIDs); err != nil {
			return failure(ReasonDanglingReference, err.Error())
		}
	}
	return nil
}
func validateReceipts(receipts []Receipt) error {
	seen := map[string]struct{}{}
	for _, receipt := range receipts {
		if !validStableID(receipt.CommandID) {
			return failure(ReasonDanglingReference, "resource receipt command ID is invalid")
		}
		if _, exists := seen[receipt.CommandID]; exists {
			return failure(ReasonDuplicateID, "duplicate resource receipt")
		}
		seen[receipt.CommandID] = struct{}{}
		if !validDigest(receipt.SnapshotDigest) {
			return failure(ReasonMismatchedDigest, "resource receipt snapshot digest is invalid")
		}
	}
	return nil
}
