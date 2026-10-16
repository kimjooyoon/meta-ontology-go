package viewlens

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestOriginProfileArchiveChainFresh1816(t *testing.T) {
	directory, request := t.TempDir(), freshHintRequest()
	request.ActorID = "profile-chain-1816"
	prior := ""
	for index := range 3 {
		loaded := Load(directory, request.ActorID, request.Audience)
		hint, _ := json.Marshal(loaded.Snapshot)
		request.ActivityID = fmt.Sprintf("urn:profile-chain1816:%d", index)
		event, _ := json.Marshal(Select(request, hint).Record)
		accepted := Record(directory, event)
		if !accepted.Persisted || accepted.ProfileArchivePath == "" {
			t.Fatalf("profile not archived: %+v", accepted)
		}
		restored := ReadArchive(directory, "event", accepted.EventDigest)
		if restored.State != "RESTORED" || !bytes.Equal(restored.Raw, event) {
			t.Fatalf("event bytes not accessible: %+v", restored)
		}
		profile := ReadArchive(directory, "profile", accepted.Snapshot.ProfileDigest)
		if profile.State != "RESTORED" || digest(profile.Raw) != accepted.Snapshot.ProfileDigest {
			t.Fatalf("profile bytes not accessible: %+v", profile)
		}
		if prior != "" {
			ancestor := ReadArchive(directory, "profile", prior)
			if ancestor.State != "RESTORED" || loaded.Snapshot.ProfileDigest != prior {
				t.Fatal("previous snapshot raw bytes were overwritten")
			}
		}
		prior = accepted.Snapshot.ProfileDigest
	}
	if got := ReadArchive(directory, "../event", prior); got.State != "UNAVAILABLE" || got.Issue != "ARCHIVE_REQUEST_INVALID" {
		t.Fatal("unknown archive domain accepted")
	}
}
