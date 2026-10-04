package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func writeOriginRecordModel(t *testing.T, variant string) string {
	t.Helper()
	name := writeSharedRecordContractModel(t, variant)
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var meta jointdecision.Metadata
	if err = json.Unmarshal(raw, &meta); err != nil {
		t.Fatal(err)
	}
	// Synthetic loader/prediction fixture, not trained origin weights.
	meta.Feature, meta.MaxBytes = jointdecision.RecordOriginSharedFeatureVersion, jointdecision.RecordOriginInputMaxBytes
	raw, _ = json.Marshal(meta)
	if err = os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestRecordOriginProjectionSeparatesReceiverTimesWithoutCases(t *testing.T) {
	body := "let copy = input0\nlet saved = copy\nlet old = copy.state\ncopy.title = \"draft\"\ncopy.state = \"wait\"\ncopy.reason = old\nreturn Candidate{title: copy.title, state: copy.state, reason: copy.reason + saved.reason + old}"
	source := flowSource(t, body)
	other := []byte(strings.Replace(string(source), `copy.title + \":\" + copy.state`, `saved.title + \":\" + saved.state`, 1))
	var arrays [2][768]float32
	for i, src := range [][]byte{source, other} {
		r, err := ExportRecordAssemblyContextWithFlow(context.Background(), "r.gooo", src, "Select", false, jointdecision.RecordOriginSharedFeatureVersion)
		if err != nil || r.Context.Status != "ENCODED" || r.Context.ValueFlowSHA256 == "" ||
			jointdecision.FeaturesIntoRecordOriginThree(r.Context.Text, &arrays[i]) != nil {
			t.Fatal("origin projection failed", r.Context, err)
		}
	}
	if arrays[0] == arrays[1] {
		t.Fatal("old receiver-time collision persists in new model input")
	}
	base, err := ExportRecordAssemblyContextWithFeature(context.Background(), "r.gooo", source, "Select", false, jointdecision.RecordOriginSharedFeatureVersion)
	if err != nil || base.ValueFlow != nil || base.ModelPredictions != 0 || base.CandidateTests != 0 {
		t.Fatal("context export executed or implicitly exported complete flow", err)
	}
	for _, changed := range [][]byte{
		[]byte(strings.ReplaceAll(string(source), "copy", "working")),
		[]byte(strings.Replace(string(source), `한글`, `변경`, 1)),
		[]byte(strings.Replace(string(source), "let saved", "\nlet saved", 1)),
	} {
		r, e := ExportRecordAssemblyContextWithFeature(context.Background(), "r.gooo", changed, "Select", false, jointdecision.RecordOriginSharedFeatureVersion)
		var next [768]float32
		if e != nil || jointdecision.FeaturesIntoRecordOriginThree(r.Context.Text, &next) != nil || next != arrays[0] {
			t.Fatal("name/format/case changes affected origin feature values", e)
		}
	}
}

func TestRecordOriginModelSingleCallFiniteContinuationAndReplay(t *testing.T) {
	source := recordUpdatesFixture(t)
	ctx := context.Background()
	for _, variant := range []string{"fp32", "ptq_ternary", "qat_ternary"} {
		t.Run(variant, func(t *testing.T) {
			g, err := NewTypedPathGenerator(writeOriginRecordModel(t, variant))
			if err != nil {
				t.Fatal(err)
			}
			result, err := g.GenerateSourceAssembly(ctx, "r.gooo", source, "Select")
			if err != nil {
				t.Fatal(err)
			}
			r := result.Report.RecordAssembly
			if r.ModelCalls != 1 || r.Context.Status != "ENCODED" || r.Context.FeatureVersion != jointdecision.RecordOriginSharedFeatureVersion ||
				r.Passed != 5 || r.FieldsPassed != 15 || result.Report.CompletenessReceipt.Scope["decision_mode"] != "local_source_origin_prediction_then_finite_tdd" {
				t.Fatal("one model prediction/finite continuation differs", r)
			}
			if realized, e := RealizeSourceAssembly(ctx, "r.gooo", source, result); e != nil || realized.ModelCalls != 0 {
				t.Fatal("origin context replay predicted again", e)
			}
			var wait sync.WaitGroup
			for range 4 {
				wait.Go(func() {
					next, e := g.GenerateSourceAssembly(ctx, "r.gooo", source, "Select")
					if e != nil || next.Source != result.Source || !reflect.DeepEqual(next.Report.RecordAssembly.Context, r.Context) {
						t.Error("concurrent generation differs", e)
					}
				})
			}
			wait.Wait()
		})
	}
}

func TestRecordOriginUnresolvedGraphContinuesDeterministically(t *testing.T) {
	body := "let copy = input0\ncopy.title = \"draft\"\ncopy.state = \"wait\"\ncopy.reason = copy.state\n" +
		strings.Repeat("copy.reason = \"x\"\n", 260) + "return copy"
	source := flowSource(t, body)
	g, err := NewTypedPathGenerator(writeOriginRecordModel(t, "qat_ternary"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "r.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.ModelCalls != 0 || r.Prediction != nil || r.Context.Status != "DECLINED_TO_DETERMINISTIC" ||
		r.Context.Reason != "SOURCE_VALUE_FLOW_UNRESOLVED:FLOW_NODE_BOUND" || r.Ranking[0] != 0 {
		t.Fatal("unresolved flow guessed or called model", r)
	}
	if _, err = RealizeSourceAssembly(context.Background(), "r.gooo", source, result); err != nil {
		t.Fatal("deterministic unresolved replay failed", err)
	}
}

func TestRecordOriginModelScalarBodyDeclinesWithoutPrediction(t *testing.T) {
	source, document := threeNativeFixture(t)
	g, err := NewTypedPathGenerator(writeOriginRecordModel(t, "fp32"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.Generate(context.Background(), "scalar.gooo", source, "ThreeTest", document, TypedPathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodyPaths
	if r.ModelContext.Reason != "FIELD_MODEL_REQUIRES_RECORD_BODY" || r.ModelContext.Status != "DECLINED_TO_DETERMINISTIC" ||
		r.Search.Selection.ModelCalls != 0 || r.FunctionalCompleteness != 100 {
		t.Fatal("origin model scalar continuation differs", r)
	}
}
