package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestReadOwnedRuntimeFirstBuildReuseAndPartialFailure(t *testing.T) {
	source, doc, prior, parent := fixture(t)
	owner := NewExecutor()
	defer owner.Close()
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	cases := []pathplan.TestCase{{Input: -4, Expected: -21}, {Input: 1, Expected: 10}}
	for call := range 2 {
		result, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, cases, tool)
		if err != nil {
			t.Fatal(err)
		}
		wire, _ := json.Marshal(result)
		before := bytes.Clone(wire)
		raw, err := DecodeRuntimeReceipt(wire)
		want, _ := json.Marshal(result.CompletenessReceipt)
		if err != nil || !bytes.Equal(raw, want) || !bytes.Equal(wire, before) ||
			result.Observation.Artifact.Reused != (call == 1) {
			t.Fatal("cannot read an owned observation without changing original bytes", err)
		}
		if call == 1 {
			testRejectChangedOwnedRecords(t, wire)
		}
	}
	failed, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, cases, t.TempDir())
	if err == nil {
		t.Fatal("expected unavailable tool failure")
	}
	wire, _ := json.Marshal(failed)
	if _, err := DecodeRuntimeReceipt(wire); err != nil {
		t.Fatal("lost a valid pre-artifact v1 failure observation", err)
	}
}

func testRejectChangedOwnedRecords(t *testing.T, wire []byte) {
	t.Helper()
	for name, mutate := range map[string]func(*Result){
		"parent-bytes":          func(r *Result) { r.ParentReceipt = append(r.ParentReceipt, '\n') },
		"future-profile":        func(r *Result) { r.CompletenessReceipt.ProfileID = "gooo/typed-path-runtime-v4" },
		"downgrade":             func(r *Result) { r.CompletenessReceipt.ProfileID = "gooo/typed-path-runtime-v1" },
		"missing-artifact":      func(r *Result) { r.Observation.Artifact = nil },
		"artifact-schema":       func(r *Result) { r.Observation.Artifact.Schema = "future/v1" },
		"owned-scope":           func(r *Result) { delete(r.CompletenessReceipt.Scope, "owned_artifact") },
		"missing-runtime-scope": func(r *Result) { delete(r.CompletenessReceipt.Scope, "runtime_scope") },
		"missing-executable-key": func(r *Result) {
			delete(r.CompletenessReceipt.Scope["runtime_scope"].(map[string]any), "executable_sha256")
		},
		"observation-digest":       func(r *Result) { r.CompletenessReceipt.Scope["runtime_observation_sha256"] = digest([]byte("other")) },
		"build-binding":            func(r *Result) { r.Observation.Artifact.SourceBuildSHA256 = digest([]byte("other")) },
		"consistent-build-binding": func(r *Result) { r.Observation.Artifact.SourceBuildSHA256 = digest([]byte("other")) },
		"consistent-future-schema": func(r *Result) { r.Observation.Artifact.Schema = "future/v1" },
		"consistent-invalid-key":   func(r *Result) { r.Observation.Artifact.KeySHA256 = "sha256:bad" },
		"consistent-build-exit": func(r *Result) {
			r.Observation.Artifact.SourceBuild.ExitCode = new(99)
			bindSourceBuild(&r.Observation)
		},
	} {
		t.Run(name, func(t *testing.T) {
			var r Result
			if err := json.Unmarshal(wire, &r); err != nil {
				t.Fatal(err)
			}
			mutate(&r)
			if strings.HasPrefix(name, "consistent-") {
				r.CompletenessReceipt.Scope["owned_artifact"] = r.Observation.Artifact
				observation, _ := json.Marshal(r.Observation)
				r.CompletenessReceipt.Scope["runtime_observation_sha256"] = digest(observation)
			}
			changed, _ := json.Marshal(r)
			if reflect.DeepEqual(changed, wire) {
				t.Fatal("mutation did not change the envelope")
			}
			if _, err := DecodeRuntimeReceipt(changed); err == nil {
				t.Fatal("accepted inconsistent or unsupported owned observation")
			}
		})
	}
}

func TestReadOwnedFailedBuildWithoutSuccessfulArtifactBinding(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX failure stub")
	}
	source, doc, prior, parent := fixture(t)
	tool := filepath.Join(t.TempDir(), "failed-go-stub")
	stub := "#!/bin/sh\nif [ \"$1\" = version ]; then echo 'go version go1.27.1 " + runtime.GOOS + "/" + runtime.GOARCH + "'; exit 0; fi\nexit 7\n"
	if err := os.WriteFile(tool, []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	owner := NewExecutor()
	defer owner.Close()
	r, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, tool)
	if err == nil || r.Observation.Stage != "BUILD" || r.Observation.Artifact == nil ||
		r.Observation.Build.ExitCode == nil || *r.Observation.Build.ExitCode != 7 || r.Observation.Artifact.SourceBuildSHA256 != "" {
		t.Fatal("missing real failed child observation", err)
	}
	wire, _ := json.Marshal(r)
	if _, err := DecodeRuntimeReceipt(wire); err != nil {
		t.Fatal("owned failure was lost because no successful build exists", err)
	}
}
