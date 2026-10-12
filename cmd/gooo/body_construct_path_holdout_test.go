package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestPathHoldoutReportSeparatesInitialFinalAndCallerExposure(t *testing.T) {
	percent := 100.0
	c := bodyexecution.JointConstruction{FinalPathHoldouts: []bodyexecution.JointPathHoldout{{
		Evaluation:     &bodycodegen.PathHoldoutReceipt{Activity: "Fee", Passed: 1, Total: 1, AccuracyPercent: &percent},
		CallerExposure: bodyexecution.PathHoldoutCallerExposure{ObservedInputs: []int64{59}},
	}}}
	raw, _ := json.Marshal(constructionFinalPathHoldoutRows(c))
	for _, s := range []string{"Final selected body holdout", "Fee: 1/1 (100.00%)", "1 observed; 0 other; 0 unknown", "measured after selection"} {
		if !strings.Contains(string(raw), s) {
			t.Fatal("report lost evaluation scope", s, string(raw))
		}
	}
	c.FinalPathHoldouts[0].Evaluation.AccuracyPercent = nil
	c.FinalPathHoldouts[0].Evaluation.Error = "zero divisor"
	if got := constructionPathHoldoutShare(c.FinalPathHoldouts[0].Evaluation); !strings.Contains(got, "unavailable: zero divisor") {
		t.Fatal(got)
	}
}

func TestSourcePathHoldoutNativeCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("fresh source path holdout example runs in CI")
	}
	root := filepath.Join(t.TempDir(), "holdout")
	base := "../../examples/source-path-holdouts/"
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}
	call := func(args ...string) bodyConstructOutput {
		t.Helper()
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK {
			t.Fatal(code, out.String(), diagnostics.String())
		}
		var result bodyConstructOutput
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	r := call("body-construct", "--source", base+"main.gooo.fixture", "--entry", "Main",
		"--construction-cases", base+"construction.json", "--cases", base+"evaluation.json", "--attempts", "2", "--go-bin", goBin, "--out", root)
	if r.Construction.Schema != "gooo/joint-construction/v8" || r.Construction.SelectedAttempt != 1 ||
		len(r.Construction.FinalPathHoldouts) != 1 || r.Construction.FinalPathHoldouts[0].Evaluation.Passed != 1 ||
		r.Evaluation.Runtime.FinitePassed != 3 || r.Evaluation.Runtime.ModelCalls != 0 {
		t.Fatal(r)
	}
	replay := call("body-construct", "--source", filepath.Join(root, "original.gooo"), "--construction", filepath.Join(root, "construction.json"),
		"--cases", base+"evaluation.json", "--go-bin", goBin)
	if replay.GeneratedNow || !replay.Evaluation.ConstructionReplayed || replay.Evaluation.NewModelCalls != 0 || replay.Evaluation.Runtime.FinitePassed != 3 {
		t.Fatal("saved selected body lost its holdout", replay)
	}
}
