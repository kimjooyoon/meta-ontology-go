package bodycodegen

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/conditiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func writeConditionContractModel(t *testing.T) string {
	t.Helper()
	m, err := conditiondecision.New([conditiondecision.ParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	// The published artifact is 72,482B. Whitespace tests the same >64KiB
	// container boundary without substituting learned quality for API coverage.
	raw = append([]byte(strings.Repeat(" ", 70<<10)), raw...)
	name := filepath.Join(t.TempDir(), "condition.json")
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func conditionModelSource() []byte {
	return []byte(`package conditionmodel
namespace conditionmodel
entity Integer id "conditionmodel://integer"
activity Choose(Integer) -> Integer computes "if input < 0 { return 0 - input } else { return input }" assembling {
 choice "comparison" operand_order at "0" intent "양수인지 확인한다. Compare whether input is positive."
 choice "branches" branch_layout at "0" intent "양수일 때 참 분기를 사용한다. Use the true branch for positive inputs."
 case "-9" -> "9"
 case "0" -> "0"
 case "9" -> "9"
 case "9007199254740993" -> "9007199254740993"
 case "-9007199254740995" -> "9007199254740995"
 case "9007199254740995" -> "9007199254740995"
 case "18014398509481990" -> "18014398509481990"
 condition_case "comparison" input "-9007199254740995" -> "false"
 condition_case "comparison" input "0" -> "false"
 condition_case "comparison" input "9007199254740995" -> "true"
 attempts "4"
}
`)
}

func TestConditionModelPreflightMatchesSourceArrayAndNativeSearch(t *testing.T) {
	ctx := context.Background()
	source := conditionModelSource()
	doc, err := DecodeSourcePathDocument(ctx, "condition.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	name := writeConditionContractModel(t)
	preflight, err := ExportTypedPathModelContext(ctx, "condition.gooo", source, "Choose", doc, name, decision.ConditionChannelFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	info := preflight.ModelCompatibility.Model
	if preflight.ModelPredictions != 0 || preflight.CandidateTests != 0 || preflight.SelectedEmission || len(preflight.Inputs) != 2 || info.ModelSchema != conditiondecision.Schema || info.ResidentTensorBytes != 24872 || info.MetadataSHA256 != "" || info.WeightsSHA256 != "" || info.ArtifactSHA256 == "" || info.ModelFingerprint == "" {
		t.Fatal("condition preflight identity or work differs", preflight)
	}
	result, err := GenerateWithTypedPathFeedback(ctx, "condition.gooo", source, "Choose", doc, name, 1, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p.Search.Status != "TRAINING_COMPLETE" || p.Search.ConditionRejected == 0 || p.Conditions.Passed != 3 || p.FunctionalCompleteness != 100 || p.Search.Selection.ModelVariant != "condition_fp32" || p.Search.Selection.ModelCalls != 3 || len(p.ConditionProgress) == 0 || len(p.Progress) != 0 || len(p.Feedback) != 0 {
		t.Fatal("condition model did not reach normal validation", p)
	}
	if !reflect.DeepEqual(p.ModelContext, preflight.Context) {
		t.Fatal("preflight used different source context")
	}
	first := p.ConditionProgress[0].Ranking
	for i, input := range preflight.Inputs {
		if input.Features == nil || input.Text != doc.Plan.Decisions[i].Intent || input.InputSHA != "sha256:"+first.FeatureSHA[i] || conditionFeatureDigest(*input.Features) != input.InputSHA || input.Bytes != 1024 {
			t.Fatal("feature bytes differ from actual neural input", i)
		}
		for _, v := range input.Features[192:] {
			if v != 0 {
				t.Fatal("future conditions leaked into first judgment")
			}
		}
	}
	if p.ConditionFeedback[0].Failure.Result.Case.Input != -9007199254740995 || !p.ConditionFeedback[0].HasFailure {
		t.Fatal("observed condition lost exact input")
	}
	// The stored selected program replays without reopening its model artifact.
	if err := os.Remove(name); err != nil {
		t.Fatal(err)
	}
	if err := VerifyTypedPathProjection(ctx, "condition.gooo", source, doc, result); err != nil {
		t.Fatal(err)
	}
	for _, row := range p.NativeCases {
		if !row.Passed {
			t.Fatal("native projection case failed", row)
		}
	}
}

func TestConditionModelRetainedRequestsAndOptions(t *testing.T) {
	ctx := context.Background()
	source := conditionModelSource()
	doc, err := DecodeSourcePathDocument(ctx, "f.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	name := writeConditionContractModel(t)
	g, err := NewTypedPathGenerator(name)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(name); err != nil {
		t.Fatal(err)
	}
	options := TypedPathOptions{StepAttempts: 1, FeedbackRounds: 3}
	first, err := g.Generate(ctx, "f.gooo", source, "Choose", doc, options)
	if err != nil {
		t.Fatal(err)
	}
	first.Report.BodyPaths.ModelContext.Inputs[0].InputSHA = "modified"
	var wg sync.WaitGroup
	for range 3 {
		wg.Go(func() {
			r, err := g.Generate(ctx, "f.gooo", source, "Choose", doc, options)
			if err != nil {
				t.Error(err)
				return
			}
			if r.Source != first.Source || r.Report.BodyPaths.Search.Selection.ModelCalls != 3 || r.Report.BodyPaths.ModelContext.Inputs[0].InputSHA == "modified" {
				t.Error("retained request borrowed mutable state")
			}
		})
	}
	wg.Wait()
	for _, bad := range []TypedPathOptions{{StepAttempts: 1, FeedbackRounds: 1, CI: &pathplan.CIHint{SourceSHA: strings.Repeat("a", 40), Status: "FAIL"}}, {StepAttempts: 1, FeedbackRounds: 1, FeedbackUnfixed: true}} {
		_, err := g.Generate(ctx, "f.gooo", source, "Choose", doc, bad)
		var failure *BodyPathError
		if !errors.As(err, &failure) || !strings.Contains(err.Error(), "outside its feature contract") || failure.Receipt.SearchStarted || failure.Receipt.Search.Selection.ModelCalls != 0 {
			t.Fatal("unsupported model context silently accepted", err)
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = g.Generate(cancelled, "f.gooo", source, "Choose", doc, options); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation ignored", err)
	}
	changed := doc
	changed.TestCases = append([]pathplan.TestCase(nil), doc.TestCases...)
	changed.TestCases[0].Expected = 99
	if _, err := GenerateWithTypedPaths(ctx, "f.gooo", source, "Choose", changed, "missing.json"); err == nil || !strings.Contains(err.Error(), "source assembling contract") {
		t.Fatal("source mismatch reached model", err)
	}
}

func TestConditionModelArtifactIdentityBoundsAndClosedFormat(t *testing.T) {
	name := writeConditionContractModel(t)
	original, err := NewTypedPathGenerator(name)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(name)
	if err := os.WriteFile(name, append([]byte("\n"), raw...), 0600); err != nil {
		t.Fatal(err)
	}
	other, err := NewTypedPathGenerator(name)
	if err != nil || other.Info().ModelFingerprint != original.Info().ModelFingerprint || other.Info().ArtifactSHA256 == original.Info().ArtifactSHA256 {
		t.Fatal("semantic and file identities confused", err)
	}
	for _, bad := range [][]byte{[]byte(strings.Repeat(" ", 513<<10)), []byte(`{"schema":"gooo/condition-candidate-decision/v1","schema":"again"}`), []byte(`{"schema":"gooo/condition-candidate-decision/v1","architecture":[256,24,2]}`)} {
		if err := os.WriteFile(name, bad, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := NewTypedPathGenerator(name); err == nil {
			t.Fatal("invalid condition artifact accepted")
		}
	}
}

func TestConditionModelUnsupportedSourceRetainsDeterministicBody(t *testing.T) {
	for _, kind := range []string{"condition", "flow"} {
		t.Run(kind, func(t *testing.T) { candidateModelUnsupportedSourceRetainsDeterministicBody(t, kind) })
	}
}

func candidateModelUnsupportedSourceRetainsDeterministicBody(t *testing.T, kind string) {
	t.Helper()
	source := []byte(`package scoped
namespace scoped
entity Integer id "scoped://integer"
activity Choose(Integer) -> Integer computes "let result = input; if input < 0 { let inside = 0 - input; result = inside } else { let inside = input; result = inside }; return result" assembling {
 choice "branches" branch_layout at "0" intent "범위별 지역 변수. Branch-local variables."
 case "-2" -> "2"
 case "2" -> "2"
 attempts "2"
}
`)
	ctx := context.Background()
	doc, err := DecodeSourcePathDocument(ctx, "s.gooo", source, "Choose", nil)
	if err != nil {
		t.Fatal(err)
	}
	base, err := GenerateWithTypedPaths(ctx, "s.gooo", source, "Choose", doc, "")
	if err != nil {
		t.Fatal(err)
	}
	name := writeConditionContractModel(t)
	if kind == "flow" {
		name = writeFlowContractModel(t, false)
	}
	r, err := GenerateWithTypedPathFeedback(ctx, "s.gooo", source, "Choose", doc, name, 1, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := r.Report.BodyPaths
	if p.ModelContext.Status != "DECLINED_TO_DETERMINISTIC" || p.Search.Selection.ModelCalls != 0 || r.Source != base.Source || len(p.ConditionProgress) != 0 || !p.ModelContext.FeedbackSkipped {
		t.Fatal("source representation decline changed execution", p)
	}
	exported, err := ExportTypedPathModelContext(ctx, "s.gooo", source, "Choose", doc, name, "")
	if err != nil || exported.ModelCompatibility.Status != "DECLINED_TO_DETERMINISTIC" || len(exported.Inputs) != 0 || exported.ModelPredictions != 0 {
		t.Fatal("decline advertised inputs", err)
	}
}
