package bodycodegen

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestSemanticSourceExportMatchesActualNativeInputs(t *testing.T) {
	for _, language := range []string{"en", "ko"} {
		source, document := conditionalPathFixture(t, language)
		before, _ := json.Marshal(document)
		exported, err := ExportTypedPathContextWithFeature(context.Background(), "fixture.gooo", source,
			"ConditionalAssign", document, decision.SemanticContextIntentFeatureVersion)
		if err != nil {
			t.Fatal(err)
		}
		modelFile := writePathContextContractModelVersion(t, decision.SemanticContextIntentFeatureVersion)
		generated, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "ConditionalAssign", document, modelFile)
		if err != nil {
			t.Fatal(err)
		}
		p := generated.Report.BodyPaths
		if exported.Schema != "gooo/compiler-path-input-export/v2" || exported.Context.Schema != semanticPathContextSchema ||
			len(exported.Inputs) != 6 || p.Search.Selection.ModelCalls != 6 || p.FunctionalCompleteness != 100 ||
			!reflect.DeepEqual(exported.Context.Inputs, p.ModelContext.Inputs) ||
			exported.Context.RankedPlanSHA != p.ModelContext.RankedPlanSHA || exported.ModelPredictions != 0 ||
			exported.CandidateTests != 0 || exported.SelectedEmission || exported.RepositoryWrites != 0 {
			t.Fatal("export/runtime contract differs", exported, p)
		}
		prepared, _ := document.Prepare()
		seen := map[string]bool{}
		for i, input := range exported.Inputs {
			choice := document.Plan.Decisions[i]
			seen[choice.Kind] = true
			fields, err := prepared.SourceFeatures(choice.ID)
			if err != nil {
				t.Fatal(err)
			}
			natural := choice.Intent[strings.LastIndex(choice.Intent, "intent: ")+8:]
			expected, err := decision.EncodeSemanticContextInput(fields, natural)
			if err != nil || input.Text != expected || input.InputSHA != digest([]byte(expected)) ||
				input.SourceFeatureSHA != digest(fields[:]) || input.Bytes != len(expected) ||
				"sha256:"+p.Search.Selection.Receipts[i].IntentSHA256 != input.InputSHA {
				t.Fatal("actual model input differs", input, err)
			}
		}
		if len(seen) != 5 {
			t.Fatal("five structural kinds were not exercised")
		}
		document.TestCases[0].Expected = 999
		document.Seed = "not-a-source-feature"
		changed, err := ExportTypedPathContextWithFeature(context.Background(), "fixture.gooo", source,
			"ConditionalAssign", document, decision.SemanticContextIntentFeatureVersion)
		if err != nil || !reflect.DeepEqual(exported.Inputs, changed.Inputs) || changed.TestSuiteSHA256 == exported.TestSuiteSHA256 {
			t.Fatal("outcome/seed leaked", err)
		}
		document.TestCases[0].Expected = generated.Report.BodyPaths.NativeCases[0].Expected
		document.Seed = ""
		after, _ := json.Marshal(document)
		if string(before) != string(after) {
			t.Fatal("caller document changed")
		}
	}
}

