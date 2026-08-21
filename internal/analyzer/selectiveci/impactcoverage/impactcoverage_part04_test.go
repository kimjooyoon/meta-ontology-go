package impactcoverage

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"strings"
	"testing"
)

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
