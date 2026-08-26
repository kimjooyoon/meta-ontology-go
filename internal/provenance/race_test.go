package provenance

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestConcurrentAppendsRemainOneCanonicalChain(t *testing.T) {
	store := New(filepath.Join(t.TempDir(), "race.jsonl"))
	const count = 32
	var wait sync.WaitGroup
	for index := range count {
		wait.Go(func() {
			record := testRecord("event/race/"+string(rune('a'+index)), "semantic/race/"+string(rune('a'+index)), StatusVerified)
			if err := store.Append(record); err != nil {
				t.Errorf("concurrent append %d: %v", index, err)
			}
		})
	}
	wait.Wait()
	snapshot, err := store.Read(ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Records) != count {
		t.Fatalf("concurrent appends lost records: got %d want %d", len(snapshot.Records), count)
	}
}
