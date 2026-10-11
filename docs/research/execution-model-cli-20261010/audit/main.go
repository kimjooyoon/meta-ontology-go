// Read saved records only. No model, compilation, evaluator or native execution.
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const producer = "9c7d5bc76735eb33db032646082240eb16a1cefc"
const model = "7827d36a64ce72ee9d880f7e6bdb9b29f3cfba34d02d435781df4bf034dd789c"
const fingerprint = "0d4e0ca1742d49d9b2af2b7e078ca7875604ee776e8201c5f318323d36231f6c"

func must(ok bool, message string) {
	if !ok {
		panic(message)
	}
}

func read(root, name string) []byte {
	f, err := os.Open(filepath.Join(root, name))
	must(err == nil, name+": open")
	defer f.Close()
	var r io.Reader = f
	if strings.HasSuffix(name, ".gz") {
		z, err := gzip.NewReader(f)
		must(err == nil, name+": gzip")
		defer z.Close()
		r = z
	}
	b, err := io.ReadAll(r)
	must(err == nil, name+": read")
	return b
}

func decode[T any](b []byte) T {
	var out T
	d := json.NewDecoder(strings.NewReader(string(b)))
	d.UseNumber()
	must(d.Decode(&out) == nil, "JSON decode")
	var extra any
	must(d.Decode(&extra) == io.EOF, "trailing JSON")
	return out
}

func digest(b []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(b)) }

func checkSearch(r bodycodegen.Result, attempts, calls int, sourceSHA string) *bodycodegen.BodyPathReceipt {
	p := r.Report.BodyPaths
	must(p != nil && r.Report.CompilerSourceSHA == producer, "generation producer")
	must(p.OriginalSourceSHA256 == sourceSHA && p.SourceBaseMatched, "source binding")
	must(p.Search.Status == "TRAINING_COMPLETE" && len(p.Search.Attempts) == attempts, "search attempts")
	must(p.Search.Selection.ModelCalls == calls && p.Search.Selection.ExternalCalls == 0, "calls")
	must(p.Search.TrainingTotal == 6 && p.Search.SelectedTrainingPassed == 6, "source outputs")
	must(p.Conditions != nil && p.Conditions.Declared == 3 && p.Conditions.Passed == 3, "conditions")
	must(p.Conditions.SelectedTreeSHA256 == p.Conditions.EmittedTreeSHA256, "emitted condition tree")
	seen := map[uint16]bool{}
	for _, a := range p.Search.Attempts {
		must(!seen[a.Mask], "repeated candidate")
		seen[a.Mask] = true
	}
	for _, row := range p.NativeCases {
		must(row.Passed && row.Actual == row.Expected, "emitted finite output")
	}
	return p
}

type constructionOutput struct {
	GeneratedNow bool                            `json:"generated_now"`
	Construction bodyexecution.JointConstruction `json:"construction"`
	Evaluation   bodyexecution.JointEvaluation   `json:"evaluation"`
}

func checkNative(r bodyexecution.CompositionRuntime, suite bodyexecution.CompositionCases) {
	must(r.Stage == "COMPLETE" && r.ProducerSourceSHA == producer, "native producer")
	must(r.GoVersion == "go version go1.27.2 darwin/arm64", "native toolchain")
	must(r.RuntimeReplayed && r.ProjectionReplayed && len(r.Runs) == 2, "native replay")
	must(r.FinitePassed == 8 && r.FiniteTotal == 8 && len(r.Traces) == 8, "native cases")
	for _, run := range r.Runs {
		must(run.Started && run.Completed && run.ExitCode != nil && *run.ExitCode == 0 && !run.Canceled && !run.TimedOut, "native process")
	}
	for i, trace := range r.Traces {
		must(trace.CaseIndex == i && len(trace.Deliveries) == 1, "native trace")
		delivery := trace.Deliveries[0]
		input := decode[int64](delivery.Input)
		expected := decode[int64](delivery.Expected)
		must(input == decode[int64](suite.Cases[i].Inputs["Main"]), "exact input")
		must(expected == decode[int64](suite.Cases[i].Expected["Main"]) && decode[int64](delivery.Actual) == expected, "exact output")
		must(delivery.Passed != nil && *delivery.Passed, "passed output")
	}
}

