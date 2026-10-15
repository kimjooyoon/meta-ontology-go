package bodycodegen

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const pathHoldoutFixture = `package withheld
namespace withheld
entity Integer id "withheld://integer"
activity Fee(Integer) -> Integer id "withheld://fee" computes "return input - 19" assembling {
 choice "direction" operand_order at "0" intent "Choose the subtraction direction."
 case "19" -> "0"
 holdout_case "23" -> "-4"
 attempts "2"
}`

func TestPathHoldoutDoesNotSelectOrRankCandidates(t *testing.T) {
	ctx := context.Background()
	model := writePathContextContractModel(t)
	var first Result
	var firstExport TypedPathContextExport
	for i, source := range []string{pathHoldoutFixture, strings.Replace(pathHoldoutFixture, `"23" -> "-4"`, `"23" -> "4"`, 1)} {
		doc, err := DecodeSourcePathDocument(ctx, "holdout.gooo", []byte(source), "Fee", nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(doc.TestCases) != 1 || doc.TestCases[0].Input != 19 {
			t.Fatal("holdout entered selection cases", doc.TestCases)
		}
		exported, err := ExportTypedPathContext(ctx, "holdout.gooo", []byte(source), "Fee", doc)
		if err != nil {
			t.Fatal(err)
		}
		r, err := GenerateWithTypedPaths(ctx, "holdout.gooo", []byte(source), "Fee", doc, model)
		if err != nil {
			t.Fatal(err)
		}
		p := r.Report.BodyPaths
		if p.Holdout == nil || p.Holdout.Total != 1 || p.Holdout.AccuracyPercent == nil || len(p.NativeCases) != 1 {
			t.Fatal("selection and holdout observations mixed", p)
		}
		if err := VerifyTypedPathProjection(ctx, "holdout.gooo", []byte(source), doc, r); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first, firstExport = r, exported
			continue
		}
		if r.Source != first.Source || !reflect.DeepEqual(p.Search.Selection.Choices, first.Report.BodyPaths.Search.Selection.Choices) ||
			p.Search.Selection.ModelCalls != first.Report.BodyPaths.Search.Selection.ModelCalls ||
			!reflect.DeepEqual(exported.Inputs, firstExport.Inputs) || exported.DocumentSHA256 != firstExport.DocumentSHA256 ||
			exported.OriginalSourceSHA256 == firstExport.OriginalSourceSHA256 || p.Holdout.Passed == first.Report.BodyPaths.Holdout.Passed {
			t.Fatal("holdout answer affected ranking or failed to change only evaluation", p.Holdout)
		}
	}
}

func TestPathHoldoutMismatchAndFaultRemainPostSelection(t *testing.T) {
	ctx := context.Background()
	for _, source := range []string{pathHoldoutFixture,
		strings.Replace(strings.Replace(pathHoldoutFixture, "return input - 19", "return 19 / input", 1), `case "19" -> "0"`, `case "19" -> "1"`, 1)} {
		fault := strings.Contains(source, "return 19 /")
		if fault {
			source = strings.Replace(source, `holdout_case "23" -> "-4"`, `holdout_case "0" -> "0"`, 1)
		}
		doc, err := DecodeSourcePathDocument(ctx, "post.gooo", []byte(source), "Fee", nil)
		if err != nil {
			t.Fatal(err)
		}
		r, err := GenerateWithTypedPaths(ctx, "post.gooo", []byte(source), "Fee", doc, "")
		if err != nil {
			t.Fatal(err)
		}
		p := r.Report.BodyPaths
		if p.FunctionalCompleteness != 100 || p.Holdout == nil || p.Holdout.Passed != 0 || len(p.Search.Attempts) != 1 {
			t.Fatal("holdout caused candidate reconsideration", p)
		}
		if fault != (p.Holdout.Error != "") || fault != (p.Holdout.AccuracyPercent == nil) {
			t.Fatal("fault was treated as a measured zero score", p.Holdout)
		}
		if err := VerifyTypedPathProjection(ctx, "post.gooo", []byte(source), doc, r); err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(r)
		var changed Result
		if err := json.Unmarshal(raw, &changed); err != nil {
			t.Fatal(err)
		}
		changed.Report.BodyPaths.Holdout.SuiteSHA256 = "changed"
		if err := VerifyTypedPathProjection(ctx, "post.gooo", []byte(source), doc, changed); err == nil {
			t.Fatal("changed holdout replayed")
		}
	}
}

func TestPathHoldoutExternalObservationCannotConsumeReservedInput(t *testing.T) {
	ctx := context.Background()
	source := []byte(pathHoldoutFixture)
	doc, err := DecodeSourcePathDocument(ctx, "external.gooo", source, "Fee", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = GenerateWithTypedPathOptions(ctx, "external.gooo", source, "Fee", doc, "missing-model.json", TypedPathOptions{
		Observation: &PathObservationOptions{Inputs: []int64{23}, MaxCandidates: 2, MaxRounds: 1}})
	if err == nil || !strings.Contains(err.Error(), "reserved by holdout_case") {
		t.Fatal(err)
	}
	if err := verifySourcePathObservation(ctx, "external.gooo", source, "Fee", &PathObservationReceipt{
		Options: PathObservationOptions{Inputs: []int64{23}, MaxCandidates: 2, MaxRounds: 1}}); err == nil {
		t.Fatal("saved observation consumed holdout")
	}
}

func TestPathHoldoutExcludedFromV4SourceAndOutputInputs(t *testing.T) {
	source := `package heldbranch
namespace heldbranch
entity Integer id "heldbranch://integer"
activity Choose(Integer) -> Integer computes "if input < 73 { return input + 5 } else { return input - 7 }" assembling {
 choice "comparison" operand_order at "0" intent "Choose comparison operands."
 choice "branch" branch_layout at "0" intent "Choose the branch layout."
 case "64" -> "69"
 case "84" -> "77"
 holdout_case "9007199254740993" -> "9007199254740986"
 attempts "4"
}`
	var baseline []ExportedPathInput
	var documentSHA string
	for i, text := range []string{source, strings.Replace(source, `"9007199254740993" -> "9007199254740986"`, `"9007199254740993" -> "-9"`, 1)} {
		doc, err := DecodeSourcePathDocument(context.Background(), "heldbranch.gooo", []byte(text), "Choose", nil)
		if err != nil {
			t.Fatal(err)
		}
		prepared, err := doc.Prepare()
		if err != nil {
			t.Fatal(err)
		}
		rows, err := exportedSourceRelativeV4Inputs(doc, prepared)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(doc)
		if len(doc.TestCases) != 2 || len(rows) != 2 {
			t.Fatal("held output entered choice inputs")
		}
		if i == 0 {
			baseline = rows
			documentSHA = digest(raw)
		} else if !reflect.DeepEqual(rows, baseline) || digest(raw) != documentSHA {
			t.Fatal("holdout changed source-relative model input")
		}
	}
}
