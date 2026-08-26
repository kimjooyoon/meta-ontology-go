package main

import (
	"fmt"
	"strings"
)

func analyzeMarkers(source []byte) ([]analyzeGeneratedRegion, []string, error) {
	var regions []analyzeGeneratedRegion
	var slots []string
	lines := strings.Split(string(source), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "//gooo:generated:start") {
			id, err := markerID(line, "//gooo:generated:start")
			if err != nil {
				return nil, nil, err
			}
			regions = append(regions, analyzeGeneratedRegion{id: id})
			name := ""
			for _, next := range lines[i+1:] {
				next = strings.TrimSpace(next)
				if next == "" || strings.HasPrefix(next, "//") {
					continue
				}
				fields := strings.Fields(next)
				if len(fields) >= 2 && (fields[0] == "type" || fields[0] == "func") {
					name = strings.Trim(fields[1], "(")
				}
				break
			}
			if name == "" {
				return nil, nil, fmt.Errorf("generated marker %q has no declaration", id)
			}
		}
		if strings.HasPrefix(line, "//gooo:slot:start") {
			id, err := markerID(line, "//gooo:slot:start")
			if err != nil {
				return nil, nil, err
			}
			slots = append(slots, id)
		}
	}
	return regions, slots, nil
}
