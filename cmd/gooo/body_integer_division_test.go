package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBodyComposeIntegerDivisionAndSavedReplay(t *testing.T) {
	root, fixtures := t.TempDir(), "../../examples/integer-division/"
	for _, replay := range []bool{false, true} {
		args := []string{"body-compose", "--source", fixtures + "source.gooo.fixture", "--cases", fixtures + "cases.json", "--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go")}
		if replay {
			args = append(args, "--composition", filepath.Join(root, "program", "composition.json"))
		} else {
			args = append(args, "--out", filepath.Join(root, "program"))
		}
		var out, stderr bytes.Buffer
		if code := run(args, &out, &stderr); code != exitOK {
			t.Fatal("division composition failed", replay, code, stderr.String(), out.String())
		}
		var observed struct {
			Generated bool `json:"generated_now"`
			Runtime   struct {
				Passed int `json:"finite_passed"`
				Total  int `json:"finite_total"`
				Calls  int `json:"model_calls"`
			} `json:"runtime"`
		}
		if err := json.Unmarshal(out.Bytes(), &observed); err != nil || observed.Generated == replay || observed.Runtime.Passed != 8 || observed.Runtime.Total != 8 || observed.Runtime.Calls != 0 {
			t.Fatal("native division values or replay differ", observed, err)
		}
	}
}
