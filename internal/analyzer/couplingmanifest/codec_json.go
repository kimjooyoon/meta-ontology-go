package couplingmanifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func field(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('|')
}

func manifestCanonical(manifest Manifest) string {
	entries := append([]ManifestEntry(nil), manifest.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].SurfaceID < entries[j].SurfaceID })
	var builder strings.Builder
	field(&builder, SchemaV1)
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
	return builder.String()
}

func stableDigest(value string) string { return semantic.StableHashString(value) }

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		return scanJSONObject(decoder)
	case '[':
		return scanJSONArray(decoder)
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
}

func scanJSONObject(decoder *json.Decoder) error {
	seen := map[string]struct{}{}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return err
		}
		name, ok := key.(string)
		if !ok {
			return fmt.Errorf("object key is not a string")
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate object field %q", name)
		}
		seen[name] = struct{}{}
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
	return err
}

func scanJSONArray(decoder *json.Decoder) error {
	for decoder.More() {
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
	return err
}
