package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func projectionSpec() ProjectionKeySpec {
	return ProjectionKeySpec{
		Domain: "billing", Version: ProjectionKeyVersion, ArtifactKind: "go-source",
		Projection: "source", HostStage: GoHostedStage,
		SemanticClosureDigest: HashBytes([]byte("semantic-closure")),
		DependencyRoot:        HashBytes([]byte("dependency-root")),
		PolicySchemaDigest:    HashBytes([]byte("policy-schema")),
		Toolchain:             "go1.26.5", Target: "darwin/arm64", BuildTags: []string{"linux"},
		OptionsDigest: mustOptionsDigest(map[string]any{"mode": "fast", "trim": true}),
	}
}

func mustOptionsDigest(value any) Digest {
	digest, err := DigestOptions(value)
	if err != nil {
		panic(err)
	}
	return digest
}

func TestProjectionKeyC1MutationMatrix(t *testing.T) {
	baseSpec := projectionSpec()
	base, err := NewProjectionKey(baseSpec)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(*ProjectionKeySpec)
	}{
		{"domain", func(s *ProjectionKeySpec) { s.Domain = "payments" }},
		{"version", func(s *ProjectionKeySpec) { s.Version = "v3" }},
		{"artifact kind", func(s *ProjectionKeySpec) { s.ArtifactKind = "docs" }},
		{"projection", func(s *ProjectionKeySpec) { s.Projection = "ir" }},
		{"host stage", func(s *ProjectionKeySpec) { s.HostStage = GoooHostedStage }},
		{"semantic closure", func(s *ProjectionKeySpec) { s.SemanticClosureDigest = HashBytes([]byte("changed")) }},
		{"dependency root", func(s *ProjectionKeySpec) { s.DependencyRoot = HashBytes([]byte("changed")) }},
		{"policy schema", func(s *ProjectionKeySpec) { s.PolicySchemaDigest = HashBytes([]byte("changed")) }},
		{"toolchain", func(s *ProjectionKeySpec) { s.Toolchain = "go1.26.6" }},
		{"target", func(s *ProjectionKeySpec) { s.Target = "linux/amd64" }},
		{"build tags", func(s *ProjectionKeySpec) { s.BuildTags = []string{"linux", "race"} }},
		{"options", func(s *ProjectionKeySpec) { s.OptionsDigest = mustOptionsDigest(map[string]any{"mode": "safe"}) }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			mutated := baseSpec
			test.mutate(&mutated)
			key, err := NewProjectionKey(mutated)
			if err != nil {
				t.Fatal(err)
			}
			if key == base {
				t.Fatalf("mutation %q retained key %s", test.name, key)
			}
		})
	}
}

func TestProjectionKeyC2PresentationStableAndCorruptionMisses(t *testing.T) {
	first := projectionSpec()
	second := projectionSpec()
	first.BuildTags = []string{"linux", "windows"}
	second.BuildTags = []string{"windows", "linux", "linux"}
	second.OptionsDigest = mustOptionsDigest(map[string]any{"trim": true, "mode": "fast"})
	firstKey, err := NewProjectionKey(first)
	if err != nil {
		t.Fatal(err)
	}
	secondKey, err := NewProjectionKey(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstKey != secondKey {
		t.Fatal("presentation-only changes altered projection key")
	}
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(firstKey, []byte("small")); err != nil {
		t.Fatal(err)
	}
	if err := cache.Put(secondKey, []byte("larger-content")); err != nil {
		t.Fatal(err)
	}
	data, metadata, err := cache.Get(firstKey)
	if err != nil || string(data) != "small" || metadata.Size != int64(len(data)) {
		t.Fatalf("immutable false-hit result = %q, metadata=%+v, err=%v", data, metadata, err)
	}
	object, err := cache.objectPath(firstKey)
	if err != nil {
		t.Fatal(err)
	}
	metadata.Size++
	raw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(object, metaFileName), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := cache.Get(firstKey); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("content-size mismatch = %v, want ErrCorrupt", err)
	}
}

func TestProjectionKeyC3UnknownIdentityFailsClosed(t *testing.T) {
	spec := projectionSpec()
	spec.DependencyRoot = ""
	if _, err := NewProjectionKey(spec); !errors.Is(err, ErrUnknownFreshness) {
		t.Fatalf("missing dependency root = %v, want ErrUnknownFreshness", err)
	}
	var unknown map[string]string
	if _, err := NewFreshness(FreshnessSpec{
		Dependencies: unknown, Provenance: map[string]any{"known": true},
	}); !errors.Is(err, ErrUnknownFreshness) {
		t.Fatalf("nil dependency input = %v, want ErrUnknownFreshness", err)
	}
	if _, err := NewFreshness(FreshnessSpec{
		Dependencies: map[string]any{"known": true}, Provenance: nil,
	}); !errors.Is(err, ErrUnknownFreshness) {
		t.Fatalf("nil provenance input = %v, want ErrUnknownFreshness", err)
	}
	if _, err := NewProjectionKey(projectionSpec()); err != nil {
		t.Fatal(err)
	}
}

func TestProjectionKeyC3OpaqueOptionsFailClosed(t *testing.T) {
	spec := projectionSpec()
	spec.Options = map[string]any{"mode": "unsafe"}
	if _, err := NewProjectionKey(spec); !errors.Is(err, ErrUnknownFreshness) {
		t.Fatalf("opaque options = %v, want ErrUnknownFreshness", err)
	}
	if _, err := DigestOptions(nil); !errors.Is(err, ErrUnknownFreshness) {
		t.Fatalf("nil options = %v, want ErrUnknownFreshness", err)
	}
	first, err := DigestOptions(map[string]any{"mode": "fast", "trim": true})
	if err != nil {
		t.Fatal(err)
	}
	second, err := DigestOptions(map[string]any{"trim": true, "mode": "fast"})
	if err != nil || first != second {
		t.Fatalf("options presentation changed digest: %s != %s (%v)", first, second, err)
	}
}

func TestProjectionKeyC3EntryInfoCannotAliasIdentity(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.PutWithInfo(key, []byte("projection"), EntryInfo{Projection: "other"}); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("metadata-only projection alias = %v, want ErrInvalidKey", err)
	}
	variant := projectionSpec()
	variant.ArtifactKind = "docs"
	variant.Projection = "ir"
	variantKey, err := NewProjectionKey(variant)
	if err != nil {
		t.Fatal(err)
	}
	if variantKey == key {
		t.Fatal("artifact/projection mutation retained identity")
	}
}
