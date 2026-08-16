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
)

func TestBuildEmitsCompleteZeroChangeAndStrictRoundTrip(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	before := testSnapshot(t, source)
	head := testSnapshot(t, source)
	authority := testAuthority(t, before, head, []surfaceFixture{{owner: source.Bindings[0].ID, suffix: "order"}})
	manifest, err := Build(Input{Before: &before, Head: &head, Authority: authority})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if manifest.Status != StatusComplete || !manifest.ObservationComplete || manifest.FullSuiteRequired || len(manifest.Entries) != 1 || len(manifest.ResolvedSurfaceIDs) != 1 {
		t.Fatalf("manifest = %#v", manifest)
	}
	if manifest.Counts != (ComponentCounts{Registered: 1, Before: 1, Head: 1, Resolved: 1}) || manifest.Work.WorkUnits != 1 {
		t.Fatalf("counts/work = %#v/%#v", manifest.Counts, manifest.Work)
	}
	if manifest.BeforeSnapshotDigest != before.Digest || manifest.HeadSnapshotDigest != head.Digest || manifest.ToolchainDigest != authority.ToolchainDigest || manifest.ProfileDigest != authority.ProfileDigest {
		t.Fatalf("authority tuple was not bound: %#v", manifest)
	}
	data, err := manifest.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON: %v", err)
	}
	decoded, err := DecodeJSON(data)
	if err != nil {
		t.Fatalf("DecodeJSON: %v", err)
	}
	if decoded.Digest != manifest.Digest || !bytes.Equal(data, mustCanonical(t, decoded)) {
		t.Fatalf("round-trip changed manifest: %#v", decoded)
	}
}

func TestBuildRepresentsAdditionsAndRemovalsWithoutDerivingDecision(t *testing.T) {
	a := testSource(t, "pkg/a.go", "A", "urn:gooo:entity:a")
	b := testSource(t, "pkg/b.go", "B", "urn:gooo:entity:b")
	before := testSnapshot(t, a)
	head := testSnapshot(t, a, b)
	authority := testAuthority(t, before, head, []surfaceFixture{{owner: a.Bindings[0].ID, suffix: "a"}, {owner: b.Bindings[0].ID, suffix: "b"}})
	manifest, err := Build(Input{Before: &before, Head: &head, Authority: authority})
	if err != nil {
		t.Fatalf("Build addition: %v", err)
	}
	if len(manifest.Entries) != 2 || !manifest.Entries[0].BeforePresent || !manifest.Entries[1].AfterPresent {
		t.Fatalf("addition entries = %#v", manifest.Entries)
	}
	removed, err := Build(Input{Before: &head, Head: &before, Authority: swapAuthoritySides(authority)})
	if err != nil {
		t.Fatalf("Build removal: %v", err)
	}
	if removed.Entries[1].BeforePresent != true || removed.Entries[1].AfterPresent != false {
		t.Fatalf("removal entry = %#v", removed.Entries[1])
	}
	if removed.Status != StatusComplete || removed.ReasonCode != "" {
		t.Fatalf("removal status = %#v", removed)
	}
}

func TestPresentationAndAuthorityPathRelocationDoNotEnterDigest(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	before := testSnapshot(t, source)
	head := testSnapshot(t, source)
	first := testAuthority(t, before, head, []surfaceFixture{{owner: source.Bindings[0].ID, suffix: "order"}})
	second := first
	second.Before = append([]SourceMapBinding(nil), first.Before...)
	second.Head = append([]SourceMapBinding(nil), first.Head...)
	second.Before[0].Path = "renamed/order.go"
	second.Head[0].Path = "renamed/order.go"
	// A path is a locator and is not part of the manifest. The binding is
	// intentionally replayed against the same snapshots to exercise digest
	// normalization without changing source authority.
	firstJSON, err := Build(Input{Before: &before, Head: &head, Authority: first})
	if err != nil {
		t.Fatalf("Build first: %v", err)
	}
	if _, err := Build(Input{Before: &before, Head: &head, Authority: second}); err == nil {
		t.Fatal("accepted a source-map path that does not resolve to the snapshot")
	}
	if strings.Contains(string(mustCanonical(t, firstJSON)), "pkg/order.go") {
		t.Fatal("canonical manifest leaked a source path")
	}
}

