package main

import "os"

func removeSplitTemporary(item stagedPart, observer splitObserver) {
	if item.temporary == "" {
		return
	}
	err := os.Remove(item.temporary)
	emitSplit(observer, "REMOVE_TEMP", item.target, item.temporary, err == nil || os.IsNotExist(err))
}

func removeSplitTarget(target string, observer splitObserver) {
	err := os.Remove(target)
	emitSplit(observer, "REMOVE_ROLLBACK", target, "", err == nil || os.IsNotExist(err))
}
