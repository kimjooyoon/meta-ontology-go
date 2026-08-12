package analyzer

import (
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const (
	policyDigestSchema    = "analyzer-semantic-policy/v1"
	toolchainDigestSchema = "analyzer-toolchain/v1"
	bindingDigestSchema   = "analyzer-semantic-binding/v1"
)

// Canonical returns the explicit analyzer-to-PROV mapping policy.
func (p MappingPolicy) Canonical() string {
	keys := make([]Relation, 0, len(p.mappings))
	for relation := range p.mappings {
		keys = append(keys, relation)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	var b strings.Builder
	b.WriteString(policyDigestSchema)
	b.WriteByte('\n')
	writeBindingField(&b, p.Revision)
	for _, key := range keys {
		mapping := p.mappings[key]
		writeBindingField(&b, string(mapping.Source))
		writeBindingField(&b, mapping.Predicate.String())
		writeBindingField(&b, mapping.SourceSubjectKind.String())
		writeBindingField(&b, mapping.SourceObjectKind.String())
		writeBindingField(&b, strconv.FormatBool(mapping.Reverse))
		origins := append([]ObservationOrigin(nil), mapping.AllowedOrigins...)
		sort.Slice(origins, func(i, j int) bool { return origins[i] < origins[j] })
		for _, origin := range origins {
			writeBindingField(&b, string(origin))
		}
	}
	return b.String()
}

func (p MappingPolicy) Digest() string { return semantic.StableHashString(p.Canonical()) }

// ToolchainDigest binds an adapter result to the caller's exact toolchain ID.
func ToolchainDigest(identity string) string {
	return semantic.StableHashString(toolchainDigestSchema + "\n" + strings.TrimSpace(identity))
}

func semanticAdapterBindingDigest(result SemanticAdapterResult) string {
	if !validDigest(result.SourceDigest) || !validDigest(result.PolicyDigest) ||
		!validDigest(result.ToolchainDigest) {
		return ""
	}
	var b strings.Builder
	b.WriteString(bindingDigestSchema)
	b.WriteByte('\n')
	writeBindingField(&b, result.SourceDigest)
	writeBindingField(&b, result.IR.StableHash())
	writeBindingField(&b, result.IR.EvidenceHash())
	writeBindingField(&b, result.IR.ProvenanceHash())
	writeBindingField(&b, result.PolicyDigest)
	writeBindingField(&b, result.ToolchainDigest)
	return semantic.StableHashString(b.String())
}

func writeBindingField(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('\n')
}

func validDigest(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}
