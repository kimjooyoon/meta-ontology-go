package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

func smokeNativeArithmeticGraph(binary, work string, input BuildInput) error {
	source, err := os.ReadFile(filepath.Join(input.Root, nativeArithmeticRoot, "graph.gooo.fixture"))
	if err != nil {
		return err
	}
	cases, err := os.ReadFile(filepath.Join(input.Root, nativeArithmeticRoot, "graph-cases.json"))
	if err != nil {
		return err
	}
	selected := ""
	for _, replay := range []bool{false, true} {
		mode := "graph-construct"
		args := []string{"body-compose", "--source", filepath.Join(nativeArithmeticRoot, "graph.gooo.fixture"), "--cases", filepath.Join(nativeArithmeticRoot, "graph-cases.json")}
		if replay {
			mode = "graph-replay"
			args = append(args, "--composition", filepath.Join(work, "native-arithmetic-graph", "composition.json"))
		} else {
			args = append(args, "--out", filepath.Join(work, "native-arithmetic-graph"))
		}
		raw, err := runNativeArithmeticSmoke(binary, input, mode, args)
		if err != nil {
			return err
		}
		selected, err = validateNativeArithmeticGraph(raw, source, cases, replay, selected)
		if err != nil {
			return fmt.Errorf("TOOLCHAIN_RELEASE_NATIVE_GRAPH: %w", err)
		}
	}
	return nil
}

func validateNativeArithmeticGraph(raw, source, cases []byte, replay bool, selected string) (string, error) {
	var r struct {
		Generated   *bool `json:"generated_now"`
		Composition struct {
			Source string `json:"original_source_sha256"`
			SHA    string `json:"generated_sha256"`
		}
		Runtime nativeSmokeRuntime
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	if !jointSmokeBool(r.Generated, !replay) || r.Composition.Source != fmt.Sprintf("sha256:%x", sha256.Sum256(source)) || r.Runtime.Source != r.Composition.Source ||
		r.Composition.SHA != r.Runtime.SHA || replay && (selected == "" || selected != r.Composition.SHA) {
		return "", fmt.Errorf("native graph source or saved replay differs")
	}
	if err := validateNativeSmokeRuntime(r.Runtime, [5]int{6, 0, 1, 1, 0}, 2); err != nil {
		return "", err
	}
	var suite struct {
		Schema string
		Cases  []jointSmokeCase
	}
	if err := json.Unmarshal(cases, &suite); err != nil {
		return "", err
	}
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) != 2 {
		return "", fmt.Errorf("native graph original cases differ")
	}
	for i, row := range r.Runtime.Traces {
		if len(row.Deliveries) != 4 {
			return "", fmt.Errorf("native graph lost an activity")
		}
		if err := validateNativeGraphRow(r.Runtime, row.Deliveries, suite.Cases[i], i); err != nil {
			return "", err
		}
	}
	return r.Composition.SHA, nil
}

func validateNativeGraphRow(r nativeSmokeRuntime, rows []nativeSmokeDelivery, original jointSmokeCase, index int) error {
	byID := map[string]nativeSmokeDelivery{}
	for _, d := range rows {
		if _, ok := byID[d.ID]; ok {
			return fmt.Errorf("native graph duplicated an activity")
		}
		byID[d.ID] = d
	}
	if err := validateNativeGraphInputs(byID, original, index); err != nil {
		return err
	}
	for _, name := range []string{"First", "Divide", "Dependent", "Independent"} {
		id := map[string]string{"First": "first", "Divide": "divide", "Dependent": "dependent", "Independent": "independent"}[name]
		d, ok := byID["faults://activity/"+id]
		if !ok || !samePackageSourceValue(d.Expected, original.Expected[name]) {
			return fmt.Errorf("native graph identity or original expectation differs")
		}
		if index == 0 && name == "Divide" {
			if !jointSmokeBool(d.Passed, false) {
				return fmt.Errorf("native graph scored a fault")
			}
			if err := validateNativeSmokeFault(r, d, "faults://activity/divide", 9007199254740993); err != nil {
				return err
			}
		} else if index == 0 && name == "Dependent" {
			if !nativeSmokeNull(d.Actual) || !nativeSmokeNull(d.Input) || !jointSmokeBool(d.Passed, false) || d.Fault != nil ||
				d.Producer != "faults://activity/divide" || !slices.Equal(d.Blocked, []string{"faults://activity/divide"}) {
				return fmt.Errorf("native graph lost dependency blocking")
			}
		} else if d.Fault != nil || len(d.Blocked) > 0 || !jointSmokeBool(d.Passed, true) || !samePackageSourceValue(d.Actual, original.Expected[name]) {
			return fmt.Errorf("native graph lost an independent actual value")
		}
	}
	return nil
}

func validateNativeGraphInputs(rows map[string]nativeSmokeDelivery, original jointSmokeCase, index int) error {
	for name, id := range map[string]string{"First": "first", "Independent": "independent"} {
		d := rows["faults://activity/"+id]
		if d.Producer != "" || len(d.Inputs) != 0 || !samePackageSourceValue(d.Input, original.Inputs[name]) {
			return fmt.Errorf("native graph changed a root input")
		}
	}
	d := rows["faults://activity/divide"]
	if len(d.Inputs) != 2 || d.Inputs[0].Producer != "faults://activity/first" || d.Inputs[1].Producer != "" ||
		!samePackageSourceValue(d.Inputs[0].Value, rows["faults://activity/first"].Actual) ||
		!samePackageSourceValue(d.Inputs[1].Value, original.Inputs["Divide.input1"]) {
		return fmt.Errorf("native graph changed a joined input")
	}
	dependent := rows["faults://activity/dependent"]
	if dependent.Producer != "faults://activity/divide" || len(dependent.Inputs) != 0 ||
		index > 0 && !samePackageSourceValue(dependent.Input, d.Actual) {
		return fmt.Errorf("native graph changed a producer delivery")
	}
	return nil
}
