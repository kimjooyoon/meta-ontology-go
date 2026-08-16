package impactcoverage

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
)

type wantVector struct {
	decision Decision
	reason   Reason
	full     bool
	changed  uint64
	covered  uint64
	open     uint64
	bindings uint64
	work     uint64
	ids      []string
	paths    []string
}

func TestLiteralFixtureMatrix(t *testing.T) {
	cases := literalCases(t)
	for _, fixture := range cases {
		t.Run(fixture.name, func(t *testing.T) {
			got := Observe(NewInput(&fixture.base, &fixture.head))
			assertVector(t, got, fixture.want)
			assertDigests(t, fixture.base, fixture.head, got)
			t.Logf("vector changed=%d covered=%d uncovered=%d bindings=%d work=%d ids=%v paths=%v",
				got.ChangedBlobCount, got.CoveredChangedBlobCount, got.UncoveredChangedBlobCount,
				got.ChangedBindingCount, got.DeterministicWorkUnits, got.ChangedStableIDs, got.UncoveredPaths)
			t.Logf("digests base=%s head=%s source-map=%s/%s registry=%s/%s input=%s output=%s",
				got.BaseSnapshotDigest, got.HeadSnapshotDigest, got.BaseSourceMapDigest,
				got.HeadSourceMapDigest, got.BaseRegistryDigest, got.HeadRegistryDigest,
				got.InputDigest, got.OutputDigest)
		})
	}
}

type fixtureCase struct {
	name       string
	base, head selectiveci.Snapshot
	want       wantVector
}

func literalCases(t *testing.T) []fixtureCase {
	t.Helper()
	a := boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a")
	b := boundSource("pkg/b.go", "b-1", "urn:gooo:entity:b")
	c := boundSource("pkg/c.go", "c-1", "urn:gooo:entity:c")
	return []fixtureCase{
		{"exact-replay", snap(t, "map", "reg", a), snap(t, "map", "reg", a), wantVector{
			DecisionExact, ReasonNoChange, false, 0, 0, 0, 0, 3, []string{}, []string{},
		}},
		{"no-change-order-permutation", snap(t, "map", "reg", a, b), snap(t, "map", "reg", b, a), wantVector{
			DecisionExact, ReasonNoChange, false, 0, 0, 0, 0, 6, []string{}, []string{},
		}},
		{"modify-add-delete", snap(t, "map", "reg", a, b), snap(t, "map", "reg",
			boundSource("pkg/a.go", "a-2", "urn:gooo:entity:a"), c), wantVector{
			DecisionExact, ReasonComplete, false, 3, 3, 0, 3, 7,
			[]string{"urn:gooo:entity:a", "urn:gooo:entity:b", "urn:gooo:entity:c"}, []string{},
		}},
		{"relocation", snap(t, "map", "reg", a), snap(t, "map", "reg",
			boundSource("pkg/new.go", "a-1", "urn:gooo:entity:a")), wantVector{
			DecisionExact, ReasonComplete, false, 2, 2, 0, 1, 4, []string{"urn:gooo:entity:a"}, []string{},
		}},
		{"binding-set-only", snap(t, "map", "reg", a), snap(t, "map", "reg",
			boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a", "urn:gooo:entity:b")), wantVector{
			DecisionExact, ReasonComplete, false, 1, 1, 0, 2, 4,
			[]string{"urn:gooo:entity:a", "urn:gooo:entity:b"}, []string{},
		}},
		{"unbound-changed-path", snap(t, "map", "reg", emptySource("pkg/u.go", "u-1")),
			snap(t, "map", "reg", emptySource("pkg/u.go", "u-2")), wantVector{
				DecisionUnknown, ReasonMissingBinding, true, 1, 0, 1, 0, 1, []string{}, []string{"pkg/u.go"},
			}},
		{"one-side-binding", snap(t, "map", "reg", a), snap(t, "map", "reg", emptySource("pkg/a.go", "a-2")), wantVector{
			DecisionExact, ReasonComplete, false, 1, 1, 0, 1, 2, []string{"urn:gooo:entity:a"}, []string{},
		}},
		{"registry-drift", snap(t, "map", "reg-a", a), snap(t, "map", "reg-b", a), wantVector{
			DecisionUnknown, ReasonAuthorityDrift, true, 0, 0, 0, 0, 3, []string{}, []string{},
		}},
		{"source-map-drift", snap(t, "map-a", "reg", a), snap(t, "map-b", "reg", a), wantVector{
			DecisionUnknown, ReasonAuthorityDrift, true, 0, 0, 0, 0, 3, []string{}, []string{},
		}},
	}
}

