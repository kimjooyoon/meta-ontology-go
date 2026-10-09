package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestTypedModelPreflightMatchesActualNativeRanking(t *testing.T) {
	for _, language := range []string{"ko", "en"} {
		for _, feature := range []string{"", decision.PositionedIntentFeatureVersion,
			decision.SplitContextIntentFeatureVersion, decision.SemanticContextIntentFeatureVersion} {
			t.Run(language+"/"+feature, func(t *testing.T) {
				source, doc := conditionalPathFixture(t, language)
				before, _ := json.Marshal(doc)
				model := writePathContextContractModelVersion(t, feature)
				out, err := ExportTypedPathModelContext(context.Background(), "f.gooo", source,
					"ConditionalAssign", doc, model, feature)
				if err != nil {
					t.Fatal(err)
				}
				assertTypedPreflightZeroWork(t, out)
				if out.ModelCompatibility.Status != "READY_FOR_RANKING" || out.ModelCompatibility.Model.FeatureVersion != feature ||
					out.ModelCompatibility.Model.ModelSchema != decision.PathMetadataSchema || out.CompleteModelInput != nil || len(out.Inputs) != 6 {
					t.Fatal("model contract differs", out)
				}
				generated, err := GenerateWithTypedPaths(context.Background(), "f.gooo", source, "ConditionalAssign", doc, model)
				if err != nil {
					t.Fatal(err)
				}
				p := generated.Report.BodyPaths
				if p.Search.Selection.ModelCalls != 6 {
					t.Fatal("native ranking did not exercise every exported input")
				}
				for i, input := range out.Inputs {
					if input.InputSHA != digest([]byte(input.Text)) || input.InputSHA != "sha256:"+p.Search.Selection.Receipts[i].IntentSHA256 {
						t.Fatal("inspection differs from consumed runtime bytes", input)
					}
				}
				if p.ModelContext != nil && !reflect.DeepEqual(p.ModelContext, out.Context) {
					t.Fatal("inspection used a different context preparation")
				}
				doc.TestCases = append([]pathplan.TestCase(nil), doc.TestCases...)
				doc.TestCases[0].Expected = 9007199254740993
				doc.Seed = "inspection-does-not-sample"
				changed, err := ExportTypedPathModelContext(context.Background(), "f.gooo", source,
					"ConditionalAssign", doc, model, feature)
				if err != nil || !reflect.DeepEqual(out.Inputs, changed.Inputs) || out.TestSuiteSHA256 == changed.TestSuiteSHA256 {
					t.Fatal("case expectation or seed influenced model input", err)
				}
				doc.TestCases[0].Expected = p.NativeCases[0].Expected
				doc.Seed = ""
				after, _ := json.Marshal(doc)
				if string(before) != string(after) {
					t.Fatal("inspection changed the caller document")
				}
			})
		}
	}
}

func assertTypedPreflightZeroWork(t *testing.T, out TypedPathContextExport) {
	t.Helper()
	if out.Schema != "gooo/compiler-path-model-input-export/v1" || out.ModelCompatibility == nil ||
		!out.ModelCompatibility.Model.Loaded || !out.SourceBinding.Equivalent || out.Context.MetadataSHA == "" ||
		out.ModelPredictions != 0 || out.CandidateTests != 0 || out.SelectedEmission || out.RepositoryWrites != 0 {
		t.Fatal("inspection identity or zero-work boundary differs", out)
	}
}

