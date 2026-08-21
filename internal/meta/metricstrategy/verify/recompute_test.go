package metricstrategyverify

import (
	"testing"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func TestReplayCandidatesAndSelectionAreIndependent(t *testing.T) {
	bindings := []strategy.Binding{
		{IndicatorID: "f", Family: "FOUNDATION", MetaOperation: "axiom", Status: "SATISFIED", EvidenceDigest: "sha256:f"},
		{IndicatorID: "c", Family: "COHERENCE", MetaOperation: "cohere", Status: "SATISFIED", EvidenceDigest: "sha256:c"},
		{IndicatorID: "r", Family: "REGRESSION", MetaOperation: "regress", Status: "SATISFIED", EvidenceDigest: "sha256:r"},
	}
	candidates, err := replayCandidates(bindings)
	if err != nil {
		t.Fatal(err)
	}
	selection := replaySelection(candidates, []metric.Projection{{Residual: 0, Status: "SATISFIED"}})
	if len(candidates) != 3 || selection.ProofChoice != "REGRESSION" || selection.Decision != "HOLD_FIXED_POINT" {
		t.Fatalf("independent strategy replay diverged: %+v %+v", candidates, selection)
	}
}

