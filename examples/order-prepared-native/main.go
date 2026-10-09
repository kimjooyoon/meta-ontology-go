// Replay frozen development requests through the retained compiler API, then
// immediately build and execute each generated result before generating another.
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
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const frozenRecords = "47315bf031f4734497f1a9f3710e93b6c0e8f43ed385f78ec7d7cc51adfa2be1"
const frozenManifest = "e336e1a6dbb7493eee3075a362f2a65674a0a32045824c700bd66800b5eab49c"
const weights = "cf00ccc83d17d28ed73fcb869366151a48ffccd3aa8ca8e635aabf19810b9e78"

type prior struct {
	ID, Request, Arm, SourceSHA, PlanSHA, GenerationSHA, RuntimeSHA string
	Budget                                                          int
}
type native struct {
	Observation struct {
		Stage string
		Runs  []json.RawMessage
		Cases []struct {
			Input, Expected, Actual int64
			Passed                  bool
		}
	}
}
type record struct {
	ID, Request, Mode, PlanSHA, GenerationSHA, RuntimeSHA string
	Budget, Trial, ModelCalls, Passed, Total, NativeRuns  int
	Reused                                                bool
	SetupMS, DecodeMS, CodegenMS, AcquireMS               float64
	GenerateNS                                            int64
	AllocatedBytes, Allocations                           uint64
	Execution                                             cost
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
func read(path string) []byte    { b, err := os.ReadFile(path); must(err); return b }
func hash(b []byte) string       { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func decode(b []byte, value any) { must(json.Unmarshal(b, value)) }
func save(path string, value any) string {
	b, err := json.MarshalIndent(value, "", "  ")
	must(err)
	b = append(b, '\n')
	must(os.WriteFile(path, b, 0644))
	return hash(b)
}

func identity() string {
	b, ok := debug.ReadBuildInfo()
	require(ok && b.GoVersion == "go1.27.2", "pinned Go build required")
	var revision string
	clean, sdk := false, false
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
	require(clean && sdk && len(revision) == 40, "clean pinned compiler and SDK required")
	return revision
}

func compare(result, old bodycodegen.Result, revision string, p prior, reused bool) {
	r, before := result.Report.BodyPaths, old.Report.BodyPaths
	require(r != nil && before != nil && r.OrderJudgment != nil && before.OrderJudgment != nil, "ranking missing")
	require(result.Report.CompilerSourceSHA == revision && r.SourceBaseMatched && r.OriginalSourceSHA256 == p.SourceSHA,
		"fresh source/compiler binding differs")
	require(result.Source == old.Source && reflect.DeepEqual(r.Search, before.Search), "source or search semantics differ")
	a, b := *r.OrderJudgment, *before.OrderJudgment
	a.PredictNS, b.PredictNS = 0, 0
	require(reflect.DeepEqual(a, b), "complete ranking differs")
	c := r.OrderPreparation
	require(c != nil && c.Reused == reused && c.PlanSHA256 == p.PlanSHA && c.CandidateCount == 8, "preparation differs")
	require(r.Search.Selection.ModelCalls == 1 && r.Search.Selection.WeightsSHA256 == weights &&
		r.Search.Selection.ExternalCallsKnown && r.Search.Selection.ExternalCalls == 0, "provider accounting differs")
}

func observe(g *bodycodegen.TypedPathGenerator, setup float64, compiler, goBin, out, revision string,
	p prior, mode string, trial int, source, recipe []byte, old bodycodegen.Result, reference native) record {
	id := fmt.Sprintf("%s-%s-%d", p.ID, mode, trial)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	started := time.Now()
	doc, err := bodycodegen.DecodeSourcePathDocument(ctx, p.Request+".gooo", source, "Compose", recipe)
	must(err)
	r := record{ID: id, Request: p.Request, Budget: p.Budget, Mode: mode, Trial: trial, PlanSHA: p.PlanSHA,
		SetupMS: setup, DecodeMS: float64(time.Since(started)) / 1e6}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	started = time.Now()
	result, err := g.Generate(ctx, p.Request+".gooo", source, "Compose", doc, bodycodegen.TypedPathOptions{})
	r.GenerateNS = time.Since(started).Nanoseconds()
	runtime.ReadMemStats(&after)
	must(err)
	r.AllocatedBytes, r.Allocations = after.TotalAlloc-before.TotalAlloc, after.Mallocs-before.Mallocs
	compare(result, old, revision, p, mode == "retained" && trial == 1)
	r.Reused, r.AcquireMS = result.Report.BodyPaths.OrderPreparation.Reused, result.Report.BodyPaths.OrderPreparation.AcquireMS
	r.ModelCalls, r.CodegenMS = result.Report.BodyPaths.Search.Selection.ModelCalls, result.Report.BodyPaths.Timing.TotalMS
	r.GenerationSHA = save(filepath.Join(out, id+"-generation.json"), result)
	// No later Generate call starts before this exact projection is executed.
	r.Execution = execute(compiler, out, id+"-runtime.json", "body-execute", "--source", p.Request+".gooo",
		"--path-plan", fmt.Sprintf("%s-b%d-recipe.json", p.Request, p.Budget), "--generation", id+"-generation.json",
		"--cases", p.Request+"-cases.json", "--go-bin", goBin)
	var actual native
	raw := read(filepath.Join(out, id+"-runtime.json"))
	decode(raw, &actual)
	require(actual.Observation.Stage == "COMPLETE" && len(actual.Observation.Runs) == 2 &&
		len(actual.Observation.Cases) == 8 && reflect.DeepEqual(actual.Observation.Cases, reference.Observation.Cases), "native replay differs")
	r.RuntimeSHA, r.NativeRuns = hash(raw), len(actual.Observation.Runs)
	for _, c := range actual.Observation.Cases {
		r.Total++
		if c.Passed {
			r.Passed++
		}
	}
	return r
}

func main() {
	compiler := flag.String("compiler", "", "clean same-revision Gooo binary")
	goBin := flag.String("go-bin", "go", "Go 1.27.2 binary")
	baseline := flag.String("baseline", "", "extracted original order-judge native observations")
	model := flag.String("model", "", "original order judge model.json")
	out := flag.String("out", "", "fresh output directory")
	flag.Parse()
	require(*compiler != "" && *baseline != "" && *model != "" && *out != "" && flag.NArg() == 0, "inputs required")
	revision := identity()
	checkCompiler(*goBin, *compiler, revision)
	recordBytes, manifestBytes := read(filepath.Join(*baseline, "records.json")), read(filepath.Join(*baseline, "manifest.json"))
	require(hash(recordBytes) == frozenRecords && hash(manifestBytes) == frozenManifest, "frozen baseline differs")
	require(hash(read(filepath.Join(filepath.Dir(*model), "weights.bin"))) == weights, "model weights differ")
	_, err := os.Stat(*out)
	require(os.IsNotExist(err), "output must be fresh")
	must(os.MkdirAll(*out, 0755))
	var priors []prior
	decode(recordBytes, &priors)
	require(len(priors) == 256, "baseline count differs")
	journal, err := os.Create(filepath.Join(*out, "progress.jsonl"))
	must(err)
	defer journal.Close()
	enc := json.NewEncoder(journal)
	var records []record
	pairs := 0
	for _, p := range priors {
		if p.Arm != "model" {
			continue
		}
		require(filepath.Base(p.Request) == p.Request && p.Request != "" && !strings.Contains(p.Request, ".."), "invalid request")
		source := read(filepath.Join(*baseline, p.Request+".gooo"))
		require("sha256:"+hash(source) == p.SourceSHA, "source differs")
		recipeName := fmt.Sprintf("%s-b%d-recipe.json", p.Request, p.Budget)
		recipe := read(filepath.Join(*baseline, recipeName))
		for _, name := range []string{p.Request + ".gooo", recipeName, p.Request + "-cases.json"} {
			must(os.WriteFile(filepath.Join(*out, name), read(filepath.Join(*baseline, name)), 0644))
		}
		gen, run := read(filepath.Join(*baseline, p.ID+"-generation.json")), read(filepath.Join(*baseline, p.ID+"-runtime.json"))
		require(hash(gen) == p.GenerationSHA && hash(run) == p.RuntimeSHA, "frozen result differs")
		var old bodycodegen.Result
		var reference native
		decode(gen, &old)
		decode(run, &reference)
		modes := []string{"fresh", "retained"}
		if pairs%2 == 1 {
			modes[0], modes[1] = modes[1], modes[0]
		}
		for _, mode := range modes {
			var g *bodycodegen.TypedPathGenerator
			for trial := range 2 {
				setup := 0.0
				if g == nil || mode == "fresh" {
					g, err = bodycodegen.NewTypedPathGenerator(*model)
					must(err)
					require(g.Info().WeightsSHA256 == weights && g.Info().ResidentTensorBytes == 16384, "model identity differs")
					setup = g.Info().SetupMS
				}
				r := observe(g, setup, *compiler, *goBin, *out, revision, p, mode, trial, source, recipe, old, reference)
				records = append(records, r)
				must(enc.Encode(r))
				must(journal.Sync())
			}
		}
		pairs++
		if pairs%8 == 0 {
			fmt.Printf("completed pairs=%d generations=%d native_runs=%d\n", pairs, len(records), len(records)*2)
		}
	}
	require(pairs == 128 && len(records) == 512, "cohort incomplete")
	save(filepath.Join(*out, "records.json"), records)
	exe, err := os.Executable()
	must(err)
	save(filepath.Join(*out, "manifest.json"), map[string]any{
		"schema": "gooo/order-prepared-native/v1", "compiler_sha": revision, "compiler_binary_sha256": hash(read(*compiler)),
		"collector_binary_sha256": hash(read(exe)), "sdk": "v0.2.20-experimental", "go": runtime.Version(),
		"os": runtime.GOOS, "arch": runtime.GOARCH, "model_weights_sha256": weights, "training_updates": 0,
		"baseline_records_sha256": frozenRecords, "baseline_manifest_sha256": frozenManifest,
		"requests": 64, "budgets": []int{1, 8}, "generations": len(records), "model_predictions": len(records), "native_runs": len(records) * 2,
		"scope": "Fresh versus retained model and candidate ownership on known development requests; two calls per pair. Generate timing excludes constructor, decoding, memory counter reads, serialization and native execution. Every generation immediately builds/runs. Prior budget-one failures are preserved. No old/new CLI timing comparison.",
	})
}