func TestSemanticSourceFeaturesIgnoreLiteralMagnitudesAndLocalSpellings(t *testing.T) {
	source, document := conditionalPathFixture(t, "en")
	before, err := ExportTypedPathContextWithFeature(context.Background(), "fixture.gooo", source,
		"ConditionalAssign", document, decision.SemanticContextIntentFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	for i := range document.Plan.Base.Expressions {
		e := &document.Plan.Base.Expressions[i]
		if e.Kind == "local" {
			e.Name = "renamed_" + e.Name
		}
		if e.Kind == "int" {
			e.Int += 11
		}
	}
	for i := range document.Plan.Base.Statements {
		s := &document.Plan.Base.Statements[i]
		if s.Name != "" {
			s.Name = "renamed_" + s.Name
		}
	}
	for i := range document.Plan.Decisions {
		for j := range document.Plan.Decisions[i].Options {
			o := &document.Plan.Decisions[i].Options[j]
			if o.Name != "" {
				o.Name = "renamed_" + o.Name
			}
		}
	}
	prepared, err := document.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(prepared.Fallback().GoooBody())
	source = []byte("package sample\nnamespace sample\nentity Integer id \"sample://entity/integer\"\n" +
		"activity ConditionalAssign(Integer) -> Integer computes " + string(body) + "\n")
	after, err := ExportTypedPathContextWithFeature(context.Background(), "fixture.gooo", source,
		"ConditionalAssign", document, decision.SemanticContextIntentFeatureVersion)
	if err != nil || len(after.Inputs) != len(before.Inputs) {
		t.Fatal(err)
	}
	for i := range before.Inputs {
		if after.Inputs[i].SourceFeatureSHA != before.Inputs[i].SourceFeatureSHA || after.Inputs[i].Text != before.Inputs[i].Text {
			t.Fatal("name/literal leaked", i)
		}
	}
	if after.OriginalSourceSHA256 == before.OriginalSourceSHA256 {
		t.Fatal("source identity was lost")
	}
}

func TestSemanticSourceReversedFallbackAndPrefixReplacement(t *testing.T) {
	_, document := conditionalPathFixture(t, "en")
	for i := range document.Plan.Decisions {
		document.Plan.Decisions[i].Fallback = document.Plan.Decisions[i].Options[1].Label
	}
	prepared, err := document.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(prepared.Fallback().GoooBody())
	source := []byte("package sample\nnamespace sample\nentity Integer id \"sample://entity/integer\"\n" +
		"activity ConditionalAssign(Integer) -> Integer computes " + string(body) + "\n")
	exported, err := ExportTypedPathContextWithFeature(context.Background(), "fixture.gooo", source,
		"ConditionalAssign", document, decision.SemanticContextIntentFeatureVersion)
	if err != nil || len(exported.Inputs) != 6 {
		t.Fatal(exported, err)
	}
	for i, choice := range document.Plan.Decisions {
		if choice.Kind == pathplan.OperandOrder || choice.Kind == pathplan.BranchLayout {
			fields, _ := prepared.SourceFeatures(choice.ID)
			if fields[44] != 128 || fields[54] != 0 || fields[53] != 0 || fields[63] != 128 {
				t.Fatal("source-relative reversal lost", fields)
			}
		}
		suffix := choice.Intent[strings.LastIndex(choice.Intent, "intent: ")+8:]
		choice.Intent = "untrusted expected=999 selected=anything intent: " + suffix
		text, input, reason := encodeCompilerPathContext(decision.SemanticContextIntentFeatureVersion, fallbackContextFacts(document.Plan), prepared, choice)
		if reason != "" || text != exported.Inputs[i].Text || input.SourceFeatureSHA != exported.Inputs[i].SourceFeatureSHA {
			t.Fatal("caller prefix became source facts")
		}
	}
}

func TestSemanticSourceDeclineContinuesDeterministicPartialBody(t *testing.T) {
	source, document := conditionalPathFixture(t, "en")
	document.TestCases[6].Expected = 999
	// A later choice overflows after earlier choices encode: decline is atomic.
	document.Plan.Decisions[4].Intent = strings.Repeat("x", 365)
	baseline, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "ConditionalAssign", document, "")
	if err != nil {
		t.Fatal(err)
	}
	document.Seed = "explicit-seed"
	result, err := GenerateWithTypedPathUnfixedFeedback(context.Background(), "fixture.gooo", source,
		"ConditionalAssign", document, writePathContextContractModelVersion(t, decision.SemanticContextIntentFeatureVersion), 8, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if result.Source != baseline.Source || p.ModelContext.Status != "DECLINED_TO_DETERMINISTIC" ||
		p.Search.Selection.ModelCalls != 0 || p.FunctionalCompleteness != 600.0/7 || p.Search.Evaluated != 64 ||
		!p.ModelContext.SeedSkipped || !p.ModelContext.FeedbackSkipped || len(p.Feedback) != 0 ||
		p.ModelContext.Inputs[4].Bytes != 513 || p.ModelContext.Inputs[4].InputSHA != "" || p.ModelContext.Inputs[4].SourceFeatureSHA == "" {
		t.Fatal("decline changed complete deterministic continuation", p)
	}
	exported, err := ExportTypedPathContextWithFeature(context.Background(), "fixture.gooo", source,
		"ConditionalAssign", document, decision.SemanticContextIntentFeatureVersion)
	if err != nil || len(exported.Inputs) != 0 {
		t.Fatal("partial text escaped decline", err)
	}
}

func TestSemanticNativeFeedbackPreservesSourceAndBounds(t *testing.T) {
	for _, long := range []bool{false, true} {
		source, document := conditionalPathFixture(t, "ko")
		document.TestCases[6].Expected = 999
		if long {
			document.Plan.Decisions[0].Intent = strings.Repeat("x", 364)
		}
		exported, err := ExportTypedPathContextWithFeature(context.Background(), "fixture.gooo", source,
			"ConditionalAssign", document, decision.SemanticContextIntentFeatureVersion)
		if err != nil {
			t.Fatal(err)
		}
		modelFile := writePathContextContractModelVersion(t, decision.SemanticContextIntentFeatureVersion)
		result, err := GenerateWithTypedPathFeedback(context.Background(), "fixture.gooo", source,
			"ConditionalAssign", document, modelFile, 8, 2, &pathplan.CIHint{SourceSHA: strings.Repeat("a", 40), Status: "FAIL"})
		if err != nil {
			t.Fatal(err)
		}
		p := result.Report.BodyPaths
		if len(p.Feedback) != 2 || p.FunctionalCompleteness != 600.0/7 || p.Search.Evaluated != 64 {
			t.Fatal("partial feedback did not finish", p)
		}
		for _, feedback := range p.Feedback {
			if long {
				if !feedback.ContextDeclined || feedback.ModelCalls != 0 || feedback.Applied || feedback.DeclinedBytes <= 512 || feedback.DeclinedInputSHA == "" {
					t.Fatal("complete oversize context was inferred", feedback)
				}
				continue
			}
			if feedback.ModelCalls != 6 || !feedback.Applied {
				t.Fatal("feedback calls hidden", feedback)
			}
			for i, judgment := range feedback.Judgments {
				if !strings.HasPrefix(judgment.Input, exported.Inputs[i].Text+"\nfeedback: ") {
					t.Fatal("source/header changed", judgment.Input)
				}
				raw, err := hex.DecodeString(judgment.Input[len("gooo;sem64=") : len("gooo;sem64=")+128])
				if err != nil || digest(raw) != exported.Inputs[i].SourceFeatureSHA {
					t.Fatal("source features changed during feedback")
				}
			}
		}
		calls := 18
		if long {
			calls = 6
		}
		if p.Search.Selection.ModelCalls != calls {
			t.Fatal("actual cumulative model count differs", p.Search.Selection.ModelCalls)
		}
	}
}

func TestSemanticRetainedConcurrentOwnershipAndSourceRejection(t *testing.T) {
	source, document := typedPathFixture(t)
	generator, err := NewTypedPathGenerator(writePathContextContractModelVersion(t, decision.SemanticContextIntentFeatureVersion))
	if err != nil {
		t.Fatal(err)
	}
	first, err := generator.Generate(context.Background(), "fixture.gooo", source, "Combined", document, TypedPathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	expected := first.Report.BodyPaths.ModelContext.Inputs[0].SourceFeatureSHA
	first.Report.BodyPaths.ModelContext.Inputs[0].SourceFeatureSHA = "caller mutation"
	var wait sync.WaitGroup
	for range 8 {
		wait.Go(func() {
			current, err := generator.Generate(context.Background(), "fixture.gooo", source, "Combined", document, TypedPathOptions{})
			if err != nil || current.Report.BodyPaths.ModelContext.Inputs[0].SourceFeatureSHA != expected {
				t.Error("retained snapshot mutated", err)
			}
		})
	}
	wait.Wait()
	for _, data := range [][]byte{nil, []byte(strings.Replace(string(source), "input + 2", "input + 99", 1))} {
		_, err = generator.Generate(context.Background(), "fixture.gooo", data, "Combined", document, TypedPathOptions{})
		failure, ok := errors.AsType[*BodyPathError](err)
		if !ok || failure.Receipt.ModelContext != nil || failure.Receipt.Search.Selection.ModelCalls != 0 {
			t.Fatal("source mismatch reached inference", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = ExportTypedPathContextWithFeature(ctx, "fixture.gooo", source, "Combined", document, decision.SemanticContextIntentFeatureVersion); !errors.Is(err, context.Canceled) {
		t.Fatal("export ignored cancellation")
	}
	if _, err = ExportTypedPathContextWithFeature(context.Background(), "fixture.gooo", source, "Combined", document, "unknown"); err == nil {
		t.Fatal("unknown feature version accepted")
	}
}
