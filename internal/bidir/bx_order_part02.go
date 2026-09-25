package bidir

import (
	"strings"
)

func spanText(span SourceSpan) string {
	var builder strings.Builder
	writeSpan(&builder, span)
	return builder.String()
}
func sequenceHash(sequence []string) string {
	var builder strings.Builder
	for _, value := range sequence {
		writePart(&builder, value)
	}
	return digest(builder.String())
}
