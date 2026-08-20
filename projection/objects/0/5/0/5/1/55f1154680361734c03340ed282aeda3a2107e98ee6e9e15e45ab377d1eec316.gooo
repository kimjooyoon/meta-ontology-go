package main

import (
	"fmt"
	"strings"
)

func analyzeMarkerAliases(source []byte) ([]analyzeMarkerAlias, error) {
	regions, _, err := analyzeMarkers(source)
	if err != nil {
		return nil, err
	}
	aliases := make([]analyzeMarkerAlias, 0, len(regions))
	for _, region := range regions {
		aliases = append(aliases, analyzeMarkerAlias{id: region.id})
	}
	lines := strings.Split(string(source), "\n")
	for i := range aliases {
		for lineIndex, line := range lines {
			if !strings.HasPrefix(strings.TrimSpace(line), "//gooo:generated:start") {
				continue
			}
			id, _ := markerID(strings.TrimSpace(line), "//gooo:generated:start")
			if id != aliases[i].id {
				continue
			}
			for _, next := range lines[lineIndex+1:] {
				next = strings.TrimSpace(next)
				if next == "" || strings.HasPrefix(next, "//") {
					continue
				}
				fields := strings.Fields(next)
				if len(fields) >= 2 {
					aliases[i].name = strings.Trim(fields[1], "(")
				}
				break
			}
		}
		if aliases[i].name == "" {
			return nil, fmt.Errorf("generated marker %q has no declaration", aliases[i].id)
		}
	}
	return aliases, nil
}
