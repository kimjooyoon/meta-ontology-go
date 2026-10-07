package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestWorkspaceInputSeparationCombinesFillAndSearch(t *testing.T) {
	manifest, suite := workspaceInputsFixture(t)
	source := &manifest.Packages[0].Sources[0].Content
	*source = strings.Replace(*source, "output = 0", "output = __GOOO_BODY_HOLE_floor__", 1)
	*source = strings.Replace(*source, "return output`", "return output` assembling {\n"+
		`search hole "floor" grammar "integer-offset-constant/v1" intent "Return zero for negative integers." max_candidates "16"
case "-2" -> "0"
case "-1" -> "0"
case "0" -> "0"
case "2" -> "2"
attempts "8"
}`, 1)
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	checkWorkspaceInputCounts(t, result.Runtime.InputSeparation, "pass")
	if result.Composition.Steps[0].Generation.Report.BodySearch == nil {
		t.Fatal("fixture did not retain the source search")
	}
}

func TestWorkspaceInputSeparationPreservesRecordTuples(t *testing.T) {
	source, err := os.ReadFile("../../../examples/body-codegen/source-ir-fill-record.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "example/records", Activity: "ReviewCandidate"},
		Packages: []packageruntime.PackageSpec{{Path: "example/records", Name: "recordfill",
			Sources: []packageruntime.Source{{Filename: "record.gooo", Content: string(source)}}}}}
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, id := range []string{"1", "2", "9007199254740992", "9007199254740993"} {
		state, decision := "ready", "accepted"
		if id == "2" {
			state, decision = "queued", "rejected"
		}
		input := json.RawMessage(`{"state":"` + state + `","candidate_id":` + id + `}`)
		expected := json.RawMessage(`{"decision":"` + decision + `","candidate_id":` + id + `}`)
		suite.Cases = append(suite.Cases, bodyexecution.CompositionCase{
			Inputs:   map[string]json.RawMessage{"example/records:ReviewCandidate": input},
			Expected: map[string]json.RawMessage{"example/records:ReviewCandidate": expected}})
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	s := result.Runtime.InputSeparation
	if s.Status != "PASS" || s.UniqueInputs != 4 || s.OverlappingInputs != 2 || s.DisjointInputs != 2 || s.DisjointCasesPassed != 2 || len(s.EarlierStages) != 1 {
		t.Fatal("record field order or large integers changed the measured input identities", s)
	}
}

func TestWorkspaceInputSeparationKeepsMissingFillEvidenceUnknown(t *testing.T) {
	manifest, suite := workspaceInputsFixture(t)
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	translated, err := translateCases(result.Program, suite)
	if err != nil {
		t.Fatal(err)
	}
	result.BodyFills[0].Generation.Report.BodyFill.SelectedCaseResults = nil
	source := []byte(result.BodyFills[0].Generation.GoooSource)
	got := measureWorkspaceInputs(context.Background(), source, result, translated)
	if got.Status != "UNKNOWN" || got.Reason != "EARLIER_CONSTRUCTION_INPUTS_UNAVAILABLE" || got.DisjointInputs != 0 {
		t.Fatal("missing fill input tuples received new-input coverage", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := measureWorkspaceInputs(ctx, source, result, translated); got.Status != "UNKNOWN" {
		t.Fatal("canceled measurement produced known coverage", got)
	}
}