func TestBuildRejectsStaleAuthorityAndUnknownChangedPath(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	before := testSnapshot(t, source)
	head := testSnapshot(t, source)
	authority := testAuthority(t, before, head, []surfaceFixture{{owner: source.Bindings[0].ID, suffix: "order"}})
	for name, mutate := range map[string]func(*selectiveci.Snapshot){
		"before": func(snapshot *selectiveci.Snapshot) { snapshot.Digest = testDigest("stale-before") },
		"head":   func(snapshot *selectiveci.Snapshot) { snapshot.Digest = testDigest("stale-head") },
	} {
		t.Run(name, func(t *testing.T) {
			staleBefore, staleHead := before, head
			if name == "before" {
				mutate(&staleBefore)
			} else {
				mutate(&staleHead)
			}
			manifest, err := Build(Input{Before: &staleBefore, Head: &staleHead, Authority: authority})
			assertStatus(t, manifest, err, StatusUnknown, CodeInvalidSnapshot)
		})
	}
	unknownPath := authority
	unknownPath.Head = append([]SourceMapBinding(nil), authority.Head...)
	unknownPath.Head[0].Path = "pkg/unknown.go"
	manifest, err := Build(Input{Before: &before, Head: &head, Authority: unknownPath})
	assertStatus(t, manifest, err, StatusUnknown, CodeUnknownChangedSurface)
	if len(manifest.ResolvedSurfaceIDs) != 0 || len(manifest.Entries) != 0 {
		t.Fatalf("unknown path retained resolved work: %#v", manifest)
	}
}

func TestBuildRejectsInventoryAndBindingConflicts(t *testing.T) {
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
			duplicate.SurfaceID = "urn:gooo:surface:other"
			duplicate.SemanticOwnerID = "urn:gooo:entity:other"
			duplicate.SourceMapID = "urn:gooo:map:other"
			a.Inventory = append(a.Inventory, duplicate)
		}, code: CodeConflictingBinding, status: StatusFailClosed},
		{name: "duplicate owner", edit: func(a *RegistrySourceMap) {
			duplicate := a.Inventory[0]
			duplicate.SurfaceID = "urn:gooo:surface:other"
			duplicate.CodeSymbolID = "urn:gooo:symbol:other"
			duplicate.SourceMapID = "urn:gooo:map:other"
			a.Inventory = append(a.Inventory, duplicate)
		}, code: CodeConflictingBinding, status: StatusFailClosed},
		{name: "extra surface", edit: func(a *RegistrySourceMap) {
			extra := a.Head[0]
			extra.SurfaceID = "urn:gooo:surface:extra"
			a.Head = append(a.Head, extra)
		}, code: CodeMalformedBinding, status: StatusFailClosed},
		{name: "candidate", edit: func(a *RegistrySourceMap) { a.CandidateBindings = append([]SourceMapBinding{}, a.Head...) }, code: CodeCandidateBinding, status: StatusUnknown},
		{name: "derived", edit: func(a *RegistrySourceMap) { a.DerivedBindings = append([]SourceMapBinding{}, a.Head...) }, code: CodeDerivedBinding, status: StatusUnknown},
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

func TestBuildRequiresCompleteInventoryAndAuthorityDigests(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	before := testSnapshot(t, source)
	head := testSnapshot(t, source)
	base := testAuthority(t, before, head, []surfaceFixture{{owner: source.Bindings[0].ID, suffix: "order"}})
	cases := []struct {
		name string
		edit func(*RegistrySourceMap)
		code ErrorCode
	}{
		{name: "nil inventory", edit: func(a *RegistrySourceMap) { a.Inventory = nil }, code: CodeMissingAuthority},
		{name: "registry drift", edit: func(a *RegistrySourceMap) { a.RegistryDigest = testDigest("drift") }, code: CodeAuthorityDrift},
		{name: "source map drift", edit: func(a *RegistrySourceMap) { a.SourceMapDigest = testDigest("drift") }, code: CodeAuthorityDrift},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := base
			tc.edit(&input)
			manifest, err := Build(Input{Before: &before, Head: &head, Authority: input})
			assertStatus(t, manifest, err, StatusUnknown, tc.code)
		})
	}
}

func TestBuildDistinguishesMissingObservationFromCompleteEmptyInventory(t *testing.T) {
	before := testSnapshot(t)
	head := testSnapshot(t)
	authority := RegistrySourceMap{
		Schema: AuthoritySchemaV1, RegistryDigest: before.RegistryDigest, SourceMapDigest: before.SourceMapDigest,
		ToolchainDigest: testDigest("toolchain"), ProfileDigest: testDigest("profile"), Inventory: []Surface{}, Before: []SourceMapBinding{}, Head: []SourceMapBinding{},
	}
	complete, err := Build(Input{Before: &before, Head: &head, Authority: authority})
	if err != nil || complete.Status != StatusComplete || !complete.ObservationComplete || len(complete.Entries) != 0 {
		t.Fatalf("complete empty inventory = %#v, err=%v", complete, err)
	}
	missing, err := Build(Input{Authority: authority})
	assertStatus(t, missing, err, StatusUnknown, CodeMissingSnapshot)
	if missing.ObservationComplete || missing.Status == StatusComplete {
		t.Fatalf("missing observation looked complete: %#v", missing)
	}
}

