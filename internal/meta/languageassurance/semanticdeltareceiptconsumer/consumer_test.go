package semanticdeltareceiptconsumer

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestConsumerRejectsUnsealedWireReceipt(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../../../..")
	input := Input{CaseID: "equivalent", SubjectSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", BeforePath: filepath.Join(root, "examples/semantic-delta-receipt/before.gooo"), AfterPath: filepath.Join(root, "examples/semantic-delta-receipt/equivalent-after.gooo")}
	verdict := AdjudicateFiles(input, Receipt{})
	if verdict.Passed || verdict.Reason != reasonReceipt || verdict.Consumer != consumerName {
		t.Fatalf("verdict=%+v", verdict)
	}
}
