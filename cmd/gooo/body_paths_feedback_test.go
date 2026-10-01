package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestTypedPathFeedbackCIHintStrictDecode(t *testing.T) {
	valid := `{"source_sha":"` + strings.Repeat("a", 40) + `","status":"PASS"}`
	if hint, err := decodePathCIHint([]byte(valid)); err != nil || hint.Status != "PASS" {
		t.Fatal("valid bounded CI hint rejected", err)
	}
	for _, raw := range []string{
		"", "null", "{}", strings.Repeat(" ", 513), valid + "{}",
		strings.Replace(valid, `"PASS"`, `null`, 1),
		strings.Replace(valid, `"PASS"`, `"pass"`, 1),
		strings.Replace(valid, strings.Repeat("a", 40), strings.Repeat("A", 40), 1),
		strings.Replace(valid, `"status":"PASS"`, `"status":"PASS","status":"FAIL"`, 1),
		strings.Replace(valid, `"status":"PASS"`, `"status":"PASS","approval":true`, 1),
		`{"status":"PASS"}`, `{"source_sha":null,"status":"PASS"}`,
	} {
		if _, err := decodePathCIHint([]byte(raw)); err == nil {
			t.Fatalf("invalid CI hint accepted: %q", raw)
		}
	}
}

// Synthetic equal logits check local calls and CLI wiring, not trained quality.
func writeCLIPathFeedbackModel(t *testing.T) string {
	t.Helper()
	labels := decision.PathLabels()
	threshold := 1.0
	metadata := decision.Metadata{Schema: decision.PathMetadataSchema, Variant: "fp32", FeatureDim: decision.FeatureDim,
		HiddenDim: decision.HiddenDim, MaxBytes: decision.InputMaxBytes, Labels: labels[:], Temperature: 1,
		WeightsFile: "weights.bin", ConfidenceThreshold: &threshold}
	var raw []byte
	for _, tensor := range []struct {
		name       string
		rows, cols int
	}{{"w1", decision.HiddenDim, decision.FeatureDim}, {"b1", 1, decision.HiddenDim},
		{"w2", decision.LabelCount, decision.HiddenDim}, {"b2", 1, decision.LabelCount}} {
		count := tensor.rows * tensor.cols
		metadata.Tensors = append(metadata.Tensors, decision.TensorMetadata{Name: tensor.name, Count: count,
			Rows: tensor.rows, Cols: tensor.cols, Encoding: "float32_le", Offset: int64(len(raw)), Bytes: int64(count * 4), Scale: 1})
		raw = append(raw, make([]byte, count*4)...)
	}
	sum := sha256.Sum256(raw)
	metadata.WeightsSHA256 = hex.EncodeToString(sum[:])
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "weights.bin"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "model.json")
	if err := os.WriteFile(path, encoded, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTypedPathFeedbackCLIIsLocalAndEmitsPartialReceipts(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("GOOO_LAYA_URL", server.URL)
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-conditional-assignment.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := os.ReadFile("../../examples/body-codegen/typed-path-conditional-assignment-ko-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	plan = bytes.Replace(plan, []byte(`"expected": 24`), []byte(`"expected": 999`), 1)
	reader := mapSourceReader{"fixture.gooo": source, "plan.json": plan,
		"hint.json": []byte(`{"source_sha":"` + strings.Repeat("a", 40) + `","status":"FAIL"}`)}
	args := []string{"--json", "--path-plan", "plan.json", "--path-model", writeCLIPathFeedbackModel(t),
		"--path-step-attempts", "8", "--path-feedback-rounds", "2", "--path-feedback-ci", "hint.json",
		"--activity", "ConditionalAssign", "fixture.gooo"}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen(args, reader, &stdout, &stderr)
	var result bodycodegen.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err, stdout.String())
	}
	if code != exitOK || stderr.Len() != 0 || result.Report.BodyPaths == nil || calls.Load() != 0 {
		t.Fatalf("local feedback failed: %d %s calls=%d", code, stderr.String(), calls.Load())
	}
	receipt := result.Report.BodyPaths
	if receipt.Search.Status != "PARTIAL" || receipt.Search.Evaluated != 64 || len(receipt.Feedback) != 2 ||
		receipt.FunctionalCompleteness != 600.0/7 || receipt.Search.Selection.ModelCalls != 18 ||
		receipt.Feedback[0].CIIsAuthority || result.Report.RepositoryWrites != 0 {
		t.Fatal("CLI lost finite partial outcome or local prediction receipts")
	}
	reader["hint.json"] = []byte(`{"status":"PASS"}`)
	stdout.Reset()
	stderr.Reset()
	if code := runBodyCodegen(args, reader, &stdout, &stderr); code != exitFailure || calls.Load() != 0 ||
		!strings.Contains(stdout.String(), `"decision":"FAIL_CLOSED"`) {
		t.Fatal("invalid caller context reached an external provider or emitted a body")
	}
}

