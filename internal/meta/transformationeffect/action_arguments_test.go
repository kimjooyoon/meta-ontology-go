package transformationeffect

import (
	"slices"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestSplitGoEvidenceFlagIsApplyOnly(t *testing.T) {
	action := generation.Action{Operation: sourcepolicy.OperationSplitGo, Subject: "fixture.go"}
	apply := actionArguments("scripts/source-splitter", "/sandbox", Options{}, generation.Plan{}, action, false)
	check := actionArguments("scripts/source-splitter", "/sandbox", Options{}, generation.Plan{}, action, true)
	if !slices.Contains(apply, "-evidence-json") {
		t.Fatal("SplitGo apply must request raw evidence")
	}
	if slices.Contains(check, "-evidence-json") || !slices.Contains(check, "-check") {
		t.Fatal("SplitGo preflight must remain a read-only check")
	}
}
