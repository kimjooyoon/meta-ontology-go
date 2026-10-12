package toolchainrelease

import (
	"encoding/json"
	"fmt"
)

func validateRelativeV3Runtime(r jointSmokeRuntime, cases []byte, total int) error {
	rows, err := jointSmokeCases(cases, total)
	if err != nil {
		return err
	}
	if r.Stage != "COMPLETE" || r.Failure != "" || !jointSmokeInt(r.Passed, total) || !jointSmokeInt(r.Total, total) ||
		!jointSmokeInt(r.Calls, 0) || !jointSmokeBool(r.Projection, true) || !jointSmokeBool(r.Replay, true) || len(r.Traces) != total {
		return fmt.Errorf("V3 native denominator or zero inference differs")
	}
	seen := make([]bool, total)
	for _, trace := range r.Traces {
		i := trace.Index
		if i < 0 || i >= total || seen[i] || len(trace.Deliveries) != 1 {
			return fmt.Errorf("V3 native index missing or duplicated")
		}
		seen[i] = true
		delivery := trace.Deliveries[0]
		q := rows[i]
		var input, want int64
		if err = json.Unmarshal(q.Inputs["Main"], &input); err != nil {
			return err
		}
		want = input + 13
		if input < 67 {
			want = input / 4
		}
		expected, err := json.Marshal(want)
		if err != nil {
			return err
		}
		if delivery.ID != "output-relations-v3://main" || !jointSmokeBool(delivery.Passed, true) ||
			!samePackageSourceValue(delivery.Input, q.Inputs["Main"]) || !samePackageSourceValue(q.Expected["Main"], expected) ||
			!samePackageSourceValue(delivery.Actual, expected) || !samePackageSourceValue(delivery.Expected, expected) {
			return fmt.Errorf("V3 exact native value differs at%d", i)
		}
	}
	return nil
}

func validateRelativeV3Envelope(r nativeSmokeRuntime, total int, schema, program, source string) error {
	if r.Schema != schema || r.Stage != "COMPLETE" || r.Failure != "" || r.SHA != program || r.Source != source ||
		!nativeSmokeDigest(program) || !nativeSmokeDigest(source) || !nativeSmokeDigest(r.ObservedProjection) || !nativeSmokeDigest(r.ObservedDriver) ||
		!jointSmokeBool(r.Projection, true) || !jointSmokeBool(r.Replay, true) || !jointSmokeInt(r.Calls, 0) ||
		!jointSmokeInt(r.Passed, total) || !jointSmokeInt(r.Total, total) || r.Outcomes == nil || len(r.Traces) != total {
		return fmt.Errorf("V3 runtime source/program identity or counters differ")
	}
	values := [5]*int{r.Outcomes.Matched, r.Outcomes.Mismatched, r.Outcomes.Faulted, r.Outcomes.Blocked, r.Outcomes.Unobserved}
	for i, value := range values {
		want := 0
		if i == 0 {
			want = total
		}
		if !jointSmokeInt(value, want) {
			return fmt.Errorf("V3 runtime outcome category differs")
		}
	}
	return nil
}
