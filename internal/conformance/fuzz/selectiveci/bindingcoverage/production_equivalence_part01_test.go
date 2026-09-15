package bindingcoverage

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/bindingcoverage"
	"testing"
)

const expectedProductionEquivalenceReceipt = "sha256:ab34297396618634dc46747a55ec5b386149df285308094005c84fd6b61eeb2c"

func TestProductionBindingCoverageEquivalence(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) != 20 {
		t.Fatalf("case count=%d", len(corpus.Cases))
	}
	seen := make(map[string]bool, len(corpus.Cases))
	receipts := make([]productionReceipt, 0, len(corpus.Cases))
	byName := make(map[string]productionReceipt, len(corpus.Cases))
	mismatchCount := 0
	for _, row := range corpus.Cases {
		if seen[row.Name] {
			mismatchCount++
			t.Errorf("PRODUCTION_EQUIVALENCE_COUNTEREXAMPLE_NO_GO case=%s duplicate corpus name", row.Name)
			continue
		}
		seen[row.Name] = true
		oracle := Evaluate(row.Input)
		translated, translateErr := translateInput(row.Input)
		if translateErr != nil {
			mismatchCount++
			t.Errorf("PRODUCTION_EQUIVALENCE_COUNTEREXAMPLE_NO_GO case=%s translator UNKNOWN: %v", row.Name, translateErr)
			continue
		}
		productionOutput := production.Observe(translated)
		receipt, compareErr := compareProductionCase(row, oracle, translated, productionOutput)
		if compareErr != nil {
			mismatchCount++
			t.Errorf("PRODUCTION_EQUIVALENCE_COUNTEREXAMPLE_NO_GO case=%s: %v\noracle=%+v\nproduction=%+v", row.Name, compareErr, oracle.Vector, productionOutput)
		}
		receipts = append(receipts, receipt)
		byName[row.Name] = receipt
	}
	if len(seen) != len(corpus.Cases) {
		mismatchCount++
		t.Errorf("corpus consumption=%d/%d", len(seen), len(corpus.Cases))
	}
	if err := comparePermutation(byName); err != nil {
		mismatchCount++
		t.Errorf("PRODUCTION_EQUIVALENCE_COUNTEREXAMPLE_NO_GO permutation: %v", err)
	}
	digest, err := receiptDigest(receipts)
	if err != nil {
		t.Fatal(err)
	}
	if digest != expectedProductionEquivalenceReceipt {
		t.Fatalf("paired receipt digest=%s want=%s", digest, expectedProductionEquivalenceReceipt)
	}
	t.Logf("binding coverage paired receipt digest=%s cases=%d mismatch_count=%d", digest, len(corpus.Cases), mismatchCount)
	if mismatchCount != 0 {
		t.Fatalf("production equivalence mismatch_count=%d", mismatchCount)
	}
}
