package eligibilityregistry

import (
	"sort"
	"unicode"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
	semanticItem := itemKind == ItemSemantic &&
		authority == AuthorityBusinessDSL &&
		projection == ProjectionSemanticIR
	structuralItem := itemKind == ItemStructural &&
		authority == AuthoritySemanticIR &&
		projection == ProjectionGeneratedGo
	return semanticItem || structuralItem
}

func validateEntry(entry RegistryEntry) bool {
	return validStableID(entry.ID) &&
		validItemKind(entry.Kind) &&
		validAuthorityKind(entry.Authority) &&
		validProjectionKind(entry.RequiredProjection) &&
		validCombination(entry.Kind, entry.Authority, entry.RequiredProjection)
}

func keyForEntry(entry RegistryEntry) registryKey {
	return registryKey{id: entry.ID, kind: entry.Kind}
}

func normalizeEntries(entries []RegistryEntry) ([]registryRecord, Reason) {
	records := make([]registryRecord, len(entries))
	for index, entry := range entries {
		if !validateEntry(entry) {
			return nil, ReasonInvalidItem
		}
		records[index] = registryRecord{entry: entry, key: keyForEntry(entry)}
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
	duplicate := false
	// Full-entry validation makes same-key altered entries invalid.
	// The explicit check below keeps that precedence local to normalization.
	// ReasonConflictingItem is reserved for a future registry version
	// that permits more than one valid shape for a key.
	for index := 1; index < len(records); index++ {
		previous, current := records[index-1], records[index]
		if previous.key != current.key {
			continue
		}
		if previous.entry != current.entry {
			return nil, ReasonInvalidItem
		}
		duplicate = true
	}
	if duplicate {
		return nil, ReasonDuplicateItem
	}
	return records, ReasonNone
}

func canonicalItemFrame(entry RegistryEntry) []byte {
	if !validateEntry(entry) {
		return nil
	}
	return encodeFrame("gooo/eligibility-registry/item/v1\x00", []frameField{
		{name: "id", tag: frameTagStableID, value: []byte(entry.ID)},
		{name: "item_kind", tag: frameTagEnum, value: itemKindSpelling(entry.Kind)},
		{name: "authority", tag: frameTagEnum, value: authorityKindSpelling(entry.Authority)},
		{
			name:  "required_projection",
			tag:   frameTagEnum,
			value: projectionKindSpelling(entry.RequiredProjection),
		},
	})
}

func canonicalRegistryFrame(sourceDigest Digest, records []registryRecord) []byte {
	if !validDigest(sourceDigest) {
		return nil
	}
	items := make([][]byte, len(records))
	for index, record := range records {
		items[index] = canonicalItemFrame(record.entry)
		if items[index] == nil {
			return nil
		}
	}
	return encodeFrame("gooo/eligibility-registry/registry/v2\x00", []frameField{
		{name: "source_snapshot_digest", tag: frameTagDigest, value: []byte(sourceDigest)},
		{name: "items", tag: frameTagRecordList, value: recordListValue(items)},
	})
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