func assertVector(t *testing.T, got Result, want wantVector) {
	t.Helper()
	if got.Decision != want.decision || got.Reason != want.reason || got.FullSuiteRequired != want.full {
		t.Fatalf("decision vector = %#v, want %s/%s/%t", got, want.decision, want.reason, want.full)
	}
	if got.ChangedBlobCount != want.changed || got.CoveredChangedBlobCount != want.covered ||
		got.UncoveredChangedBlobCount != want.open || got.ChangedBindingCount != want.bindings ||
		got.DeterministicWorkUnits != want.work {
		t.Fatalf("numeric vector = %#v, want changed=%d covered=%d open=%d bindings=%d work=%d",
			got, want.changed, want.covered, want.open, want.bindings, want.work)
	}
	if !reflect.DeepEqual(got.ChangedStableIDs, want.ids) || !reflect.DeepEqual(got.UncoveredPaths, want.paths) {
		t.Fatalf("set vector = %#v/%#v, want %#v/%#v", got.ChangedStableIDs, got.UncoveredPaths, want.ids, want.paths)
	}
}

func assertDigests(t *testing.T, base, head selectiveci.Snapshot, got Result) {
	t.Helper()
	input := NewInput(&base, &head)
	if got.InputDigest != input.Digest() || got.OutputDigest != got.StableDigest() {
		t.Fatalf("digest vector = input %q/%q output %q/%q", got.InputDigest,
			input.Digest(), got.OutputDigest, got.StableDigest())
	}
	if got.BaseSnapshotDigest != base.Digest || got.HeadSnapshotDigest != head.Digest ||
		got.BaseSourceMapDigest != base.SourceMapDigest || got.HeadSourceMapDigest != head.SourceMapDigest ||
		got.BaseRegistryDigest != base.RegistryDigest || got.HeadRegistryDigest != head.RegistryDigest {
		t.Fatalf("authority digest vector = %#v", got)
	}
}

