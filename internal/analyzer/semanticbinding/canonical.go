package semanticbinding

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

const canonicalSchema = "gooo/semantic-binding/v1"

func makeResult(bindings []Binding, obligations []Obligation) Result {
	sortRecords(bindings, obligations)
	for index := range bindings {
		bindings[index].Digest = digestString(bindings[index].Canonical())
		bindings[index].CanonicalDigest = bindings[index].Digest
	}
	for index := range obligations {
		obligations[index].Digest = digestString(obligations[index].Canonical())
		obligations[index].CanonicalDigest = obligations[index].Digest
	}
	result := Result{Status: StatusBound, Bindings: bindings, Obligations: obligations}
	result.Digest = digestString(result.Canonical())
	result.CanonicalDigest = result.Digest
	return result
}

// Canonical returns the versioned, order-independent record encoding.
func (r Result) Canonical() string {
	bindings := append([]Binding(nil), r.Bindings...)
	obligations := append([]Obligation(nil), r.Obligations...)
	sortRecords(bindings, obligations)
	var builder strings.Builder
	writeField(&builder, canonicalSchema)
	writeField(&builder, strconv.Itoa(len(bindings)))
	for _, binding := range bindings {
		writeField(&builder, "binding")
		writeField(&builder, binding.Canonical())
	}
	writeField(&builder, strconv.Itoa(len(obligations)))
	for _, obligation := range obligations {
		writeField(&builder, "obligation")
		writeField(&builder, obligation.Canonical())
	}
	return builder.String()
}

// StableHash returns the canonical SHA-256 digest of the result.
func (r Result) StableHash() string { return digestString(r.Canonical()) }

// Canonical returns the canonical encoding of a binding without its digest.
func (b Binding) Canonical() string {
	var builder strings.Builder
	writeField(&builder, "binding")
	writeField(&builder, b.ID)
	writeField(&builder, string(b.Role))
	writeField(&builder, b.PackagePath)
	writeField(&builder, b.DeclarationKey)
	writeSpan(&builder, b.Span)
	writeSpan(&builder, b.DirectiveSpan)
	return builder.String()
}

// StableHash returns the canonical SHA-256 digest of a binding.
func (b Binding) StableHash() string { return digestString(b.Canonical()) }

// Canonical returns the canonical encoding of an obligation without its
// digest.
func (o Obligation) Canonical() string {
	var builder strings.Builder
	writeField(&builder, "obligation")
	writeField(&builder, o.ID)
	writeField(&builder, o.Subject)
	writeField(&builder, o.Pressure)
	writeField(&builder, o.PackagePath)
	writeField(&builder, o.DeclarationKey)
	writeSpan(&builder, o.Span)
	writeSpan(&builder, o.DirectiveSpan)
	return builder.String()
}

// StableHash returns the canonical SHA-256 digest of an obligation.
func (o Obligation) StableHash() string { return digestString(o.Canonical()) }

func writeField(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('\n')
}

func writeSpan(builder *strings.Builder, span Span) {
	writeField(builder, span.Filename)
	writeField(builder, strconv.Itoa(span.Start.Offset))
	writeField(builder, strconv.Itoa(span.Start.Line))
	writeField(builder, strconv.Itoa(span.Start.Column))
	writeField(builder, strconv.Itoa(span.End.Offset))
	writeField(builder, strconv.Itoa(span.End.Line))
	writeField(builder, strconv.Itoa(span.End.Column))
}

func digestString(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func sortRecords(bindings []Binding, obligations []Obligation) {
	sort.SliceStable(bindings, func(left, right int) bool {
		return compareBinding(bindings[left], bindings[right]) < 0
	})
	sort.SliceStable(obligations, func(left, right int) bool {
		return compareObligation(obligations[left], obligations[right]) < 0
	})
}

func compareBinding(left, right Binding) int {
	if value := strings.Compare(left.PackagePath, right.PackagePath); value != 0 {
		return value
	}
	if value := strings.Compare(left.DeclarationKey, right.DeclarationKey); value != 0 {
		return value
	}
	if value := strings.Compare(left.Span.Filename, right.Span.Filename); value != 0 {
		return value
	}
	if left.Span.Start.Offset != right.Span.Start.Offset {
		return compareInt(left.Span.Start.Offset, right.Span.Start.Offset)
	}
	if value := strings.Compare(left.ID, right.ID); value != 0 {
		return value
	}
	return strings.Compare(string(left.Role), string(right.Role))
}

func compareObligation(left, right Obligation) int {
	if value := strings.Compare(left.PackagePath, right.PackagePath); value != 0 {
		return value
	}
	if value := strings.Compare(left.DeclarationKey, right.DeclarationKey); value != 0 {
		return value
	}
	if value := strings.Compare(left.Span.Filename, right.Span.Filename); value != 0 {
		return value
	}
	if left.Span.Start.Offset != right.Span.Start.Offset {
		return compareInt(left.Span.Start.Offset, right.Span.Start.Offset)
	}
	return strings.Compare(left.ID, right.ID)
}

func compareInt(left, right int) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
