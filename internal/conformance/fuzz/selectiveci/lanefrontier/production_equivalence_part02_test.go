package lanefrontier

import (
	"reflect"
	"testing"
)

func TestProductionLaneFrontierEquivalence(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) != 16 {
		t.Fatalf("lane frontier corpus cases=%d, want 16", len(corpus.Cases))
	}
	seen := map[string]bool{}
	receipts := make([]pairedReceipt, 0, len(corpus.Cases))
	mismatches := 0
	for _, fixture := range corpus.Cases {
		if seen[fixture.Name] {
			t.Fatalf("duplicate lane frontier case %q", fixture.Name)
		}
		seen[fixture.Name] = true
		receipt, err := compareCase(fixture)
		if err != nil {
			mismatches++
			t.Errorf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO case=%s error=%v", fixture.Name, err)
			continue
		}
		receipts = append(receipts, receipt)
		if err := validatePartition(receipt.OracleVector); err != nil {
			mismatches++
			t.Errorf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO case=%s oracle=%s production=%s partition=%v", fixture.Name, vectorJSON(receipt.OracleVector), vectorJSON(receipt.ProductionVector), err)
		}
		if err := validatePartition(receipt.ProductionVector); err != nil {
			mismatches++
			t.Errorf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO case=%s oracle=%s production=%s partition=%v", fixture.Name, vectorJSON(receipt.OracleVector), vectorJSON(receipt.ProductionVector), err)
		}
		if !reflect.DeepEqual(receipt.OracleVector, receipt.ProductionVector) || !receipt.OraclePermutationEqual || !receipt.ProductionPermutationEqual {
			mismatches++
			t.Errorf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO case=%s oracle=%s production=%s", fixture.Name, vectorJSON(receipt.OracleVector), vectorJSON(receipt.ProductionVector))
		}
	}
	if len(seen) != len(corpus.Cases) || len(receipts)+mismatches < len(corpus.Cases) {
		t.Fatalf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO consumed=%d receipts=%d cases=%d", len(seen), len(receipts), len(corpus.Cases))
	}
	digest := pairedReceiptDigest(receipts)
	t.Logf("lane frontier paired receipt digest=%s mismatch_count=%d", digest, mismatches)
	if mismatches != 0 {
		t.Fatalf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO mismatch_count=%d paired_receipt_digest=%s", mismatches, digest)
	}
}
