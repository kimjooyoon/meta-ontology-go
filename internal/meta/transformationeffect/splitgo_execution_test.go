package transformationeffect

import (
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestSplitGoUnknownEvidenceLowersProductionResolution(t *testing.T) {
	action := generation.Action{
		Operation:            sourcepolicy.OperationSplitGo,
		RequiredIndicatorIDs: splitGoTestIndicatorIDs,
		ProofChoice:          generation.ProofFoundation,
	}
	receipts, artifact, evidence, err := operationObservations(filepath.Clean("../../.."), action, []byte(`{}`), "unused")
	if err != nil {
		t.Fatal(err)
	}
	if artifact == nil || artifact.Resolution != "LOWER_RESOLUTION" {
		t.Fatalf("artifact=%v, want lower-resolution SplitGo evaluation", artifact)
	}
	if evidence != hashJSON(*artifact) || len(receipts) != len(splitGoTestIndicatorIDs) {
		t.Fatal("production receipts are not bound to the replayable artifact")
	}
	for _, receipt := range receipts {
		if splitGoReceiptField(receipt, "Verdict") != "UNKNOWN" {
			t.Fatal("unknown actor evidence must not become a passing receipt")
		}
	}
	if err := ValidateSplitGoEvaluation(*artifact); err != nil {
		t.Fatal(err)
	}
}
