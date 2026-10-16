package bodycodegen

import (
	"context"
	"testing"
)

func TestFixedRecordSkipsRankingFresh1757(t *testing.T) {
	// Metadata-only sentinel: no weights are loaded and no inference is invoked.
	generator := &TypedPathGenerator{}
	generator.info.Loaded = true
	receipt := &RecordAssemblyReceipt{Ranking: []uint16{0}}
	if err := generator.rankRecordAssemblyContext(context.Background(), recordAssemblyPlan{}, receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.ModelRequested || receipt.ModelCalls != 0 || receipt.Model != nil || receipt.Context != nil ||
		len(receipt.Ranking) != 1 || receipt.Ranking[0] != 0 {
		t.Fatal("fixed source body manufactured a prediction or extra candidate")
	}
}
