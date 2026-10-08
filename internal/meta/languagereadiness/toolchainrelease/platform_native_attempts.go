package toolchainrelease

import (
	"encoding/json"
	"fmt"
)

func validateNativeArithmeticAttempts(r jointSmokeOutput, fault nativeSmokeRuntime, cases []jointSmokeCase) error {
	c := r.Construction
	plan := c.Initial.Preparations[0].Generation.Report.Fill.PlanSHA
	for i, a := range c.Attempts {
		if len(a.FillCandidates) != 1 || a.FillCandidates[0].PlanSHA != plan {
			return fmt.Errorf("native arithmetic candidate plan differs")
		}
		if i == 1 || i == 2 {
			if err := validateJointRejectedFillCount(a, i, 6, c.OriginalSHA); err != nil {
				return err
			}
			continue
		}
		local := map[int]int{0: 0, 3: 1, 4: 3, 5: 2}[i]
		if err := validateJointFillLocalAt(a, local, i, 6, c.OriginalSHA); err != nil {
			return err
		}
		if i == 4 {
			if fault.Source != a.FillCandidates[0].SelectedSHA {
				return fmt.Errorf("native fault runtime belongs to a different selected source")
			}
			if err := validateNativeCallerFault(fault, cases); err != nil {
				return err
			}
			continue
		}
		actual := []json.RawMessage{json.RawMessage([]string{"9", "-7", "-8"}[local])}
		if err := validateJointFillRuntime(a.Runtime, cases, actual); err != nil {
			return err
		}
		if err := validateNativeSmokeRuns(a.Runtime.Runs); err != nil {
			return err
		}
	}
	return nil
}

func validateNativeCallerFault(r nativeSmokeRuntime, cases []jointSmokeCase) error {
	if err := validateNativeSmokeRuntime(r, [5]int{0, 0, 1, 0, 0}, 1); err != nil {
		return err
	}
	if len(r.Traces[0].Deliveries) != 1 {
		return fmt.Errorf("native caller failure row differs")
	}
	d := r.Traces[0].Deliveries[0]
	if d.ID != "budgetplan://activity/main" || !samePackageSourceValue(d.Input, cases[0].Inputs["Main"]) ||
		!samePackageSourceValue(d.Expected, cases[0].Expected["Main"]) || !jointSmokeBool(d.Passed, false) || len(d.Blocked) > 0 {
		return fmt.Errorf("native caller failure changed original input or expectation")
	}
	return validateNativeSmokeFault(r, d, "budgetplan://activity/plan-budget", 8)
}