func TestTypedPathFeedbackCLIContinuesAfterContextDecline(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-conditional-assignment.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/typed-path-conditional-assignment-ko-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	var document pathplan.Document
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	document.TestCases[6].Expected = 999
	for len(document.Plan.Decisions[0].Intent)+len(" 설명") <= 480 {
		document.Plan.Decisions[0].Intent += " 설명"
	}
	raw, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	reader := mapSourceReader{"fixture.gooo": source, "plan.json": raw}
	args := []string{"--json", "--path-plan", "plan.json", "--path-model", writeCLIPathFeedbackModel(t),
		"--path-step-attempts", "8", "--path-feedback-rounds", "2", "--activity", "ConditionalAssign", "fixture.gooo"}
	var stdout, stderr bytes.Buffer
	if code := runBodyCodegen(args, reader, &stdout, &stderr); code != exitOK || stderr.Len() != 0 {
		t.Fatalf("context decline stopped CLI: %d %s %s", code, stderr.String(), stdout.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if result.Source == "" || p == nil || p.Search.Evaluated != 64 || p.Search.Selection.ModelCalls != 6 ||
		p.FunctionalCompleteness != 600.0/7 || len(p.Feedback) != 2 || !p.Feedback[0].ContextDeclined ||
		!p.Feedback[1].ContextDeclined || p.Feedback[0].ModelCalls != 0 || p.Feedback[1].ModelCalls != 0 {
		t.Fatal("CLI did not retain zero-call declines and verified partial body")
	}
}

func solePathFeedbackReader(t *testing.T, intent string) mapSourceReader {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-conditional-assignment.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/typed-path-conditional-assignment-ko-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	var document pathplan.Document
	if err = json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	document.Plan.Decisions = document.Plan.Decisions[:1]
	document.Plan.Decisions[0].Intent = intent
	document.MaxAttempts = 2
	document.TestCases = []pathplan.TestCase{{Input: 3, Expected: 999}}
	raw, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return mapSourceReader{"fixture.gooo": source, "plan.json": raw,
		"hint.json": []byte(`{"source_sha":"` + strings.Repeat("a", 40) + `","status":"FAIL"}`)}
}

func TestTypedPathFeedbackCLISoleRemainingCandidateNeedsNoPrediction(t *testing.T) {
	for _, intent := range []string{"Use reverse operand order.", "피연산자 순서를 반대로 사용해라."} {
		t.Run(intent, func(t *testing.T) {
			reader := solePathFeedbackReader(t, intent)
			args := []string{"--json", "--path-plan", "plan.json", "--path-model", writeCLIPathFeedbackModel(t),
				"--path-step-attempts", "1", "--activity", "ConditionalAssign"}
			var baseline bodycodegen.Result
			for _, feedback := range []bool{false, true} {
				flags := append([]string(nil), args...)
				if feedback {
					flags = append(flags, "--path-feedback-rounds", "1", "--path-feedback-ci", "hint.json")
				}
				flags = append(flags, "fixture.gooo")
				var out, stderr bytes.Buffer
				if code := runBodyCodegen(flags, reader, &out, &stderr); code != exitOK || stderr.Len() != 0 {
					t.Fatalf("CLI %d %s %s", code, out.String(), stderr.String())
				}
				var result bodycodegen.Result
				if err := json.Unmarshal(out.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				p := result.Report.BodyPaths
				if p == nil || p.Search.Evaluated != 2 || p.Search.DeclaredCombinations != 2 || p.Search.Selection.ModelCalls != 1 || p.Search.Selection.ExternalCalls != 0 || p.FunctionalCompleteness != 0 || result.Source == "" || result.Report.RepositoryWrites != 0 {
					t.Fatal("finite partial outcome or call accounting lost")
				}
				if !feedback {
					baseline = result
					continue
				}
				if len(p.Feedback) != 1 {
					t.Fatal("zero-call receipt missing")
				}
				f := p.Feedback[0]
				if !f.RankingUnnecessary || f.ModelCalls != 0 || f.CumulativeCalls != 1 || f.Applied || f.ContextDeclined || len(f.Judgments) != 0 || f.FirstFailure == nil || f.CI == nil || f.CI.Status != "FAIL" || f.CIIsAuthority || f.SHA == "" || f.FromProgressSHA == "" {
					t.Fatal("sole candidate receipt differs")
				}
				if result.Source != baseline.Source || p.Search.SelectedTrainingPassed != baseline.Report.BodyPaths.Search.SelectedTrainingPassed {
					t.Fatal("zero-call continuation changed selected body")
				}
			}
		})
	}
}

func TestTypedPathFeedbackCLIRejectsAmbiguousModes(t *testing.T) {
	base := []string{"--path-plan", "plan.json", "--path-model", "model.json", "--path-step-attempts", "8"}
	for _, flags := range [][]string{
		{"--path-feedback-rounds", "1"},
		{"--path-plan", "plan.json", "--path-feedback-rounds", "1"},
		{"--path-plan", "plan.json", "--path-model", "model.json", "--path-feedback-rounds", "1"},
		append(append([]string{}, base...), "--path-feedback-rounds", "0"),
		append(append([]string{}, base...), "--path-feedback-rounds", "17"),
		append(append([]string{}, base...), "--path-feedback-rounds", "NaN"),
		append(append([]string{}, base...), "--path-feedback-rounds", "1", "--path-feedback-rounds", "2"),
		append(append([]string{}, base...), "--path-feedback-ci", "hint.json"),
		append(append([]string{}, base...), "--path-feedback-rounds", "1", "--path-feedback-ci"),
		append(append([]string{}, base...), "--path-feedback-rounds", "1", "--path-feedback-ci", "--json"),
		append(append([]string{}, base...), "--path-feedback-rounds", "1", "--path-feedback-ci", "hint.json", "--path-feedback-ci", "hint.json"),
	} {
		args := append(flags, "--activity", "Combined", "fixture.gooo")
		var stdout, stderr bytes.Buffer
		if code := runBodyCodegen(args, mapSourceReader{}, &stdout, &stderr); code != exitUsage {
			t.Fatalf("ambiguous feedback accepted: %v: %d %s", flags, code, stderr.String())
		}
	}
}
