package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const packageTextManifest = "examples/text-operations/gooo.workspace.json"
const packageTextInputs = "examples/text-operations/inputs.json"

func smokePackageText(binary string, input BuildInput) error {
	receipt := filepath.Join(input.OutputDir, input.Target.ID+"-package-text-construct.json")
	selected := ""
	for _, replay := range []bool{false, true} {
		mode, output := "execute", receipt
		args := []string{"package", mode, "--json", "--inputs", packageTextInputs}
		if replay {
			args[1] = "replay"
			args = append(args, "--receipt", receipt)
			output = filepath.Join(input.OutputDir, input.Target.ID+"-package-text-replay.json")
		}
		raw, err := commandOutput(input.Root, nil, binary, append(args, packageTextManifest)...)
		if len(raw) > 0 {
			if writeErr := os.WriteFile(output, raw, 0o644); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_TEXT: %w", err)
		}
		selected, err = validatePackageTextSmoke(raw, replay, selected)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_TEXT: %w", err)
		}
	}
	return nil
}

type packageTextValue struct {
	Source *bool   `json:"source"`
	Stem   *string `json:"stem"`
	Bytes  *int64  `json:"bytes"`
}

func validatePackageTextSmoke(raw []byte, replay bool, selected string) (string, error) {
	var r struct {
		Schema, Decision, Error string
		From                    string `json:"replayed_from_sha256"`
		Result                  struct {
			Composition struct {
				SHA string `json:"generated_sha256"`
			} `json:"composition"`
			Replay *struct {
				Calls *int `json:"model_calls"`
			}
			Runtime struct {
				Calls      *int `json:"model_calls"`
				Passed     *int `json:"finite_passed"`
				Total      *int `json:"finite_total"`
				Projection bool `json:"projection_replayed"`
				Replay     bool `json:"runtime_replayed"`
				Traces     []struct {
					Index      int `json:"case_index"`
					Deliveries []struct {
						ID       string           `json:"activity_id"`
						Actual   packageTextValue `json:"actual"`
						Expected json.RawMessage  `json:"expected"`
					} `json:"deliveries"`
				} `json:"traces"`
			} `json:"runtime"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	x := r.Result.Runtime
	zero := func(v *int) bool { return v != nil && *v == 0 }
	if r.Schema != "gooo/workspace-body-execution-receipt/v1" || r.Decision != "OBSERVED" || r.Error != "" ||
		!zero(x.Calls) || !zero(x.Passed) || !zero(x.Total) || !x.Projection || !x.Replay || len(x.Traces) != 3 ||
		r.Result.Composition.SHA == "" {
		return "", fmt.Errorf("input-only package observation is incomplete")
	}
	if replay && (selected == "" || r.Result.Composition.SHA != selected || r.From == "" ||
		r.Result.Replay == nil || !zero(r.Result.Replay.Calls)) {
		return "", fmt.Errorf("package replay is unbound or predicted again")
	}
	wantStem, wantBytes, wantSource := []string{"한🙂", ".hidden", "notes.txt"}, []int64{12, 12, 9}, []bool{true, false, false}
	seen := [3]bool{}
	for _, trace := range x.Traces {
		i := trace.Index
		if i < 0 || i >= len(seen) || seen[i] || len(trace.Deliveries) != 1 {
			return "", fmt.Errorf("package case identity differs")
		}
		seen[i] = true
		d := trace.Deliveries[0]
		v := d.Actual
		if d.ID != "gooo-workspace://activity/gooo-package4activity-classify" || len(d.Expected) != 0 ||
			v.Source == nil || *v.Source != wantSource[i] || v.Stem == nil || *v.Stem != wantStem[i] ||
			v.Bytes == nil || *v.Bytes != wantBytes[i] {
			return "", fmt.Errorf("package filename output differs at case %d", i)
		}
	}
	return r.Result.Composition.SHA, nil
}
