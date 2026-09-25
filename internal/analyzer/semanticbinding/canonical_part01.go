package semanticbinding

import (
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
	return builder.String()
}

// StableHash returns the canonical SHA-256 digest of an obligation.
func (o Obligation) StableHash() string { return digestString(o.Canonical()) }
