package couplingmanifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	detectorcoupling "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestBuildMatchesDetectorManifestContractAndZeroChange(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	before := testSnapshot(t, source)
	head := testSnapshot(t, source)
	authority := testAuthority(t, before, head, []surfaceFixture{{owner: source.Bindings[0].ID, suffix: "order"}})
	manifest, err := Build(Input{Before: &before, Head: &head, Authority: authority})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if manifest.Schema != SchemaV1 || !manifest.Complete || !manifest.ZeroChange || manifest.Status != StatusComplete {
		t.Fatalf("manifest state = %#v", manifest)
	}
	if manifest.BeforeSnapshotDigest != rawTestDigest(before.Digest) || manifest.AfterSnapshotDigest != rawTestDigest(head.Digest) {
		t.Fatalf("snapshot digests = %#v", manifest)
	}
	if manifest.HeadSnapshotDigest != manifest.AfterSnapshotDigest || manifest.SourceMapDigest != rawTestDigest(authority.SourceMapDigest) {
		t.Fatalf("authority aliases = %#v", manifest)
	}
	if len(manifest.Entries) != 1 || !manifest.Entries[0].BeforePresent || !manifest.Entries[0].AfterPresent {
		t.Fatalf("entries = %#v", manifest.Entries)
	}
	data, err := manifest.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON: %v", err)
	}
	for _, forbidden := range []string{"\"status\"", "\"source_map_digest\"", "\"counts\"", "\"work\"", "\"resolved_surface_ids\""} {
		if bytes.Contains(data, []byte(forbidden)) {
			t.Fatalf("detector wire leaked adapter field %s: %s", forbidden, data)
		}
	}
	for _, required := range []string{"\"complete\"", "\"zero_change\"", "\"after_snapshot_digest\"", "\"before_source_path\"", "\"after_source_path\""} {
		if !bytes.Contains(data, []byte(required)) {
			t.Fatalf("detector wire omitted field %s: %s", required, data)
		}
	}
	decoded, err := DecodeJSON(data)
	if err != nil {
		t.Fatalf("DecodeJSON: %v", err)
	}
	if decoded.Digest != manifest.Digest || !bytes.Equal(data, mustCanonical(t, decoded)) {
		t.Fatalf("round-trip changed manifest: %#v", decoded)
	}
	var detectorManifest detectorcoupling.ChangeManifest
	if err := json.Unmarshal(data, &detectorManifest); err != nil {
		t.Fatalf("detector ChangeManifest decode: %v", err)
	}
	detectorData, err := json.Marshal(detectorManifest)
	if err != nil {
		t.Fatalf("detector ChangeManifest encode: %v", err)
	}
	if !bytes.Equal(data, detectorData) {
		t.Fatalf("adapter bytes differ from detector contract:\nadapter: %s\ndetector: %s", data, detectorData)
	}
}

func TestBuildRepresentsAdditionsAndRemovalsInFullInventory(t *testing.T) {
	a := testSource(t, "pkg/a.go", "A", "urn:gooo:entity:a")
	b := testSource(t, "pkg/b.go", "B", "urn:gooo:entity:b")
	before := testSnapshot(t, a)
	head := testSnapshot(t, a, b)
	authority := testAuthority(t, before, head, []surfaceFixture{{owner: a.Bindings[0].ID, suffix: "a"}, {owner: b.Bindings[0].ID, suffix: "b"}})
	added, err := Build(Input{Before: &before, Head: &head, Authority: authority})
	if err != nil {
		t.Fatalf("Build addition: %v", err)
	}
	if len(added.Entries) != 2 || added.Entries[1].BeforePresent || !added.Entries[1].AfterPresent || added.ZeroChange {
		t.Fatalf("addition entries = %#v", added.Entries)
	}
	if added.Entries[1].BeforeBindingDigest != absentDigest || added.Entries[1].BeforeBlobDigest != absentDigest || added.Entries[1].BeforeSourcePath != absentPath {
		t.Fatalf("addition absence encoding = %#v", added.Entries[1])
	}

	removedBefore, removedHead := head, before
	removedAuthority := swapAuthoritySides(authority)
	removed, err := Build(Input{Before: &removedBefore, Head: &removedHead, Authority: removedAuthority})
	if err != nil {
		t.Fatalf("Build removal: %v", err)
	}
	if removed.Entries[1].BeforePresent != true || removed.Entries[1].AfterPresent || removed.Entries[1].AfterBlobDigest != absentDigest || removed.Entries[1].AfterSourcePath != absentPath {
		t.Fatalf("removal entry = %#v", removed.Entries[1])
	}
}

