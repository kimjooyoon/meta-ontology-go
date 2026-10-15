package toolchainrelease

import (
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
)

//go:embed testdata/relative-v2-native-20261012/*.json.gz
var relativeV2RetainedFixtures embed.FS

// Reads the first new zero-model native observation; no inference or execution.
func TestRelativeV2PortableRetainedNativeReader(t *testing.T) {
	root := "testdata/relative-v2-native-20261012"
	pins := map[string]string{
		"preflight.json.gz": "sha256:f3f7f1bdea1c76416f6ac72a1800c3a0299eb3ae44929d17d6effbb723479551",
		"construct.json.gz": "sha256:f8d883ace34e716c08700716ee4833145b416aef2274c94ce94eb2317c4da9aa",
		"replay.json.gz":    "sha256:677aee2681bff99bebe56ba708a2fe287f0cdc0507690264cb42c522a0d1066e",
	}
	read := func(name string) []byte {
		t.Helper()
		raw, err := os.ReadFile(name)
		if strings.HasSuffix(name, ".gz") {
			raw, err = relativeV2RetainedFixtures.ReadFile(name)
		}
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(name, ".gz") {
			r, err := gzip.NewReader(bytes.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			raw, err = io.ReadAll(io.LimitReader(r, (1<<20)+1))
			if err != nil || len(raw) > 1<<20 {
				t.Fatal("bounded retained JSON required", err)
			}
			if scalarPreflightDigest(raw) != pins[filepath.Base(name)] {
				t.Fatal("original native fixture bytes changed")
			}
		}
		return raw
	}
	source := read(filepath.Join("../../../../", relativeV2Root+"cases.gooo.fixture"))
	feedback := read(filepath.Join("../../../../", relativeV2Root+"construction.json"))
	cases := read(filepath.Join("../../../../", relativeV2Root+"evaluation.json"))
	m, err := contractdecision.NewSourceRelativeV2([contractdecision.InteractionRequirementParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	model, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	preflight, err := validateRelativeV2Preflight(read(filepath.Join(root, "preflight.json.gz")), source, model)
	if err != nil {
		t.Fatal(err)
	}
	selected := ""
	for _, replay := range []bool{false, true} {
		name := "construct"
		if replay {
			name = "replay"
		}
		raw := read(filepath.Join(root, name+".json.gz"))
		selected, err = validateRelativeV2Smoke(raw, source, feedback, cases, preflight, "7f57c454603d40465ac3e3764451fe26f5143062", replay, selected)
		if err != nil {
			t.Fatal(err)
		}
		obj := func(v any) map[string]any { return v.(map[string]any) }
		body := func(r map[string]any) map[string]any {
			initial := obj(obj(r["construction"])["initial"])
			preparation := obj(initial["preparations"].([]any)[0])
			return obj(obj(obj(preparation["generation"])["report"])["body_paths"])
		}
		changes := []func(map[string]any){
			func(r map[string]any) {
				obj(body(r)["declared_contract_ranking"])["local_model_predictions"] = json.Number("2")
			},
			func(r map[string]any) {
				obj(body(r)["declared_contract_ranking"])["source_feature_sha256"].([]any)[0] = "changed"
			},
			func(r map[string]any) {
				attempt := obj(obj(body(r)["search"])["attempts"].([]any)[0])
				condition := obj(attempt["condition_results"].([]any)[0])
				obj(condition["observation"])["reached"] = false
			},
			func(r map[string]any) {
				selected := obj(obj(r["construction"])["selected"])
				selected["source"] = selected["source"].(string) + "changed"
			},
			func(r map[string]any) { delete(obj(r["evaluation"]), "new_model_calls") },
			func(r map[string]any) { obj(body(r)["model_retention"])["artifact_sha256"] = "changed" },
			func(r map[string]any) { obj(r["construction"])["original_source_sha256"] = "changed" },
			func(r map[string]any) { obj(r["construction"])["program_budget"] = json.Number("5") },
			func(r map[string]any) {
				runtime := obj(obj(r["evaluation"])["runtime"])
				obj(runtime["outcomes"])["faulted"] = json.Number("1")
			},
			func(r map[string]any) {
				runtime := obj(obj(r["evaluation"])["runtime"])
				delete(obj(runtime["outcomes"]), "unobserved")
			},
			func(r map[string]any) {
				runtime := obj(obj(r["evaluation"])["runtime"])
				traces := runtime["traces"].([]any)
				obj(traces[1])["case_index"] = json.Number("0")
			},
			func(r map[string]any) {
				runtime := obj(obj(r["evaluation"])["runtime"])
				trace := obj(runtime["traces"].([]any)[7])
				delivery := obj(trace["deliveries"].([]any)[0])
				delivery["actual"] = json.Number("9007199254741001")
			},
		}
		// Preserve unknown runtime fields and exact JSON numbers in each mutation.
		for index := -1; index < len(changes); index++ {
			var r map[string]any
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err = decoder.Decode(&r); err != nil {
				t.Fatal(err)
			}
			if index >= 0 {
				changes[index](r)
			}
			bad, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			_, err = validateRelativeV2Smoke(bad, source, feedback, cases, preflight, "7f57c454603d40465ac3e3764451fe26f5143062", replay, selected)
			if index < 0 && err != nil {
				t.Fatal("unaltered JSON roundtrip rejected", err)
			}
			if index >= 0 && err == nil {
				t.Fatal("changed retained construction accepted", index)
			}
		}
	}
}
