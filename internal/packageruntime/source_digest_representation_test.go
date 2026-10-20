package packageruntime

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestPackageSourceDigestUsesJSONStringBytesAndRejectsRawDigest(t *testing.T) {
	manifest := interfaceFixture()
	manifest.Packages[1].Sources[1].Content += "// 한국어 \"quoted\"\n"
	source := manifest.Packages[1].Sources[1]
	image, err := Build(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var entry *SourceImage
	for packageIndex := range image.Packages {
		if image.Packages[packageIndex].Path != "records" {
			continue
		}
		for sourceIndex := range image.Packages[packageIndex].Sources {
			candidate := &image.Packages[packageIndex].Sources[sourceIndex]
			if candidate.Filename == source.Filename {
				entry = candidate
			}
		}
	}
	if entry == nil {
		t.Fatal("source image entry missing")
	}
	jsonString, err := json.Marshal(source.Content)
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("sha256:%x", sha256.Sum256(jsonString))
	raw := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(source.Content)))
	if entry.SourceDigest != want || raw == want {
		t.Fatalf("source entry does not distinguish JSON string bytes from raw source bytes: stored=%q json=%q raw=%q", entry.SourceDigest, want, raw)
	}
	if _, err := ParseResolvedPackageSource(image, "records", source); err != nil {
		t.Fatalf("matching JSON string digest should replay: %v", err)
	}
	entry.SourceDigest = raw
	if _, err := ParseResolvedPackageSource(image, "records", source); err == nil ||
		!strings.Contains(err.Error(), "image.packages[].sources[].source_digest") ||
		!strings.Contains(err.Error(), "sha256(json.Marshal(source.Content))") {
		t.Fatalf("raw source digest did not fail closed with its representation: %v", err)
	}
}
