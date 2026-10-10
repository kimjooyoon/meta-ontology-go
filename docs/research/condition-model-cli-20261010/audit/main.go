// Read the saved observation only: no model loading, inference, or native run.
package main

import (
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

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const producer = "35851648ee237837841905bb4c35c35db3298f00"
const model = "a16696ed44c668f38cc2e7ee1dc6ff3f4716649d57e59df7c9490d1c670edc48"
const fingerprint = "456a3528de6abbcfc9ef63feb66264d59e31d7807b54e38f3a6ce08f0daf278a"

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
	must(p.Search.Selection.ModelCalls == calls && p.Search.Selection.ExternalCalls == 0, "call counts")
	must(p.Search.TrainingTotal == 9 && p.Search.SelectedTrainingPassed == 9, "source output cases")
	must(p.Conditions != nil && p.Conditions.Declared == 3 && p.Conditions.Passed == 3, "source conditions")
	must(p.Conditions.SelectedTreeSHA256 == p.Conditions.EmittedTreeSHA256, "selected/emitted condition tree")
	seen := map[uint16]bool{}
	for _, a := range p.Search.Attempts {
		must(!seen[a.Mask], "repeated mask")
		seen[a.Mask] = true
	}
	for _, c := range p.NativeCases {
		must(c.Passed && c.Actual == c.Expected, "emitted finite case")
	}
	return p
}

func main() {
	must(len(os.Args) == 2, "usage: audit RESULT_DIRECTORY")
	root := os.Args[1]
	build := decode[map[string]any](read(root, "build.json"))
	must(build["compiler_source_sha"] == producer && build["vcs_modified"] == "false", "clean producer")
	must(build["go_version"] == "go1.27.2", "Go version")
	sdk := build["decision_runtime"].(map[string]any)
	must(sdk["version"] == "v0.2.30-experimental" && sdk["replaced"] == false, "public SDK")
	must(strings.TrimSpace(string(read(root, "model.sha256"))) == model, "model file identity")
	sourceSHA := digest(read(root, "source.gooo"))
	pre := decode[bodycodegen.TypedPathContextExport](read(root, "preflight.json.gz"))
	must(pre.OriginalSourceSHA256 == sourceSHA && pre.ModelPredictions == 0 && pre.CandidateTests == 0, "preflight")
	must(pre.Context != nil && len(pre.Inputs) == 2 && len(pre.Context.Inputs) == 2, "preflight features")
	d := checkSearch(decode[bodycodegen.Result](read(root, "deterministic.json.gz")), 4, 0, sourceSHA)
	r := checkSearch(decode[bodycodegen.Result](read(root, "condition.json.gz")), 2, 2, sourceSHA)
	must(r.ModelContext != nil && r.ModelContext.ArtifactSHA == "sha256:"+model, "artifact identity")
	must(r.ModelContext.ModelFingerprint == "sha256:"+fingerprint, "semantic model identity")
	must(len(r.ConditionProgress) == 4 && len(r.ConditionFeedback) == 1, "observed feedback count")
	first, feedback := r.ConditionProgress[0].Ranking, r.ConditionFeedback[0]
	must(first.ModelFingerprint == fingerprint && feedback.ModelFingerprint == fingerprint, "same weights")
	must(!first.HasFailure && first.Attempted == 0 && first.Calls == 1 && first.Proposed == 0, "initial judgment")
	for i, input := range pre.Inputs {
		must(input.Features != nil, "feature array")
		for _, value := range input.Features[192:] {
			must(value == 0, "no future condition observation in initial input")
		}
		var b [1024]byte
		for j, value := range input.Features {
			binary.LittleEndian.PutUint32(b[j*4:], math.Float32bits(value))
		}
		sha := digest(b[:])
		must(sha == input.InputSHA && sha == "sha256:"+first.FeatureSHA[i], "exported/actual input bytes")
		must(sha == pre.Context.Inputs[i].InputSHA && sha == r.ModelContext.Inputs[i].InputSHA, "context input identity")
		must(digest([]byte(input.Text)) == input.OriginalIntentSHA, "original full intent")
	}
	must(feedback.HasFailure && feedback.Attempted == 1 && feedback.Calls == 1 && feedback.Proposed == 3, "feedback")
	must(reflect.DeepEqual(feedback.Failure.Result, r.Search.Attempts[0].Conditions[0]), "actual failed condition")
	must(feedback.Failure.Result.Case.Input == -9007199254740995, "exact condition int64")
	must(!feedback.Failure.Result.Case.Expected && feedback.Failure.Result.Observation.Value, "Boolean mismatch")
	must(first.PredictNS > 0 && feedback.PredictNS > 0, "observed model timing")
	excerpt := decode[struct {
		OriginalSHA string                    `json:"original_native_sha256"`
		Omitted     []string                  `json:"omitted_fields"`
		Observation bodyexecution.Observation `json:"observation"`
	}](read(root, "native-observation.json"))
	must(len(excerpt.OriginalSHA) == 71 && reflect.DeepEqual(excerpt.Omitted, []string{"observation.go_tool_path", "parent_receipt_bytes", "completeness_receipt"}), "explicit excerpt scope")
	n := excerpt.Observation
	must(n.Stage == "COMPLETE" && n.ProducerSourceSHA == producer && n.OriginalSourceSHA256 == sourceSHA, "native identity")
	must(n.GoToolPath == "" && n.GoVersion == "go version go1.27.2 darwin/arm64", "native toolchain")
	must(n.RuntimeReplayed && n.ProjectionReplayed && len(n.Runs) == 2, "actual saved replay")
	for _, p := range append([]bodyexecution.ProcessObservation{n.Toolchain, n.Build}, n.Runs...) {
		must(p.Started && p.Completed && p.ExitCode != nil && *p.ExitCode == 0 && !p.Canceled && !p.TimedOut, "actual process completion")
	}
	suite := decode[struct {
		Cases []pathplan.TestCase `json:"cases"`
	}](read(root, "runtime-cases.json")).Cases
	must(len(suite) == 10 && len(n.Cases) == 10 && n.DeclaredCases == 10, "runtime count")
	selectionInputs := map[int64]bool{}
	for _, c := range d.Search.Attempts[0].Results {
		selectionInputs[c.Input] = true
	}
	disjoint := 0
	for i, c := range n.Cases {
		must(c.Input == suite[i].Input && c.Expected == suite[i].Expected && c.Actual == c.Expected && c.Passed, "exact native int64")
		if !selectionInputs[c.Input] {
			disjoint++
		}
	}
	must(disjoint == 4 && n.SelectionDisjointInputs == disjoint, "source-disjoint input count")
	summary := map[string]any{"status": "PASS", "scope": "Read-only recount of saved observations; native excerpt omits the local tool path and parent/completeness payloads. No new inference or execution.",
		"producer_source_sha": producer, "new_model_calls": 0, "new_native_runs": 0,
		"deterministic_attempts": 4, "condition_attempts": 2, "observed_model_calls": 2,
		"initial_predict_ns": first.PredictNS, "feedback_predict_ns": feedback.PredictNS,
		"source_outputs_passed": 9, "source_conditions_passed": 3,
		"native_cases_passed": 10, "selection_disjoint_inputs": disjoint, "observed_native_runs": 2}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	must(e.Encode(summary) == nil, "summary")
}
