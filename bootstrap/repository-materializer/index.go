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
		fmt.Fprintf(&records, "%s %s\t%s%c", gitMode(item), oid, item.Logical, byte(0))
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

func gitMode(item entry) string {
	if item.Kind == "symlink" {
		return "120000"
	}
	if item.Mode&0o111 != 0 {
		return "100755"
	}
	return "100644"
}

func unexpectedPhysical(tracked map[string]string, model manifest) int {
	expected := map[string]bool{"projection/catalog/manifest.json": true}
	for _, item := range model.Entries {
		expected[item.Backing] = true
	}
	unexpected := 0
	for name := range tracked {
		if !expected[name] {
			unexpected++
		}
	}
	return unexpected
}
