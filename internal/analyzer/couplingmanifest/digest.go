package couplingmanifest

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// detectorManifestDigest constructs the detector's published digest over the
// exact detector ChangeManifest fields. It is a constructor helper only; all
// acceptance and validation is delegated to detector.Evaluate.
func detectorManifestDigest(manifest detector.ChangeManifest) string {
	entries := append([]detector.ManifestEntry(nil), manifest.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].SurfaceID < entries[j].SurfaceID })
	var builder strings.Builder
	field(&builder, detector.ManifestSchemaV1)
	field(&builder, strconv.FormatBool(manifest.Complete))
	field(&builder, strconv.FormatBool(manifest.ZeroChange))
	field(&builder, manifest.RegistryDigest)
	field(&builder, manifest.ToolchainDigest)
	field(&builder, manifest.ProfileDigest)
	field(&builder, manifest.BeforeSnapshotDigest)
	field(&builder, manifest.AfterSnapshotDigest)
	for _, entry := range entries {
		field(&builder, entry.SurfaceID.String())
		field(&builder, entry.CodeSymbolID.String())
		field(&builder, entry.SemanticOwnerID.String())
		field(&builder, entry.BeforeBindingDigest)
		field(&builder, entry.AfterBindingDigest)
		field(&builder, entry.BeforeBlobDigest)
		field(&builder, entry.AfterBlobDigest)
	}
	return stableDigest(builder.String())
}

func field(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('|')
}

func stableDigest(value string) string { return semantic.StableHashString(value) }

func canonicalID(value semantic.ID) (semantic.ID, error) {
	if value == "" {
		return "", fmt.Errorf("ID is empty")
	}
	parsed, err := semantic.ParseIdentity(value.String())
	if err != nil || parsed != value {
		return "", fmt.Errorf("ID %q is not canonical", value)
	}
	return parsed, nil
}

func canonicalIDString(value string) (semantic.ID, error) { return canonicalID(semantic.ID(value)) }

func rawDigest(value string) (string, error) {
	value = strings.TrimPrefix(value, "sha256:")
	if len(value) != 64 || value != strings.ToLower(value) {
		return "", fmt.Errorf("digest is not lowercase SHA-256")
	}
	for _, char := range value {
		if char < '0' || char > '9' && char < 'a' || char > 'f' {
			return "", fmt.Errorf("digest is not lowercase SHA-256")
		}
	}
	return value, nil
}
