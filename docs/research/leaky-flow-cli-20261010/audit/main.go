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

const producer = "fff144361186f24e0fc3008100d11f19cdcfdebd"
const model = "d5f9c4ffb204852f4e1b3b1b1c491c7d1fb0753d0df2c1e0f0ffead80d9ac47e"
const fingerprint = "8fcbd396665411b02a8b2a3d0d62b7d00c9e327ab794dc06b2eeda36ad99cf48"

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

func checkSavedConstruction(root string, attempts int) {
	suite := decode[bodyexecution.CompositionCases](read(root, "evaluation-cases.json"))
	built := decode[constructionOutput](read(root, "construction.json.gz"))
	replayed := decode[constructionOutput](read(root, "replay.json.gz"))
	must(built.GeneratedNow && !replayed.GeneratedNow && replayed.Evaluation.ConstructionReplayed && replayed.Evaluation.NewModelCalls == 0, "model-free saved replay")
	retained := built.Construction.Initial.Preparations[0].Generation.Report.BodyPaths.ModelRetention
	checkSearch(built.Construction.Initial.Preparations[0].Generation, attempts, 1, digest(read(root, "source.gooo.gz")))
	must(retained.ModelSchema == "gooo/flow-candidate-decision/v2" && retained.FeatureVersion == decision.RelationalFlowFeatureVersion && retained.ResidentTensorBytes == 37160 && retained.ModelFingerprint == "sha256:"+fingerprint, "retained flow weights")
	must(reflect.DeepEqual(built.Construction, replayed.Construction), "unchanged saved construction")
	saved := decode[bodyexecution.JointConstruction](read(root, "saved-construction.json.gz"))
	savedJSON, savedErr := json.Marshal(saved)
	builtJSON, builtErr := json.Marshal(built.Construction)
	must(savedErr == nil && builtErr == nil && bytes.Equal(savedJSON, builtJSON), "saved file matches stdout after JSON whitespace normalization")
	checkNative(built.Evaluation.Runtime, suite)
	checkNative(replayed.Evaluation.Runtime, suite)
	must(built.Evaluation.Runtime.Build.Started && !replayed.Evaluation.Runtime.Build.Started, "saved executable reused")
}

func checkInputs(pre bodycodegen.TypedPathContextExport, p *bodycodegen.BodyPathReceipt, normalized bool) {
	must(reflect.DeepEqual(pre.Context, p.ModelContext), "preflight and actual context")
	must(p.ModelContext.ArtifactSHA == "sha256:"+model && p.ModelContext.ModelFingerprint == "sha256:"+fingerprint, "exact model")
	must(p.ModelContext.FeatureVersion == decision.RelationalFlowFeatureVersion, "v6 context")
	for i, input := range pre.Inputs {
		must(input.FlowFeatures != nil && input.Features == nil && input.ExecutionFeatures == nil && input.Bytes == 1536, "v6 array")
		must(input.SemanticFlow != nil && input.SemanticFlow.Normalized == normalized, "source normalization")
		must(input.SourceFeatureSHA == digest(input.SemanticFlow.Source[:]), "normalized source digest")
		if normalized {
			must(input.FlowFeatures[255] == 3, "normalized input marker")
		}
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
