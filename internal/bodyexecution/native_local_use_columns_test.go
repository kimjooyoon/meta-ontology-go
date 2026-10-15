package bodyexecution

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestRecordedUseColumnsRetainVersionReferences(t *testing.T) {
	f, err := os.Open("../../examples/local-version-origins/recorded-runtime.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	var runtime CompositionRuntime
	if err := json.Unmarshal(raw, &runtime); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	scope := "sha256:" + hex.EncodeToString(sum[:])
	columns, report, err := SelectRecordedLocalUses(runtime, NativeOriginQuery{Kind: "local-uses", Offset: 14, Limit: 2}, scope)
	if err != nil || columns.ReceiptSHA256 != scope || columns.SourceSHA256 != runtime.SelectedSourceSHA256 || columns.JournalSHA256 != runtime.LocalObservation.SHA256 {
		t.Fatal("saved scope", report, err)
	}
	if report.TotalRows != 32 || report.ReturnedRows != 2 || report.ConsumerRows != 1 || report.InputShape != "gooo/local-use-columns/v1" || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
		t.Fatal("recorded page scope", report)
	}
	for i, assignment := range []int{0, 7} {
		if columns.Cases[i] != 1 || columns.UseIndices[i] != 6+i || columns.Frames[i] != 0 || columns.Assignments[i] != assignment || columns.VersionFrames[i] != 0 || !columns.VersionCompleted[i] || columns.VersionActivities[i] != columns.Activities[i] {
			t.Fatal("original case/use/version numbering changed", i)
		}
	}
	if _, _, err := SelectRecordedLocalUses(runtime, NativeOriginQuery{Kind: "local-writes", Limit: 1}, scope); err == nil {
		t.Fatal("write rows accepted as use columns")
	}
	if _, err := localUseTraceIndices([]CompositionTrace{{CaseIndex: 9}, {CaseIndex: 9}}); err == nil {
		t.Fatal("ambiguous case table accepted")
	}
}
