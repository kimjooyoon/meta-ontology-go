package viewlens

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"unsafe"
)

func TestDetachedArchiveGrowthFresh1810(t *testing.T) {
	directory, request := t.TempDir(), freshHintRequest()
	request.ActorID = "ledger-growth-1810"
	rawTotal, storedTotal := 0, 0
	for index := range 40 {
		loaded := Load(directory, request.ActorID, request.Audience)
		hint, _ := json.Marshal(loaded.Snapshot)
		request.ActivityID = fmt.Sprintf("urn:fresh:archive1810:activity:%d", index)
		request.View = []string{"full", "inputs"}[index%2]
		event, _ := json.Marshal(Select(request, hint).Record)
		result := Record(directory, event)
		if !result.Persisted || result.Snapshot.Revision != int64(index+1) {
			t.Fatalf("record %d: %+v", index, result)
		}
		restored, size, err := readPacked(result.ArchivePath)
		if err != nil || !bytes.Equal(restored, event) || digest(restored) != result.EventDigest {
			t.Fatalf("archive %d lost raw bytes: %v", index, err)
		}
		rawTotal, storedTotal = rawTotal+len(event), storedTotal+size
	}
	files, err := filepath.Glob(filepath.Join(directory, "events", "*.gopack"))
	if err != nil || len(files) != 40 {
		t.Fatalf("older origins lost: %v %d", err, len(files))
	}
	path, _ := profilePath(directory, request.ActorID, request.Audience)
	profile, raw, err := readProfile(path, request.ActorID, request.Audience)
	if err != nil || len(profile.Recent) != 32 || profile.Snapshot.Revision != 40 {
		t.Fatalf("bounded summary: %v %+v", err, profile)
	}
	_, size, err := readPacked(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ORIGIN_LAYOUT_METRIC %s", mustMetric1810(map[string]any{
		"corpus": "detached-growth-1810", "events": 40, "raw_event_bytes": rawTotal,
		"stored_event_bytes": storedTotal, "raw_profile_bytes": len(raw), "stored_profile_bytes": size,
		"restored_event_byte_matches": 40, "recent_summary_rows": len(profile.Recent),
		"snapshot_inline_bytes": unsafe.Sizeof(Snapshot{}), "choice_inline_bytes": unsafe.Sizeof(Choice{}),
		"layout_scope": "Go sizeof excludes referenced string/slice backing storage; no latency or RSS claim",
	}))
	if err := os.WriteFile(path, []byte("broken optional storage"), 0600); err != nil {
		t.Fatal(err)
	}
	if result := Load(directory, request.ActorID, request.Audience); result.State != "UNAVAILABLE" {
		t.Fatal("corrupt profile accepted")
	}
	if result := Select(request, nil); result.Record == nil || result.Choice.ViewID != NamedInputs {
		t.Fatal("corrupt detached storage affected pure selection")
	}
}

func mustMetric1810(value any) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}