func TestTypedModelPreflightExportsCompleteJointRuntimeInput(t *testing.T) {
	for _, kind := range []string{"joint", "three", "shared_fp32", "shared_ternary", "shared_bag"} {
		t.Run(kind, func(t *testing.T) {
			source, doc := typedPathFixture(t)
			model := ""
			switch kind {
			case "joint":
				source, doc = jointNativeFixture(t)
				model = writeJointContractModel(t)
			case "three":
				model = writeThreeContractModel(t)
			case "shared_fp32", "shared_bag":
				model = writeSharedThreeContractModel(t, "fp32")
			case "shared_ternary":
				model = writeSharedThreeContractModel(t, "qat_ternary")
			}
			if kind == "shared_bag" {
				raw, _ := os.ReadFile(model)
				raw = []byte(strings.ReplaceAll(string(raw), jointdecision.ThreeFeatureVersion, jointdecision.ThreeBagFeatureVersion))
				if err := os.WriteFile(model, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			out, err := ExportTypedPathModelContext(context.Background(), "f.gooo", source, "Combined", doc, model, "")
			if err != nil {
				t.Fatal(err)
			}
			assertTypedPreflightZeroWork(t, out)
			generated, err := GenerateWithTypedPaths(context.Background(), "f.gooo", source, "Combined", doc, model)
			if err != nil {
				t.Fatal(err)
			}
			selection := generated.Report.BodyPaths.Search.Selection
			text := ""
			if selection.Three != nil {
				text = selection.Three.Input
			} else {
				text = selection.Joint.Input
			}
			if selection.ModelCalls != 1 || out.ModelCompatibility.Status != "READY_FOR_RANKING" ||
				out.CompleteModelInput == nil || out.CompleteModelInput.Text != text || out.CompleteModelInput.SHA256 != digest([]byte(text)) ||
				out.CompleteModelInput.Bytes != len(text) || !reflect.DeepEqual(out.Context, generated.Report.BodyPaths.ModelContext) {
				t.Fatal("joint preflight differs from complete native framing", out)
			}
		})
	}
}

func TestTypedModelPreflightRetainsNativeDeclinesAtomically(t *testing.T) {
	for _, reason := range []string{"THREE_DECISION_COUNT_UNSUPPORTED", "FIELD_MODEL_REQUIRES_RECORD_BODY",
		"COMBINED_CONTEXT_INTENT_EXCEEDS_MODEL_BOUND"} {
		t.Run(reason, func(t *testing.T) {
			source, doc := typedPathFixture(t)
			model := writeSharedThreeContractModel(t, "qat_ternary")
			switch reason {
			case "THREE_DECISION_COUNT_UNSUPPORTED":
				doc.Plan.Decisions = doc.Plan.Decisions[:1]
			case "FIELD_MODEL_REQUIRES_RECORD_BODY":
				model = "../../examples/scalar-identity/model/model.json"
			case "COMBINED_CONTEXT_INTENT_EXCEEDS_MODEL_BOUND":
				model = writePathContextContractModel(t)
				doc.Plan.Decisions[1].Intent = strings.Repeat("가", 150)
			}
			out, err := ExportTypedPathModelContext(context.Background(), "f.gooo", source, "Combined", doc, model, "")
			if err != nil {
				t.Fatal(err)
			}
			assertTypedPreflightZeroWork(t, out)
			if out.ModelCompatibility.Status != "DECLINED_TO_DETERMINISTIC" || out.ModelCompatibility.Reason != reason ||
				len(out.Inputs) != 0 || out.CompleteModelInput != nil {
				t.Fatal("decline exposed partial model input", out)
			}
			generated, err := GenerateWithTypedPaths(context.Background(), "f.gooo", source, "Combined", doc, model)
			if err != nil || generated.Report.BodyPaths.ModelContext.Reason != reason || generated.Report.BodyPaths.Search.Selection.ModelCalls != 0 {
				t.Fatal("preflight decline differs from native construction", err)
			}
		})
	}
}

func TestTypedModelPreflightRejectsInvalidInputsBeforeAdvertisingCompatibility(t *testing.T) {
	source, doc := typedPathFixture(t)
	model := writePathContextContractModel(t)
	for _, explicit := range []string{decision.SemanticContextIntentFeatureVersion, "unknown"} {
		if out, err := ExportTypedPathModelContext(context.Background(), "f.gooo", source, "Combined", doc, model, explicit); err == nil || out.ModelCompatibility != nil {
			t.Fatal("mismatched model feature advertised compatibility")
		}
	}
	for _, modelPath := range []string{"", model + ".missing", writeOrderModel(t)} {
		if out, err := ExportTypedPathModelContext(context.Background(), "f.gooo", source, "Combined", doc, modelPath, ""); err == nil || out.ModelCompatibility != nil {
			t.Fatal("unsupported artifact advertised compatibility")
		}
	}
	changed := []byte(strings.Replace(string(source), "input + 2", "input + 99", 1))
	_, err := ExportTypedPathModelContext(context.Background(), "f.gooo", changed, "Combined", doc, model+".missing", "")
	var failure *BodyPathError
	if !errors.As(err, &failure) || strings.Contains(err.Error(), "load retained") || failure.Receipt.ModelContext != nil {
		t.Fatal("model inspected before source binding", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if out, err := ExportTypedPathModelContext(ctx, "f.gooo", source, "Combined", doc, model, ""); !errors.Is(err, context.Canceled) || out.ModelCompatibility != nil {
		t.Fatal("cancelled request advertised compatibility", err)
	}
}
