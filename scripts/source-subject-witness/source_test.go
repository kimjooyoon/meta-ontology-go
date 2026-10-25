package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSourceBindsAuthoritativeWorkflowDiscoveryPolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source-metrics.json")
	data := []byte(`{"meta":{"policy":{"exempt_workflow_discovery_root":true}}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := loadSource(path)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Meta.Policy.ExemptWorkflowDiscoveryRoot {
		t.Fatal("workflow discovery root policy was not bound")
	}
}

func TestUnknownSourceFieldsRemainUnsupported(t *testing.T) {
	for _, test := range []struct{ raw, reason string }{
		{`{"files":[{"path":"a.go","module_origin_new":"unrecognized"}]}`, "SOURCE_SCHEMA_UNSUPPORTED"},
		{`{"files":`, "SOURCE_JSON_MALFORMED"},
		{`{} {}`, "SOURCE_JSON_TRAILING_DATA"},
	} {
		path := filepath.Join(t.TempDir(), "metrics.json")
		if err := os.WriteFile(path, []byte(test.raw), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := loadSource(path)
		var failure *sourceValidationError
		if !errors.As(err, &failure) || failure.Reason != test.reason {
			t.Fatalf("unexpected refusal: %v", err)
		}
		if test.reason == "SOURCE_SCHEMA_UNSUPPORTED" && failure.NextOperation != "update-source-metrics-consumer" {
			t.Fatal("unknown fields blamed on source JSON", failure)
		}
	}
}
