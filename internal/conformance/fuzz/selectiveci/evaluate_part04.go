package selectiveci

import (
	"math"
	"sort"
)

func buildResult(digest string, selection changeSelection, commands map[string]Command) Result {
	ids := make([]string, 0, len(selection.owners))
	for id := range selection.owners {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	argv := make(map[string][]string, len(ids))
	var cpu, memory uint64
	for _, id := range ids {
		command := commands[id]
		if math.MaxUint64-cpu < command.CPUUnits {
			return fallbackResult(digest, cpuOverflowReason)
		}
		cpu += command.CPUUnits
		if math.MaxUint64-memory < command.MemoryCeiling {
			return fallbackResult(digest, memoryOverflowReason)
		}
		memory += command.MemoryCeiling
		argv[id] = append([]string(nil), command.Argv...)
	}
	return Result{
		Decision:        Selective,
		Reason:          completeReason,
		CommandIDs:      ids,
		Argv:            argv,
		CPUUnits:        cpu,
		MemoryCeiling:   memory,
		PathCount:       len(selection.changedPaths),
		CanonicalDigest: digest,
	}
}
func changeMatchesEvidence(change PathChange, path PathEvidence) bool {
	switch change.Kind {
	case ChangeDelete:
		return !path.Present && change.Blob == ""
	case ChangeAdd, ChangeModify, ChangeRelocate:
		return path.Present && change.Blob != "" && change.Blob == path.Blob
	default:
		return false
	}
}
