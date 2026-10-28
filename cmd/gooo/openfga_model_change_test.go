package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestOpenFGAModelChangeSourceCompositionAndReplay(t *testing.T) {
	root := t.TempDir()
	for _, replay := range []bool{false, true} {
		name := "first"
		args := []string{"body-compose", "--source", "../../examples/openfga-model-change/main.gooo",
			"--cases", "../../examples/openfga-model-change/cases.json"}
		if replay {
			name = "replay"
			args = append(args, "--composition", filepath.Join(root, "first", "composition.json"))
		} else {
			args = append(args, "--entry", "Compare")
		}
		args = append(args, "--out", filepath.Join(root, name))
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != exitOK {
			t.Fatalf("%s: %d / %s", name, code, stderr.String())
		}
		var result bodyCompositionOutput
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.GeneratedNow == replay || result.Runtime.FinitePassed != 21 || result.Runtime.FiniteTotal != 21 ||
			!result.Runtime.RuntimeReplayed || result.Runtime.ModelCalls != 0 || len(result.Runtime.Traces) != 21 {
			t.Fatalf("%s lost its finite scope, deterministic execution or replay identity", name)
		}
		var unknown, renamed struct {
			Change       string `json:"change"`
			IntentResult string `json:"intent_result"`
			BeforeState  string `json:"before_state"`
			AfterState   string `json:"after_state"`
			BeforeError  string `json:"before_error"`
			AfterError   string `json:"after_error"`
		}
		if err := json.Unmarshal(result.Runtime.Traces[4].Deliveries[0].Actual, &unknown); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(result.Runtime.Traces[20].Deliveries[0].Actual, &renamed); err != nil {
			t.Fatal(err)
		}
		if unknown.Change != "UNKNOWN" || unknown.IntentResult != "UNASSESSED" ||
			unknown.BeforeState != "ERROR" || unknown.AfterState != "ERROR" ||
			unknown.BeforeError != "Code(2000)" || unknown.AfterError != "Code(2000)" ||
			renamed.Change != "EXPANSION" || renamed.IntentResult != "EXPECTED" {
			t.Fatal("error response became a denial or display label changed source intent", unknown, renamed)
		}
	}
}
