package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolveGoToolFixedLocalOrderAndExplicitPriority(t *testing.T) {
	root := t.TempDir()
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	rootTool := filepath.Join(root, "bin", name)
	pathTool := filepath.Join(root, "path", name)
	explicit := filepath.Join(root, "custom", name)
	for _, tc := range []struct {
		name, requested, path, goroot, want, origin string
		pathSupported, rootSupported                bool
		wantError                                   bool
	}{
		{name: "explicit wins", requested: explicit, path: pathTool, goroot: root,
			want: explicit, origin: "explicit", rootSupported: true},
		{name: "missing explicit never substitutes", requested: "missing", path: pathTool,
			goroot: root, rootSupported: true, wantError: true},
		{name: "matching PATH wins", path: pathTool, goroot: root,
			want: pathTool, origin: "path", pathSupported: true, rootSupported: true},
		{name: "wrong PATH uses compiler root", path: pathTool, goroot: root,
			want: rootTool, origin: "compiler_goroot", rootSupported: true},
		{name: "missing PATH uses compiler root", goroot: root,
			want: rootTool, origin: "compiler_goroot", rootSupported: true},
		{name: "unavailable root retains PATH observation", path: pathTool, goroot: root,
			want: pathTool, origin: "path_fallback"},
		{name: "empty root retains PATH observation", path: pathTool,
			want: pathTool, origin: "path_fallback"},
		{name: "no local executable", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var inspected []string
			lookup := func(input string) (string, error) {
				switch {
				case input == "go" && tc.path != "":
					return tc.path, nil
				case input == rootTool && tc.rootSupported:
					return rootTool, nil
				case input == explicit:
					return explicit, nil
				}
				return "", os.ErrNotExist
			}
			check := func(path string) bool {
				inspected = append(inspected, path)
				return path == pathTool && tc.pathSupported || path == rootTool && tc.rootSupported
			}
			got, origin, err := resolveGoTool(tc.requested, tc.goroot, "", lookup, check)
			if (err != nil) != tc.wantError || got != tc.want || origin != tc.origin {
				t.Fatalf("selection = %q, %q, %v", got, origin, err)
			}
			if tc.requested != "" && len(inspected) != 0 {
				t.Fatal("explicit selection inspected alternatives", inspected)
			}
			if len(inspected) > 2 || tc.pathSupported && len(inspected) != 1 {
				t.Fatal("default selection expanded the fixed local search", inspected)
			}
		})
	}
}

func TestDefaultGoToolExecutesWithWrongPATHAndPreservesExplicitFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX wrong-version executable fixture")
	}
	root := t.TempDir()
	wrong := filepath.Join(root, "go")
	stub := "#!/bin/sh\necho 'go version go1.26.5 " + runtime.GOOS + "/" + runtime.GOARCH + "'\n"
	if err := os.WriteFile(wrong, []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
	source, doc, prior, parent := fixture(t)
	owner := NewExecutor()
	defer owner.Close()
	for i := range 2 {
		r, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, "")
		if err != nil || r.Observation.Stage != "COMPLETE" || r.Observation.GoToolPath != nativeTool() ||
			r.Observation.GoToolSelection != "compiler_goroot" || len(r.Observation.Runs) != 2 ||
			r.Observation.Artifact.Reused != (i == 1) || r.Observation.ToolchainReference.Reused != (i == 1) {
			t.Fatalf("default execution: %+v, %v", r.Observation, err)
		}
		for _, c := range r.Observation.Cases {
			if !c.Passed {
				t.Fatal("default selection changed finite behavior", c)
			}
		}
		verifyReceipt(t, r)
		wire, _ := json.Marshal(r)
		if _, err := DecodeRuntimeReceipt(wire); err != nil {
			t.Fatal("new fields cannot be read", err)
		}
		r.Observation.GoToolPath = wrong
		changed, _ := json.Marshal(r)
		if _, err := DecodeRuntimeReceipt(changed); err == nil {
			t.Fatal("changed selection path retained the old observation binding")
		}
		// Earlier envelopes omitted both optional selection fields. Their original
		// observation digest remains valid when decoded by the current reader.
		r.Observation.GoToolPath, r.Observation.GoToolSelection = "", ""
		observation, _ := json.Marshal(r.Observation)
		r.CompletenessReceipt.Scope["runtime_observation_sha256"] = digest(observation)
		legacy, _ := json.Marshal(r)
		if _, err := DecodeRuntimeReceipt(legacy); err != nil {
			t.Fatal("legacy observation without selection fields rejected", err)
		}
	}
	r, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, wrong)
	if err == nil || !strings.Contains(err.Error(), "go1.26.5") || r.Observation.GoToolSelection != "explicit" ||
		r.Observation.GoToolPath != wrong || !r.Observation.Toolchain.Completed || len(r.Observation.Runs) != 0 {
		t.Fatal("explicit version mismatch was silently replaced", r.Observation, err)
	}
}
