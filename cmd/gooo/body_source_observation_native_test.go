package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSourceOwnedObservationNativeCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("fresh native Gooo example runs in CI")
	}
	root := filepath.Join(t.TempDir(), "composition")
	source := "../../examples/source-owned-observation/main.gooo.fixture"
	cases := "../../examples/source-owned-observation/evaluation.json"
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}
	call := func(args ...string) bodyCompositionOutput {
		t.Helper()
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK {
			t.Fatal(code, out.String(), diagnostics.String())
		}
		var result bodyCompositionOutput
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	first := call("body-compose", "--source", source, "--entry", "Main", "--cases", cases,
		"--go-bin", goBin, "--out", root)
	if first.Runtime.FinitePassed != 4 || first.Runtime.FiniteTotal != 4 {
		t.Fatal("Gooo source observation failed separate native inputs")
	}
	observations := 0
	for _, preparation := range first.Composition.Preparations {
		p := preparation.Generation.Report.BodyPaths
		if p != nil && p.Observation != nil {
			observations++
			if p.Observation.Status != "ONE_SURVIVING_CANDIDATE" || p.Observation.EffectiveCases != 2 ||
				p.DeclaredTestCases != 1 || p.LocalModelRequested {
				t.Fatal("native preparation lost source questions or introduced a model")
			}
		}
	}
	if observations != 1 {
		t.Fatal("expected exactly one source-owned observation contract", observations)
	}
	replay := call("body-compose", "--source", filepath.Join(root, "original.gooo"),
		"--composition", filepath.Join(root, "composition.json"), "--cases", cases, "--go-bin", goBin)
	if replay.GeneratedNow || !replay.Runtime.RuntimeReplayed || replay.Runtime.ModelCalls != 0 ||
		replay.Runtime.FinitePassed != 4 ||
		replay.Composition.GeneratedSHA256 != first.Composition.GeneratedSHA256 {
		t.Fatal("saved source observation changed during native replay")
	}
}
