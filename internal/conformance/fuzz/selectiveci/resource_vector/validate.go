package resourcevector

import (
	"strings"
	"unicode"
)

func validate(input Input) validationFailure {
	if failure := validateBase(input); failure.reason != "" {
		return failure
	}
	commands, failure := validateCommands(input)
	if failure.reason != "" {
		return failure
	}
	if failure := validateSelections(input, commands); failure.reason != "" {
		return failure
	}
	return validatePaths(input, commands)
}

func validateBase(input Input) validationFailure {
	if input.Schema != SchemaV1 || !validToken(input.FixtureID) || !validRoot(input.Root) ||
		input.Commands == nil || len(input.Commands) == 0 || input.Paths == nil || len(input.Paths) == 0 ||
		input.SelectedCommandIDs == nil || len(input.SelectedCommandIDs) == 0 ||
		input.FullCommandIDs == nil || len(input.FullCommandIDs) == 0 {
		return validationFailure{DecisionUnknown, ReasonMissingInput}
	}
	return validateCeilings(input.Ceilings)
}

func validateCommands(input Input) (map[string]CommandRecord, validationFailure) {
	commands := make(map[string]CommandRecord, len(input.Commands))
	pressureIDs := map[string]struct{}{}
	for _, command := range input.Commands {
		if failure := validateCommand(input.Root, command, commands, pressureIDs); failure.reason != "" {
			return nil, failure
		}
		commands[command.ID] = command
	}
	return commands, validationFailure{}
}

func validateCommand(root string, command CommandRecord, commands map[string]CommandRecord, pressureIDs map[string]struct{}) validationFailure {
	if !validToken(command.ID) || !validToken(command.Path) {
		return validationFailure{DecisionFailClosed, ReasonInvalidPath}
	}
	if _, exists := commands[command.ID]; exists {
		return validationFailure{DecisionFailClosed, ReasonDuplicateID}
	}
	if _, ok := canonicalRelativePath(root, command.Path); !ok {
		return validationFailure{DecisionFailClosed, ReasonInvalidPath}
	}
	if command.CPUCoreNS == nil || command.MemoryBytes == nil || command.PeakMemoryBytes == nil || command.WorkUnits == nil {
		return validationFailure{DecisionUnknown, ReasonMissingResource}
	}
	if command.Pressures == nil {
		return validationFailure{DecisionUnknown, ReasonMissingResource}
	}
	local := map[string]struct{}{}
	for _, pressure := range command.Pressures {
		if failure := validatePressure(pressure, local, pressureIDs); failure.reason != "" {
			return failure
		}
	}
	return validationFailure{}
}

func validatePressure(pressure PressureRecord, local, global map[string]struct{}) validationFailure {
	if !validToken(pressure.ID) || !validToken(pressure.IndependenceGroupID) {
		return validationFailure{DecisionFailClosed, ReasonInvalidPressure}
	}
	if pressure.Applicable == nil {
		return validationFailure{DecisionUnknown, ReasonMissingResource}
	}
	if _, exists := local[pressure.ID]; exists {
		return validationFailure{DecisionFailClosed, ReasonDuplicateID}
	}
	if _, exists := global[pressure.ID]; exists {
		return validationFailure{DecisionFailClosed, ReasonDuplicateID}
	}
	local[pressure.ID], global[pressure.ID] = struct{}{}, struct{}{}
	return validationFailure{}
}

func validateSelections(input Input, commands map[string]CommandRecord) validationFailure {
	if failure := validateSelection(input.SelectedCommandIDs, commands, nil); failure.reason != "" {
		return failure
	}
	if failure := validateSelection(input.FullCommandIDs, commands, input.SelectedCommandIDs); failure.reason != "" {
		return failure
	}
	full := map[string]struct{}{}
	for _, id := range input.FullCommandIDs {
		full[id] = struct{}{}
	}
	if len(full) != len(commands) {
		return validationFailure{DecisionFailClosed, ReasonSelectionInvalid}
	}
	for id := range commands {
		if _, exists := full[id]; !exists {
			return validationFailure{DecisionFailClosed, ReasonSelectionInvalid}
		}
	}
	return validationFailure{}
}

