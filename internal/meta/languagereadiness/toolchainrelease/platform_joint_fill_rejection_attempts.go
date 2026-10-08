package toolchainrelease

import (
	"encoding/json"
	"fmt"
)

func validateJointFillRejectionAttempts(r jointSmokeOutput, construction []jointSmokeCase) error {
	c := r.Construction
	plan := c.Initial.Preparations[0].Generation.Report.Fill.PlanSHA
	for i, a := range c.Attempts {
		if len(a.FillCandidates) != 1 || a.FillCandidates[0].PlanSHA != plan {
			return fmt.Errorf("rejected-fill plan or candidate count differs")
		}
		if i == 1 || i == 2 {
			if err := validateJointRejectedFill(a, i, c.OriginalSHA); err != nil {
				return err
			}
			continue
		}
		local := i
		if i > 0 {
			local -= 2
		}
		if err := validateJointFillLocalAt(a, local, i, 5, c.OriginalSHA); err != nil {
			return err
		}
		actual := []json.RawMessage{json.RawMessage([]string{"9", "-7", "-8"}[local])}
		if err := validateJointFillRuntime(a.Runtime, construction, actual); err != nil {
			return err
		}
	}
	return nil
}
