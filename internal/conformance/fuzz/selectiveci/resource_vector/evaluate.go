package resourcevector

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

type validationFailure struct {
	decision Decision
	reason   Reason
}

type indexedRecords struct {
	commands map[string]CommandRecord
	paths    map[string]PathRecord
	byCmd    map[string][]PathRecord
}

type vectorMeta struct {
	pathCount int
	allFinite bool
}

// Evaluate strictly replays canonical command and path records. It never
// trusts an expected vector and never imports a production selector.
func Evaluate(input Input) Output {
	output := Output{
		Schema: input.Schema, FixtureID: input.FixtureID, InputDigest: CanonicalInputDigest(input),
		LimitFailures: []string{}, FullSuiteRequired: true,
	}
	if failure := validate(input); failure.reason != "" {
		return finish(output, failure.decision, failure.reason)
	}
	indexed := index(input)
	selected, selectedMeta, ok := replayVector(indexed, input.SelectedCommandIDs)
	if !ok {
		return finish(output, DecisionUnknown, ReasonOverflow)
	}
	full, _, ok := replayVector(indexed, input.FullCommandIDs)
	if !ok {
		return finish(output, DecisionUnknown, ReasonOverflow)
	}
	output.Selected, output.Full = &selected, &full
	output.LimitFailures = append(output.LimitFailures, compareCeilings(selected, input.Ceilings.Selected, "selected")...)
	output.LimitFailures = append(output.LimitFailures, compareCeilings(full, input.Ceilings.Full, "full")...)
	if len(output.LimitFailures) != 0 {
		return finish(output, DecisionUnknown, ReasonCeilingExceeded)
	}
	output.Decision, output.Reason, output.FullSuiteRequired = DecisionPass, ReasonNone, false
	output.ProofValid = selectedMeta.pathCount > 0 && selectedMeta.allFinite &&
		selected.FinitePROVPaths == uint64(selectedMeta.pathCount) &&
		selected.ClosureDenominator > 0 && selected.ClosureNumerator == selected.ClosureDenominator
	return finish(output, output.Decision, output.Reason)
}

func Replay(input Input) Output { return Evaluate(input) }

func finish(output Output, decision Decision, reason Reason) Output {
	output.Decision, output.Reason = decision, reason
	if decision != DecisionPass {
		output.FullSuiteRequired = true
		output.ProofValid = false
	}
	output.CanonicalOutputDigest = CanonicalOutputDigest(output)
	output.ReplayDigest = ReplayDigest(output.InputDigest, output.CanonicalOutputDigest)
	return output
}

