package bodycodegen

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func jointNativeFixture(t *testing.T) ([]byte, pathplan.Document) {
	t.Helper()
	source, doc := typedPathFixture(t)
	doc.Plan.Decisions = doc.Plan.Decisions[:2]
	doc.MaxAttempts = 4
	return source, doc
}

// Synthetic constant logits exercise ABI/call accounting, not trained quality.
func writeJointContractModel(t *testing.T) string {
	t.Helper()
	meta := jointdecision.Metadata{Schema: jointdecision.Schema, Feature: jointdecision.FeatureVersion, Variant: "fp32", FeatureDim: 512, HiddenDim: 24, MaxBytes: 1088, Labels: []string{"mask_0", "mask_1", "mask_2", "mask_3"}, Temperature: 1, WeightsFile: "weights.bin"}
	var raw []byte
	for i, name := range [4]string{"w1", "b1", "w2", "b2"} {
		rows, cols := [4]int{24, 1, 4, 1}[i], [4]int{512, 24, 24, 4}[i]
		count := rows * cols
		meta.Tensors = append(meta.Tensors, decision.TensorMetadata{Name: name, Rows: rows, Cols: cols, Count: count, Encoding: "float32_le", Offset: int64(len(raw)), Bytes: int64(count * 4), Scale: 1})
		raw = append(raw, make([]byte, count*4)...)
	}
	sum := sha256.Sum256(raw)
	meta.WeightsSHA = hex.EncodeToString(sum[:])
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
func TestJointNativeSourceInputsAndOneInitialPrediction(t *testing.T) {
	source, doc := jointNativeFixture(t)
	exported, err := ExportTypedPathContextWithFeature(context.Background(), "fixture.gooo", source, "Combined", doc, decision.SemanticContextIntentFeatureVersion)
	if err != nil {
		t.Fatal(err)
	}
	model := writeJointContractModel(t)
	result, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", doc, model)
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodyPaths
	j := r.Search.Selection.Joint
	if j == nil || r.Search.Selection.ModelCalls != 1 || j.Calls != 1 || r.FunctionalCompleteness != 100 || !r.SourceBaseMatched || len(r.Progress) < 2 || r.Progress[0].Attempted != 0 || r.Progress[0].Selection.ModelCalls != 1 {
		t.Fatal("joint initial/native contract", r)
	}
	text, err := jointdecision.Encode([2]string{exported.Inputs[0].Text, exported.Inputs[1].Text})
	if err != nil || j.Input != text || "sha256:"+j.InputSHA != digest([]byte(text)) {
		t.Fatal("source/export/native joint bytes differ", err)
	}
	if r.ModelContext.Schema != "gooo/compiler-joint-path-context/v1" || r.ModelContext.FeatureVersion != jointdecision.FeatureVersion || result.Report.RepositoryWrites != 0 {
		t.Fatal("joint ABI or mutation")
	}
	for i, c := range r.NativeCases {
		if c.Input != doc.TestCases[i].Input || c.Expected != doc.TestCases[i].Expected || !c.Passed {
			t.Fatal("native case mismatch")
		}
	}
	doc.Plan.Base.Expressions[1].Int = 999
	_, err = GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", doc, "missing-joint-model.json")
	var failure *BodyPathError
	if !errors.As(err, &failure) || failure.Receipt.SourceBaseMatched || failure.Receipt.Timing.ModelLoadMS != 0 || failure.Receipt.Search.Selection.ModelCalls != 0 {
		t.Fatal("joint load preceded source binding", err)
	}
}
func TestJointNativeFeedbackAndRepresentationDecline(t *testing.T) {
	source, doc := jointNativeFixture(t)
	model := writeJointContractModel(t)
	hint := &pathplan.CIHint{SourceSHA: strings.Repeat("a", 40), Status: "PASS"}
	result, err := GenerateWithTypedPathUnfixedFeedback(context.Background(), "fixture.gooo", source, "Combined", doc, model, 1, 3, hint)
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.BodyPaths
	calls := 1
	for _, f := range r.Feedback {
		calls += f.ModelCalls
		if f.CIIsAuthority {
			t.Fatal("CI hint became authority")
		}
		if f.RankingUnnecessary && f.ModelCalls != 0 {
			t.Fatal("sole path inferred")
		}
	}
	if r.Search.Selection.ModelCalls != calls || r.FunctionalCompleteness != 100 || calls > 3 {
		t.Fatal("joint feedback totals", calls)
	}
	doc.Seed = "skip-seed"
	for i := range doc.Plan.Decisions {
		doc.Plan.Decisions[i].Intent = "intent: " + strings.Repeat("x", 400)
	}
	result, err = GenerateWithTypedPathFeedback(context.Background(), "fixture.gooo", source, "Combined", doc, model, 1, 3, hint)
	if err != nil {
		t.Fatal(err)
	}
	r = result.Report.BodyPaths
	if r.ModelContext.Status != "DECLINED_TO_DETERMINISTIC" || !r.ModelContext.SeedSkipped || !r.ModelContext.FeedbackSkipped || r.Search.Selection.ModelCalls != 0 || len(r.Feedback) != 0 || r.FunctionalCompleteness != 100 {
		t.Fatal("source representation overflow lost deterministic continuation", r)
	}
	source, doc = typedPathFixture(t)
	result, err = GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", doc, model)
	if err != nil || result.Report.BodyPaths.Search.Selection.ModelCalls != 0 || result.Report.BodyPaths.ModelContext.Reason != "JOINT_DECISION_COUNT_UNSUPPORTED" {
		t.Fatal("unsupported joint arity did not decline", err)
	}
}
func TestJointRetainedGeneratorConcurrentCallerOwnedSessions(t *testing.T) {
	source, doc := jointNativeFixture(t)
	g, err := NewTypedPathGenerator(writeJointContractModel(t))
	if err != nil {
		t.Fatal(err)
	}
	info := g.Info()
	if !info.Loaded || info.ModelSchema != jointdecision.Schema || info.ResidentTensorBytes != 49648 {
		t.Fatal("retained joint model layout", info)
	}
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			result, err := g.Generate(context.Background(), "fixture.gooo", source, "Combined", doc, TypedPathOptions{StepAttempts: 1})
			if err != nil {
				t.Error(err)
				return
			}
			r := result.Report.BodyPaths
			if r.Timing.ModelLoadMS != 0 || r.Search.Selection.ModelCalls != 1 || r.FunctionalCompleteness != 100 {
				t.Error("retained joint request shares mutable session")
			}
		})
	}
	wg.Wait()
}
