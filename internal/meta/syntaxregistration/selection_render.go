package syntaxregistration

import (
	"fmt"
	"go/format"
	"slices"
	"strings"
)

const selectionChunkSize = 16

func renderSelection(name, kind, empty string, entries []selectionEntry) (string, string) {
	slices.SortFunc(entries, func(a, b selectionEntry) int { return a.capability - b.capability })
	var wrapper, chunks strings.Builder
	fmt.Fprintf(&wrapper, "func %s() %s {\n", name, kind)
	index := 0
	for group := range slices.Chunk(entries, selectionChunkSize) {
		call := fmt.Sprintf("%sChunk%d(languagesyntax.FixedCapabilityTotal)", name, index)
		if (index+1)*selectionChunkSize >= len(entries) {
			fmt.Fprintf(&wrapper, "return %s\n", call)
		} else {
			fmt.Fprintf(&wrapper, "if value := %s; value != %s {return value}\n", call, empty)
		}
		fmt.Fprintf(&chunks, "\nfunc %sChunk%d(total int) %s {\nswitch total {\n", name, index, kind)
		for _, entry := range group {
			fmt.Fprintf(&chunks, "case %d: return %s\n", entry.capability, entry.result)
		}
		fmt.Fprintf(&chunks, "}\nreturn %s\n}\n", empty)
		index++
	}
	wrapper.WriteString("}\n")
	return wrapper.String(), chunks.String()
}

func formatSelection(declaration string) string {
	formatted, err := format.Source([]byte("package selection\n" + declaration))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(string(formatted), "package selection\n"))
}
