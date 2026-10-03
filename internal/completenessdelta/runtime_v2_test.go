package completenessdelta

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestOwnedRuntimeComparisonsKeepParentScopeAndCurrentResourceDenominators(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	rawPlan, err := os.ReadFile("../../examples/body-codegen/typed-path-compound-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := bodyexecution.DecodePlan(rawPlan)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := bodycodegen.GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", doc, "")
	if err != nil {
		t.Fatal(err)
	}
	generation, _ := json.Marshal(prior)
	prior, parent, err := bodyexecution.DecodeGeneration(generation)
	if err != nil {
		t.Fatal(err)
	}
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	owner := bodyexecution.NewExecutor()
	defer owner.Close()
	cases := []pathplan.TestCase{{Input: -4, Expected: -21}, {Input: 1, Expected: 10}}
	var observations [][]byte
	for call := range 3 {
		if call == 2 {
			cases[0].Expected++
		}
		result, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, cases, tool)
		if err != nil {
			t.Fatal(err)
		}
		wire, _ := json.Marshal(result)
		copy := bytes.Clone(wire)
		comparison, err := Compare(generation, wire)
		if err != nil || comparison.Relation != "PARENT_RUNTIME_CONTINUATION" || !bytes.Equal(wire, copy) {
			t.Fatal("owned runtime did not preserve exact parent continuation", err)
		}
		for _, axis := range comparison.Dimensions {
			if axis.CountMagnitude != nil {
				t.Fatal("profile transition created a numeric improvement", axis)
			}
		}
		observations = append(observations, wire)
		if call == 1 {
			bare, _ := json.Marshal(result.CompletenessReceipt)
			if comparison, err := Compare(bare, bare); err != nil || comparison.Relation != "SAME_MEASUREMENT_SCOPE" {
				t.Fatal("valid bare owned receipt was not comparable", err)
			}
			delete(result.CompletenessReceipt.Scope, "owned_artifact")
			missing, _ := json.Marshal(result.CompletenessReceipt)
			if comparison, err := Compare(bare, missing); err != nil || comparison.Relation != "INCOMPARABLE_SCOPE" {
				t.Fatal("bare owned receipt lost its required contract", err)
			}
		}
	}
	comparison, err := Compare(observations[0], observations[1])
	if err != nil || comparison.Relation != "SAME_MEASUREMENT_SCOPE" {
		t.Fatal("first build and current reuse changed finite measurement scope", err)
	}
	finiteSeen, resourceSeen := false, false
	for _, axis := range comparison.Dimensions {
		switch axis.ID {
		case "runtime_finite_accuracy":
			finiteSeen = axis.CountMagnitude != nil && *axis.CountMagnitude == 0
		case "runtime_child_resources":
			resourceSeen = axis.CountMagnitude == nil && axis.Before.Denominator == 4 &&
				axis.After.Denominator == 2 && axis.Before.Unit != axis.After.Unit
		}
	}
	if !finiteSeen || !resourceSeen {
		t.Fatal("finite counts or changed resource denominator misrepresented", finiteSeen, resourceSeen)
	}
	comparison, err = Compare(observations[1], observations[2])
	if err != nil || comparison.Relation != "INCOMPARABLE_SCOPE" {
		t.Fatal("changed runtime expectations compared as same measurement", err)
	}
}