func TestInvalidAndDuplicateSnapshotRecordsFailClosed(t *testing.T) {
	base := snap(t, "map", "reg", boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a"))
	stale := base
	stale.Digest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	got := Observe(NewInput(&base, &stale))
	if got.Decision != DecisionUnknown || got.Reason != ReasonInvalidSnapshot ||
		!got.FullSuiteRequired || len(got.ChangedStableIDs) != 0 {
		t.Fatalf("stale snapshot result = %#v", got)
	}
	duplicate := base
	duplicate.Sources = append(append([]selectiveci.Source{}, base.Sources...), base.Sources[0])
	got = Observe(NewInput(&base, &duplicate))
	if got.Decision != DecisionUnknown || got.Reason != ReasonInvalidSnapshot || len(got.ChangedStableIDs) != 0 {
		t.Fatalf("duplicate snapshot result = %#v", got)
	}
}

func TestRootLocationAndOrderingAreInvariant(t *testing.T) {
	base := snap(t, "map", "reg", boundSource("./pkg/a.go", "a-1", "urn:gooo:entity:a"))
	head := snap(t, "map", "reg", boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a"))
	got := Observe(NewInput(&base, &head))
	if got.Decision != DecisionExact || got.Reason != ReasonNoChange || got.ChangedBlobCount != 0 {
		t.Fatalf("root-location result = %#v", got)
	}
	baseJSON, err := base.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	headJSON, err := head.CanonicalJSON()
	if err != nil || !bytes.Equal(baseJSON, headJSON) {
		t.Fatalf("root-location canonical bytes differ: %s/%s/%v", baseJSON, headJSON, err)
	}
}

func TestNilAuthorityAndOverflowFailClosed(t *testing.T) {
	valid := snap(t, "map", "reg", boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a"))
	for name, input := range map[string]Input{
		"nil-base": NewInput(nil, &valid), "nil-head": NewInput(&valid, nil),
	} {
		t.Run(name, func(t *testing.T) {
			got := Observe(input)
			if got.Decision != DecisionUnknown || got.Reason != ReasonInvalidSnapshot || !got.FullSuiteRequired {
				t.Fatalf("nil authority result = %#v", got)
			}
		})
	}
	if _, err := checkedAdd(^uint64(0), 1); err == nil {
		t.Fatal("work-unit overflow was accepted")
	}
	if units, err := checkedAdd(4, 5); err != nil || units != 9 {
		t.Fatalf("checked work units = %d/%v", units, err)
	}
}

func TestStrictJSONAndExplicitEmptyBoundary(t *testing.T) {
	base := snap(t, "map", "reg", emptySource("pkg/u.go", "u-1"))
	head := snap(t, "map", "reg", emptySource("pkg/u.go", "u-2"))
	input := NewInput(&base, &head)
	canonical, err := input.CanonicalJSON()
	if err != nil || !bytes.Contains(canonical, []byte(`"bindings":[]`)) {
		t.Fatalf("explicit empty canonical input = %s/%v", canonical, err)
	}
	if _, err := DecodeInput(canonical); err != nil {
		t.Fatalf("DecodeInput canonical: %v", err)
	}
	encoded, err := EncodeInputJSON(input)
	if err != nil || !bytes.Equal(encoded, canonical) {
		t.Fatalf("EncodeInputJSON = %s/%v, want canonical bytes", encoded, err)
	}
	null := bytes.Replace(canonical, []byte(`"bindings":[]`), []byte(`"bindings":null`), 1)
	if _, err := DecodeInput(null); err == nil {
		t.Fatal("JSON null bindings accepted")
	}
	duplicateRecord := duplicateBaseSource(t, input)
	if _, err := DecodeInput(duplicateRecord); err == nil {
		t.Fatal("duplicate snapshot source record accepted")
	}
	for name, data := range map[string][]byte{
		"duplicate":    []byte(strings.Replace(string(canonical), `{"schema":`, `{"schema":"duplicate","schema":`, 1)),
		"unknown":      []byte(strings.Replace(string(canonical), `,"head":`, `,"unknown":true,"head":`, 1)),
		"trailing":     append(append([]byte{}, canonical...), []byte(`{}`)...),
		"presentation": append([]byte(" \n"), append(canonical, '\n')...),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeInput(data); err == nil {
				t.Fatalf("accepted malformed input %s", data)
			}
		})
	}
	if base.Digest == head.Digest {
		t.Fatal("explicit empty blob change did not change snapshot digest")
	}
	inputWithNilBindings := selectiveci.SnapshotInput{
		Sources:         []selectiveci.SourceInput{{Path: "pkg/u.go", BlobDigest: blobDigest("u-1")}},
		SourceMapDigest: digest("map"), RegistryDigest: digest("reg"), RegisteredIDs: []string{},
	}
	unknown, err := selectiveci.Build(inputWithNilBindings)
	if err == nil || unknown.Status != selectiveci.StatusUnknown || !unknown.FullSuiteFallback {
		t.Fatal("nil binding array accepted")
	}
	nonEmpty := snap(t, "map", "reg", boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a"))
	nonEmptyBytes, err := nonEmpty.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := selectiveci.DecodeSnapshot(nonEmptyBytes)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, err := decoded.CanonicalJSON()
	if err != nil || !bytes.Equal(nonEmptyBytes, reencoded) {
		t.Fatalf("non-empty snapshot bytes changed: %s/%s/%v", nonEmptyBytes, reencoded, err)
	}
}

func duplicateBaseSource(t *testing.T, input Input) []byte {
	t.Helper()
	data, err := input.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	base, err := input.Base.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	baseText := string(base)
	start := strings.Index(baseText, `,"sources":[`) + len(`,"sources":[`)
	end := strings.Index(baseText[start:], `],"digest":`) + start
	record := baseText[start:end]
	duplicate := strings.Replace(baseText, `],"digest":`, `,`+record+`],"digest":`, 1)
	return []byte(strings.Replace(string(data), baseText, duplicate, 1))
}

func TestOutputJSONRoundTripAndExpectedLabelIsolation(t *testing.T) {
	base := snap(t, "map", "reg", boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a"))
	head := snap(t, "map", "reg", boundSource("pkg/a.go", "a-2", "urn:gooo:entity:a"))
	fixture := struct {
		Input Input
		Label string
	}{NewInput(&base, &head), "expected:COMPLETE"}
	first := Observe(fixture.Input)
	firstEncoded, err := EncodeJSON(first)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeJSON(firstEncoded)
	if err != nil || !reflect.DeepEqual(first, decoded) {
		t.Fatalf("output round trip = %#v/%v", decoded, err)
	}
	fixture.Label = "expected:UNKNOWN"
	second := Observe(fixture.Input)
	if first.InputDigest != second.InputDigest || first.OutputDigest != second.OutputDigest ||
		!reflect.DeepEqual(first, second) {
		t.Fatalf("expected label changed observation: %#v/%#v", first, second)
	}
}
