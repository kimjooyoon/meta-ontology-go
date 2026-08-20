package main

import "go/format"

func compactDensity(name string, original []byte) ([]byte, int, error) {
	current := append([]byte(nil), original...)
	operations := 0
	for densityLines(current) > 75 {
		spans, err := guardCandidates(name, current)
		if err != nil {
			return nil, operations, err
		}
		reduced := false
		for _, span := range spans {
			inline, ok := oneLineTokens(current[span.start:span.end])
			if !ok {
				continue
			}
			candidate := make([]byte, 0, len(current))
			candidate = append(candidate, current[:span.start]...)
			candidate = append(candidate, inline...)
			candidate = append(candidate, current[span.end:]...)
			formatted, formatErr := format.Source(candidate)
			if formatErr != nil || densityLines(formatted) >= densityLines(current) {
				continue
			}
			current = formatted
			operations++
			reduced = true
			break
		}
		if !reduced {
			break
		}
	}
	return current, operations, nil
}
