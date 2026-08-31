package writeset

func validSnapshot(snapshot Snapshot) bool {
	if snapshot.Schema != SnapshotSchema || snapshot.RootDigest != digestEntries(snapshot.Entries) {
		return false
	}
	previous := ""
	for index, entry := range snapshot.Entries {
		paths, valid := normalizePaths([]string{entry.Path})
		if !valid || len(paths) != 1 || paths[0] != entry.Path || entry.Digest == "" {
			return false
		}
		if index > 0 && entry.Path <= previous {
			return false
		}
		previous = entry.Path
	}
	return true
}
