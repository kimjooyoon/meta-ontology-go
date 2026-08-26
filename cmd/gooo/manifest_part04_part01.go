package main

import (
	"fmt"
	"strings"
)

func ownedSlotBodies(source []byte) (map[string][]byte, error) {
	result := make(map[string][]byte)
	var open *ownedSlot
	for lineStart := 0; lineStart < len(source); {
		lineEnd := lineStart + 1
		for lineEnd < len(source) && source[lineEnd-1] != '\n' {
			lineEnd++
		}
		line := strings.TrimSpace(string(source[lineStart:lineEnd]))
		switch {
		case strings.HasPrefix(line, "//gooo:slot:start"):
			if open != nil {
				return nil, fmt.Errorf("nested slot %q", open.id)
			}
			id, err := markerID(line, "//gooo:slot:start")
			if err != nil {
				return nil, err
			}
			open = &ownedSlot{id: id, bodyStart: lineEnd}
		case strings.HasPrefix(line, "//gooo:slot:end"):
			if open == nil {
				return nil, fmt.Errorf("slot end without start")
			}
			id, err := markerID(line, "//gooo:slot:end")
			if err != nil {
				return nil, err
			}
			if id != open.id {
				return nil, fmt.Errorf("slot %q closes as %q", open.id, id)
			}
			if _, exists := result[id]; exists {
				return nil, fmt.Errorf("duplicate slot %q", id)
			}
			result[id] = append([]byte(nil), source[open.bodyStart:lineStart]...)
			open = nil
		}
		lineStart = lineEnd
	}
	if open != nil {
		return nil, fmt.Errorf("unterminated slot %q", open.id)
	}
	return result, nil
}
