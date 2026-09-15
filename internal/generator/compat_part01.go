package generator

import (
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
