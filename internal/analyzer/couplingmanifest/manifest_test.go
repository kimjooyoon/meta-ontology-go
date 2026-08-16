package couplingmanifest

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	detectorcoupling "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
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
			duplicate.SurfaceID, duplicate.SemanticOwnerID, duplicate.Binding.SourceMapID = fixtureID("surface-other"), fixtureID("owner-other"), fixtureID("map-other")
			duplicate.Binding.BindingDigest = sourceMapBindingDigest(duplicate)
			a.Inventory = append(a.Inventory, duplicate)
		}, code: CodeConflictingBinding, status: StatusFailClosed},
		{name: "duplicate owner", edit: func(a *RegistrySourceMap) {
			duplicate := a.Inventory[0]
			duplicate.SurfaceID, duplicate.CodeSymbolID, duplicate.Binding.SourceMapID = fixtureID("surface-other"), fixtureID("symbol-other"), fixtureID("map-other")
			duplicate.Binding.BindingDigest = sourceMapBindingDigest(duplicate)
			a.Inventory = append(a.Inventory, duplicate)
		}, code: CodeConflictingBinding, status: StatusFailClosed},
		{name: "duplicate source map", edit: func(a *RegistrySourceMap) {
			duplicate := a.Inventory[0]
			duplicate.SurfaceID, duplicate.CodeSymbolID, duplicate.SemanticOwnerID = fixtureID("surface-other"), fixtureID("symbol-other"), fixtureID("owner-other")
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
