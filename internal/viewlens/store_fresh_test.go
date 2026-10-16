package viewlens

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDetachedOriginLedgerFresh1802(t *testing.T) {
	directory, request := t.TempDir(), freshHintRequest()
	request.Prefer = "inputs"
	first := Select(request, nil)
	raw, _ := json.Marshal(first.Record)
	accepted := Record(directory, raw)
	if !accepted.Persisted || accepted.State != "SAVED" || accepted.Snapshot.Revision != 1 {
		t.Fatalf("record: %+v", accepted)
	}
	if replay := Record(directory, raw); replay.State != "ALREADY_RECORDED" || replay.Snapshot.Revision != 1 {
		t.Fatalf("idempotent last record: %+v", replay)
	}
	request.View = "full"
	stale, _ := json.Marshal(Select(request, nil).Record)
	if got := Record(directory, stale); got.Persisted || got.Issue != "PROFILE_SNAPSHOT_STALE" {
		t.Fatalf("stale record accepted: %+v", got)
	}
	loaded := Load(directory, request.ActorID, request.Audience)
	hint, _ := json.Marshal(loaded.Snapshot)
	request.Prefer = ""
	second, _ := json.Marshal(Select(request, hint).Record)
	saved := Record(directory, second)
	if !saved.Persisted || saved.Snapshot.Revision != 2 || saved.Snapshot.FullSelections != 1 ||
		saved.Snapshot.NamedSelections != 1 || saved.Snapshot.Preferred != NamedInputs {
		t.Fatalf("second record: %+v", saved)
	}
	for _, record := range []StoreResult{accepted, saved} {
		restored, size, err := readPacked(record.ArchivePath)
		if err != nil || digest(restored) != record.EventDigest || size != record.StoredBytes {
			t.Fatalf("raw event restoration: %v %+v", err, record)
		}
	}
	if restored, _, _ := readPacked(accepted.ArchivePath); !bytes.Equal(restored, raw) {
		t.Fatal("first canonical event bytes changed")
	}
}

func TestDetachedBusyHintFresh1803(t *testing.T) {
	directory, request := t.TempDir(), freshHintRequest()
	path, err := profilePath(directory, request.ActorID, request.Audience)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if locked, err := tryLock(file); err != nil || !locked {
		t.Skipf("native lock unavailable: %v", err)
	}
	raw, _ := json.Marshal(Select(request, nil).Record)
	if got := Record(directory, raw); got.Persisted || got.Issue != "PROFILE_BUSY" {
		t.Fatalf("busy store: %+v", got)
	}
	if got := Select(request, []byte("broken hint")); got.Choice.ViewID != NamedInputs {
		t.Fatalf("busy store affected pure selection: %+v", got)
	}
	if files, _ := filepath.Glob(filepath.Join(directory, "events", "*")); len(files) != 0 {
		t.Fatal("busy store wrote an event")
	}
}