func TestCanonicalJSONRejectsMutationsAndTrustedCounts(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	snapshot := testSnapshot(t, source)
	authority := testAuthority(t, snapshot, snapshot, []surfaceFixture{{owner: source.Bindings[0].ID, suffix: "order"}})
	manifest, err := Build(Input{Before: &snapshot, Head: &snapshot, Authority: authority})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	canonical := mustCanonical(t, manifest)
	mutations := map[string][]byte{
		"duplicate key":  append(bytes.Replace(canonical, []byte(`"status"`), []byte(`"status":"COMPLETE","status"`), 1), []byte{}...),
		"unknown field":  bytes.Replace(canonical, []byte(`"digest"`), []byte(`"extra":true,"digest"`), 1),
		"trailing value": append(append([]byte{}, canonical...), []byte(`{}`)...),
		"wrong schema":   bytes.Replace(canonical, []byte(SchemaV1), []byte("wrong/v9"), 1),
		"counts":         bytes.Replace(canonical, []byte(`"resolved":1`), []byte(`"resolved":0`), 1),
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

func TestBuildIsPureAndDoesNotMutateInputs(t *testing.T) {
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
	if manifest.Status != status || !manifest.FullSuiteRequired || len(manifest.Entries) != 0 || len(manifest.ResolvedSurfaceIDs) != 0 {
		t.Fatalf("manifest = %#v", manifest)
	}
	var typed *Error
	if !errors.As(err, &typed) || typed.Code != code || typed.Status != status {
		t.Fatalf("error = %v, want %s/%s", err, status, code)
	}
}

type surfaceFixture struct {
	owner  string
	suffix string
}

func testAuthority(t *testing.T, before, head selectiveci.Snapshot, surfaces []surfaceFixture) RegistrySourceMap {
	t.Helper()
	authority := RegistrySourceMap{
		Schema: AuthoritySchemaV1, RegistryDigest: before.RegistryDigest, SourceMapDigest: before.SourceMapDigest,
		ToolchainDigest: testDigest("toolchain"), ProfileDigest: testDigest("profile"), Inventory: make([]Surface, 0, len(surfaces)),
		Before: make([]SourceMapBinding, 0), Head: make([]SourceMapBinding, 0),
	}
	byOwner := make(map[string]surfaceFixture, len(surfaces))
	for _, fixture := range surfaces {
		byOwner[fixture.owner] = fixture
		authority.Inventory = append(authority.Inventory, Surface{SurfaceID: "urn:gooo:surface:" + fixture.suffix, CodeSymbolID: "urn:gooo:symbol:" + fixture.suffix, SemanticOwnerID: fixture.owner, SourceMapID: "urn:gooo:map:" + fixture.suffix})
	}
	authority.Before = authorityBindings(t, before, authority.Inventory, byOwner)
	authority.Head = authorityBindings(t, head, authority.Inventory, byOwner)
	return authority
}

func authorityBindings(t *testing.T, snapshot selectiveci.Snapshot, inventory []Surface, fixtures map[string]surfaceFixture) []SourceMapBinding {
	t.Helper()
	byOwner := make(map[string]Surface, len(inventory))
	for _, surface := range inventory {
		byOwner[surface.SemanticOwnerID] = surface
	}
	result := make([]SourceMapBinding, 0)
	for _, source := range snapshot.Sources {
		for _, binding := range source.Bindings {
			surface := byOwner[binding.ID]
			_ = fixtures[binding.ID]
			result = append(result, SourceMapBinding{SurfaceID: surface.SurfaceID, CodeSymbolID: surface.CodeSymbolID, SemanticOwnerID: surface.SemanticOwnerID, SourceMapID: surface.SourceMapID, Role: binding.Role, Path: source.Path, BlobDigest: source.BlobDigest, BindingDigest: binding.BindingDigest, SourceMapBindingDigest: binding.BindingDigest})
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

func mustCanonical(t *testing.T, manifest Manifest) []byte {
	t.Helper()
	data, err := manifest.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON: %v", err)
	}
	return data
}
