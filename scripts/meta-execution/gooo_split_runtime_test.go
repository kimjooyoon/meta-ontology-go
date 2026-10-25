package main

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func TestGoooSplitRuntimeRequiresExactFinitePopulation(t *testing.T) {
	digest := digestBytes([]byte("independent cases"))
	for _, mode := range []string{"complete", "different-cases", "denominator", "failed-case", "duplicate-case",
		"missing-trace", "model-called", "projection", "replay", "missing-tool"} {
		runtime := bodyexecution.CompositionRuntime{GeneratedSHA256: digest, GoToolSHA256: digest,
			FiniteTotal: 2, FinitePassed: 2, ProjectionReplayed: true, RuntimeReplayed: true,
			Traces: []bodyexecution.CompositionTrace{{CaseIndex: 0}, {CaseIndex: 1}}}
		casesDigest := digest
		switch mode {
		case "different-cases":
			casesDigest = digestBytes([]byte("other cases"))
		case "denominator":
			runtime.FiniteTotal++
		case "failed-case":
			runtime.FinitePassed--
		case "duplicate-case":
			runtime.Traces[1].CaseIndex = 0
		case "missing-trace":
			runtime.Traces = runtime.Traces[:1]
		case "model-called":
			runtime.ModelCalls = 1
		case "projection":
			runtime.ProjectionReplayed = false
		case "replay":
			runtime.RuntimeReplayed = false
		case "missing-tool":
			runtime.GoToolSHA256 = ""
		}
		raw, err := json.Marshal(map[string]any{"decision": "PASS", "cases_digest": casesDigest,
			"result": workspaceexecution.Result{Runtime: runtime}})
		if err != nil {
			t.Fatal(err)
		}
		_, err = decodeGoooSplitRuntime(raw, digest, 2)
		if (err == nil) != (mode == "complete") {
			t.Fatal("finite evidence mismatch", mode, err)
		}
	}
}
