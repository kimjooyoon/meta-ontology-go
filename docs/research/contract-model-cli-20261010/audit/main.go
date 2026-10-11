// audit reads the original records; it never predicts, compiles or runs a body.
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
	"regexp"
	"strconv"
	"strings"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const producer = "4d4accf51ff8227293365ca92e4f3ddf2a5e8c57"
const fingerprint = "7533d7eb6889e03f6b9901be75dd66777a945a693f061b40af1f550015826254"

var names = []string{"bound-k7-r0-w0-direct-g0", "bound-k7-r0-w0-direct-g1",
	"offset-k11-r1-w0-assignment-g0", "offset-k11-r1-w0-assignment-g1"}

func must(ok bool, message string) {
	if !ok {
		panic(message)
	}
}

func read(path string) []byte {
	f, err := os.Open(path)
	must(err == nil, "open "+path)
	defer f.Close()
	var reader io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		z, err := gzip.NewReader(f)
		must(err == nil, "gzip "+path)
		defer z.Close()
		reader = z
	}
	raw, err := io.ReadAll(reader)
	must(err == nil, "read "+path)
	return raw
}

func decode[T any](path string) T {
	var value T
	d := json.NewDecoder(strings.NewReader(string(read(path))))
	d.UseNumber()
	must(d.Decode(&value) == nil, "decode "+path)
	var extra any
	must(d.Decode(&extra) == io.EOF, "extra JSON")
	return value
}

func digest(raw []byte) string { return fmt.Sprintf("%x", sha256.Sum256(raw)) }

func jsonSHA(v any) string {
	raw, err := json.Marshal(v)
	must(err == nil, "marshal")
	return digest(raw)
}

func featureBytes(values []float32) []byte {
	raw := make([]byte, len(values)*4)
	for i, v := range values {
		binary.LittleEndian.PutUint32(raw[i*4:], math.Float32bits(v))
	}
	return raw
}

func checkSearch(r bodycodegen.Result, source []byte, calls int) *bodycodegen.BodyPathReceipt {
	p := r.Report.BodyPaths
	must(p != nil && r.Report.CompilerSourceSHA == producer, "compiler source")
	must(p.OriginalSourceSHA256 == "sha256:"+digest(source) && p.SourceBaseMatched, "source binding")
	must(p.Search.Status == "TRAINING_COMPLETE" && p.Search.SelectedTrainingPassed == 8 && p.Search.TrainingTotal == 8, "finite complete")
	must(p.Search.Selection.ModelCalls == calls && p.Search.Selection.ExternalCalls == 0, "model calls")
	must(p.Conditions != nil && p.Conditions.Declared == 3 && p.Conditions.Passed == 3, "source conditions")
	must(len(p.NativeCases) == 8 && len(p.Search.Attempts) > 0, "finite outputs")
	seen := map[uint16]bool{}
	for _, a := range p.Search.Attempts {
		must(!seen[a.Mask], "repeated candidate")
		seen[a.Mask] = true
	}
	for _, c := range p.NativeCases {
		must(c.Passed && c.Actual == c.Expected, "exact emitted case")
	}
	return p
}

func checkInputs(pre bodycodegen.TypedPathContextExport, p *bodycodegen.BodyPathReceipt, doc pathplan.Document) {
	must(pre.ModelPredictions == 0 && pre.CandidateTests == 0 && pre.ContractCases != nil, "no preflight execution")
	must(reflect.DeepEqual(pre.Context, p.ModelContext), "preflight context differs")
	r := p.ContractRanking
	must(r != nil && r.ModelFingerprint == fingerprint && r.Calls == 1 && r.Applied && !r.Declined, "contract ranking")
	must(r.CaseSHA == jsonSHA(doc.TestCases) && r.CaseSHA == pre.ContractCases.CaseSHA && r.CaseCount == 8, "all declared cases")
	var caseBytes []byte
	for i, c := range doc.TestCases {
		must(p.NativeCases[i].Input == c.Input && p.NativeCases[i].Expected == c.Expected && p.NativeCases[i].Actual == c.Expected, "emitted output differs from declared int64")
		var want [decision.DeclaredCaseFeatureDim]float32
		must(decision.DeclaredCaseFeaturesInto(c.Input, c.Expected, &want) == nil && want == pre.ContractCases.Features[i], "exact case features")
		caseBytes = append(caseBytes, featureBytes(want[:])...)
	}
	must(pre.ContractCases.FeatureSHA == "sha256:"+digest(caseBytes), "case feature digest")
	for i, input := range pre.Inputs {
		must(input.FlowFeatures != nil && input.InputSHA == "sha256:"+r.FeatureSHA[i], "actual source feature digest")
		must(digest(featureBytes(input.FlowFeatures[:])) == r.FeatureSHA[i], "source array bytes")
	}
	checkProgress(p)
}

