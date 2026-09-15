package resourcevector

import (
	"sort"
)

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
