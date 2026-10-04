package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func writeSharedRecordContractModel(t *testing.T, variant string) string {
	t.Helper()
	name := writeSharedThreeContractModel(t, variant)
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var meta jointdecision.Metadata
	if err = json.Unmarshal(raw, &meta); err != nil {
		t.Fatal(err)
	}
	meta.Feature, meta.Arithmetic = jointdecision.RecordSharedFeatureVersion, jointdecision.SeparateArithmeticVersion
	raw, err = json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestSharedRecordFieldContractFiniteContinuationAndRetainedConcurrency(t *testing.T) {
	source := recordAssemblyFixture(t)
	ctx := context.Background()
	exported, err := ExportRecordAssemblyContextWithFeature(ctx, "record.gooo", source, "Select", false, jointdecision.RecordSharedFeatureVersion)
	if err != nil || exported.ModelPredictions != 0 || exported.CandidateTests != 0 || exported.ExpandedPlan != nil {
		t.Fatal("shared source-only export", err)
	}
	dense, err := ExportRecordAssemblyContextWithFeature(ctx, "record.gooo", source, "Select", false, jointdecision.RecordFieldFeatureVersion)
	if err != nil || dense.Context.Text != exported.Context.Text || exported.Context.FeatureVersion != jointdecision.RecordSharedFeatureVersion {
		t.Fatal("shared projection changed complete source context", err)
	}
	for _, variant := range []string{"fp32", "ptq_ternary", "qat_ternary"} {
		t.Run(variant, func(t *testing.T) {
			name := writeSharedRecordContractModel(t, variant)
			g, err := NewTypedPathGenerator(name)
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{name, filepath.Join(filepath.Dir(name), "weights.bin")} {
				if err = os.Rename(file, file+".retained"); err != nil {
					t.Fatal(err)
				}
			}
			result, err := g.GenerateSourceAssembly(ctx, "record.gooo", source, "Select")
			if err != nil {
				t.Fatal(err)
			}
			r := result.Report.RecordAssembly
			if r.ModelCalls != 1 || r.PredictNS < 1 || r.Prediction == nil || r.Context.Text != exported.Context.Text ||
				r.SelectedMask != 7 || r.FieldsPassed != 15 || len(r.Attempts) != 8 || r.Attempts[0].Mask != 0 ||
				result.Report.CompletenessReceipt.Scope["decision_mode"] != "local_shared_field_prediction_then_finite_tdd" {
				t.Fatal("shared single prediction/finite continuation differs", r)
			}
			realized, err := RealizeSourceAssembly(ctx, "record.gooo", source, result)
			if err != nil || realized.ModelCalls != 0 {
				t.Fatal("shared deterministic reconstruction", err)
			}
			partialSource := []byte(strings.Replace(string(source), "attempts \"8\"", "attempts \"1\"", 1))
			partial, err := g.GenerateSourceAssembly(ctx, "partial.gooo", partialSource, "Select")
			if err != nil || partial.Report.RecordAssembly.Status != "PARTIAL_FINITE" || partial.Report.RecordAssembly.ModelCalls != 1 ||
				partial.Report.RecordAssembly.FieldsPassed != 6 || len(partial.Report.RecordAssembly.Attempts) != 1 {
				t.Fatal("partial score lost", err)
			}
			var wait sync.WaitGroup
			for range 4 {
				wait.Go(func() {
					next, e := g.GenerateSourceAssembly(ctx, "record.gooo", source, "Select")
					if e != nil {
						t.Error(e)
						return
					}
					if next.Source != result.Source || next.Report.RecordAssembly.ModelCalls != 1 {
						t.Error("shared request state differs")
					}
					if _, e = RealizeSourceAssembly(ctx, "record.gooo", source, next); e != nil {
						t.Error(e)
					}
				})
			}
			wait.Wait()
			large := []byte(strings.Replace(string(source), "Keep the original title.", strings.Repeat("한", 140), 1))
			declined, err := g.GenerateSourceAssembly(ctx, "large.gooo", large, "Select")
			if err != nil || declined.Report.RecordAssembly.Context.Status != "DECLINED_TO_DETERMINISTIC" ||
				declined.Report.RecordAssembly.ModelCalls != 0 || declined.Report.RecordAssembly.FieldsPassed != 15 {
				t.Fatal("whole shared context decline", err)
			}
			c := recordModelContext(r.Choices[:2], jointdecision.RecordSharedFeatureVersion)
			if c.Status != "DECLINED_TO_DETERMINISTIC" || len(c.Parts) != 2 {
				t.Fatal("shared arity decline lost parts")
			}
		})
	}
}

func TestSharedRecordModelScalarBodyContinuesWithoutPrediction(t *testing.T) {
	source, doc := threeNativeFixture(t)
	g, err := NewTypedPathGenerator(writeSharedRecordContractModel(t, "fp32"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.Generate(context.Background(), "scalar.gooo", source, "ThreeTest", doc, TypedPathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodyPaths
	if r.ModelContext.Status != "DECLINED_TO_DETERMINISTIC" || r.ModelContext.Reason != "FIELD_MODEL_REQUIRES_RECORD_BODY" ||
		r.Search.Selection.ModelCalls != 0 || r.FunctionalCompleteness != 100 {
		t.Fatal("record model scalar continuation", r)
	}
}