func checkProgress(p *bodycodegen.BodyPathReceipt) {
	r := *p.ContractRanking
	r.SHA = ""
	must(jsonSHA(r) == p.ContractRanking.SHA, "ranking digest")
	var attempts []pathplan.SearchAttempt
	previous := ""
	for i, progress := range p.ContractProgress {
		must(progress.Sequence == i+1 && progress.RankingSHA == p.ContractRanking.SHA && progress.PreviousSHA == previous, "progress links")
		must(progress.PredictionsThisAdvance == 0 && progress.Selection.ModelCalls == 1, "additional predictions")
		raw := progress
		raw.SHA = ""
		must(jsonSHA(raw) == progress.SHA, "contract progress digest")
		core := progress.SessionProgress
		previous, core.SHA = core.SHA, ""
		must(jsonSHA(core) == previous, "ordinary progress digest")
		attempts = append(attempts, progress.NewAttempts...)
	}
	must(len(p.ContractProgress) >= 2 && reflect.DeepEqual(attempts, p.Search.Attempts), "complete attempt chain")
}

type constructionOutput struct {
	GeneratedNow bool                            `json:"generated_now"`
	Construction bodyexecution.JointConstruction `json:"construction"`
	Evaluation   bodyexecution.JointEvaluation   `json:"evaluation"`
}

func checkNative(r bodyexecution.CompositionRuntime, cases bodyexecution.CompositionCases) {
	must(r.Stage == "COMPLETE" && r.ProducerSourceSHA == producer && r.FinitePassed == 8 && r.FiniteTotal == 8, "native completion")
	must(len(r.Traces) == 8 && len(r.Runs) == 2 && r.RuntimeReplayed && r.ProjectionReplayed, "native independent runs")
	for _, run := range r.Runs {
		must(run.Started && run.Completed && run.ExitCode != nil && *run.ExitCode == 0 && !run.Canceled && !run.TimedOut, "native exit")
	}
	for i, trace := range r.Traces {
		must(trace.CaseIndex == i && len(trace.Deliveries) == 1, "native trace")
		d := trace.Deliveries[0]
		var input, expected, actual, declaredInput, declaredExpected int64
		must(json.Unmarshal(d.Input, &input) == nil && json.Unmarshal(d.Expected, &expected) == nil && json.Unmarshal(d.Actual, &actual) == nil, "exact native int64")
		must(json.Unmarshal(cases.Cases[i].Inputs["Main"], &declaredInput) == nil && json.Unmarshal(cases.Cases[i].Expected["Main"], &declaredExpected) == nil, "exact declared int64")
		must(input == declaredInput && actual == declaredExpected && expected == declaredExpected && d.Passed != nil && *d.Passed, "native exact output")
	}
}

