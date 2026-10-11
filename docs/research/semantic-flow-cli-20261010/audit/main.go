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

const producer = "449b45c103e23903e5839f38f2ba42e945d8f993"
const model = "8127de6776d13b9c06710d1da762bb254401c1d8ff7e26ee6bb3c2d50d5cb8db"
const fingerprint = "eb53c0fdef003f151aed582c56ce285947f70361b25199dba2eb882f0d14a16a"

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
	must(p.Search.TrainingTotal == 8 && p.Search.SelectedTrainingPassed == 8, "source outputs")
	must(p.Conditions != nil && p.Conditions.Declared == 3 && p.Conditions.Passed == 3, "conditions")
	must(p.Conditions.SelectedTreeSHA256 == p.Conditions.EmittedTreeSHA256, "emitted condition tree")
	seen := map[uint16]bool{}
	for _, a := range p.Search.Attempts {
		must(!seen[a.Mask], "repeated candidate")
		seen[a.Mask] = true
	}
	must(len(p.NativeCases) == 8, "all emitted finite outputs")
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

func checkSavedConstruction(root string) {
	suite := decode[bodyexecution.CompositionCases](read(root, "evaluation-cases.json"))
	built := decode[constructionOutput](read(root, "construction.json.gz"))
	replayed := decode[constructionOutput](read(root, "replay.json.gz"))
	must(built.GeneratedNow && !replayed.GeneratedNow && replayed.Evaluation.ConstructionReplayed && replayed.Evaluation.NewModelCalls == 0, "model-free saved replay")
	retained := built.Construction.Initial.Preparations[0].Generation.Report.BodyPaths.ModelRetention
	checkSearch(built.Construction.Initial.Preparations[0].Generation, 2, 1, digest(read(root, "source.gooo.gz")))
	must(retained.FeatureVersion == decision.SemanticFlowFeatureVersion && retained.ResidentTensorBytes == 37160 && retained.ModelFingerprint == "sha256:"+fingerprint, "retained flow weights")
	must(reflect.DeepEqual(built.Construction, replayed.Construction), "unchanged saved construction")
	saved := decode[bodyexecution.JointConstruction](read(root, "saved-construction.json.gz"))
	savedJSON, savedErr := json.Marshal(saved)
	builtJSON, builtErr := json.Marshal(built.Construction)
	must(savedErr == nil && builtErr == nil && bytes.Equal(savedJSON, builtJSON), "saved file matches stdout after JSON whitespace normalization")
	checkNative(built.Evaluation.Runtime, suite)
	checkNative(replayed.Evaluation.Runtime, suite)
	must(built.Evaluation.Runtime.Build.Started && !replayed.Evaluation.Runtime.Build.Started, "saved executable reused")
}

func checkInputs(pre bodycodegen.TypedPathContextExport, p *bodycodegen.BodyPathReceipt) {
	must(reflect.DeepEqual(pre.Context, p.ModelContext), "preflight and actual context")
	must(p.ModelContext.ArtifactSHA == "sha256:"+model && p.ModelContext.ModelFingerprint == "sha256:"+fingerprint, "exact model")
	must(p.ModelContext.FeatureVersion == decision.SemanticFlowFeatureVersion, "v5 context")
	for i, input := range pre.Inputs {
		must(input.FlowFeatures != nil && input.Features == nil && input.ExecutionFeatures == nil && input.Bytes == 1536, "v5 array")
		must(input.SemanticFlow != nil && input.SemanticFlow.Normalized && input.SemanticFlow.Reason == "SINGLE_BRANCH_KNOWN_ATOMS", "source normalization")
		must(input.SourceFeatureSHA == digest(input.SemanticFlow.Source[:]), "normalized source digest")
		must(input.FlowFeatures[255] == 2, "normalized input marker")
		var raw [1536]byte
		for j, x := range input.FlowFeatures {
			binary.LittleEndian.PutUint32(raw[j*4:], math.Float32bits(x))
		}
		must(digest(raw[:]) == input.InputSHA && input.InputSHA == "sha256:"+p.ConditionProgress[0].Ranking.FeatureSHA[i], "actual array identity")
		for _, segment := range [][]float32{input.FlowFeatures[192:236], input.FlowFeatures[256:320]} {
			for _, v := range segment {
				must(v == 0, "future observation in initial input")
			}
		}
		must(digest([]byte(input.Text)) == input.OriginalIntentSHA, "source intent")
	}
}

