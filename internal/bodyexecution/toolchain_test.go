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

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestOwnedNativeToolchainReusesVersionAndKeepsOriginalProcess(t *testing.T) {
	source, doc, prior, parent := fixture(t)
	owner := NewExecutor()
	defer owner.Close()
	first, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	if first.Observation.ToolchainReference == nil || first.Observation.ToolchainReference.Reused || !first.Observation.Toolchain.Started ||
		first.CompletenessReceipt.ProfileID != RuntimeProfileV3 {
		t.Fatal("missing real first version observation")
	}
	savedOutput := bytes.Clone(first.Observation.ToolchainReference.SourceOutput)
	savedProcess := copyProcess(first.Observation.ToolchainReference.SourceCheck)
	*first.Observation.ToolchainReference.SourceCheck.ExitCode = 99
	first.Observation.ToolchainReference.SourceOutput[0] = '?'
	second, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	o := second.Observation
	if !o.ToolchainReference.Reused || !o.ToolchainReference.BytesVerified || !o.ToolchainReference.BuildInfoVerified ||
		!reflect.DeepEqual(o.Toolchain, ProcessObservation{}) || !o.Artifact.Reused || len(o.Runs) != 2 ||
		!bytes.Equal(o.ToolchainReference.SourceOutput, savedOutput) || !reflect.DeepEqual(o.ToolchainReference.SourceCheck, savedProcess) {
		t.Fatal("version reuse changed history or invented current work")
	}
	if dimension(t, second, "runtime_child_resources").Denominator != 2 {
		t.Fatal("retained toolchain/build costs counted as current")
	}
	wire, _ := json.Marshal(second)
	if _, err := DecodeRuntimeReceipt(wire); err != nil {
		t.Fatal(err)
	}
	testToolchainMutations(t, wire, prior)
	if err := owner.Close(); err != nil || owner.toolchain != nil {
		t.Fatal("Close retained owned toolchain", err)
	}
}

func testToolchainMutations(t *testing.T, wire []byte, prior bodycodegen.Result) {
	t.Helper()
	for name, change := range map[string]func(*Result){
		"missing":             func(r *Result) { r.Observation.ToolchainReference = nil },
		"schema":              func(r *Result) { r.Observation.ToolchainReference.Schema = "future/v1" },
		"key":                 func(r *Result) { r.Observation.ToolchainReference.KeySHA256 = "bad" },
		"tool":                func(r *Result) { r.Observation.ToolchainReference.GoToolSHA256 = digest([]byte("wrong tool")) },
		"observed-tool":       func(r *Result) { r.Observation.GoToolSHA256 = digest([]byte("wrong observed tool")) },
		"observed-version":    func(r *Result) { r.Observation.GoVersion += "changed" },
		"bytes-unverified":    func(r *Result) { r.Observation.ToolchainReference.BytesVerified = false },
		"metadata-unverified": func(r *Result) { r.Observation.ToolchainReference.BuildInfoVerified = false },
		"version":             func(r *Result) { r.Observation.ToolchainReference.GoVersion = "go version go1.27.0 linux/amd64" },
		"output":              func(r *Result) { r.Observation.ToolchainReference.SourceOutput[0] = '?' },
		"source-process-exit": func(r *Result) {
			r.Observation.ToolchainReference.SourceCheck.ExitCode = new(7)
			bindToolchain(r.Observation.ToolchainReference)
		},
		"source-process-stdout": func(r *Result) {
			r.Observation.ToolchainReference.SourceCheck.StdoutSHA256 = digest(nil)
			bindToolchain(r.Observation.ToolchainReference)
		},
		"source-process-stderr": func(r *Result) {
			r.Observation.ToolchainReference.SourceCheck.StderrSHA256 = digest([]byte("changed"))
			bindToolchain(r.Observation.ToolchainReference)
		},
		"binding":    func(r *Result) { r.Observation.ToolchainReference.SourceCheckSHA256 = digest(nil) },
		"reuse-flag": func(r *Result) { r.Observation.ToolchainReference.Reused = false },
		"zero-wall": func(r *Result) {
			r.Observation.ToolchainReference.SourceCheck.WallNS = 0
			bindToolchain(r.Observation.ToolchainReference)
		},
		"negative-cpu": func(r *Result) {
			r.Observation.ToolchainReference.SourceCheck.UserNS = -1
			bindToolchain(r.Observation.ToolchainReference)
		},
		"negative-rss": func(r *Result) {
			r.Observation.ToolchainReference.SourceCheck.PeakRSSBytes = new(int64(-1))
			bindToolchain(r.Observation.ToolchainReference)
		},
		"invented-current-process": func(r *Result) { r.Observation.Toolchain = copyProcess(r.Observation.ToolchainReference.SourceCheck) },
	} {
		t.Run(name, func(t *testing.T) {
			var r Result
			if err := json.Unmarshal(wire, &r); err != nil {
				t.Fatal(err)
			}
			change(&r)
			// Recompute the enclosing hashes to check the nested contract itself.
			r.CompletenessReceipt = runtimeCompleteness(prior, r)
			if name == "missing" {
				r.CompletenessReceipt.ProfileID = RuntimeProfileV3
			}
			changed, _ := json.Marshal(r)
			if _, err := DecodeRuntimeReceipt(changed); err == nil {
				t.Fatal("accepted inconsistent toolchain observation")
			}
		})
	}
}

