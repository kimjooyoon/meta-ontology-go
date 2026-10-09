package toolchainrelease

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func validatePlanScoredTraces(r planInputRuntime, plan bodyexecution.CompositionPlan, total int) error {
	count := 0
	for i, trace := range r.Traces {
		if trace.CaseIndex != i || len(trace.Deliveries) != len(plan.Activities) {
			return fmt.Errorf("scored input delivery count or order differs")
		}
		for j, d := range trace.Deliveries {
			if d.ActivityID != plan.Activities[j].ID || d.Passed == nil || !*d.Passed || d.Fault != nil ||
				len(d.BlockedBy) != 0 || !planInputEqualValue(d.Actual, d.Expected) {
				return fmt.Errorf("scored input native result differs from its supplied expectation")
			}
			count++
		}
	}
	if count != total {
		return fmt.Errorf("scored input named expectation count differs")
	}
	return nil
}

func planInputEqualValue(a, b json.RawMessage) bool {
	if nativeSmokeNull(a) || nativeSmokeNull(b) {
		return false
	}
	decode := func(raw []byte) (any, error) {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var value any
		err := decoder.Decode(&value)
		return value, err
	}
	actual, actualErr := decode(a)
	expected, expectedErr := decode(b)
	return actualErr == nil && expectedErr == nil && reflect.DeepEqual(actual, expected)
}