func validate(input Input) validationFailure {
	if input.Schema != SchemaV1 || !validToken(input.FixtureID) || !validRoot(input.Root) ||
		input.Commands == nil || len(input.Commands) == 0 || input.Paths == nil || len(input.Paths) == 0 ||
		input.SelectedCommandIDs == nil || len(input.SelectedCommandIDs) == 0 ||
		input.FullCommandIDs == nil || len(input.FullCommandIDs) == 0 {
		return validationFailure{DecisionUnknown, ReasonMissingInput}
	}
	if failure := validateCeilings(input.Ceilings); failure.reason != "" {
		return failure
	}
	commands := make(map[string]CommandRecord, len(input.Commands))
	allPressureIDs := map[string]struct{}{}
	for _, command := range input.Commands {
		if !validToken(command.ID) || !validToken(command.Path) {
			return validationFailure{DecisionFailClosed, ReasonInvalidPath}
		}
		if _, exists := commands[command.ID]; exists {
			return validationFailure{DecisionFailClosed, ReasonDuplicateID}
		}
		if _, ok := canonicalRelativePath(input.Root, command.Path); !ok {
			return validationFailure{DecisionFailClosed, ReasonInvalidPath}
		}
		if command.CPUCoreNS == nil || command.MemoryBytes == nil || command.PeakMemoryBytes == nil || command.WorkUnits == nil {
			return validationFailure{DecisionUnknown, ReasonMissingResource}
		}
		if command.Pressures == nil {
			return validationFailure{DecisionUnknown, ReasonMissingResource}
		}
		pressureIDs := map[string]struct{}{}
		for _, pressure := range command.Pressures {
			if !validToken(pressure.ID) || !validToken(pressure.IndependenceGroupID) {
				return validationFailure{DecisionFailClosed, ReasonInvalidPressure}
			}
			if pressure.Applicable == nil {
				return validationFailure{DecisionUnknown, ReasonMissingResource}
			}
			if _, exists := pressureIDs[pressure.ID]; exists {
				return validationFailure{DecisionFailClosed, ReasonDuplicateID}
			}
			if _, exists := allPressureIDs[pressure.ID]; exists {
				return validationFailure{DecisionFailClosed, ReasonDuplicateID}
			}
			pressureIDs[pressure.ID] = struct{}{}
			allPressureIDs[pressure.ID] = struct{}{}
		}
		commands[command.ID] = command
	}
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
	paths := map[string]PathRecord{}
	pathNames := map[string]struct{}{}
	recordIDs := map[string]struct{}{}
	byCommand := map[string]int{}
	for _, path := range input.Paths {
		if !validToken(path.ID) || !validToken(path.Path) {
			return validationFailure{DecisionFailClosed, ReasonInvalidPath}
		}
		canonicalPath, ok := canonicalRelativePath(input.Root, path.Path)
		if !ok {
			return validationFailure{DecisionFailClosed, ReasonInvalidPath}
		}
		if _, exists := paths[path.ID]; exists {
			return validationFailure{DecisionFailClosed, ReasonDuplicatePath}
		}
		if _, exists := pathNames[canonicalPath]; exists {
			return validationFailure{DecisionFailClosed, ReasonDuplicatePath}
		}
		if _, exists := commands[path.CommandID]; !exists {
			return validationFailure{DecisionFailClosed, ReasonDanglingID}
		}
		if path.RecordIDs == nil || len(path.RecordIDs) == 0 || path.Finite == nil ||
			path.ClosureNumerator == nil || path.ClosureDenominator == nil {
			return validationFailure{DecisionUnknown, ReasonMissingPROV}
		}
		if *path.ClosureNumerator > *path.ClosureDenominator {
			return validationFailure{DecisionFailClosed, ReasonClosureInvalid}
		}
		localRecords := map[string]struct{}{}
		for _, recordID := range path.RecordIDs {
			if !validToken(recordID) {
				return validationFailure{DecisionFailClosed, ReasonInvalidPath}
			}
			if _, exists := localRecords[recordID]; exists {
				return validationFailure{DecisionFailClosed, ReasonDuplicateRecord}
			}
			if _, exists := recordIDs[recordID]; exists {
				return validationFailure{DecisionFailClosed, ReasonDuplicateRecord}
			}
			localRecords[recordID] = struct{}{}
			recordIDs[recordID] = struct{}{}
		}
		paths[path.ID] = path
		pathNames[canonicalPath] = struct{}{}
		byCommand[path.CommandID]++
	}
	for id := range commands {
		if byCommand[id] == 0 {
			return validationFailure{DecisionUnknown, ReasonMissingPROV}
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
	if subset != nil {
		allowed := map[string]struct{}{}
		for _, id := range values {
			allowed[id] = struct{}{}
		}
		for _, id := range subset {
			if _, exists := allowed[id]; !exists {
				return validationFailure{DecisionFailClosed, ReasonSelectionInvalid}
			}
		}
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

func index(input Input) indexedRecords {
	result := indexedRecords{commands: map[string]CommandRecord{}, paths: map[string]PathRecord{}, byCmd: map[string][]PathRecord{}}
	for _, command := range input.Commands {
		result.commands[command.ID] = command
	}
	paths := append([]PathRecord(nil), input.Paths...)
	sort.Slice(paths, func(left, right int) bool { return paths[left].ID < paths[right].ID })
	for _, path := range paths {
		result.paths[path.ID] = path
		result.byCmd[path.CommandID] = append(result.byCmd[path.CommandID], path)
	}
	return result
}

func replayVector(records indexedRecords, selectedIDs []string) (Vector, vectorMeta, bool) {
	ids := sortedStrings(selectedIDs)
	selected := map[string]struct{}{}
	for _, id := range ids {
		selected[id] = struct{}{}
	}
	vector := Vector{PeakMemoryBytes: 0}
	groups := map[string]struct{}{}
	for _, id := range ids {
		command := records.commands[id]
		var ok bool
		if vector.CPUCoreNS, ok = add(vector.CPUCoreNS, *command.CPUCoreNS); !ok {
			return Vector{}, vectorMeta{}, false
		}
		if vector.MemoryBytes, ok = add(vector.MemoryBytes, *command.MemoryBytes); !ok {
			return Vector{}, vectorMeta{}, false
		}
		if *command.PeakMemoryBytes > vector.PeakMemoryBytes {
			vector.PeakMemoryBytes = *command.PeakMemoryBytes
		}
		if vector.WorkUnits, ok = add(vector.WorkUnits, *command.WorkUnits); !ok {
			return Vector{}, vectorMeta{}, false
		}
		if vector.AffectedStableIDs, ok = add(vector.AffectedStableIDs, 1); !ok {
			return Vector{}, vectorMeta{}, false
		}
		for _, pressure := range command.Pressures {
			if !*pressure.Applicable {
				continue
			}
			if vector.ApplicablePressures, ok = add(vector.ApplicablePressures, 1); !ok {
				return Vector{}, vectorMeta{}, false
			}
			groups[pressure.IndependenceGroupID] = struct{}{}
		}
	}
	vector.IndependentGroups = uint64(len(groups))
	meta := vectorMeta{allFinite: true}
	for _, id := range ids {
		for _, path := range records.byCmd[id] {
			meta.pathCount++
			var ok bool
			if vector.UniquePROVRecords, ok = add(vector.UniquePROVRecords, uint64(len(path.RecordIDs))); !ok {
				return Vector{}, vectorMeta{}, false
			}
			if path.Finite == nil || !*path.Finite {
				meta.allFinite = false
				continue
			}
			vector.FinitePROVPaths, ok = add(vector.FinitePROVPaths, 1)
			if !ok {
				return Vector{}, vectorMeta{}, false
			}
			if vector.ClosureNumerator, ok = add(vector.ClosureNumerator, *path.ClosureNumerator); !ok {
				return Vector{}, vectorMeta{}, false
			}
			if vector.ClosureDenominator, ok = add(vector.ClosureDenominator, *path.ClosureDenominator); !ok {
				return Vector{}, vectorMeta{}, false
			}
		}
	}
	return vector, meta, true
}

func compareCeilings(vector Vector, ceiling CeilingSet, prefix string) []string {
	failures := make([]string, 0, 11)
	checks := []struct {
		name string
		got  uint64
		max  uint64
	}{
		{"cpu_core_ns", vector.CPUCoreNS, *ceiling.CPUCoreNS},
		{"memory_bytes", vector.MemoryBytes, *ceiling.MemoryBytes},
		{"peak_memory_bytes", vector.PeakMemoryBytes, *ceiling.PeakMemoryBytes},
		{"work_units", vector.WorkUnits, *ceiling.WorkUnits},
		{"affected_stable_ids", vector.AffectedStableIDs, *ceiling.AffectedStableIDs},
		{"applicable_pressures", vector.ApplicablePressures, *ceiling.ApplicablePressures},
		{"independent_groups", vector.IndependentGroups, *ceiling.IndependentGroups},
		{"unique_prov_records", vector.UniquePROVRecords, *ceiling.UniquePROVRecords},
		{"finite_prov_paths", vector.FinitePROVPaths, *ceiling.FinitePROVPaths},
		{"closure_numerator", vector.ClosureNumerator, *ceiling.ClosureNumerator},
		{"closure_denominator", vector.ClosureDenominator, *ceiling.ClosureDenominator},
	}
	for _, check := range checks {
		if check.got > check.max {
			failures = append(failures, prefix+":"+check.name)
		}
	}
	sort.Strings(failures)
	return failures
}

func add(left, right uint64) (uint64, bool) {
	if math.MaxUint64-left < right {
		return 0, false
	}
	return left + right, true
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
