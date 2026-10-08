package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func graphSourceFeatures(t *testing.T, source []byte, activity string) [jointdecision.ThreeFeatureDim]float32 {
	t.Helper()
	exported, err := ExportRecordAssemblyContextWithFlow(context.Background(), "graph.gooo", source, activity, false,
		jointdecision.RecordGraphSharedFeatureVersion)
	if err != nil || exported.Context.Status != "ENCODED" || exported.ValueFlow.Status != "RESOLVED" ||
		exported.ModelPredictions != 0 || exported.CandidateTests != 0 {
		t.Fatal("source graph context", exported.Context, err)
	}
	var features [jointdecision.ThreeFeatureDim]float32
	if err := jointdecision.FeaturesIntoRecordGraphThree(exported.Context.Text, &features); err != nil {
		t.Fatal(err)
	}
	return features
}

func TestRecordGraphFilenameChoiceOrdersProduceEightInputs(t *testing.T) {
	raw, err := os.ReadFile("../../examples/text-operations/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[[jointdecision.ThreeFeatureDim]float32]int{}
	for order := range 8 {
		source := string(raw)
		for bit, pair := range [][4]string{
			{"source: suffix || visible", "source: suffix && visible", "alternative \"suffix && visible\"", "alternative \"suffix || visible\""},
			{"stem: input", "stem: stem", "alternative \"stem\"", "alternative \"input\""},
			{"bytes: 0", "bytes: bytes", "alternative \"bytes\"", "alternative \"0\""},
		} {
			if order&(1<<bit) != 0 {
				source = strings.Replace(source, pair[0], pair[1], 1)
				source = strings.Replace(source, pair[2], pair[3], 1)
			}
		}
		features := graphSourceFeatures(t, []byte(source), "Classify")
		if prior, exists := seen[features]; exists {
			t.Fatalf("source arrangements %d and %d collided", prior, order)
		}
		seen[features] = order
	}
}

func TestRecordGraphSourceLiteralViewsAndNames(t *testing.T) {
	source := helperFlowSource("return \"a\" + input", "input", "Keep(input)")
	want := graphSourceFeatures(t, source, "Select")
	renamed := []byte(strings.ReplaceAll(string(source), "Keep(", "RenamedHelper("))
	if graphSourceFeatures(t, renamed, "Select") != want {
		t.Fatal("helper display name changed source value features")
	}
	changed := []byte(strings.Replace(string(source), "\"a\" + input", "\"b\" + input", 1))
	if graphSourceFeatures(t, changed, "Select") == want {
		t.Fatal("alternative-only helper literal disappeared")
	}
	literal := helperFlowSource("return input", "\"x\"", "\"y\"")
	if graphSourceFeatures(t, literal, "Select") == graphSourceFeatures(t,
		[]byte(strings.Replace(string(literal), `alternative "\"y\""`, `alternative "\"z\""`, 1)), "Select") {
		t.Fatal("alternative expression literal disappeared")
	}
	base := recordAssemblyFixture(t)
	original := graphSourceFeatures(t, base, "Select")
	otherCases := []byte(strings.Replace(string(base), "attempts \"8\"", "attempts \"1\"", 1))
	if graphSourceFeatures(t, otherCases, "Select") != original {
		t.Fatal("candidate attempt budget leaked into source features")
	}
	changedCase := []byte(strings.Replace(string(base), "한글", "다른 사례", 1))
	if graphSourceFeatures(t, changedCase, "Select") != original {
		t.Fatal("finite input case leaked into source features")
	}
}

func writeGraphContractModel(t *testing.T, variant string) string {
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
	meta.Feature, meta.MaxBytes = jointdecision.RecordGraphSharedFeatureVersion, jointdecision.RecordGraphInputMaxBytes
	raw, _ = json.Marshal(meta)
	if err = os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestRecordGraphModelConstructionAndReplay(t *testing.T) {
	ctx := context.Background()
	source := recordAssemblyFixture(t)
	for _, variant := range []string{"fp32", "ptq_ternary", "qat_ternary"} {
		t.Run(variant, func(t *testing.T) {
			g, err := NewTypedPathGenerator(writeGraphContractModel(t, variant))
			if err != nil {
				t.Fatal(err)
			}
			r, err := g.GenerateSourceAssembly(ctx, "graph.gooo", source, "Select")
			if err != nil {
				t.Fatal(err)
			}
			assembly := r.Report.RecordAssembly
			if assembly.ModelCalls != 1 || assembly.Context.FeatureVersion != jointdecision.RecordGraphSharedFeatureVersion ||
				assembly.FieldsPassed != 15 || assembly.SelectedMask != 7 ||
				r.Report.CompletenessReceipt.Scope["decision_mode"] != "local_source_graph_prediction_then_finite_tdd" {
				t.Fatal("graph contract did not rank then construct", assembly)
			}
			replay, err := RealizeSourceAssembly(ctx, "graph.gooo", source, r)
			if err != nil || replay.ModelCalls != 0 {
				t.Fatal("graph contract replay made a prediction", err)
			}
			large := []byte(strings.Replace(string(source), `"draft"`, `"`+strings.Repeat("한", 350)+`"`, 1))
			declined, err := g.GenerateSourceAssembly(ctx, "large.gooo", large, "Select")
			if err != nil || declined.Report.RecordAssembly.Context.Status != "DECLINED_TO_DETERMINISTIC" ||
				declined.Report.RecordAssembly.ModelCalls != 0 || declined.Report.RecordAssembly.FieldsPassed != 15 {
				t.Fatal("oversize graph context did not use deterministic candidates", err)
			}
		})
	}
}
