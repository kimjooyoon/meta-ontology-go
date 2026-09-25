package resourcevector

func replayVector(records indexedRecords, selectedIDs []string) (Vector, vectorMeta, bool) {
	ids := sortedStrings(selectedIDs)
	selected := map[string]struct{}{}
	for _, id := range ids {
		selected[id] = struct{}{}
	}
	vector := Vector{PeakMemoryBytes: 0}
	affected := map[string]struct{}{}
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
		for _, stableID := range command.AffectedStableIDs {
			affected[stableID] = struct{}{}
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
	vector.AffectedStableIDs = uint64(len(affected))
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