func audit(root string) map[string]any {
	build := decode[map[string]any](read(root, "producer-version.json"))
	must(build["compiler_source_sha"] == producer && build["source_status"] == "CLEAN_VCS" && build["go_version"] == "go1.27.2", "clean producer")
	sdk := build["decision_runtime"].(map[string]any)
	must(sdk["version"] == "v0.2.32-experimental" && sdk["replaced"] == false, "public SDK")
	must(strings.TrimSpace(string(read(root, "model.sha256"))) == model, "model bytes identity")
	must(string(read(root, "completed-commands.txt")) == "preflight\ndeterministic\ninitial\nfeedback\nconstruction\nreplay\n", "six completed original commands")
	sourceSHA := digest(read(root, "source.gooo.gz"))
	pre := decode[bodycodegen.TypedPathContextExport](read(root, "preflight.json.gz"))
	must(pre.OriginalSourceSHA256 == sourceSHA && pre.ModelPredictions == 0 && pre.CandidateTests == 0 && len(pre.Inputs) == 2, "preflight")
	d := checkSearch(decode[bodycodegen.Result](read(root, "deterministic.json.gz")), 3, 0, sourceSHA)
	initial := checkSearch(decode[bodycodegen.Result](read(root, "initial.json.gz")), 1, 1, sourceSHA)
	feedback := checkSearch(decode[bodycodegen.Result](read(root, "feedback.json.gz")), 1, 1, sourceSHA)
	for _, p := range []*bodycodegen.BodyPathReceipt{initial, feedback} {
		must(p.ModelContext.ArtifactSHA == "sha256:"+model && p.ModelContext.ModelFingerprint == "sha256:"+fingerprint, "model identities")
		must(len(p.ConditionFeedback) == 0 && len(p.Search.Attempts) == 1, "no extra feedback call on first success")
		first := p.ConditionProgress[0].Ranking
		must(first.FeatureVersion == decision.ExecutionFeatureVersion && first.ModelFingerprint == fingerprint && first.Proposed == 2 && first.Calls == 1 && first.PredictNS > 0, "initial judgment")
		must(first.OutputFailure == nil && !first.HasFailure, "future observation absent")
		for i, input := range pre.Inputs {
			must(input.Features == nil && input.ExecutionFeatures != nil && input.Bytes == 1280, "320 features")
			var raw [1280]byte
			for j, value := range input.ExecutionFeatures {
				binary.LittleEndian.PutUint32(raw[j*4:], math.Float32bits(value))
				if j >= 256 {
					must(value == 0, "future output feature")
				}
			}
			must(input.InputSHA == digest(raw[:]) && input.InputSHA == "sha256:"+first.FeatureSHA[i], "actual neural input")
			must(digest([]byte(input.Text)) == input.OriginalIntentSHA, "authored intent")
		}
	}
	suite := decode[bodyexecution.CompositionCases](read(root, "evaluation-cases.json"))
	built := decode[constructionOutput](read(root, "construction.json.gz"))
	replayed := decode[constructionOutput](read(root, "replay.json.gz"))
	must(built.GeneratedNow && !replayed.GeneratedNow && replayed.Evaluation.ConstructionReplayed && replayed.Evaluation.NewModelCalls == 0, "model-free saved replay")
	must(reflect.DeepEqual(built.Construction, replayed.Construction), "unchanged saved construction")
	saved := decode[bodyexecution.JointConstruction](read(root, "saved-construction.json.gz"))
	savedJSON, savedErr := json.Marshal(saved)
	builtJSON, builtErr := json.Marshal(built.Construction)
	must(savedErr == nil && builtErr == nil && bytes.Equal(savedJSON, builtJSON), "saved file matches stdout after JSON whitespace normalization")
	checkNative(built.Evaluation.Runtime, suite)
	checkNative(replayed.Evaluation.Runtime, suite)
	return map[string]any{"status": "PASS", "scope": "Read-only audit of six original commands; no new inference, evaluator or native process.", "producer": producer,
		"deterministic_attempts": len(d.Search.Attempts), "model_attempts": len(initial.Search.Attempts), "feedback_additional_calls": 0,
		"initial_predict_ns": initial.ConditionProgress[0].Ranking.PredictNS, "feedback_mode_initial_predict_ns": feedback.ConditionProgress[0].Ranking.PredictNS,
		"source_outputs": 6, "source_conditions": 3, "native_cases": 8, "new_audit_model_calls": 0, "new_audit_native_runs": 0}
}

func main() {
	must(len(os.Args) == 2, "usage: audit RESULT_DIRECTORY")
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	must(e.Encode(audit(os.Args[1])) == nil, "summary")
}
