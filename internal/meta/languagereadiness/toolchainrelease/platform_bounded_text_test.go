package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func boundedTextSmokeFixture(t *testing.T) (map[string]any, []byte, []byte) {
	t.Helper()
	manifest, err := os.ReadFile("../../../../examples/bounded-text-search/gooo.workspace.json")
	if err != nil {
		t.Fatal(err)
	}
	cases, err := os.ReadFile("../../../../examples/bounded-text-search/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	values, err := boundedTextReference(cases)
	if err != nil {
		t.Fatal(err)
	}
	traces := []any{}
	for i, value := range values {
		traces = append(traces, map[string]any{"case_index": i, "deliveries": []any{map[string]any{
			"activity_id": "gooo://bounded-search/app/main", "actual": value, "expected": value, "passed": true}}})
	}
	r := map[string]any{"schema": "gooo/workspace-body-execution-receipt/v1", "decision": "PASS",
		"manifest_digest": fmt.Sprintf("sha256:%x", sha256.Sum256(manifest)),
		"cases_digest":    fmt.Sprintf("sha256:%x", sha256.Sum256(cases)),
		"result": map[string]any{"program": map[string]any{"entry": map[string]any{"package_path": "app/search", "activity": "Main"}},
			"composition": map[string]any{"generated_sha256": "sha256:" + strings.Repeat("a", 64)},
			"runtime": map[string]any{"model_calls": 0, "finite_passed": 10, "finite_total": 10,
				"projection_replayed": true, "runtime_replayed": true, "traces": traces}}}
	return r, manifest, cases
}

func TestBoundedTextReleaseRejectsUnboundOrIncompleteObservations(t *testing.T) {
	r, manifest, cases := boundedTextSmokeFixture(t)
	raw, _ := json.Marshal(r)
	if _, err := validateBoundedTextSmoke(raw, manifest, cases, false, boundedTextSmokeBinding{}); err != nil {
		t.Fatal(err)
	}
	mutations := []func(map[string]any){
		func(r map[string]any) { r["manifest_digest"] = "other" },
		func(r map[string]any) { r["cases_digest"] = "other" },
		func(r map[string]any) { r["decision"] = "OBSERVED" },
		func(r map[string]any) { r["error"] = "failed" },
		func(r map[string]any) { delete(boundedTextRuntime(r), "model_calls") },
		func(r map[string]any) { boundedTextRuntime(r)["model_calls"] = 1 },
		func(r map[string]any) { boundedTextRuntime(r)["finite_passed"] = 9 },
		func(r map[string]any) { boundedTextRuntime(r)["runtime_replayed"] = false },
		func(r map[string]any) { boundedTextRuntime(r)["traces"] = []any{} },
		func(r map[string]any) { boundedTextTrace(r)["case_index"] = 1 },
		func(r map[string]any) { boundedTextDelivery(r)["actual"] = 4 },
		func(r map[string]any) { boundedTextDelivery(r)["expected"] = 4 },
		func(r map[string]any) { boundedTextDelivery(r)["passed"] = false },
		func(r map[string]any) { boundedTextDelivery(r)["activity_id"] = "other" },
	}
	for _, mutate := range mutations {
		r, _, _ := boundedTextSmokeFixture(t)
		mutate(r)
		bad, _ := json.Marshal(r)
		if _, err := validateBoundedTextSmoke(bad, manifest, cases, false, boundedTextSmokeBinding{}); err == nil {
			t.Fatal("incomplete native observation accepted")
		}
	}
}

func TestBoundedTextReleaseRequiresSavedProgramWithoutPrediction(t *testing.T) {
	r, manifest, cases := boundedTextSmokeFixture(t)
	original, _ := json.Marshal(r)
	selected, err := validateBoundedTextSmoke(original, manifest, cases, false, boundedTextSmokeBinding{})
	if err != nil {
		t.Fatal(err)
	}
	r["replayed_from_sha256"] = selected.Receipt
	r["result"].(map[string]any)["replay"] = map[string]any{"model_calls": 0}
	raw, _ := json.Marshal(r)
	if _, err := validateBoundedTextSmoke(raw, manifest, cases, true, selected); err != nil {
		t.Fatal(err)
	}
	for _, selected := range []boundedTextSmokeBinding{{},
		{Program: "sha256:" + strings.Repeat("c", 64), Receipt: selected.Receipt},
		{Program: selected.Program, Receipt: "sha256:" + strings.Repeat("b", 64)}} {
		if _, err := validateBoundedTextSmoke(raw, manifest, cases, true, selected); err == nil {
			t.Fatal("changed saved program accepted")
		}
	}
	delete(r, "replayed_from_sha256")
	raw, _ = json.Marshal(r)
	if _, err := validateBoundedTextSmoke(raw, manifest, cases, true, selected); err == nil {
		t.Fatal("missing original receipt accepted")
	}
}

func boundedTextRuntime(r map[string]any) map[string]any {
	return r["result"].(map[string]any)["runtime"].(map[string]any)
}
func boundedTextTrace(r map[string]any) map[string]any {
	return boundedTextRuntime(r)["traces"].([]any)[0].(map[string]any)
}
func boundedTextDelivery(r map[string]any) map[string]any {
	return boundedTextTrace(r)["deliveries"].([]any)[0].(map[string]any)
}
