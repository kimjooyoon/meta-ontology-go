package toolchainrelease

import (
	"encoding/json"
	"fmt"
)

// The release witness checks known actual values without adding an oracle to the
// user's input-only receipt. Construction and later labelled evaluation stay separate.
func validatePackageInputRuntime(evaluation, inputs []byte, budget int, replay bool) error {
	var suite struct {
		Schema string
		Inputs []map[string]json.RawMessage
	}
	var result struct{ Runtime nativeSmokeRuntime }
	if err := json.Unmarshal(inputs, &suite); err != nil {
		return err
	}
	if err := json.Unmarshal(evaluation, &result); err != nil {
		return err
	}
	r := result.Runtime
	schema := "gooo/body-composition-runtime/v1"
	if replay {
		schema = "gooo/body-composition-runtime/v2"
	}
	if suite.Schema != "gooo/body-composition-inputs/v1" || len(suite.Inputs) != 4 ||
		r.Schema != schema || r.Stage != "COMPLETE" || r.Failure != "" ||
		!nativeSmokeDigest(r.SHA) || !jointSmokeInt(r.Passed, 0) || !jointSmokeInt(r.Total, 0) || !jointSmokeInt(r.Calls, 0) ||
		!jointSmokeBool(r.Projection, true) || !jointSmokeBool(r.Replay, true) || len(r.Traces) != len(suite.Inputs) {
		return fmt.Errorf("package input observations are incomplete or claim an evaluation score")
	}
	actual := []string{"3", "-10", "-9007199254740994", "1"}
	if budget == 5 {
		actual = []string{"3", "-20", "9007199254740995", "1"}
	}
	for i, row := range r.Traces {
		if row.Index != i || len(row.Deliveries) != 1 || len(suite.Inputs[i]) != 1 {
			return fmt.Errorf("package input row identity differs")
		}
		d := row.Deliveries[0]
		if d.ID != packageCallerID || d.Passed != nil || len(d.Expected) != 0 || d.Fault != nil || len(d.Blocked) != 0 ||
			!samePackageSourceValue(d.Input, suite.Inputs[i][packageCallerKey]) || !samePackageSourceValue(d.Actual, []byte(actual[i])) {
			return fmt.Errorf("package input row %d changed its value or invented an expectation", i)
		}
	}
	return nil
}
