package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestVersionBuildReportsCompiledPureBodyLimits(t *testing.T) {
	for _, args := range [][]string{{"--build", "--json"}, {"--json", "--build"}} {
		var output, stderr bytes.Buffer
		if status := runVersion(args, &output, &stderr); status != exitOK || stderr.Len() != 0 {
			t.Fatal("build limits unavailable", status, stderr.String())
		}
		var report struct {
			Schema string         `json:"schema"`
			Limits map[string]int `json:"pure_body_limits"`
		}
		if err := json.Unmarshal(output.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		want := map[string]int{
			"max_record_types": 16, "max_fields_per_record": 32, "max_elements_per_collection": 256,
			"max_activity_calls_per_invocation": 4096, "max_iterations_per_counted_loop": 1024,
			"max_activities_per_call_closure": 32, "max_activity_call_depth": 16,
			"max_syntactic_collection_iteration_product": 65536,
		}
		if report.Schema != "gooo/build-identity/v1" || !reflect.DeepEqual(report.Limits, want) {
			t.Fatal("compiled limit population differs", report)
		}
		var typed versionBuildInfo
		if err := json.Unmarshal(output.Bytes(), &typed); err != nil || typed.PureBodyLimits != bodycodegen.CurrentProfileLimits() {
			t.Fatal("discovery differs from source validation constants", err, typed.PureBodyLimits)
		}
	}
}

func TestVersionBodyLimitsDoNotManufactureSourceBinding(t *testing.T) {
	identity := versionBuildMetadata(nil, "observed-runtime")
	if identity.PureBodyLimits != bodycodegen.CurrentProfileLimits() || identity.BuildInfoObserved ||
		identity.CompilerSourceSHA != "UNBOUND_LOCAL_SOURCE" || identity.SourceStatus != "UNOBSERVED_VCS" {
		t.Fatal("compiled bounds fabricated source origin", identity)
	}
	var output, stderr bytes.Buffer
	if status := runVersion([]string{"--json"}, &output, &stderr); status != exitOK ||
		bytes.Contains(output.Bytes(), []byte("pure_body_limits")) {
		t.Fatal("legacy contract output changed", status, output.String())
	}
}
