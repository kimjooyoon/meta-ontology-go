package operationconformance

import (
	"os"
	"testing"
)

func TestBehavioralCorpusDenominator(t *testing.T) {
	raw, err := os.ReadFile("testdata/split-go-behavior-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	corpus, err := ParseBehavioralCorpus(raw)
	if err != nil {
		t.Fatal(err)
	}
	if corpus.CaseCount != 8 || len(corpus.Cases) != 8 {
		t.Fatalf("corpus=%+v", corpus)
	}
}
