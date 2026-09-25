package generator

import (
	"fmt"
	"go/format"
	"sort"
	"strings"
)

func restoreSlotBodies(source []byte, original map[string]parsedSlot) ([]byte, error) {
	if len(original) == 0 {
		return source, nil
	}
	formatted, err := parseMarkers(source)
	if err != nil {
		return nil, fmt.Errorf("generator: parse formatted slot markers: %w", err)
	}
	slots := make([]parsedSlot, 0, len(original))
	for id, previous := range original {
		current, exists := formatted.Slots[id]
		if !exists {
			return nil, fmt.Errorf("generator: formatted block lost slot %q", id)
		}
		current.Body = previous.Body
		slots = append(slots, current)
	}
	sort.Slice(slots, func(i, j int) bool { return slots[i].Start > slots[j].Start })
	result := append([]byte(nil), source...)
	for _, slot := range slots {
		result = replaceBytes(result, slot.Start, slot.End, slot.Body)
	}
	return result, nil
}
func replaceBytes(source []byte, start, end int, replacement []byte) []byte {
	result := make([]byte, 0, len(source)-end+start+len(replacement))
	result = append(result, source[:start]...)
	result = append(result, replacement...)
	result = append(result, source[end:]...)
	return result
}
func (g Generator) renderNewFile(ir SemanticIR, blocks map[string][]byte, order []string) ([]byte, error) {
	header := g.Options.Header
	if header == "" {
		header = defaultHeader
	}
	var source strings.Builder
	source.WriteString(header)
	source.WriteString("\npackage ")
	source.WriteString(ir.Package)
	source.WriteString("\n")
	if len(ir.Imports) > 0 {
		source.WriteString("\nimport (\n")
		for _, item := range ir.Imports {
			if item.Name == "" {
				fmt.Fprintf(&source, "\t%s\n", quotePath(item.Path))
			} else {
				fmt.Fprintf(&source, "\t%s %s\n", item.Name, quotePath(item.Path))
			}
		}
		source.WriteString(")\n")
	}
	for _, id := range order {
		source.WriteString("\n")
		source.Write(blocks[id])
	}
	formatted, err := format.Source([]byte(source.String()))
	if err != nil {
		return nil, fmt.Errorf("generator: format new file: %w", err)
	}
	return formatted, nil
}
