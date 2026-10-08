package toolchainrelease

import (
	"encoding/json"
	"testing"
)

func packageTextFixture() map[string]any {
	var traces []any
	for i, value := range []map[string]any{
		{"source": true, "stem": "한🙂", "bytes": 12},
		{"source": false, "stem": ".hidden", "bytes": 12},
		{"source": false, "stem": "notes.txt", "bytes": 9},
	} {
		traces = append(traces, map[string]any{"case_index": i, "deliveries": []any{map[string]any{
			"activity_id": "gooo-workspace://activity/gooo-package4activity-classify", "actual": value}}})
	}
	return map[string]any{"schema": "gooo/workspace-body-execution-receipt/v1", "decision": "OBSERVED",
		"result": map[string]any{"composition": map[string]any{"generated_sha256": "sha256:program"},
			"runtime": map[string]any{"model_calls": 0, "finite_passed": 0, "finite_total": 0,
				"projection_replayed": true, "runtime_replayed": true, "traces": traces}}}
}

func TestPackageTextSmokeRequiresActualInputOnlyValues(t *testing.T) {
	valid, _ := json.Marshal(packageTextFixture())
	if _, err := validatePackageTextSmoke(valid, false, ""); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(map[string]any){
		func(r map[string]any) { r["schema"] = "other" },
		func(r map[string]any) { r["decision"] = "PASS" },
		func(r map[string]any) { r["error"] = "failed" },
		func(r map[string]any) { delete(packageTextRuntime(r), "model_calls") },
		func(r map[string]any) { packageTextRuntime(r)["model_calls"] = 1 },
		func(r map[string]any) { packageTextRuntime(r)["finite_total"] = 3 },
		func(r map[string]any) { packageTextRuntime(r)["finite_passed"] = 3 },
		func(r map[string]any) { packageTextRuntime(r)["runtime_replayed"] = false },
		func(r map[string]any) { packageTextRuntime(r)["projection_replayed"] = false },
		func(r map[string]any) { packageTextRuntime(r)["traces"] = []any{} },
		func(r map[string]any) { packageTextTrace(r)["case_index"] = 1 },
		func(r map[string]any) { packageTextDelivery(r)["activity_id"] = "other" },
		func(r map[string]any) { packageTextDelivery(r)["expected"] = map[string]any{} },
		func(r map[string]any) { packageTextDelivery(r)["actual"].(map[string]any)["bytes"] = 7 },
		func(r map[string]any) { delete(packageTextDelivery(r)["actual"].(map[string]any), "source") },
	} {
		r := packageTextFixture()
		mutate(r)
		raw, _ := json.Marshal(r)
		if _, err := validatePackageTextSmoke(raw, false, ""); err == nil {
			t.Fatal("incomplete package observation accepted", string(raw))
		}
	}
}

func TestPackageTextSmokeBindsSavedReplayWithoutNewPrediction(t *testing.T) {
	r := packageTextFixture()
	r["replayed_from_sha256"] = "sha256:prior"
	r["result"].(map[string]any)["replay"] = map[string]any{"model_calls": 0}
	raw, _ := json.Marshal(r)
	if _, err := validatePackageTextSmoke(raw, true, "sha256:program"); err != nil {
		t.Fatal(err)
	}
	for _, selected := range []string{"", "sha256:other"} {
		if _, err := validatePackageTextSmoke(raw, true, selected); err == nil {
			t.Fatal("changed saved program accepted")
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
		if _, err := validatePackageTextSmoke(bad, true, "sha256:program"); err == nil {
			t.Fatal("unbound replay accepted")
		}
	}
}

func packageTextRuntime(r map[string]any) map[string]any {
	return r["result"].(map[string]any)["runtime"].(map[string]any)
}
func packageTextTrace(r map[string]any) map[string]any {
	return packageTextRuntime(r)["traces"].([]any)[0].(map[string]any)
}
func packageTextDelivery(r map[string]any) map[string]any {
	return packageTextTrace(r)["deliveries"].([]any)[0].(map[string]any)
}
