package couplingmanifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestCanonicalJSONRejectsMutationsAndTrustedMetadata(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	snapshot := testSnapshot(t, source)
	authority := testAuthority(t, snapshot, snapshot, []surfaceFixture{{owner: source.Bindings[0].ID, suffix: "order"}})
	manifest, err := Build(Input{Before: &snapshot, Head: &snapshot, Authority: authority})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	canonical := mustCanonical(t, manifest)
	mutations := map[string][]byte{
		"duplicate key":    bytes.Replace(canonical, []byte(`"complete":true`), []byte(`"complete":true,"complete":true`), 1),
		"unknown field":    bytes.Replace(canonical, []byte(`"digest"`), []byte(`"extra":true,"digest"`), 1),
		"trailing value":   append(append([]byte{}, canonical...), []byte(`{}`)...),
		"wrong schema":     bytes.Replace(canonical, []byte(SchemaV1), []byte("wrong/v9"), 1),
		"wrong zero claim": bytes.Replace(canonical, []byte(`"zero_change":true`), []byte(`"zero_change":false`), 1),
	}
	for name, data := range mutations {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeJSON(data); err == nil {
				t.Fatal("accepted mutated manifest")
			}
		})
	}
	mutated := manifest
	mutated.Counts.Resolved = 0
	if _, err := mutated.CanonicalJSON(); err == nil {
		t.Fatal("CanonicalJSON trusted mutated counts")
	}
}

func TestCanonicalIdentityIsPermutationIndependent(t *testing.T) {
	a := testSource(t, "pkg/a.go", "A", "urn:gooo:entity:a")
	b := testSource(t, "pkg/b.go", "B", "urn:gooo:entity:b")
	firstBefore := testSnapshot(t, a, b)
	firstHead := testSnapshot(t, a, b)
	firstAuthority := testAuthority(t, firstBefore, firstHead, []surfaceFixture{{owner: a.Bindings[0].ID, suffix: "a"}, {owner: b.Bindings[0].ID, suffix: "b"}})
	first, err := Build(Input{Before: &firstBefore, Head: &firstHead, Authority: firstAuthority})
	if err != nil {
		t.Fatalf("Build first: %v", err)
	}
	secondBefore := testSnapshot(t, b, a)
	secondHead := testSnapshot(t, b, a)
	secondAuthority := testAuthority(t, secondBefore, secondHead, []surfaceFixture{{owner: b.Bindings[0].ID, suffix: "b"}, {owner: a.Bindings[0].ID, suffix: "a"}})
	second, err := Build(Input{Before: &secondBefore, Head: &secondHead, Authority: secondAuthority})
	if err != nil {
		t.Fatalf("Build second: %v", err)
	}
	if first.Digest != second.Digest || !bytes.Equal(mustCanonical(t, first), mustCanonical(t, second)) {
		t.Fatalf("permutation changed identity: %s/%s", first.Digest, second.Digest)
	}
}

