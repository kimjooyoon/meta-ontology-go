package generator

import (
	"fmt"
	"strings"
)

type legacyBlock struct {
	id    string
	start int
	end   int
	text  string
}

func legacyBlocks(source string) (map[string]legacyBlock, error) {
	result := make(map[string]legacyBlock)
	var open *legacyBlock
	for _, line := range splitSourceLines([]byte(source)) {
		kind, id, ok, err := parseLegacyMarker(line.text)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if kind == "start" {
			if open != nil {
				return nil, fmt.Errorf("generator: nested legacy generated block %q", open.id)
			}
			open = &legacyBlock{id: id, start: line.start}
			continue
		}
		if open == nil {
			return nil, fmt.Errorf("generator: legacy generated end without start")
		}
		if _, exists := result[open.id]; exists {
			return nil, fmt.Errorf("generator: duplicate legacy generated block %q", open.id)
		}
		result[open.id] = legacyBlock{id: open.id, start: open.start, end: line.end, text: source[open.start:line.end]}
		open = nil
	}
	if open != nil {
		return nil, fmt.Errorf("generator: legacy generated block %q is unterminated", open.id)
	}
	return result, nil
}
func parseLegacyMarker(line string) (string, string, bool, error) {
	trimmed := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(trimmed, BeginMarker):
		rest := trimmed[len(BeginMarker):]
		if rest == "" || (rest[0] != ' ' && rest[0] != '\t') {
			return "", "", false, fmt.Errorf("generator: legacy generated marker has invalid line boundary")
		}
		id := strings.TrimSpace(rest)
		if id == "" || len(strings.Fields(id)) != 1 {
			return "", "", false, fmt.Errorf("generator: legacy generated marker has invalid ID")
		}
		return "start", id, true, nil
	case strings.HasPrefix(trimmed, EndMarker):
		if strings.TrimSpace(trimmed[len(EndMarker):]) != "" {
			return "", "", false, fmt.Errorf("generator: legacy generated end marker has unexpected attributes")
		}
		return "end", "", true, nil
	default:
		return "", "", false, nil
	}
}