func TestPathsAndPresentationDoNotChangeCanonicalDigest(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	before := testSnapshot(t, source)
	head := testSnapshot(t, source)
	authority := testAuthority(t, before, head, []surfaceFixture{{owner: source.Bindings[0].ID, suffix: "order"}})
	first, err := Build(Input{Before: &before, Head: &head, Authority: authority})
	if err != nil {
		t.Fatalf("Build first: %v", err)
	}
	second := first
	second.Entries = append([]ManifestEntry(nil), first.Entries...)
	second.Entries[0].BeforeSourcePath = "relocated/old.go"
	second.Entries[0].AfterSourcePath = "relocated/new.go"
	if second.Digest != first.Digest {
		t.Fatalf("path mutation changed digest before encoding: %s/%s", first.Digest, second.Digest)
	}
	if _, err := second.CanonicalJSON(); err != nil {
		t.Fatalf("path-only manifest rejected: %v", err)
	}
	if first.Canonical() == second.Canonical() {
		t.Fatal("path-only presentation mutation was discarded from JSON")
	}
	labels := authority
	labels.Inventory = append([]Surface(nil), authority.Inventory...)
	labels.Inventory[0].PresentationLabel = "Renamed"
	labels.Inventory[0].Binding.PackageLabel = "relocated.package"
	labels.Inventory[0].Binding.FileLabel = "relocated.go"
	third, err := Build(Input{Before: &before, Head: &head, Authority: labels})
	if err != nil || third.Digest != first.Digest {
		t.Fatalf("presentation mutation changed identity: %#v, %v", third, err)
	}
}

func TestBuildUnknownsStaleSnapshotsAuthorityDriftAndUnknownPath(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	before := testSnapshot(t, source)
	head := testSnapshot(t, source)
	authority := testAuthority(t, before, head, []surfaceFixture{{owner: source.Bindings[0].ID, suffix: "order"}})
	cases := []struct {
		name string
		edit func(*selectiveci.Snapshot, *selectiveci.Snapshot, *RegistrySourceMap)
		code ErrorCode
	}{
		{name: "stale before digest", edit: func(b, _ *selectiveci.Snapshot, _ *RegistrySourceMap) { b.Digest = testDigest("stale-before") }, code: CodeInvalidSnapshot},
		{name: "stale head digest", edit: func(_, h *selectiveci.Snapshot, _ *RegistrySourceMap) { h.Digest = testDigest("stale-head") }, code: CodeInvalidSnapshot},
		{name: "registry drift", edit: func(_, _ *selectiveci.Snapshot, a *RegistrySourceMap) {
			a.RegistryDigest = testDigest("registry-drift")
		}, code: CodeAuthorityDrift},
		{name: "source-map drift", edit: func(_, _ *selectiveci.Snapshot, a *RegistrySourceMap) {
			a.SourceMapDigest = testDigest("source-map-drift")
		}, code: CodeAuthorityDrift},
		{name: "unknown changed path", edit: func(_, _ *selectiveci.Snapshot, a *RegistrySourceMap) {
			a.Head = append([]SourceMapObservation(nil), a.Head...)
			a.Head[0].Path = "pkg/unknown.go"
		}, code: CodeUnknownChangedSurface},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, h, a := before, head, authority
			tc.edit(&b, &h, &a)
			manifest, err := Build(Input{Before: &b, Head: &h, Authority: a})
			assertStatus(t, manifest, err, StatusUnknown, tc.code)
		})
	}
}

