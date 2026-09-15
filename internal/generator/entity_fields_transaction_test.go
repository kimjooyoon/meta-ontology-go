package generator

import (
	"bytes"
	"os"
	"testing"
)

func TestEntityFieldsRejectedGenerationLeavesFilesystemSnapshotUnchanged(t *testing.T) {
	root := t.TempDir()
	sentinel := root + "/sentinel.txt"
	if err := os.WriteFile(sentinel, []byte("preserved"), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot := func() ([]string, []byte) {
		entries, err := os.ReadDir(root)
		if err != nil {
			t.Fatal(err)
		}
		names := make([]string, len(entries))
		for index, entry := range entries {
			names[index] = entry.Name()
		}
		data, err := os.ReadFile(sentinel)
		if err != nil {
			t.Fatal(err)
		}
		return names, data
	}
	beforeNames, beforeData := snapshot()
	ir := entityFieldsFixture()
	ir.Entities[0].Fields[0].Presence = "optional"
	result, err := New(Options{}).generateWithEntityFieldsSupport(ir, nil, supportedEntityFieldsForTest())
	if err == nil || result.Source != nil || result.SourceMap.Mappings != nil {
		t.Fatalf("rejected generation produced artifacts: result=%#v err=%v", result, err)
	}
	afterNames, afterData := snapshot()
	if len(beforeNames) != len(afterNames) || !bytes.Equal(beforeData, afterData) {
		t.Fatalf("rejected generation changed filesystem snapshot: before=%v/%q after=%v/%q", beforeNames, beforeData, afterNames, afterData)
	}
	for index := range beforeNames {
		if beforeNames[index] != afterNames[index] {
			t.Fatalf("rejected generation changed filesystem entries: before=%v after=%v", beforeNames, afterNames)
		}
	}
}