func TestBuildIsPureAndDoesNotWrite(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	before := testSnapshot(t, source)
	head := testSnapshot(t, source)
	authority := testAuthority(t, before, head, []surfaceFixture{{owner: source.Bindings[0].ID, suffix: "order"}})
	originalBefore, originalHead, originalAuthority := before, head, authority
	if _, err := Build(Input{Before: &before, Head: &head, Authority: authority}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !reflect.DeepEqual(before, originalBefore) || !reflect.DeepEqual(head, originalHead) || !reflect.DeepEqual(authority, originalAuthority) {
		t.Fatal("Build mutated an input authority or snapshot")
	}
}

func assertStatus(t *testing.T, manifest Manifest, err error, status Status, code ErrorCode) {
	t.Helper()
	if manifest.Status != status || !manifest.FullSuiteRequired || manifest.Complete || len(manifest.Entries) != 0 || len(manifest.ResolvedSurfaceIDs) != 0 {
		t.Fatalf("manifest = %#v", manifest)
	}
	var typed *Error
	if !errors.As(err, &typed) || typed.Code != code || typed.Status != status || !typed.FullSuiteRequired {
		t.Fatalf("error = %v, want %s/%s", err, status, code)
	}
}

type surfaceFixture struct {
	owner  string
	suffix string
}

func fixtureID(name string) semantic.ID { return semantic.MustIdentity("urn:gooo:coupling:" + name) }

func testAuthority(t *testing.T, before, head selectiveci.Snapshot, surfaces []surfaceFixture) RegistrySourceMap {
	t.Helper()
	authority := RegistrySourceMap{
		Schema: AuthoritySchemaV1, RegistryDigest: before.RegistryDigest, SourceMapDigest: before.SourceMapDigest,
		ToolchainDigest: testDigest("toolchain"), ProfileDigest: testDigest("profile"), Inventory: make([]Surface, 0, len(surfaces)),
		Before: make([]SourceMapObservation, 0), Head: make([]SourceMapObservation, 0),
	}
	for _, fixture := range surfaces {
		owner := semantic.MustIdentity(fixture.owner)
		surface := Surface{SurfaceID: fixtureID("surface-" + fixture.suffix), CodeSymbolID: fixtureID("symbol-" + fixture.suffix), SemanticOwnerID: owner, Binding: SourceMapBinding{SourceMapID: fixtureID("map-" + fixture.suffix), PackageLabel: "fixture", FileLabel: fixture.suffix + ".go"}, PresentationLabel: "Surface-" + fixture.suffix}
		surface.Binding.BindingDigest = sourceMapBindingDigest(surface)
		authority.Inventory = append(authority.Inventory, surface)
	}
	authority.Before = authorityObservations(t, before, authority.Inventory)
	authority.Head = authorityObservations(t, head, authority.Inventory)
	return authority
}

func authorityObservations(t *testing.T, snapshot selectiveci.Snapshot, inventory []Surface) []SourceMapObservation {
	t.Helper()
	byOwner := make(map[semantic.ID]Surface, len(inventory))
	for _, surface := range inventory {
		byOwner[surface.SemanticOwnerID] = surface
	}
	result := make([]SourceMapObservation, 0)
	for _, source := range snapshot.Sources {
		for _, binding := range source.Bindings {
			owner := semantic.MustIdentity(binding.ID)
			surface, ok := byOwner[owner]
			if !ok {
				continue
			}
			result = append(result, SourceMapObservation{SurfaceID: surface.SurfaceID, CodeSymbolID: surface.CodeSymbolID, SemanticOwnerID: surface.SemanticOwnerID, SourceMapID: surface.Binding.SourceMapID, Role: binding.Role, Path: source.Path, BlobDigest: source.BlobDigest, BindingDigest: binding.BindingDigest, SourceMapBindingDigest: surface.Binding.BindingDigest})
		}
	}
	return result
}

func swapAuthoritySides(authority RegistrySourceMap) RegistrySourceMap {
	authority.Before, authority.Head = authority.Head, authority.Before
	return authority
}

func testSnapshot(t *testing.T, sources ...selectiveci.SourceInput) selectiveci.Snapshot {
	t.Helper()
	ids := make([]string, 0)
	for _, source := range sources {
		for _, binding := range source.Bindings {
			ids = append(ids, binding.ID)
		}
	}
	result, err := selectiveci.Build(selectiveci.SnapshotInput{Sources: append([]selectiveci.SourceInput{}, sources...), SourceMapDigest: testDigest("source-map"), RegistryDigest: testDigest("registry"), RegisteredIDs: ids})
	if err != nil {
		t.Fatalf("selectiveci.Build: %v", err)
	}
	return result
}

func testSource(t *testing.T, filename, name, id string) selectiveci.SourceInput {
	t.Helper()
	source := []byte(fmt.Sprintf("package fixture\n\n//gooo:bind id=%q role=\"HANDWRITTEN_IMPL\"\nfunc %s() {}\n", id, name))
	result, err := semanticbinding.Extract(semanticbinding.Input{Sources: []semanticbinding.SourceFile{{Filename: filename, PackagePath: "fixture", Source: source}}})
	if err != nil || result.Status != semanticbinding.StatusBound || len(result.Bindings) != 1 {
		t.Fatalf("semanticbinding.Extract = %#v, err=%v", result, err)
	}
	return selectiveci.SourceInput{Path: filename, BlobDigest: testDigest(string(source)), Bindings: result.Bindings}
}

func testDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func rawTestDigest(value string) string { return strings.TrimPrefix(value, "sha256:") }

func mustCanonical(t *testing.T, manifest Manifest) []byte {
	t.Helper()
	data, err := manifest.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON: %v", err)
	}
	return data
}
