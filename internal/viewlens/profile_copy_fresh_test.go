package viewlens

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestCopiedProfileOriginPreservationFresh1820(t *testing.T) {
	original, copied, request := t.TempDir(), t.TempDir(), freshHintRequest()
	request.ActorID = "copied-profile-1820"
	event, _ := json.Marshal(Select(request, nil).Record)
	first := Record(original, event)
	if !first.Persisted {
		t.Fatalf("initial origin: %+v", first)
	}
	oldPath, _ := profilePath(original, request.ActorID, request.Audience)
	newPath, _ := profilePath(copied, request.ActorID, request.Audience)
	packed, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, packed, 0600); err != nil {
		t.Fatal(err)
	}
	loaded := Load(copied, request.ActorID, request.Audience)
	hint, _ := json.Marshal(loaded.Snapshot)
	request.ActivityID = "urn:fresh:copied-profile1820:next"
	next, _ := json.Marshal(Select(request, hint).Record)
	second := Record(copied, next)
	if !second.Persisted {
		t.Fatalf("copied profile update: %+v", second)
	}
	ancestor := ReadArchive(copied, "profile", first.Snapshot.ProfileDigest)
	want := ReadArchive(original, "profile", first.Snapshot.ProfileDigest)
	if ancestor.State != "RESTORED" || !bytes.Equal(ancestor.Raw, want.Raw) {
		t.Fatal("overwriting a copied current profile lost its raw origin")
	}
	if missing := ReadArchive(copied, "event", first.EventDigest); missing.State != "UNAVAILABLE" {
		t.Fatal("partial copy was incorrectly described as a complete history")
	}
}
