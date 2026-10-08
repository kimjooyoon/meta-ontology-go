package toolchainrelease

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

const packageSourceRoot = "examples/package-body-calls"

func smokePackageSource(binary string, input BuildInput) error {
	cases, err := os.ReadFile(filepath.Join(input.Root, packageSourceRoot, "cases.json"))
	if err != nil {
		return err
	}
	selected := ""
	for _, variant := range []string{"explicit", "source"} {
		manifest := "gooo.workspace.json"
		if variant == "source" {
			manifest = "source.workspace.json"
		}
		receipt := filepath.Join(input.OutputDir, input.Target.ID+"-package-"+variant+"-construct.json")
		for _, replay := range []bool{false, true} {
			output := receipt
			args := []string{"package", "execute", "--json", "--cases", filepath.Join(packageSourceRoot, "cases.json")}
			if replay {
				args[1] = "replay"
				args = append(args, "--receipt", receipt)
				output = filepath.Join(input.OutputDir, input.Target.ID+"-package-"+variant+"-replay.json")
			}
			raw, err := commandOutput(input.Root, nil, binary, append(args, filepath.Join(packageSourceRoot, manifest))...)
			if len(raw) > 0 {
				if writeErr := os.WriteFile(output, raw, 0o644); writeErr != nil {
					return writeErr
				}
			}
			if err != nil {
				return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_SOURCE %s: %w", variant, err)
			}
			selected, err = validatePackageSourceSmoke(raw, cases, replay, selected)
			if err != nil {
				return fmt.Errorf("TOOLCHAIN_RELEASE_PACKAGE_SOURCE %s: %w", variant, err)
			}
		}
	}
	return nil
}

type packageSourceDelivery struct {
	ID       string          `json:"activity_id"`
	Actual   json.RawMessage `json:"actual"`
	Expected json.RawMessage `json:"expected"`
	Passed   *bool           `json:"passed"`
}

func validatePackageSourceSmoke(raw, cases []byte, replay bool, selected string) (string, error) {
	var reference struct {
		Cases []struct {
			Expected map[string]json.RawMessage
		}
	}
	var r struct {
		Schema, Decision, Error string
		From                    string `json:"replayed_from_sha256"`
		Result                  struct {
			Composition struct {
				SHA string `json:"generated_sha256"`
			}
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
					Deliveries []packageSourceDelivery
				}
			}
		}
	}
	if err := json.Unmarshal(cases, &reference); err != nil {
		return "", err
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	x, sha := r.Result.Runtime, r.Result.Composition.SHA
	zero := func(v *int) bool { return v != nil && *v == 0 }
	if len(reference.Cases) != 4 || len(x.Traces) != 4 ||
		r.Schema != "gooo/workspace-body-execution-receipt/v1" || r.Decision != "PASS" || r.Error != "" ||
		x.Passed == nil || *x.Passed != 8 || x.Total == nil || *x.Total != 8 ||
		!zero(x.Calls) || !x.Projection || !x.Replay || sha == "" || (selected != "" && selected != sha) {
		return "", fmt.Errorf("package construction, native results or selected program differ")
	}
	if replay && (selected == "" || r.From == "" || r.Result.Replay == nil || !zero(r.Result.Replay.Calls)) {
		return "", fmt.Errorf("source package replay is unbound or predicted again")
	}
	ids := map[string]string{
		"gooo-workspace://activity/gooo-package2activity-diagnose": "tools/diagnostics:Diagnose",
		"gooo-workspace://activity/gooo-package3activity-main":     "app/explain:Main",
	}
	seen := [4]bool{}
	for _, trace := range x.Traces {
		if trace.Index < 0 || trace.Index >= len(seen) || seen[trace.Index] || len(trace.Deliveries) != 2 {
			return "", fmt.Errorf("source package input row differs")
		}
		seen[trace.Index] = true
		expected := reference.Cases[trace.Index].Expected
		delivered := map[string]bool{}
		for _, d := range trace.Deliveries {
			name := ids[d.ID]
			if name == "" || delivered[name] || len(expected) != 2 || d.Passed == nil || !*d.Passed ||
				!samePackageSourceValue(d.Actual, expected[name]) || !samePackageSourceValue(d.Expected, expected[name]) {
				return "", fmt.Errorf("source package output differs at row %d", trace.Index)
			}
			delivered[name] = true
		}
	}
	return sha, nil
}

func samePackageSourceValue(a, b []byte) bool {
	decode := func(raw []byte) (any, error) {
		var value any
		d := json.NewDecoder(bytes.NewReader(raw))
		d.UseNumber()
		err := d.Decode(&value)
		return value, err
	}
	x, xe := decode(a)
	y, ye := decode(b)
	return xe == nil && ye == nil && reflect.DeepEqual(x, y)
}
