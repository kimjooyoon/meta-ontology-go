package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func buildLogicalIndex(settings config, model manifest, destination, head string) (indexState, error) {
	state := indexState{}
	tracked, err := trackedBlobs(settings.root)
	if err != nil {
		return state, err
	}
	if err := os.MkdirAll(filepath.Dir(settings.index), 0o755); err != nil {
		return state, err
	}
	if _, err := gitBytes(settings.root, settings.index, destination, true, nil, "read-tree", "--empty"); err != nil {
		return state, err
	}
	var records bytes.Buffer
	for _, item := range model.Entries {
		oid, ok := tracked[item.Backing]
		if !ok && !samePath(settings.root, settings.physical) {
			oid, ok = tracked[item.Logical]
		}
		if !ok {
			state.Unbound++
			continue
		}
		fmt.Fprintf(&records, "%s %s\t%s%c", item.gitMode(), oid, item.Logical, byte(0))
	}
	if state.Unbound == 0 {
		_, err = gitBytes(settings.root, settings.index, destination, true, records.Bytes(), "update-index", "-z", "--index-info")
	}
	if err != nil {
		return state, err
	}
	state.TreeOID, err = gitText(settings.root, settings.index, destination, true, nil, "write-tree")
	if err != nil {
		return state, err
	}
	state.Replacement, err = bindLogicalHead(settings.root, settings.index, destination, head, state.TreeOID)
	if err != nil {
		return state, err
	}
	status, err := gitText(settings.root, settings.index, destination, false, nil, "status", "--porcelain")
	if err != nil {
		return state, err
	}
	if status != "" {
		state.Dirty = 1
	}
	if samePath(settings.root, settings.physical) {
		state.Unexpected = unexpectedPhysical(tracked, model)
	}
	return state, nil
}