func validateSelection(values []string, commands map[string]CommandRecord, subset []string) validationFailure {
	seen := map[string]struct{}{}
	for _, id := range values {
		if !validToken(id) {
			return validationFailure{DecisionFailClosed, ReasonSelectionInvalid}
		}
		if _, exists := seen[id]; exists {
			return validationFailure{DecisionFailClosed, ReasonDuplicateID}
		}
		if _, exists := commands[id]; !exists {
			return validationFailure{DecisionFailClosed, ReasonDanglingID}
		}
		seen[id] = struct{}{}
	}
	if subset == nil {
		return validationFailure{}
	}
	for _, id := range subset {
		if _, exists := seen[id]; !exists {
			return validationFailure{DecisionFailClosed, ReasonSelectionInvalid}
		}
	}
	return validationFailure{}
}

func validatePaths(input Input, commands map[string]CommandRecord) validationFailure {
	paths := map[string]PathRecord{}
	pathNames, recordIDs := map[string]struct{}{}, map[string]struct{}{}
	byCommand := map[string]int{}
	for _, path := range input.Paths {
		canonical, failure := validatePath(input.Root, path, commands, paths, pathNames, recordIDs)
		if failure.reason != "" {
			return failure
		}
		paths[path.ID], pathNames[canonical] = path, struct{}{}
		byCommand[path.CommandID]++
	}
	for id := range commands {
		if byCommand[id] == 0 {
			return validationFailure{DecisionUnknown, ReasonMissingPROV}
		}
	}
	return validationFailure{}
}

func validatePath(root string, path PathRecord, commands map[string]CommandRecord, paths map[string]PathRecord, names, records map[string]struct{}) (string, validationFailure) {
	if !validToken(path.ID) || !validToken(path.Path) {
		return "", validationFailure{DecisionFailClosed, ReasonInvalidPath}
	}
	canonical, ok := canonicalRelativePath(root, path.Path)
	if !ok {
		return "", validationFailure{DecisionFailClosed, ReasonInvalidPath}
	}
	if _, exists := paths[path.ID]; exists {
		return "", validationFailure{DecisionFailClosed, ReasonDuplicatePath}
	}
	if _, exists := names[canonical]; exists {
		return "", validationFailure{DecisionFailClosed, ReasonDuplicatePath}
	}
	if _, exists := commands[path.CommandID]; !exists {
		return "", validationFailure{DecisionFailClosed, ReasonDanglingID}
	}
	if failure := validatePathFields(path); failure.reason != "" {
		return "", failure
	}
	if failure := validateRecordIDs(path.RecordIDs, records); failure.reason != "" {
		return "", failure
	}
	return canonical, validationFailure{}
}

func validatePathFields(path PathRecord) validationFailure {
	if path.RecordIDs == nil || len(path.RecordIDs) == 0 || path.Finite == nil ||
		path.ClosureNumerator == nil || path.ClosureDenominator == nil {
		return validationFailure{DecisionUnknown, ReasonMissingPROV}
	}
	if *path.ClosureNumerator > *path.ClosureDenominator {
		return validationFailure{DecisionFailClosed, ReasonClosureInvalid}
	}
	return validationFailure{}
}

func validateRecordIDs(values []string, global map[string]struct{}) validationFailure {
	local := map[string]struct{}{}
	for _, recordID := range values {
		if !validToken(recordID) {
			return validationFailure{DecisionFailClosed, ReasonInvalidPath}
		}
		if _, exists := local[recordID]; exists {
			return validationFailure{DecisionFailClosed, ReasonDuplicateRecord}
		}
		if _, exists := global[recordID]; exists {
			return validationFailure{DecisionFailClosed, ReasonDuplicateRecord}
		}
		local[recordID], global[recordID] = struct{}{}, struct{}{}
	}
	return validationFailure{}
}

func validateCeilings(ceilings ResourceCeilings) validationFailure {
	if !ceilingComplete(ceilings.Selected) || !ceilingComplete(ceilings.Full) {
		return validationFailure{DecisionUnknown, ReasonMissingInput}
	}
	return validationFailure{}
}

func ceilingComplete(ceiling CeilingSet) bool {
	return ceiling.CPUCoreNS != nil && ceiling.MemoryBytes != nil && ceiling.PeakMemoryBytes != nil &&
		ceiling.WorkUnits != nil && ceiling.AffectedStableIDs != nil && ceiling.ApplicablePressures != nil &&
		ceiling.IndependentGroups != nil && ceiling.UniquePROVRecords != nil && ceiling.FinitePROVPaths != nil &&
		ceiling.ClosureNumerator != nil && ceiling.ClosureDenominator != nil
}

func validRoot(root string) bool {
	return root != "" && root == strings.TrimSpace(root) && !strings.ContainsAny(root, "\x00")
}

func validToken(value string) bool {
	if value == "" || value != strings.TrimSpace(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}
