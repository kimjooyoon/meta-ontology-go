package toolchaincli

func snapshot(root string) (treeSnapshot, error) {
	records, err := scanTree(root)
	if err != nil {
		return treeSnapshot{}, err
	}
	byPath := make(map[string]treeRecord, len(records))
	for _, record := range records {
		byPath[record.Path] = record
	}
	return treeSnapshot{Digest: digestJSON(records), Records: byPath}, nil
}

func changedFiles(before, after treeSnapshot) int {
	changed := 0
	seen := make(map[string]bool, len(before.Records)+len(after.Records))
	for path := range before.Records {
		seen[path] = true
	}
	for path := range after.Records {
		seen[path] = true
	}
	for path := range seen {
		if before.Records[path] != after.Records[path] {
			changed++
		}
	}
	return changed
}
