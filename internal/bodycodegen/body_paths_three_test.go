package bodycodegen

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func threeNativeSource(t *testing.T, doc pathplan.Document) []byte {
	t.Helper()
	p, err := doc.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	return []byte(fmt.Sprintf("package sample\nnamespace sample\nentity Integer id \"sample://entity/integer\"\n"+
		"activity ThreeTest(Integer) -> Integer computes %q\n"+
		"activity Unrelated(Integer) -> Integer computes \"return input\"\n", strings.TrimSuffix(p.Fallback().GoooBody(), "\n")))
}

func threeNativeFixture(t *testing.T) ([]byte, pathplan.Document) {
	t.Helper()
	plan := pathplan.Plan{Schema: pathplan.Schema, Base: bodyplan.Plan{
		Schema: bodyplan.Schema, ID: "three-native-test", Name: "ThreeTest", ResultType: decision.TypeInt,
		Expressions: []bodyplan.Expr{{Kind: bodyplan.ExprInput, Name: "input"}, {Kind: bodyplan.ExprInt, Int: 1},
			{Kind: bodyplan.ExprBinary, Operation: "subtract", Left: 0, Right: 1},
			{Kind: bodyplan.ExprLocal, Name: "a"}, {Kind: bodyplan.ExprInt, Int: 3},
			{Kind: bodyplan.ExprBinary, Operation: "subtract", Left: 3, Right: 4},
			{Kind: bodyplan.ExprLocal, Name: "b"}, {Kind: bodyplan.ExprInt, Int: 9},
			{Kind: bodyplan.ExprBinary, Operation: "subtract", Left: 6, Right: 7}},
		Statements: []bodyplan.Stmt{{Kind: bodyplan.StmtLet, Name: "a", Expr: 2},
			{Kind: bodyplan.StmtLet, Name: "b", Expr: 5}, {Kind: bodyplan.StmtReturn, Expr: 8}}, Root: []int{0, 1, 2}}}
	for i, intent := range []string{"첫 번째 피연산자를 반대로 배치해라.", "Reverse the second operand order.", "세 번째 피연산자를 반대로 배치해라."} {
		plan.Decisions = append(plan.Decisions, pathplan.Choice{ID: fmt.Sprintf("choice_%d", i),
			Kind: pathplan.OperandOrder, Target: 2 + i*3, Intent: intent, Fallback: "layout_forward",
			Options: []pathplan.Option{{Label: "layout_forward"}, {Label: "layout_reverse", Reverse: true}}})
	}
	// Independent ordinary arithmetic: 9-(3-(1-input)) = 7-input.
	doc := pathplan.Document{Schema: pathplan.DocumentSchema, Plan: plan, MaxAttempts: 8,
		TestCases: []pathplan.TestCase{{Input: -8, Expected: 15}, {Input: -1, Expected: 8},
			{Input: 0, Expected: 7}, {Input: 3, Expected: 4}, {Input: 4, Expected: 3}, {Input: 9, Expected: -2}}}
	return threeNativeSource(t, doc), doc
}

