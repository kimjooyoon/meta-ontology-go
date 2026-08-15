package generator

import (
	"fmt"
	"sort"
	"strings"
)

// BeginMarker and EndMarker are the legacy marker spellings accepted by
// MergeGenerated.  New output uses the attribute-bearing start/end markers in
// markers.go; accepting this form keeps early generated files migratable.
const (
	BeginMarker = "//gooo:generated begin"
	EndMarker   = "//gooo:generated end"
)

// MergeGenerated replaces legacy generated blocks in existing with blocks of
// the same identity from fresh.  Text outside generated blocks is untouched.
// It is intentionally independent of the semantic renderer so it can be used
// as a safe migration step for outputs made by an earlier prototype.
func MergeGenerated(existing, fresh string) (string, error) {
	existingBlocks, err := legacyBlocks(existing)
	if err != nil {
		return "", err
	}
	freshBlocks, err := legacyBlocks(fresh)
	if err != nil {
		return "", err
	}
	if len(freshBlocks) == 0 {
		return existing, nil
	}
	existingOrdered := make([]legacyBlock, 0, len(existingBlocks))
	for _, block := range existingBlocks {
		existingOrdered = append(existingOrdered, block)
	}
	sort.Slice(existingOrdered, func(i, j int) bool { return existingOrdered[i].start < existingOrdered[j].start })
	var output strings.Builder
	cursor := 0
	for _, block := range existingOrdered {
		output.WriteString(existing[cursor:block.start])
		if replacement, ok := freshBlocks[block.id]; ok {
			output.WriteString(replacement.text)
		} else {
			output.WriteString(existing[block.start:block.end])
		}
		cursor = block.end
	}
	output.WriteString(existing[cursor:])
	freshIDs := make([]string, 0, len(freshBlocks))
	for id := range freshBlocks {
		freshIDs = append(freshIDs, id)
	}
	sort.Strings(freshIDs)
	for _, id := range freshIDs {
		block := freshBlocks[id]
		found := false
		for _, existingBlock := range existingOrdered {
			if existingBlock.id == id {
				found = true
				break
			}
		}
		if !found {
			if output.Len() > 0 && !strings.HasSuffix(output.String(), "\n") {
				output.WriteByte('\n')
			}
			output.WriteString(block.text)
		}
	}
	return output.String(), nil
}

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
