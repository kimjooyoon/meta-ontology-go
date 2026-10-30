package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestVersionRuntimeDescriptionsAreAnExplicitJSONView(t *testing.T) {
	var output, diagnostics bytes.Buffer
	if code := runVersion([]string{"--build", "--json", "--describe-runtime"}, &output, &diagnostics); code != exitOK {
		t.Fatalf("description view code=%d stderr=%s", code, &diagnostics)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(output.Bytes(), &root); err != nil {
		t.Fatal(err)
	}
	var view struct {
		TargetSchema string `json:"target_schema"`
		ActivityID   string `json:"activity_id"`
		Entries      []struct {
			Path        string `json:"path"`
			Role        string `json:"role"`
			InputDomain string `json:"input_domain"`
			Meaning     string `json:"meaning"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(root["runtime_observation_fields"], &view); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"runtime_suite_sha256": "CASE_CONTENT_DIGEST", "artifact.key_sha256": "ARTIFACT_REUSE_IDENTITY",
		"executable_sha256": "EXECUTABLE_BYTE_DIGEST", "artifact.source_build_sha256": "HISTORICAL_BUILD_BINDING_DIGEST",
		"build": "CURRENT_BUILD_OBSERVATION", "runs": "CURRENT_RUN_OBSERVATIONS"}
	if view.TargetSchema != "gooo/body-composition-runtime/v2" || view.ActivityID != "urn:gooo:runtime-identity:describe" || len(view.Entries) != len(want) {
		t.Fatalf("runtime field descriptions=%+v", view)
	}
	for _, field := range view.Entries {
		if want[field.Path] != field.Role || field.Meaning == "" || field.InputDomain == "" {
			t.Fatalf("unknown or incomplete field description=%+v", field)
		}
		delete(want, field.Path)
	}
	if len(want) != 0 {
		t.Fatalf("missing field descriptions=%v", want)
	}
}

func TestRuntimeDescriptionFlagRequiresAnExplicitBuildJSONRequest(t *testing.T) {
	for _, args := range [][]string{{"--describe-runtime"}, {"--build", "--describe-runtime"},
		{"--json", "--describe-runtime"}, {"--build", "--json", "--describe-runtime", "--describe-runtime"}} {
		var output, diagnostics bytes.Buffer
		if runVersion(args, &output, &diagnostics) != exitUsage || output.Len() != 0 {
			t.Fatalf("description modifier accepted incomplete request: %v, stdout=%s", args, &output)
		}
	}
}

func TestRuntimeDescriptionsDoNotChangeOrdinaryVersionOutput(t *testing.T) {
	for _, args := range [][]string{{"--json"}, {"--build", "--json"}} {
		var output, diagnostics bytes.Buffer
		if runVersion(args, &output, &diagnostics) != exitOK || bytes.Contains(output.Bytes(), []byte("runtime_observation_fields")) {
			t.Fatalf("ordinary version changed: args=%v stdout=%s stderr=%s", args, &output, &diagnostics)
		}
	}
}