// Controlled full-mask logits require eight candidates; these are not trained weights.
func writeThreeContractModel(t *testing.T) string {
	t.Helper()
	meta := jointdecision.Metadata{Schema: jointdecision.ThreeSchema, Feature: jointdecision.ThreeFeatureVersion,
		Variant: "fp32", FeatureDim: 768, HiddenDim: 24, MaxBytes: 1600, Temperature: 1, WeightsFile: "weights.bin"}
	var raw []byte
	for i, name := range [4]string{"w1", "b1", "w2", "b2"} {
		rows, cols := [4]int{24, 1, 8, 1}[i], [4]int{768, 24, 24, 8}[i]
		count := rows * cols
		meta.Tensors = append(meta.Tensors, decision.TensorMetadata{Name: name, Rows: rows, Cols: cols, Count: count,
			Encoding: "float32_le", Offset: int64(len(raw)), Bytes: int64(count * 4), Scale: 1})
		raw = append(raw, make([]byte, count*4)...)
	}
	for mask := range 8 {
		meta.Labels = append(meta.Labels, fmt.Sprintf("mask_%d", mask))
		binary.LittleEndian.PutUint32(raw[len(raw)-32+mask*4:], math.Float32bits(float32(7-mask)))
	}
	meta.WeightsSHA = strings.TrimPrefix(digest(raw), "sha256:")
	dir := t.TempDir()
	encoded, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(dir, "model.json")
	if err = os.WriteFile(name, encoded, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "weights.bin"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestThreeNativeSourceInputsFiniteTDDAndIndependentGo(t *testing.T) {
	source, doc := threeNativeFixture(t)
	original := string(source)
	ctx := context.Background()
	exported, err := ExportTypedPathContextWithFeature(ctx, "fixture.gooo", source, "ThreeTest", doc,
		decision.SemanticContextIntentFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPathUnfixedFeedback(ctx, "fixture.gooo", source, "ThreeTest", doc,
		writeThreeContractModel(t), 1, 7, &pathplan.CIHint{SourceSHA: strings.Repeat("a", 40), Status: "FAIL"})
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodyPaths
	consumePathCompleteness(t, result.Report.CompletenessReceipt)
	if result.Report.CompletenessReceipt.Scope["typed_path"].(map[string]any)["local_model_predictions"] != 7 {
		t.Fatal("common receipt lost actual three-choice feedback calls")
	}
	three := r.Search.Selection.Three
	if three == nil || !three.PredictionValid || r.Search.Selection.Joint != nil ||
		r.Search.Selection.ModelCalls != 7 || len(r.Search.Attempts) != 8 || r.Search.TypeRejected != 0 ||
		r.FunctionalCompleteness != 100 || !r.SourceBaseMatched || r.Progress[0].Attempted != 0 ||
		r.Progress[0].Selection.ModelCalls != 1 || result.Report.RepositoryWrites != 0 || string(source) != original {
		t.Fatal("three native TDD or initial prediction accounting differs")
	}
	text, err := jointdecision.EncodeThree([3]string{exported.Inputs[0].Text, exported.Inputs[1].Text, exported.Inputs[2].Text})
	if err != nil || three.Input != text || "sha256:"+three.InputSHA != digest([]byte(text)) ||
		r.ModelContext.Schema != "gooo/compiler-three-choice-path-context/v1" {
		t.Fatal("source export and actual model input differ", err)
	}
	for i, attempt := range r.Search.Attempts {
		if attempt.Mask != uint16(i) {
			t.Fatal("three native full-space ranking dropped a candidate")
		}
	}
	if len(r.Feedback) != 7 || !r.Feedback[6].RankingUnnecessary || r.Feedback[6].ModelCalls != 0 {
		t.Fatal("native sole mask made an unnecessary prediction")
	}
	for _, f := range r.Feedback {
		if f.CIIsAuthority || f.CI == nil || f.CI.Status != "FAIL" {
			t.Fatal("CI feedback became authority or was dropped")
		}
	}
	oracle := "package main\nimport \"fmt\"\n" + withoutPackage(t, result.Source) + "\nfunc main() {\n"
	var expected []string
	for _, c := range doc.TestCases {
		oracle += fmt.Sprintf("fmt.Println(ThreeTest(%d))\n", c.Input)
		expected = append(expected, fmt.Sprint(c.Expected))
	}
	if got := runBodyFillGoOracle(t, oracle+"}\n"); strings.TrimSpace(got) != strings.Join(expected, "\n") {
		t.Fatal("independently compiled generated Go differs from authored arithmetic", got)
	}
	doc.Plan.Base.Expressions[1].Int = 999
	_, err = GenerateWithTypedPaths(ctx, "fixture.gooo", source, "ThreeTest", doc, "missing-three-model.json")
	var failure *BodyPathError
	if !errors.As(err, &failure) || failure.Receipt.SourceBaseMatched ||
		failure.Receipt.Timing.ModelLoadMS != 0 || failure.Receipt.Search.Selection.ModelCalls != 0 {
		t.Fatal("model opened before native source binding", err)
	}
}

func TestThreeNativeFullOverflowAndUnsupportedArityKeepDeterministicPaths(t *testing.T) {
	model := writeThreeContractModel(t)
	ctx := context.Background()
	for _, kind := range []string{"feedback-overflow", "initial-overflow", "two", "four"} {
		t.Run(kind, func(t *testing.T) {
			source, doc := threeNativeDeclineFixture(t, kind)
			result, err := GenerateWithTypedPathUnfixedFeedback(ctx, "fixture.gooo", source, "ThreeTest", doc, model, 1, 7, nil)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "feedback-overflow" {
				consumePathCompleteness(t, result.Report.CompletenessReceipt)
				assertThreeFeedbackOverflow(t, result.Report.BodyPaths)
				return
			}
			assertThreeDeterministicDecline(t, source, doc, result)
			consumePathCompleteness(t, result.Report.CompletenessReceipt)
			observed := result.Report.CompletenessReceipt.Scope["typed_path"].(map[string]any)
			if observed["local_model_predictions"] != 0 || observed["provider_counts_known"] != true ||
				observed["external_provider_calls"] != 0 {
				t.Fatal("decline lost exact accounting")
			}
		})
	}
}

func threeNativeDeclineFixture(t *testing.T, kind string) ([]byte, pathplan.Document) {
	t.Helper()
	source, doc := threeNativeFixture(t)
	switch kind {
	case "feedback-overflow", "initial-overflow":
		length := 364
		if kind == "initial-overflow" {
			length = 400
		}
		for i := range doc.Plan.Decisions {
			doc.Plan.Decisions[i].Intent = strings.Repeat("x", length)
		}
	case "two":
		doc.Plan.Decisions = doc.Plan.Decisions[:2]
	case "four":
		doc.Plan.Base.Expressions = append(doc.Plan.Base.Expressions, bodyplan.Expr{Kind: bodyplan.ExprInt, Int: 2},
			bodyplan.Expr{Kind: bodyplan.ExprBinary, Operation: "subtract", Left: 8, Right: 9})
		doc.Plan.Base.Statements[2].Expr = 10
		doc.Plan.Decisions = append(doc.Plan.Decisions, pathplan.Choice{ID: "fourth", Kind: pathplan.OperandOrder,
			Target: 10, Intent: "Retain the complete fourth intention.", Fallback: "layout_forward",
			Options: []pathplan.Option{{Label: "layout_forward"}, {Label: "layout_reverse", Reverse: true}}})
		doc.MaxAttempts = 16
		source = threeNativeSource(t, doc)
	}
	if kind != "feedback-overflow" {
		doc.Seed = "must-skip-seed"
	}
	return source, doc
}

func assertThreeFeedbackOverflow(t *testing.T, r *BodyPathReceipt) {
	t.Helper()
	if r.Search.Selection.ModelCalls != 1 || r.FunctionalCompleteness != 100 || len(r.Feedback) != 7 {
		t.Fatal("native overflow continuation differs")
	}
	for _, feedback := range r.Feedback[:6] {
		if !feedback.ContextDeclined || feedback.ModelCalls != 0 || feedback.Three == nil ||
			feedback.Three.Bytes <= 1560 || feedback.Three.PartSHA[2] == "" {
			t.Fatal("native overflow lost complete zero-call evidence")
		}
	}
}

func assertThreeDeterministicDecline(t *testing.T, source []byte, doc pathplan.Document, result Result) {
	t.Helper()
	r := result.Report.BodyPaths
	declared := r.ModelContext.DeclaredInputs
	if r.Search.Selection.ModelCalls != 0 || len(r.Feedback) != 0 || !r.ModelContext.SeedSkipped ||
		!r.ModelContext.FeedbackSkipped || declared == nil || declared.Decisions != len(doc.Plan.Decisions) ||
		declared.Bytes != len(declared.Text) || declared.SHA256 != digest([]byte(declared.Text)) {
		t.Fatal("native decline inferred or lost full original input")
	}
	for _, c := range doc.Plan.Decisions {
		if !strings.Contains(declared.Text, c.Intent) {
			t.Fatal("original unsupported input dropped", c.ID)
		}
	}
	doc.Seed = ""
	baseline, err := GenerateWithTypedPathBatches(context.Background(), "fixture.gooo", source, "ThreeTest", doc, "", 1)
	if err != nil || baseline.Source != result.Source ||
		!reflect.DeepEqual(baseline.Report.BodyPaths.Search.Attempts, r.Search.Attempts) {
		t.Fatal("native decline changed disconnected ordering or emitted source", err)
	}
}

func TestThreeRetainedModelSurvivesFilesMovingAndConcurrentRequests(t *testing.T) {
	source, doc := threeNativeFixture(t)
	model := writeThreeContractModel(t)
	g, err := NewTypedPathGenerator(model)
	if err != nil {
		t.Fatal(err)
	}
	info := g.Info()
	if info.ModelSchema != jointdecision.ThreeSchema || info.ResidentTensorBytes != 74624 || !info.Loaded {
		t.Fatal("retained three model identity or layout differs")
	}
	for _, name := range []string{model, filepath.Join(filepath.Dir(model), "weights.bin")} {
		if err = os.Rename(name, name+".moved"); err != nil {
			t.Fatal(err)
		}
	}
	options := TypedPathOptions{StepAttempts: 1, FeedbackRounds: 7, FeedbackUnfixed: true}
	seeded := doc
	seeded.Seed = "repeatable-three-seed"
	first, err := g.Generate(context.Background(), "fixture.gooo", source, "ThreeTest", seeded, options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := g.Generate(context.Background(), "fixture.gooo", source, "ThreeTest", seeded, options)
	if err != nil || first.Source != second.Source ||
		first.Report.BodyPaths.Search.Selection.Three.Sampled != second.Report.BodyPaths.Search.Selection.Three.Sampled ||
		strings.Contains(first.Report.BodyPaths.Search.Selection.Three.Input, seeded.Seed) {
		t.Fatal("native seed replay changed source or leaked into model text", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = g.Generate(canceled, "fixture.gooo", source, "ThreeTest", doc, options)
	var failure *BodyPathError
	if !errors.As(err, &failure) || failure.Receipt.Search.Selection.ModelCalls != 0 || !errors.Is(err, context.Canceled) {
		t.Fatal("canceled retained request called a model", err)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			result, err := g.Generate(context.Background(), "fixture.gooo", source, "ThreeTest", doc, options)
			if err != nil {
				t.Error(err)
				return
			}
			r := result.Report.BodyPaths
			if r.FunctionalCompleteness != 100 || r.Search.Selection.ModelCalls != 7 ||
				r.Timing.ModelLoadMS != 0 || *r.ModelRetention != info || result.Report.RepositoryWrites != 0 {
				t.Error("retained three request shared state or reloaded weights")
			}
		})
	}
	wg.Wait()
}
