package main

func publishAtomicWrites(writes []atomicWrite, changed []bool, temps []string, snapshots []outputSnapshot, ops atomicFileOps) error {
	committed := make([]int, 0, len(writes))
	for index, write := range writes {
		if !changed[index] {
			continue
		}
		if err := ops.rename(temps[index], write.path); err != nil {
			return atomicPublishError(writes, snapshots, committed, write.path, err)
		}
		temps[index] = ""
		committed = append(committed, index)
	}
	for directory := range atomicWriteDirectories(writes) {
		if err := ops.syncDir(directory); err != nil {
			return atomicPublishError(writes, snapshots, committed, directory, err)
		}
	}
	return nil
}
