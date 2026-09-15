package main

import (
	"encoding/json"
)

func hasErrorFixDiagnostic(diagnostics []fixPlanDiagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == "error" {
			return true
		}
	}
	return false
}
func marshalFixPlan(plan fixPlan) ([]byte, error) {
	payload, err := json.Marshal(plan)
	if err != nil {
		return nil, err
	}
	if len(payload) > maxDiagnosticBytes {
		return nil, errDiagnosticLimit
	}
	return append(payload, '\n'), nil
}
