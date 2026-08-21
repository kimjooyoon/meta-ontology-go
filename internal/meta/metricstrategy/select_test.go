package metricstrategy

import (
	"testing"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
)

func TestChooseTerminatesAtVerifiedFixedPoint(t *testing.T) {
	candidates := []Candidate{{ProofChoice: "FOUNDATION", Admissible: true}, {ProofChoice: "COHERENCE", Admissible: true}, {ProofChoice: "REGRESSION", Admissible: true, EvidenceDigest: "sha256:r", MetaOperations: []string{"replay-counterfactual"}}}
	selection := choose(candidates, []metric.Projection{{Residual: 0, Status: "SATISFIED"}}, true)
	if selection.ProofChoice != "REGRESSION" || selection.Decision != "HOLD_FIXED_POINT" || selection.MetaOperation != "terminate-at-fixed-point" {
		t.Fatalf("unexpected fixed-point selection: %+v", selection)
	}
}

func TestChooseRepairsFirstCanonicalFailure(t *testing.T) {
	candidates := []Candidate{{ProofChoice: "FOUNDATION", UnsatisfiedCount: 1, MetaOperations: []string{"bind-exact-source-metrics"}}, {ProofChoice: "COHERENCE"}, {ProofChoice: "REGRESSION"}}
	selection := choose(candidates, nil, false)
	if selection.ProofChoice != "FOUNDATION" || selection.Decision != "REPAIR" || selection.MetaOperation != "bind-exact-source-metrics" {
		t.Fatalf("unexpected repair selection: %+v", selection)
	}
}

