package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBodyComposeCandidateLocalsAndReplay(t *testing.T) {
	root, fixtures := t.TempDir(), "../../examples/candidate-locals/"
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	for _, replay := range []bool{false, true} {
		args := []string{"body-compose", "--source", fixtures + "retry.gooo.fixture", "--cases", fixtures + "cases.json", "--go-bin", tool}
		if replay {
			args = append(args, "--composition", filepath.Join(root, "program", "composition.json"))
		} else {
			args = append(args, "--out", filepath.Join(root, "program"))
		}
		var out, stderr bytes.Buffer
		if code := run(args, &out, &stderr); code != exitOK {
			t.Fatal("candidate-local execution failed", replay, code, stderr.String(), out.String())
		}
		var observed bodyCompositionOutput
		if err := json.Unmarshal(out.Bytes(), &observed); err != nil || observed.GeneratedNow == replay ||
			observed.Runtime.FinitePassed != 12 || observed.Runtime.FiniteTotal != 12 || observed.Runtime.ModelCalls != 0 {
			t.Fatal("native values or saved replay changed", observed.Runtime, err)
		}
	}
}
