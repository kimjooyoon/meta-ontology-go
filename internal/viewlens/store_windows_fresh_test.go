//go:build windows

package viewlens

import (
	"encoding/json"
	"testing"
)

func TestWindowsAdvisoryStoreFresh1818(t *testing.T) {
	directory, request := t.TempDir(), freshHintRequest()
	request.ActorID, request.View = "windows-native-1818", "full"
	for index := range 2 {
		loaded := Load(directory, request.ActorID, request.Audience)
		hint, _ := json.Marshal(loaded.Snapshot)
		event, _ := json.Marshal(Select(request, hint).Record)
		accepted := Record(directory, event)
		if !accepted.Persisted || accepted.Snapshot.Revision != int64(index+1) {
			t.Fatalf("native Windows lock/profile replacement: %+v", accepted)
		}
		if got := ReadArchive(directory, "profile", accepted.Snapshot.ProfileDigest); got.State != "RESTORED" {
			t.Fatalf("native Windows profile archive: %+v", got)
		}
	}
}
