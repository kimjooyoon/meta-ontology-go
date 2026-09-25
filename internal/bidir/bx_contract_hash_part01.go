package bidir

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func documentDigest(document Document) string {
	return digest(documentCanonical(document))
}
func documentCanonical(document Document) string {
	var builder strings.Builder
	writePart(&builder, document.Package)
	writePart(&builder, document.Namespace)
	for _, declaration := range document.Declarations {
		writePart(&builder, string(declaration.Kind))
		writePart(&builder, string(declaration.ID))
		writePart(&builder, declaration.Name)
		if len(declaration.Fields) > 0 {
			writeFields(&builder, declaration.Fields)
		}
		writeMapFingerprint(&builder, declaration.Attributes)
		writeSpan(&builder, declaration.Span)
		writeReferences(&builder, declaration.Inputs)
		writeReferences(&builder, declaration.Outputs)
	}
	for _, relation := range document.Relations {
		writePart(&builder, string(relation.Kind))
		writePart(&builder, string(relation.Source))
		writePart(&builder, string(relation.Target))
		writeMapFingerprint(&builder, relation.Attributes)
		writeSpan(&builder, relation.Span)
	}
	return builder.String()
}
func writeFields(builder *strings.Builder, fields []Field) {
	fmt.Fprintf(builder, "%d|", len(fields))
	for _, field := range fields {
		writePart(builder, string(field.ID))
		writePart(builder, string(field.Parent))
		writePart(builder, field.Name)
		fmt.Fprintf(builder, "%d|", len(field.Aliases))
		for _, alias := range field.Aliases {
			writePart(builder, alias)
		}
		writePart(builder, string(field.TypeRef.ID))
		writePart(builder, string(field.TypeRef.Namespace))
		writePart(builder, field.TypeRef.Name)
		writePart(builder, string(field.Origin))
		writePart(builder, string(field.TypeRefUse.Form))
		writePart(builder, field.TypeRefUse.Spelling)
		writePart(builder, string(field.TypeRefUse.ResolvedID))
		writeSpan(builder, field.TypeRefUse.Span)
		writePart(builder, string(field.Presence))
		writePart(builder, string(field.Cardinality))
		writeSpan(builder, field.Span)
		writeSpan(builder, field.IDSpan)
		writeSpan(builder, field.NameSpan)
		writeSpan(builder, field.TypeRefSpan)
		writeSpan(builder, field.PresenceSpan)
		writeSpan(builder, field.CardinalitySpan)
	}
}
