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
	if len(receipt.ClaimTransitions) != 3 || receipt.ClaimTransitions[0].Kind != ClaimKindBounded || receipt.ClaimTransitions[0].ToStatus != StatusDischarged {
		t.Fatalf("transitions=%+v", receipt.ClaimTransitions)
	}
	for _, transition := range receipt.ClaimTransitions[1:] {
		if transition.Kind != ClaimKindPreserve || transition.ToStatus != StatusDischarged {
			t.Fatalf("presentation preservation=%+v", receipt.ClaimTransitions)
		}
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
	if len(receipt.ClaimTransitions) != 4 || receipt.ClaimTransitions[0].ToStatus != StatusRefuted {
		t.Fatalf("transitions=%+v", receipt.ClaimTransitions)
	}
	preserved, refuted, observed := 0, 0, 0
	for _, transition := range receipt.ClaimTransitions[1:] {
		if transition.Kind == ClaimKindPreserve && transition.ToStatus == StatusDischarged {
			preserved++
		}
		if transition.Kind == ClaimKindPreserve && transition.ToStatus == StatusRefuted {
			refuted++
		}
		if transition.Kind == ClaimKindObject && transition.ToStatus == StatusDischarged {
			observed++
		}
	}
	if preserved != 1 || refuted != 1 || observed != 1 {
		t.Fatalf("semantic ledger transitions=%+v", receipt.ClaimTransitions)
	}
}
