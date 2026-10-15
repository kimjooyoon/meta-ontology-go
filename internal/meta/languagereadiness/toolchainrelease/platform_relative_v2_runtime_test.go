package toolchainrelease

import (
	"encoding/json"
	"testing"
)

// Synthetic runtime records exercise the reader without native or model execution.
func TestRelativeV2NativeReaderRejectsMissingWrongAndDuplicateValues(t *testing.T) {
	cases := []byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":9007199254740993},"expected":{"Main":9007199254741000}}]}`)
	raw := []byte(`{"stage":"COMPLETE","generated_sha256":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","finite_passed":1,"finite_total":1,"model_calls":0,"projection_replayed":true,"runtime_replayed":true,"traces":[{"case_index":0,"deliveries":[{"activity_id":"division-layout://main","input":9007199254740993,"actual":9007199254741000,"expected":9007199254741000,"passed":true}]}]}`)
	var r jointSmokeRuntime
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	if err := validateRelativeV2Runtime(r, cases, 1); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*jointSmokeRuntime){
		func(r *jointSmokeRuntime) { r.Traces[0].Deliveries[0].Actual = []byte("9007199254741001") },
		func(r *jointSmokeRuntime) { r.Traces[0].Index = 1 },
		func(r *jointSmokeRuntime) { r.Traces = append(r.Traces, r.Traces[0]) },
		func(r *jointSmokeRuntime) { r.Calls = nil },
		func(r *jointSmokeRuntime) { *r.Calls = 1 },
	} {
		if err := json.Unmarshal(raw, &r); err != nil {
			t.Fatal(err)
		}
		change(&r)
		if err := validateRelativeV2Runtime(r, cases, 1); err == nil {
			t.Fatal("altered native record accepted")
		}
	}
}
