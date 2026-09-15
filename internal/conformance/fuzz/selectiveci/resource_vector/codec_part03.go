package resourcevector

import (
	"sort"
)

func canonicalInput(input Input) canonicalInputView {
	commands := append([]CommandRecord(nil), input.Commands...)
	for index := range commands {
		if path, ok := canonicalRelativePath(input.Root, commands[index].Path); ok {
			commands[index].Path = path
		}
		commands[index].Pressures = append([]PressureRecord(nil), commands[index].Pressures...)
		commands[index].AffectedStableIDs = sortedStrings(commands[index].AffectedStableIDs)
		sort.Slice(commands[index].Pressures, func(left, right int) bool {
			return pressureKey(commands[index].Pressures[left]) < pressureKey(commands[index].Pressures[right])
		})
	}
	sort.Slice(commands, func(left, right int) bool { return commands[left].ID < commands[right].ID })
	paths := append([]PathRecord(nil), input.Paths...)
	for index := range paths {
		if path, ok := canonicalRelativePath(input.Root, paths[index].Path); ok {
			paths[index].Path = path
		}
		paths[index].RecordIDs = sortedStrings(paths[index].RecordIDs)
	}
	sort.Slice(paths, func(left, right int) bool {
		if paths[left].ID != paths[right].ID {
			return paths[left].ID < paths[right].ID
		}
		if paths[left].Path != paths[right].Path {
			return paths[left].Path < paths[right].Path
		}
		return paths[left].CommandID < paths[right].CommandID
	})
	return canonicalInputView{
		Schema: input.Schema, FixtureID: input.FixtureID, Commands: commands, Paths: paths,
		AffectedStableIDs:  sortedStrings(input.AffectedStableIDs),
		SelectedCommandIDs: sortedStrings(input.SelectedCommandIDs), FullCommandIDs: sortedStrings(input.FullCommandIDs),
		Ceilings: input.Ceilings,
	}
}

type canonicalInputView struct {
	Schema             string           `json:"schema"`
	FixtureID          string           `json:"fixture_id"`
	Commands           []CommandRecord  `json:"commands"`
	Paths              []PathRecord     `json:"paths"`
	AffectedStableIDs  []string         `json:"affected_stable_ids"`
	SelectedCommandIDs []string         `json:"selected_command_ids"`
	FullCommandIDs     []string         `json:"full_command_ids"`
	Ceilings           ResourceCeilings `json:"ceilings"`
}
