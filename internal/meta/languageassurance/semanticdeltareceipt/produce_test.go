package semanticdeltareceipt

import (
	"path/filepath"
	"runtime"
	"testing"
)

const candidateSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func fixture(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../../../..")
	return filepath.Join(root, "examples", "semantic-delta-receipt", name)
}

func TestProducerReadsCanonicalFixturePair(t *testing.T) {
	receipt, err := ProduceFiles(Input{CaseID: "equivalent", SubjectSHA: candidateSHA, BeforePath: fixture("before.gooo"), AfterPath: fixture("equivalent-after.gooo")})
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Classification != ClassPreserved || receipt.RawDecision != RawChanged || receipt.SemanticDecision != SemanticPreserved || receipt.StructuralDelta.Status != "KNOWN" {
		t.Fatalf("receipt=%+v", receipt)
	}
	if receipt.StructuralDelta.AddedNodes != nil || receipt.SemanticClaimDelta.Changed != nil {
		t.Fatalf("presentation change was treated as semantic: %+v", receipt)
	}
	if len(receipt.ClaimTransitions) != 1 || receipt.ClaimTransitions[0].ToStatus != StatusDischarged {
		t.Fatalf("transitions=%+v", receipt.ClaimTransitions)
	}
}

func TestProducerRecordsSemanticClaimRefutation(t *testing.T) {
	receipt, err := ProduceFiles(Input{CaseID: "semantic-change", SubjectSHA: candidateSHA, BeforePath: fixture("before.gooo"), AfterPath: fixture("semantic-after.gooo")})
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Classification != ClassChanged || receipt.SemanticDecision != SemanticChanged || len(receipt.SemanticClaimDelta.Changed) != 1 {
		t.Fatalf("receipt=%+v", receipt)
	}
	if receipt.ClaimTransitions[0].ToStatus != StatusRefuted {
		t.Fatalf("transitions=%+v", receipt.ClaimTransitions)
	}
}
