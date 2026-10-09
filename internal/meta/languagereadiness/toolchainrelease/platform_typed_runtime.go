package toolchainrelease

import (
	"fmt"
)

func validateTypedRuntime(r jointSmokeRuntime, cases []byte) error {
	rows, err := jointSmokeCases(cases, 3)
	if err != nil {
		return err
	}
	if r.Stage != "COMPLETE" || r.Failure != "" || r.SHA == "" || !jointSmokeInt(r.Passed, 3) ||
		!jointSmokeInt(r.Total, 3) || !jointSmokeInt(r.Calls, 0) || !jointSmokeBool(r.Projection, true) ||
		!jointSmokeBool(r.Replay, true) || len(r.Traces) != 3 {
		return fmt.Errorf("typed native evaluation counts or saved replay differs")
	}
	seen := [3]bool{}
	for _, trace := range r.Traces {
		i := trace.Index
		if i < 0 || i >= len(rows) || seen[i] || len(trace.Deliveries) != 1 {
			return fmt.Errorf("typed native evaluation lost an input row")
		}
		seen[i] = true
		d, row := trace.Deliveries[0], rows[i]
		if d.ID != "callerpaths://activity/main" || !jointSmokeBool(d.Passed, true) ||
			!samePackageSourceValue(d.Input, row.Inputs["Main"]) ||
			!samePackageSourceValue(d.Actual, row.Expected["Main"]) ||
			!samePackageSourceValue(d.Expected, row.Expected["Main"]) {
			return fmt.Errorf("typed exact native input or value differs at row %d", i)
		}
	}
	return nil
}
