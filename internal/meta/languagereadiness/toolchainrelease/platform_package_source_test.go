package toolchainrelease

import (
	"encoding/json"
	"os"
	"testing"
)

func sourcePackageFixture(t *testing.T) (map[string]any, []byte) {
	t.Helper()
	cases, err := os.ReadFile("../../../../examples/package-body-calls/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference struct {
		Cases []struct {
			Expected map[string]json.RawMessage
		}
	}
	if err := json.Unmarshal(cases, &reference); err != nil {
		t.Fatal(err)
	}
	traces := []any{}
	for i, row := range reference.Cases {
		deliveries := []any{}
		for _, pair := range [][2]string{
			{"tools/diagnostics:Diagnose", "gooo-workspace://activity/gooo-package2activity-diagnose"},
			{"app/explain:Main", "gooo-workspace://activity/gooo-package3activity-main"},
		} {
			deliveries = append(deliveries, map[string]any{"activity_id": pair[1], "passed": true,
				"actual": row.Expected[pair[0]], "expected": row.Expected[pair[0]]})
		}
		traces = append(traces, map[string]any{"case_index": i, "deliveries": deliveries})
	}
	return map[string]any{"schema": "gooo/workspace-body-execution-receipt/v1", "decision": "PASS",
		"result": map[string]any{"composition": map[string]any{"generated_sha256": "sha256:program"},
			"runtime": map[string]any{"model_calls": 0, "finite_passed": 8, "finite_total": 8,
				"projection_replayed": true, "runtime_replayed": true, "traces": traces}}}, cases
}

func TestPackageSourceSmokeComparesEveryNamedOutput(t *testing.T) {
	fixture, cases := sourcePackageFixture(t)
	raw, _ := json.Marshal(fixture)
	if _, err := validatePackageSourceSmoke(raw, cases, false, ""); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(map[string]any){
		func(r map[string]any) { r["decision"] = "PROGRESS" },
		func(r map[string]any) { r["schema"] = "other" },
		func(r map[string]any) { r["error"] = "failed" },
		func(r map[string]any) { delete(packageTextRuntime(r), "model_calls") },
		func(r map[string]any) { packageTextRuntime(r)["model_calls"] = 1 },
		func(r map[string]any) { packageTextRuntime(r)["finite_passed"] = 7 },
		func(r map[string]any) { packageTextRuntime(r)["finite_total"] = 4 },
		func(r map[string]any) { packageTextRuntime(r)["runtime_replayed"] = false },
		func(r map[string]any) { packageTextRuntime(r)["projection_replayed"] = false },
		func(r map[string]any) { packageTextRuntime(r)["traces"] = []any{} },
		func(r map[string]any) { packageTextTrace(r)["case_index"] = 1 },
		func(r map[string]any) { packageTextDelivery(r)["activity_id"] = "other" },
		func(r map[string]any) { packageTextDelivery(r)["passed"] = false },
		func(r map[string]any) { delete(packageTextDelivery(r), "passed") },
		func(r map[string]any) { packageTextDelivery(r)["actual"] = "wrong" },
		func(r map[string]any) { packageTextDelivery(r)["expected"] = "wrong" },
		func(r map[string]any) {
			d := packageTextDelivery(r)
			packageTextTrace(r)["deliveries"] = []any{d, d}
		},
	} {
		r, _ := sourcePackageFixture(t)
		mutate(r)
		bad, _ := json.Marshal(r)
		if _, err := validatePackageSourceSmoke(bad, cases, false, ""); err == nil {
			t.Fatal("changed package result accepted", string(bad))
		}
	}
	if _, err := validatePackageSourceSmoke(raw, []byte(`{"cases":[]}`), false, ""); err == nil {
		t.Fatal("absent independent expectations accepted")
	}
}

func TestPackageSourceSmokePreservesProgramAndReplay(t *testing.T) {
	r, cases := sourcePackageFixture(t)
	raw, _ := json.Marshal(r)
	if _, err := validatePackageSourceSmoke(raw, cases, false, "sha256:other"); err == nil {
		t.Fatal("source metadata changed the selected program")
	}
	r["replayed_from_sha256"] = "sha256:prior"
	r["result"].(map[string]any)["replay"] = map[string]any{"model_calls": 0}
	raw, _ = json.Marshal(r)
	if _, err := validatePackageSourceSmoke(raw, cases, true, "sha256:program"); err != nil {
		t.Fatal(err)
	}
	for _, selected := range []string{"", "sha256:other"} {
		if _, err := validatePackageSourceSmoke(raw, cases, true, selected); err == nil {
			t.Fatal("unbound saved program accepted")
		}
	}
	for _, mutate := range []func(map[string]any){
		func(r map[string]any) { delete(r, "replayed_from_sha256") },
		func(r map[string]any) { delete(r["result"].(map[string]any), "replay") },
		func(r map[string]any) { r["result"].(map[string]any)["replay"] = map[string]any{} },
		func(r map[string]any) { r["result"].(map[string]any)["replay"] = map[string]any{"model_calls": 1} },
	} {
		fresh := map[string]any{}
		if err := json.Unmarshal(raw, &fresh); err != nil {
			t.Fatal(err)
		}
		mutate(fresh)
		bad, _ := json.Marshal(fresh)
		if _, err := validatePackageSourceSmoke(bad, cases, true, "sha256:program"); err == nil {
			t.Fatal("replay lost its prior program or performed new inference")
		}
	}
}
