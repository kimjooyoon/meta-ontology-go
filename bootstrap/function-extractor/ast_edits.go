package main

import (
	"bytes"
	"sort"
)

type astEdit struct {
	start, end int
	replacement []byte
}

func applyGenericEdits(source []byte, edits []astEdit) ([]byte, error) {
	sort.SliceStable(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	var out bytes.Buffer
	last := 0
	for _, edit := range edits {
		if edit.start < last || edit.start < 0 || edit.end < edit.start || edit.end > len(source) {
			return nil, extractionError("rewrite-source", "apply-edits", "WRITE_SET_ESCAPE", "KNOWN_CONTRADICTION", "report-contradiction", []string{})
		}
		out.Write(source[last:edit.start])
		out.Write(edit.replacement)
		last = edit.end
	}
	out.Write(source[last:])
	return out.Bytes(), nil
}