func TestOwnedToolchainEnvironmentAndProjectionChanges(t *testing.T) {
	source, doc, prior, parent := fixture(t)
	owner := NewExecutor()
	defer owner.Close()
	first, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", t.TempDir())
	second, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool())
	if err != nil || second.Observation.ToolchainReference.Reused || second.Observation.Artifact.Reused ||
		second.Observation.ToolchainReference.KeySHA256 == first.Observation.ToolchainReference.KeySHA256 {
		t.Fatal("changed environment reused identity", err)
	}
	source = bytes.Replace(source, []byte("package sample"), []byte("package renamed"), 1)
	prior, err = bodycodegen.GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", doc, "")
	if err != nil {
		t.Fatal(err)
	}
	parent, _ = json.Marshal(prior.Report.CompletenessReceipt)
	third, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool())
	if err != nil || !third.Observation.ToolchainReference.Reused || third.Observation.Artifact.Reused || third.Observation.Artifact.MissReason != "key_changed" {
		t.Fatal("projection replacement lost valid version reference or reused old build", err)
	}
	if dimension(t, third, "runtime_child_resources").Denominator != 3 {
		t.Fatal("current rebuild/retained toolchain denominator")
	}
	fourth, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool())
	if err != nil || !fourth.Observation.ToolchainReference.Reused || !fourth.Observation.Artifact.Reused {
		t.Fatal("replacement dropped version owner", err)
	}
}

func TestShellToolchainNeverReusesVersionFromWrapperBytes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX wrapper")
	}
	source, doc, prior, parent := fixture(t)
	tool := filepath.Join(t.TempDir(), "go-wrapper")
	quoted := "'" + strings.ReplaceAll(nativeTool(), "'", "'\"'\"'") + "'"
	stub := "#!/bin/sh\nexec " + quoted + " \"$@\"\n"
	if err := os.WriteFile(tool, []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	if nativeGoBuildInfo(tool) {
		t.Fatal("shell metadata accepted")
	}
	owner := NewExecutor()
	defer owner.Close()
	for range 2 {
		r, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, tool)
		if err != nil || !r.Observation.Toolchain.Started || r.Observation.ToolchainReference != nil || r.CompletenessReceipt.ProfileID != RuntimeProfileV2 {
			t.Fatal("wrapper's unknown dependencies reused", err)
		}
		wire, _ := json.Marshal(r)
		if _, err := DecodeRuntimeReceipt(wire); err != nil {
			t.Fatal("legacy v2 wire", err)
		}
	}
}