func TestBuildRejectsConflictsAndNonAuthoritativeBindings(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	before := testSnapshot(t, source)
	head := testSnapshot(t, source)
	base := testAuthority(t, before, head, []surfaceFixture{{owner: source.Bindings[0].ID, suffix: "order"}})
	cases := []struct {
		name   string
		edit   func(*RegistrySourceMap)
		code   ErrorCode
		status Status
	}{
		{name: "duplicate surface", edit: func(a *RegistrySourceMap) { a.Inventory = append(a.Inventory, a.Inventory[0]) }, code: CodeDuplicateBinding, status: StatusFailClosed},
		{name: "duplicate code symbol", edit: func(a *RegistrySourceMap) {
			duplicate := a.Inventory[0]
			duplicate.SurfaceID = fixtureID("surface-other")
			duplicate.SemanticOwnerID = fixtureID("owner-other")
			duplicate.Binding.SourceMapID = fixtureID("map-other")
			duplicate.Binding.BindingDigest = sourceMapBindingDigest(duplicate)
			a.Inventory = append(a.Inventory, duplicate)
		}, code: CodeConflictingBinding, status: StatusFailClosed},
		{name: "duplicate owner", edit: func(a *RegistrySourceMap) {
			duplicate := a.Inventory[0]
			duplicate.SurfaceID = fixtureID("surface-other")
			duplicate.CodeSymbolID = fixtureID("symbol-other")
			duplicate.Binding.SourceMapID = fixtureID("map-other")
			duplicate.Binding.BindingDigest = sourceMapBindingDigest(duplicate)
			a.Inventory = append(a.Inventory, duplicate)
		}, code: CodeConflictingBinding, status: StatusFailClosed},
		{name: "duplicate source map", edit: func(a *RegistrySourceMap) {
			duplicate := a.Inventory[0]
			duplicate.SurfaceID = fixtureID("surface-other")
			duplicate.CodeSymbolID = fixtureID("symbol-other")
			duplicate.SemanticOwnerID = fixtureID("owner-other")
			duplicate.Binding.BindingDigest = sourceMapBindingDigest(duplicate)
			a.Inventory = append(a.Inventory, duplicate)
		}, code: CodeConflictingBinding, status: StatusFailClosed},
		{name: "extra observation", edit: func(a *RegistrySourceMap) {
			extra := a.Head[0]
			extra.SurfaceID = fixtureID("surface-extra")
			a.Head = append(a.Head, extra)
		}, code: CodeMalformedBinding, status: StatusFailClosed},
		{name: "candidate", edit: func(a *RegistrySourceMap) { a.CandidateBindings = append([]SourceMapObservation{}, a.Head...) }, code: CodeCandidateBinding, status: StatusUnknown},
		{name: "derived", edit: func(a *RegistrySourceMap) { a.DerivedBindings = append([]SourceMapObservation{}, a.Head...) }, code: CodeDerivedBinding, status: StatusUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := base
			tc.edit(&input)
			manifest, err := Build(Input{Before: &before, Head: &head, Authority: input})
			assertStatus(t, manifest, err, tc.status, tc.code)
		})
	}
}

func TestBuildDistinguishesMissingObservationFromExplicitEmptyInventory(t *testing.T) {
	before := testSnapshot(t)
	head := testSnapshot(t)
	authority := RegistrySourceMap{
		Schema: AuthoritySchemaV1, RegistryDigest: before.RegistryDigest, SourceMapDigest: before.SourceMapDigest,
		ToolchainDigest: testDigest("toolchain"), ProfileDigest: testDigest("profile"), Inventory: []Surface{},
		Before: []SourceMapObservation{}, Head: []SourceMapObservation{},
	}
	complete, err := Build(Input{Before: &before, Head: &head, Authority: authority})
	if err != nil || complete.Status != StatusComplete || !complete.Complete || !complete.ZeroChange || len(complete.Entries) != 0 {
		t.Fatalf("complete empty inventory = %#v, err=%v", complete, err)
	}
	missing, err := Build(Input{Authority: authority})
	assertStatus(t, missing, err, StatusUnknown, CodeMissingSnapshot)
}

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

func rawTestDigest(value string) string {
	return strings.TrimPrefix(value, "sha256:")
}

func mustCanonical(t *testing.T, manifest Manifest) []byte {
	t.Helper()
	data, err := manifest.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON: %v", err)
	}
	return data
}
