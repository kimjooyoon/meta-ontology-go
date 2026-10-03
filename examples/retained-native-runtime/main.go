// Compare one-shot and owned native execution with fresh generation on frozen
// development requests. Every generated body is immediately executed twice.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const frozenRecords = "47315bf031f4734497f1a9f3710e93b6c0e8f43ed385f78ec7d7cc51adfa2be1"
const frozenManifest = "e336e1a6dbb7493eee3075a362f2a65674a0a32045824c700bd66800b5eab49c"
const weights = "cf00ccc83d17d28ed73fcb869366151a48ffccd3aa8ca8e635aabf19810b9e78"

type prior struct {
	ID, Request, Arm, SourceSHA, PlanSHA, GenerationSHA, RuntimeSHA string
	Budget                                                          int
}
type record struct {
	ID, Request, Mode, GenerationSHA, RuntimeSHA                                         string
	Budget, Trial, ModelCalls, Passed, Total, NativeRuns                                 int
	Reused, Built                                                                        bool
	DecodeNS, GenerateNS, ExecuteNS, BuildNS, FirstRunNS, SecondRunNS, CurrentChildCPUNS int64
	MaxCurrentChildRSSBytes                                                              int64
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func require(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func read(path string) []byte { b, err := os.ReadFile(path); must(err); return b }
func hash(b []byte) string    { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func decode(b []byte, v any)  { must(json.Unmarshal(b, v)) }
func save(path string, v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	must(err)
	b = append(b, '\n')
	must(os.WriteFile(path, b, 0644))
	return hash(b)
}
func identity() string {
	b, ok := debug.ReadBuildInfo()
	require(ok && b.GoVersion == "go1.27.1", "Go 1.27.1 required")
	revision, clean, sdk := "", false, false
	for _, s := range b.Settings {
		if s.Key == "vcs.revision" {
			revision = s.Value
		}
		if s.Key == "vcs.modified" {
			clean = s.Value == "false"
		}
	}
	for _, d := range b.Deps {
		if d.Path == "github.com/kimjooyoon/gooo-decision-runtime" {
			sdk = d.Version == "v0.2.20-experimental" && d.Replace == nil
		}
	}
	require(clean && sdk && len(revision) == 40, "clean pinned compiler/SDK required")
	return revision
}
func compare(result, old bodycodegen.Result, revision string, p prior, reused bool) {
	r, before := result.Report.BodyPaths, old.Report.BodyPaths
	require(r != nil && before != nil && r.OrderJudgment != nil && before.OrderJudgment != nil, "ranking missing")
	require(result.Report.CompilerSourceSHA == revision && r.SourceBaseMatched && r.OriginalSourceSHA256 == p.SourceSHA, "source binding differs")
	require(result.Source == old.Source && reflect.DeepEqual(r.Search, before.Search), "source/search differs")
	a, b := *r.OrderJudgment, *before.OrderJudgment
	a.PredictNS, b.PredictNS = 0, 0
	require(reflect.DeepEqual(a, b), "ranking differs")
	c := r.OrderPreparation
	require(c != nil && c.Reused == reused && c.PlanSHA256 == p.PlanSHA && c.CandidateCount == 8, "preparation differs")
	require(r.Search.Selection.ModelCalls == 1 && r.Search.Selection.WeightsSHA256 == weights && r.Search.Selection.ExternalCallsKnown && r.Search.Selection.ExternalCalls == 0, "prediction accounting differs")
}
func observe(g *bodycodegen.TypedPathGenerator, owner *bodyexecution.Executor, out, revision, goBin string,
	p prior, mode string, trial int, source, recipe, suite []byte, old bodycodegen.Result, reference bodyexecution.Result) record {
	id := fmt.Sprintf("%s-%s-%d", p.ID, mode, trial)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r := record{ID: id, Request: p.Request, Budget: p.Budget, Mode: mode, Trial: trial}
	start := time.Now()
	doc, err := bodycodegen.DecodeSourcePathDocument(ctx, p.Request+".gooo", source, "Compose", recipe)
	must(err)
	r.DecodeNS = time.Since(start).Nanoseconds()
	start = time.Now()
	generated, err := g.Generate(ctx, p.Request+".gooo", source, "Compose", doc, bodycodegen.TypedPathOptions{})
	must(err)
	r.GenerateNS = time.Since(start).Nanoseconds()
	compare(generated, old, revision, p, trial == 1)
	r.ModelCalls = generated.Report.BodyPaths.Search.Selection.ModelCalls
	r.GenerationSHA = save(filepath.Join(out, id+"-generation.json"), generated)
	parent, err := json.Marshal(generated.Report.CompletenessReceipt)
	must(err)
	cases, err := bodyexecution.DecodeCases(suite)
	must(err)
	start = time.Now()
	var observed bodyexecution.Result
	if owner == nil {
		observed, err = bodyexecution.Execute(ctx, p.Request+".gooo", source, doc, generated, parent, cases, goBin)
	} else {
		observed, err = owner.Execute(ctx, p.Request+".gooo", source, doc, generated, parent, cases, goBin)
	}
	r.ExecuteNS = time.Since(start).Nanoseconds()
	r.RuntimeSHA = save(filepath.Join(out, id+"-runtime.json"), observed)
	must(err)
	o := observed.Observation
	require(o.ProducerSourceSHA == revision && o.Stage == "COMPLETE" && o.RuntimeReplayed && len(o.Runs) == 2 && len(o.Cases) == 8 && reflect.DeepEqual(o.Cases, reference.Observation.Cases), "native outcomes differ")
	r.Built, r.BuildNS = o.Build.Completed, o.Build.WallNS
	if owner != nil {
		require(o.Artifact != nil && o.Artifact.Reused == (trial == 1), "reuse differs")
		r.Reused = o.Artifact.Reused
	}
	r.FirstRunNS, r.SecondRunNS, r.NativeRuns = o.Runs[0].WallNS, o.Runs[1].WallNS, len(o.Runs)
	for _, c := range o.Cases {
		r.Total++
		if c.Passed {
			r.Passed++
		}
	}
	for _, c := range append([]bodyexecution.ProcessObservation{o.Toolchain, o.Build}, o.Runs...) {
		r.CurrentChildCPUNS += c.UserNS + c.SystemNS
		if c.PeakRSSBytes != nil && *c.PeakRSSBytes > r.MaxCurrentChildRSSBytes {
			r.MaxCurrentChildRSSBytes = *c.PeakRSSBytes
		}
	}
	return r
}
func main() {
	baseline := flag.String("baseline", "", "extracted frozen initial native observations")
	model := flag.String("model", "", "original public model.json")
	goBin := flag.String("go-bin", "go", "Go 1.27.1 tool")
	out := flag.String("out", "", "fresh output directory")
	flag.Parse()
	require(*baseline != "" && *model != "" && *out != "" && flag.NArg() == 0, "inputs required")
	revision := identity()
	priorBytes, manifest := read(filepath.Join(*baseline, "records.json")), read(filepath.Join(*baseline, "manifest.json"))
	require(hash(priorBytes) == frozenRecords && hash(manifest) == frozenManifest, "frozen baseline differs")
	require(hash(read(filepath.Join(filepath.Dir(*model), "weights.bin"))) == weights, "weights differ")
	must(os.Mkdir(*out, 0755))
	journal, err := os.Create(filepath.Join(*out, "progress.jsonl"))
	must(err)
	defer journal.Close()
	var priors []prior
	decode(priorBytes, &priors)
	require(len(priors) == 256, "baseline count differs")
	var records []record
	pairs := 0
	for _, p := range priors {
		if p.Arm != "model" {
			continue
		}
		require(filepath.Base(p.Request) == p.Request && p.Request != "", "invalid request")
		source := read(filepath.Join(*baseline, p.Request+".gooo"))
		require("sha256:"+hash(source) == p.SourceSHA, "source differs")
		recipeName := fmt.Sprintf("%s-b%d-recipe.json", p.Request, p.Budget)
		recipe, suite := read(filepath.Join(*baseline, recipeName)), read(filepath.Join(*baseline, p.Request+"-cases.json"))
		for name, data := range map[string][]byte{p.Request + ".gooo": source, recipeName: recipe, p.Request + "-cases.json": suite} {
			must(os.WriteFile(filepath.Join(*out, name), data, 0644))
		}
		gen, run := read(filepath.Join(*baseline, p.ID+"-generation.json")), read(filepath.Join(*baseline, p.ID+"-runtime.json"))
		require(hash(gen) == p.GenerationSHA && hash(run) == p.RuntimeSHA, "frozen observations differ")
		var old bodycodegen.Result
		var reference bodyexecution.Result
		decode(gen, &old)
		decode(run, &reference)
		modes := []string{"fresh", "retained"}
		if pairs%2 == 1 {
			modes[0], modes[1] = modes[1], modes[0]
		}
		for _, mode := range modes {
			g, err := bodycodegen.NewTypedPathGenerator(*model)
			must(err)
			require(g.Info().WeightsSHA256 == weights && g.Info().ResidentTensorBytes == 16384, "model differs")
			var owner *bodyexecution.Executor
			if mode == "retained" {
				owner = bodyexecution.NewExecutor()
			}
			for trial := range 2 {
				r := observe(g, owner, *out, revision, *goBin, p, mode, trial, source, recipe, suite, old, reference)
				records = append(records, r)
				must(json.NewEncoder(journal).Encode(r))
				must(journal.Sync())
			}
			must(owner.Close())
		}
		pairs++
		if pairs%8 == 0 {
			fmt.Printf("completed pairs=%d generations=%d\n", pairs, len(records))
		}
	}
	require(pairs == 128 && len(records) == 512, "cohort incomplete")
	save(filepath.Join(*out, "records.json"), records)
	exe, err := os.Executable()
	must(err)
	save(filepath.Join(*out, "manifest.json"), map[string]any{
		"schema": "gooo/retained-native-comparison/v1", "compiler_sha": revision, "collector_binary_sha256": hash(read(exe)),
		"sdk": "v0.2.20-experimental", "go": runtime.Version(), "os": runtime.GOOS, "arch": runtime.GOARCH,
		"model_weights_sha256": weights, "training_updates": 0, "baseline_records_sha256": frozenRecords, "baseline_manifest_sha256": frozenManifest,
		"requests": 64, "budgets": []int{1, 8}, "generations": len(records), "model_predictions": len(records), "native_runs": len(records) * 2,
		"scope": "Known EN/KO development requests. Both arms use fresh source decoding and a retained generator within their two calls; only executable ownership differs. Immediate native executions match all original ordered outcomes, including partial failures. Current-child CPU and max single-child peak RSS exclude retained build history. No whole-host utilization or new generalization claim.",
	})
}
