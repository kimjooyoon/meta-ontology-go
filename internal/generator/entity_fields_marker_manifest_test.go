package generator

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestEntityFieldsSupportedCanonicalMarkerManifestReplay(t *testing.T) {
	ir := entityFieldsFixture()
	first := supportedEntityFieldsResult(t, ir, nil)
	second := supportedEntityFieldsResult(t, ir, nil)
	firstManifest := entityFieldsCanonicalMarkerManifest(t, first, ir)
	secondManifest := entityFieldsCanonicalMarkerManifest(t, second, ir)
	if !bytes.Equal(firstManifest, secondManifest) {
		t.Fatal("EntityFields marker manifest changed across clean replay")
	}
	digest := sha256.Sum256(firstManifest)
	const wantDigest = "0651ed593dfd439e5196fe9223f13b90a48e3fd24db7ebd841ec70c0778dcb83"
	if hex.EncodeToString(digest[:]) != wantDigest {
		t.Fatalf("EntityFields marker manifest digest changed: got %s want %s", hex.EncodeToString(digest[:]), wantDigest)
	}
}

func entityFieldsCanonicalMarkerManifest(t *testing.T, result Result, ir SemanticIR) []byte {
	t.Helper()
	markers, err := parseMarkers(result.Source)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := canonicalMarkerManifestV1(result.Source, markers, ir)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}
