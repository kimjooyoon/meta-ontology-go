package generator

import (
	"strings"
	"testing"
)

func TestMergeGeneratedRequiresLegacyMarkerLineBoundaries(t *testing.T) {
	existing := "package generated\nvar marker = \"//gooo:generated begin old\"\n"
	fresh := BeginMarker + " old\nnew body\n" + EndMarker + "\n"
	merged, err := MergeGenerated(existing, fresh)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(merged, "var marker") || !strings.Contains(merged, "new body") {
		t.Fatalf("legacy marker in string was treated as a block: %s", merged)
	}
}

func TestMergeGeneratedRejectsMalformedFreshWithoutWritingInputs(t *testing.T) {
	existing := BeginMarker + "old\nold body\n" + EndMarker + "\n"
	fresh := BeginMarker + "old\nnew body\n"
	existingBefore, freshBefore := existing, fresh
	merged, err := MergeGenerated(existing, fresh)
	if err == nil || merged != "" {
		t.Fatalf("expected malformed fresh source rejection, got merged=%q err=%v", merged, err)
	}
	if existing != existingBefore || fresh != freshBefore {
		t.Fatal("legacy rejection changed caller-owned source")
	}
}

func TestMergeGeneratedRejectsLegacyEndAttributes(t *testing.T) {
	malformed := BeginMarker + "old\nbody\n" + EndMarker + " extra\n"
	if _, err := MergeGenerated(malformed, malformed); err == nil {
		t.Fatal("legacy end attributes were accepted")
	}
}
