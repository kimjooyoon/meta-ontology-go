// Measure recipe decoding in two clean compiler revisions, without model calls
// or generation of selected candidates. Native outcomes are collected separately.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const frozen = "47315bf031f4734497f1a9f3710e93b6c0e8f43ed385f78ec7d7cc51adfa2be1"

type prior struct {
	ID, Request, Arm, SourceSHA, PlanSHA string
	Budget                               int
}

type observation struct {
	ID, Request, PlanSHA, DocumentSHA string
	Budget, Repeat                    int
	DecodeNS                          int64
	AllocatedBytes, Allocations       uint64
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func require(ok bool, reason string) {
	if !ok {
		panic(reason)
	}
}

func read(name string) []byte { raw, err := os.ReadFile(name); check(err); return raw }
func hash(raw []byte) string  { return fmt.Sprintf("%x", sha256.Sum256(raw)) }

func identity() string {
	info, ok := debug.ReadBuildInfo()
	require(ok && info.GoVersion == "go1.27.1", "Go1.27.1 build required")
	var revision string
	clean, sdk := false, false
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			revision = setting.Value
		}
		if setting.Key == "vcs.modified" {
			clean = setting.Value == "false"
		}
	}
	for _, dependency := range info.Deps {
		if dependency.Path == "github.com/kimjooyoon/gooo-decision-runtime" {
			sdk = dependency.Version == "v0.2.20-experimental" && dependency.Replace == nil
		}
	}
	require(clean && sdk && len(revision) == 40, "clean public SDK/compiler build required")
	return revision
}

func measure(p prior, source, recipe []byte, repeat int) observation {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	started := time.Now()
	doc, err := bodycodegen.DecodeSourcePathDocument(ctx, p.Request+".gooo", source, "Compose", recipe)
	elapsed := time.Since(started).Nanoseconds()
	runtime.ReadMemStats(&after)
	check(err)
	prepared, err := doc.Prepare()
	check(err)
	require(prepared.PlanSHA256() == p.PlanSHA, "frozen plan differs")
	raw, err := json.Marshal(doc)
	check(err)
	return observation{ID: p.ID, Request: p.Request, Budget: p.Budget, Repeat: repeat,
		PlanSHA: p.PlanSHA, DocumentSHA: hash(raw), DecodeNS: elapsed,
		AllocatedBytes: after.TotalAlloc - before.TotalAlloc, Allocations: after.Mallocs - before.Mallocs}
}

func main() {
	baseline := flag.String("baseline", "", "frozen original native collection")
	out := flag.String("out", "", "fresh output directory")
	repeats := flag.Int("repeats", 8, "measured decodes per request (1..32)")
	flag.Parse()
	require(flag.NArg() == 0 && *baseline != "" && *out != "" && *repeats >= 1 && *repeats <= 32, "bounded arguments required")
	revision := identity()
	raw := read(filepath.Join(*baseline, "records.json"))
	require(hash(raw) == frozen, "frozen records differ")
	var priors []prior
	check(json.Unmarshal(raw, &priors))
	_, err := os.Stat(*out)
	require(os.IsNotExist(err), "output must be fresh")
	check(os.MkdirAll(*out, 0700))
	journal, err := os.Create(filepath.Join(*out, "records.jsonl"))
	check(err)
	defer journal.Close()
	encoder := json.NewEncoder(journal)
	count := 0
	for _, p := range priors {
		if p.Arm != "model" {
			continue
		}
		require(p.Request != "" && filepath.Base(p.Request) == p.Request, "invalid request")
		source := read(filepath.Join(*baseline, p.Request+".gooo"))
		require("sha256:"+hash(source) == p.SourceSHA, "source differs")
		recipe := read(filepath.Join(*baseline, fmt.Sprintf("%s-b%d-recipe.json", p.Request, p.Budget)))
		// One unmeasured warm-up per request; every measured call still decodes fresh.
		_ = measure(p, source, recipe, -1)
		for repeat := range *repeats {
			check(encoder.Encode(measure(p, source, recipe, repeat)))
			count++
		}
	}
	require(count == 128**repeats, "cohort incomplete")
	check(journal.Sync())
	file, err := os.Create(filepath.Join(*out, "manifest.json"))
	check(err)
	defer file.Close()
	check(json.NewEncoder(file).Encode(map[string]any{"schema": "gooo/source-recipe-cost/v2",
		"compiler_sha": revision, "baseline_records_sha256": frozen, "requests": 64, "request_budget_pairs": 128,
		"measured_decodes": count, "repeats": *repeats, "model_predictions": 0,
		"scope": "Fresh recipe decode wall and allocation deltas; excludes input reads, memory counter reads, final plan/digest checks and serialization; known requests only"}))
}