func auditNative(root string, source []byte, pre bodycodegen.TypedPathContextExport, doc pathplan.Document) {
	built := decode[constructionOutput](filepath.Join(root, "construction.json.gz"))
	replay := decode[constructionOutput](filepath.Join(root, "replay.json.gz"))
	suite := decode[bodyexecution.CompositionCases](filepath.Join(root, "evaluation-cases.json"))
	must(built.GeneratedNow && !replay.GeneratedNow && replay.Evaluation.ConstructionReplayed && replay.Evaluation.NewModelCalls == 0, "model-free replay")
	p := checkSearch(built.Construction.Initial.Preparations[0].Generation, source, 1)
	checkInputs(pre, p, doc)
	must(p.ModelRetention.ModelSchema == "gooo/contract-candidate-decision/v1" && p.ModelRetention.ResidentTensorBytes == 38984, "retained model")
	must(reflect.DeepEqual(built.Construction, replay.Construction), "saved construction changed")
	saved := decode[bodyexecution.JointConstruction](filepath.Join(root, "saved-construction.json.gz"))
	must(jsonSHA(saved) == jsonSHA(built.Construction), "saved file differs from stdout after JSON whitespace normalization")
	checkNative(built.Evaluation.Runtime, suite)
	checkNative(replay.Evaluation.Runtime, suite)
	must(built.Evaluation.Runtime.Build.Started && !replay.Evaluation.Runtime.Build.Started, "cached native replay")
}

func resources(root string) map[string]any {
	var wallNS int64
	var seconds [3]float64
	var rss int64
	times := regexp.MustCompile(`(?m)^(real|user|sys) ([0-9.]+)$`)
	memory := regexp.MustCompile(`([0-9]+)\s+maximum resident set size`)
	for _, name := range []string{"preflight", "deterministic", "initial", "construction", "replay"} {
		command := decode[struct {
			Args   []string `json:"args"`
			Exit   int      `json:"exit_code"`
			WallNS int64    `json:"wall_ns"`
			Error  string   `json:"error"`
		}](filepath.Join(root, name+".command.json"))
		must(command.Exit == 0 && command.Error == "" && command.WallNS > 0 && len(command.Args) > 0, "original command completion")
		wallNS += command.WallNS
		text := string(read(filepath.Join(root, name+".time-stderr")))
		t, m := times.FindAllStringSubmatch(text, -1), memory.FindStringSubmatch(text)
		must(len(t) == 3 && len(m) == 2, "original resource measurements")
		for i, item := range t {
			must(item[1] == []string{"real", "user", "sys"}[i], "time order")
			v, err := strconv.ParseFloat(item[2], 64)
			must(err == nil && v >= 0, "process time")
			seconds[i] += v
		}
		bytes, err := strconv.ParseInt(m[1], 10, 64)
		must(err == nil && bytes > 0, "RSS")
		rss = max(rss, bytes)
	}
	return map[string]any{"sum_command_wall_ns": wallNS, "rounded_wall_user_system_seconds": seconds, "max_rss_bytes": rss}
}

func main() {
	must(len(os.Args) == 2, "usage: audit RESULT_DIRECTORY")
	rows := make([]map[string]any, 0, len(names))
	for _, name := range names {
		root := filepath.Join(os.Args[1], name)
		source := read(filepath.Join(root, "source.gooo.gz"))
		doc := decode[pathplan.Document](filepath.Join(root, "original-document.json.gz"))
		pre := decode[bodycodegen.TypedPathContextExport](filepath.Join(root, "preflight.json.gz"))
		d := checkSearch(decode[bodycodegen.Result](filepath.Join(root, "deterministic.json.gz")), source, 0)
		m := checkSearch(decode[bodycodegen.Result](filepath.Join(root, "initial.json.gz")), source, 1)
		checkInputs(pre, m, doc)
		auditNative(root, source, pre, doc)
		rows = append(rows, map[string]any{"source": name, "deterministic_attempts": len(d.Search.Attempts), "model_attempts": len(m.Search.Attempts),
			"deterministic_first_complete": len(d.Search.Attempts) == 1, "model_first_complete": len(m.Search.Attempts) == 1,
			"predict_ns": m.ContractRanking.PredictNS, "deterministic_timing": d.Timing, "model_timing": m.Timing, "resources": resources(root)})
	}
	must(json.NewEncoder(os.Stdout).Encode(map[string]any{"schema": "gooo/contract-cli-audit/v1", "producer": producer, "rows": rows,
		"commands": 20, "model_calls": 8, "native_processes": 16, "construction_cases_passed": 32, "saved_replay_cases_passed": 32,
		"audit_predictions": 0, "audit_native_executions": 0}) == nil, "write report")
}
