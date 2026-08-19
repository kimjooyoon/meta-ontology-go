package generator

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestCanonicalMarkerManifestV1HasCheckedInEncoding(t *testing.T) {
	ir := acceptanceFixture()
	result := mustAcceptanceResult(t, ir, nil)
	markers, err := parseMarkers(result.Source)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := canonicalMarkerManifestV1(result.Source, markers, ir)
	if err != nil {
		t.Fatal(err)
	}
	const header = "schema\tgooo.generator.marker-manifest\tprofile\tgenerated-regions-and-slots\tversion\t1\tencoding\tutf-8-byte-length-hex\tterminal-newline\tLF\n"
	if !bytes.HasPrefix(manifest, []byte(header)) || !bytes.HasSuffix(manifest, []byte{'\n'}) {
		t.Fatalf("manifest does not use the checked-in LF format:\n%s", manifest)
	}
	if strings.Contains(string(manifest), "\r") || !strings.Contains(string(manifest), "regions\t4\n") || !strings.Contains(string(manifest), "slot\t") {
		t.Fatalf("manifest omitted required typed records:\n%s", manifest)
	}
	digest := sha256.Sum256(manifest)
	const wantDigest = "f35f03e7bb130e1dadfb65ed40b5fda91eed61620f2d2f3fface06e08f971d23"
	if hex.EncodeToString(digest[:]) != wantDigest {
		t.Fatalf("canonical marker manifest digest changed: got %s want %s", hex.EncodeToString(digest[:]), wantDigest)
	}
}
func TestCanonicalMarkerManifestV1IsIdempotentAcrossObservationOrder(t *testing.T) {
	ir := acceptanceFixture()
	result := mustAcceptanceResult(t, ir, nil)
	markers, err := parseMarkers(result.Source)
	if err != nil {
		t.Fatal(err)
	}
	first, err := canonicalMarkerManifestV1(result.Source, markers, ir)
	if err != nil {
		t.Fatal(err)
	}
	permuted := cloneParsedMarkersV1(markers)
	for left, right := 0, len(permuted.Regions)-1; left < right; left, right = left+1, right-1 {
		permuted.Regions[left], permuted.Regions[right] = permuted.Regions[right], permuted.Regions[left]
	}
	for index := range permuted.Regions {
		for left, right := 0, len(permuted.Regions[index].Slots)-1; left < right; left, right = left+1, right-1 {
			permuted.Regions[index].Slots[left], permuted.Regions[index].Slots[right] = permuted.Regions[index].Slots[right], permuted.Regions[index].Slots[left]
		}
	}
	second, err := canonicalMarkerManifestV1(result.Source, permuted, ir)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("equivalent marker observations produced different canonical bytes")
	}
	changed := cloneParsedMarkersV1(markers)
	changed.Regions[0].ID, changed.Regions[1].ID = changed.Regions[1].ID, changed.Regions[0].ID
	third, err := canonicalMarkerManifestV1(result.Source, changed, ir)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(first, third) {
		t.Fatal("marker identity change did not change canonical bytes")
	}
}
