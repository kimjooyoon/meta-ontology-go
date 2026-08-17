package eligibilityregistry

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"unicode"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type frameField struct {
	name  string
	tag   byte
	value []byte
}

type registryKey struct {
	id   StableID
	kind ItemKind
}

type registryRecord struct {
	entry RegistryEntry
	key   registryKey
}

func validStableID(id StableID) bool {
	raw := string(id)
	if !utf8.ValidString(raw) || raw == "" || len(raw) > 256 {
		return false
	}
	for _, value := range raw {
		if unicode.IsSpace(value) || unicode.IsControl(value) {
			return false
		}
	}
	parsed, err := semantic.ParseIdentity(raw)
	return err == nil && parsed.String() == raw
}

func validDigest(digest Digest) bool {
	raw := string(digest)
	if len(raw) != 71 || raw[:7] != "sha256:" {
		return false
	}
	for index := 7; index < len(raw); index++ {
		if (raw[index] < '0' || raw[index] > '9') && (raw[index] < 'a' || raw[index] > 'f') {
			return false
		}
	}
	return true
}

func validItemKind(kind ItemKind) bool {
	return kind == ItemSemantic || kind == ItemStructural
}

func validAuthorityKind(kind AuthorityKind) bool {
	return kind == AuthorityBusinessDSL || kind == AuthoritySemanticIR
}

func validProjectionKind(kind ProjectionKind) bool {
	return kind == ProjectionSemanticIR || kind == ProjectionGeneratedGo
}

func validCombination(itemKind ItemKind, authority AuthorityKind, projection ProjectionKind) bool {
	return itemKind == ItemSemantic && authority == AuthorityBusinessDSL && projection == ProjectionSemanticIR ||
		itemKind == ItemStructural && authority == AuthoritySemanticIR && projection == ProjectionGeneratedGo
}

func validateEntry(entry RegistryEntry) bool {
	return validStableID(entry.ID) && validItemKind(entry.Kind) && validAuthorityKind(entry.Authority) &&
		validProjectionKind(entry.RequiredProjection) &&
		validCombination(entry.Kind, entry.Authority, entry.RequiredProjection)
}

func keyForEntry(entry RegistryEntry) registryKey {
	return registryKey{id: entry.ID, kind: entry.Kind}
}

func normalizeEntries(entries []RegistryEntry) ([]registryRecord, Reason) {
	records := make([]registryRecord, 0, len(entries))
	for _, entry := range entries {
		if !validateEntry(entry) {
			return nil, ReasonInvalidItem
		}
		records = append(records, registryRecord{entry: entry, key: keyForEntry(entry)})
	}
	sort.Slice(records, func(left, right int) bool {
		a, b := records[left].entry, records[right].entry
		if a.ID != b.ID {
			return a.ID < b.ID
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Authority != b.Authority {
			return a.Authority < b.Authority
		}
		return a.RequiredProjection < b.RequiredProjection
	})
	duplicate, conflict := false, false
	for index := 1; index < len(records); index++ {
		previous, current := records[index-1], records[index]
		if previous.key != current.key {
			continue
		}
		if previous.entry == current.entry {
			duplicate = true
		} else {
			conflict = true
		}
	}
	if conflict {
		return nil, ReasonConflictingItem
	}
	if duplicate {
		return nil, ReasonDuplicateItem
	}
	return records, ReasonNone
}

func canonicalItemFrame(entry RegistryEntry) []byte {
	return encodeFrame("gooo/eligibility-registry/item/v1\x00", []frameField{
		{name: "id", tag: 0x02, value: []byte(entry.ID)},
		{name: "item_kind", tag: 0x05, value: itemKindSpelling(entry.Kind)},
		{name: "authority", tag: 0x05, value: authorityKindSpelling(entry.Authority)},
		{name: "required_projection", tag: 0x05, value: projectionKindSpelling(entry.RequiredProjection)},
	})
}

func canonicalRegistryFrame(sourceDigest Digest, records []registryRecord) []byte {
	items := make([][]byte, len(records))
	for index, record := range records {
		items[index] = canonicalItemFrame(record.entry)
		if items[index] == nil {
			return nil
		}
	}
	return encodeFrame("gooo/eligibility-registry/registry/v2\x00", []frameField{
		{name: "source_snapshot_digest", tag: 0x03, value: []byte(sourceDigest)},
		{name: "items", tag: 0x09, value: recordListValue(items)},
	})
}

func recordListValue(records [][]byte) []byte {
	value := appendU64BE(make([]byte, 0, 8), uint64(len(records)))
	for _, record := range records {
		if record == nil {
			return nil
		}
		value = appendU64BE(value, uint64(len(record)))
		value = append(value, record...)
	}
	return value
}

func encodeFrame(domain string, fields []frameField) []byte {
	frame := appendU64BE(nil, uint64(len(domain)))
	frame = append(frame, domain...)
	frame = appendU64BE(frame, uint64(len(fields)))
	for _, field := range fields {
		encoded := encodeField(field)
		if encoded == nil {
			return nil
		}
		frame = append(frame, encoded...)
	}
	return frame
}

func encodeField(field frameField) []byte {
	if field.value == nil {
		return nil
	}
	encoded := appendU64BE(nil, uint64(len(field.name)))
	encoded = append(encoded, field.name...)
	encoded = append(encoded, field.tag)
	encoded = appendU64BE(encoded, uint64(len(field.value)))
	return append(encoded, field.value...)
}

func appendU64BE(dst []byte, value uint64) []byte {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	return append(dst, encoded[:]...)
}

func itemKindSpelling(kind ItemKind) []byte {
	switch kind {
	case ItemSemantic:
		return []byte("SEMANTIC")
	case ItemStructural:
		return []byte("STRUCTURAL")
	default:
		return nil
	}
}

func authorityKindSpelling(kind AuthorityKind) []byte {
	switch kind {
	case AuthorityBusinessDSL:
		return []byte("BUSINESS_DSL")
	case AuthoritySemanticIR:
		return []byte("SEMANTIC_IR")
	default:
		return nil
	}
}

func projectionKindSpelling(kind ProjectionKind) []byte {
	switch kind {
	case ProjectionSemanticIR:
		return []byte("SEMANTIC_IR")
	case ProjectionGeneratedGo:
		return []byte("GENERATED_GO")
	default:
		return nil
	}
}

func digestBytes(payload []byte) Digest {
	sum := sha256.Sum256(payload)
	encoded := make([]byte, 71)
	copy(encoded, "sha256:")
	const digits = "0123456789abcdef"
	for index, value := range sum {
		encoded[7+index*2] = digits[value>>4]
		encoded[8+index*2] = digits[value&0x0F]
	}
	return Digest(encoded)
}

func CanonicalRegistry(sourceDigest Digest, entries []RegistryEntry) ([]byte, bool) {
	if !validDigest(sourceDigest) {
		return nil, false
	}
	records, reason := normalizeEntries(entries)
	if reason != ReasonNone {
		return nil, false
	}
	frame := canonicalRegistryFrame(sourceDigest, records)
	return frame, frame != nil
}

func RegistryDigest(sourceDigest Digest, entries []RegistryEntry) (Digest, bool) {
	frame, ok := CanonicalRegistry(sourceDigest, entries)
	if !ok {
		return "", false
	}
	return digestBytes(frame), true
}
