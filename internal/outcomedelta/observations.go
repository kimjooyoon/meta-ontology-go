package outcomedelta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

type observation struct {
	record                      map[string]any
	outcome, expected, identity string
	state                       string
	passed                      *bool
}

type group struct {
	activity     string
	inputs       []map[string]any
	observations []observation
}

func groups(runtime bodyexecution.CompositionRuntime) (map[string]*group, error) {
	result := map[string]*group{}
	indices := map[int]bool{}
	passed, total, deliveries := 0, 0, 0
	for _, trace := range runtime.Traces {
		if trace.CaseIndex < 0 || indices[trace.CaseIndex] {
			return nil, fmt.Errorf("duplicate or invalid case index")
		}
		indices[trace.CaseIndex] = true
		roots, rootKey, err := rootInputs(trace)
		if err != nil {
			return nil, err
		}
		for _, delivery := range trace.Deliveries {
			deliveries++
			if deliveries > 8192 {
				return nil, fmt.Errorf("runtime exceeds 8192 delivery observations")
			}
			observed, err := snapshot(trace.CaseIndex, delivery)
			if err != nil {
				return nil, fmt.Errorf("case %d activity %s: %w", trace.CaseIndex, delivery.ActivityID, err)
			}
			if observed.expected != "" {
				total++
			}
			if observed.passed != nil {
				if *observed.passed {
					passed++
				}
			}
			keyBytes, _ := json.Marshal([]string{rootKey, delivery.ActivityID})
			key := string(keyBytes)
			if result[key] == nil {
				result[key] = &group{activity: delivery.ActivityID, inputs: roots}
			}
			result[key].observations = append(result[key].observations, observed)
		}
	}
	if runtime.Stage == "COMPLETE" && (passed != runtime.FinitePassed || total != runtime.FiniteTotal) {
		return nil, fmt.Errorf("complete runtime totals differ from observed expectations")
	}
	return result, nil
}

// Join on the entire caller input tuple, not a row index or intermediate input.
// Bound producer values may change after a rule change and must not change this key.
func rootInputs(trace bodyexecution.CompositionTrace) ([]map[string]any, string, error) {
	roots := []map[string]any{}
	ids := map[string]bool{}
	for _, d := range trace.Deliveries {
		if d.ActivityID == "" || ids[d.ActivityID] {
			return nil, "", fmt.Errorf("missing or duplicate activity ID")
		}
		ids[d.ActivityID] = true
	}
	add := func(id, port, producer string, raw json.RawMessage) error {
		if producer != "" {
			if !ids[producer] {
				return fmt.Errorf("unknown input producer %s", producer)
			}
			return nil
		}
		if len(raw) == 0 {
			return fmt.Errorf("caller input is missing")
		}
		value, _, err := canonical(raw)
		if err != nil {
			return err
		}
		roots = append(roots, map[string]any{"activity_id": id, "port": port, "value": value})
		return nil
	}
	for _, d := range trace.Deliveries {
		if len(d.Inputs) == 0 {
			if err := add(d.ActivityID, "input", d.ProducerID, d.Input); err != nil {
				return nil, "", err
			}
		} else {
			if len(d.Input) != 0 || d.ProducerID != "" {
				return nil, "", fmt.Errorf("ambiguous scalar and port inputs")
			}
			ports := map[string]bool{}
			for _, p := range d.Inputs {
				if p.Port == "" || ports[p.Port] {
					return nil, "", fmt.Errorf("missing or duplicate input port")
				}
				ports[p.Port] = true
				if err := add(d.ActivityID, p.Port, p.ProducerID, p.Value); err != nil {
					return nil, "", err
				}
			}
		}
	}
	if len(roots) == 0 {
		return nil, "", fmt.Errorf("trace has no observable caller inputs")
	}
	sort.Slice(roots, func(i, j int) bool {
		a, b := roots[i], roots[j]
		if a["activity_id"] != b["activity_id"] {
			return a["activity_id"].(string) < b["activity_id"].(string)
		}
		return a["port"].(string) < b["port"].(string)
	})
	key, _ := json.Marshal(roots)
	return roots, string(key), nil
}

func snapshot(index int, d bodyexecution.CompositionDelivery) (observation, error) {
	o := observation{state: "UNOBSERVED", record: map[string]any{"case_index": index, "expected_present": len(d.Expected) > 0}}
	if d.Fault != nil && len(d.BlockedBy) > 0 {
		return o, fmt.Errorf("conflicting fault and blocked outcome")
	}
	switch {
	case d.Fault != nil:
		o.state = "FAULT"
		o.record["fault"] = d.Fault
		b, _ := json.Marshal([]any{d.Fault.Kind, d.Fault.Site.ExpressionID, d.Fault.Site.Operator, d.Fault.Left, d.Fault.Right})
		o.outcome = string(b)
	case len(d.BlockedBy) > 0:
		o.state = "BLOCKED"
		blocked := slices.Clone(d.BlockedBy)
		slices.Sort(blocked)
		o.record["blocked_by"] = blocked
		b, _ := json.Marshal(blocked)
		o.outcome = string(b)
	case len(d.Actual) > 0:
		value, key, err := canonical(d.Actual)
		if err != nil {
			return o, err
		}
		o.state, o.outcome, o.record["actual"] = "VALUE", key, value
	}
	if (o.state == "FAULT" || o.state == "BLOCKED") && len(d.Actual) > 0 && !bytes.Equal(bytes.TrimSpace(d.Actual), []byte("null")) {
		return o, fmt.Errorf("fault or blocked outcome also claims a value")
	}
	if len(d.Expected) > 0 {
		value, key, err := canonical(d.Expected)
		if err != nil {
			return o, err
		}
		o.expected, o.record["expected"] = key, value
		if o.state != "UNOBSERVED" {
			pass := o.state == "VALUE" && o.outcome == key
			o.passed = &pass
		}
	}
	if (o.passed == nil) != (d.Passed == nil) || o.passed != nil && *o.passed != *d.Passed {
		return o, fmt.Errorf("reported pass differs from actual and expected values")
	}
	o.record["state"], o.record["passed"] = o.state, o.passed
	b, _ := json.Marshal([]any{o.state, o.outcome, o.expected})
	o.identity = string(b)
	return o, nil
}
