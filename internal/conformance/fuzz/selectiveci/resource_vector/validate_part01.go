package resourcevector

func validate(input Input) validationFailure {
	if failure := validateBase(input); failure.reason != "" {
		return failure
	}
	commands, failure := validateCommands(input)
	if failure.reason != "" {
		return failure
	}
	if failure := validateAffected(input, commands); failure.reason != "" {
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