func checkForm(root string) (bodycodegen.TypedPathContextExport, map[string]any) {
	sourceSHA := digest(read(root, "source.gooo.gz"))
	pre := decode[bodycodegen.TypedPathContextExport](read(root, "preflight.json.gz"))
	must(pre.OriginalSourceSHA256 == sourceSHA && pre.ModelPredictions == 0 && pre.CandidateTests == 0 && len(pre.Inputs) == 2, "preflight")
	d := checkSearch(decode[bodycodegen.Result](read(root, "deterministic.json.gz")), 1, 0, sourceSHA)
	initial := checkSearch(decode[bodycodegen.Result](read(root, "initial.json.gz")), 2, 1, sourceSHA)
	fb := checkSearch(decode[bodycodegen.Result](read(root, "feedback.json.gz")), 2, 2, sourceSHA)
	must(d.Search.Attempts[0].Mask == 0, "baseline succeeds first")
	for _, p := range []*bodycodegen.BodyPathReceipt{initial, fb} {
		must(p.Search.Attempts[0].Mask == 2 && p.Search.Attempts[1].Mask == 0, "observed model failure then continuation")
		checkInputs(pre, p)
	}
	must(len(fb.ConditionFeedback) == 1, "one actual feedback call")
	f := fb.ConditionFeedback[0]
	must(f.Calls == 1 && f.OutputFailure != nil && !f.HasFailure && f.Proposed == 2 && f.Applied && !f.AddedMask, "repeated highest proposal")
	must(f.OutputFailure.Result.Actual == -9007199254740995 && f.OutputFailure.Result.Expected == 11, "exact output failure")
	checkSavedConstruction(root)
	return pre, map[string]any{"deterministic_attempts": 1, "initial_attempts": 2, "feedback_attempts": 2,
		"deterministic_total_ms": d.Timing.TotalMS, "initial_total_ms": initial.Timing.TotalMS, "feedback_total_ms": fb.Timing.TotalMS,
		"initial_predict_ns": initial.ConditionProgress[0].Ranking.PredictNS, "feedback_initial_predict_ns": fb.ConditionProgress[0].Ranking.PredictNS, "feedback_predict_ns": f.PredictNS,
		"source_outputs": 8, "source_conditions": 3, "construction_native_cases": 8, "model_free_replay_cases": 8, "repeated_candidate_executions": 0}
}

func audit(root string) map[string]any {
	build := decode[map[string]any](read(root, "producer-version.json"))
	must(build["compiler_source_sha"] == producer && build["source_status"] == "CLEAN_VCS" && build["go_version"] == "go1.27.2", "clean producer")
	sdk := build["decision_runtime"].(map[string]any)
	must(sdk["version"] == "v0.2.34-experimental" && sdk["replaced"] == false, "public SDK")
	must(strings.TrimSpace(string(read(root, "model.sha256"))) == model, "public model hash")
	var completed strings.Builder
	for _, form := range []string{"direct", "copy"} {
		for _, name := range []string{"preflight", "deterministic", "initial", "feedback", "construction", "replay"} {
			fmt.Fprintf(&completed, "%s\t%s\n", form, name)
		}
	}
	must(string(read(root, "completed-commands.txt")) == completed.String(), "twelve original commands")
	a, ar := checkForm(filepath.Join(root, "direct"))
	b, br := checkForm(filepath.Join(root, "copy"))
	must(a.OriginalSourceSHA256 != b.OriginalSourceSHA256, "distinct source spelling")
	for i, input := range a.Inputs {
		must(input.InputSHA == b.Inputs[i].InputSHA && *input.FlowFeatures == *b.Inputs[i].FlowFeatures, "same actual model arrays")
	}
	return map[string]any{"status": "PASS", "producer": producer, "direct": ar, "copy": br, "paired_choices": 2, "native_processes": 8, "new_audit_model_calls": 0, "new_audit_native_runs": 0}
}

func main() {
	must(len(os.Args) == 2, "usage: audit RESULT_DIRECTORY")
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	must(e.Encode(audit(os.Args[1])) == nil, "summary")
}
