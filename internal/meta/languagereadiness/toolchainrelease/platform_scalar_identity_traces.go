package toolchainrelease

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

var scalarInputIDs = [3]string{"urn:gooo:type:integer", "urn:gooo:type:boolean", "urn:gooo:type:string"}

func validateScalarIdentityTraces(traces []bodyexecution.CompositionTrace) error {
	rows := [4][3]string{
		{"9007199254740993", "true", `"해보자"`}, {"-9007199254740993", "true", `"negative"`},
		{"4", "false", `"🙂"`}, {"12", "true", `"context"`},
	}
	text := [4]string{`"해보자!"`, `"negative"`, `"🙂"`, `"context!"`}
	active := [4]string{"true", "false", "false", "true"}
	if len(traces) != 4 {
		return fmt.Errorf("scalar identity native trace count differs")
	}
	for i, trace := range traces {
		if trace.CaseIndex != i || len(trace.Deliveries) != 1 {
			return fmt.Errorf("scalar identity native trace order differs")
		}
		d := trace.Deliveries[0]
		if d.ActivityID != scalarDescribeID || len(d.Inputs) != 3 || d.Fault != nil || d.Passed == nil || !*d.Passed {
			return fmt.Errorf("scalar identity native delivery differs")
		}
		for j, input := range d.Inputs {
			if input.Port != fmt.Sprintf("input%d", j) || input.EntityID != scalarInputIDs[j] || string(input.Value) != rows[i][j] {
				return fmt.Errorf("scalar identity exact typed input differs")
			}
		}
		want := [3]string{text[i], active[i], rows[i][0]}
		if err := validateScalarIdentityFields(d, want); err != nil {
			return err
		}
	}
	return nil
}

func validateScalarIdentityFields(d bodyexecution.CompositionDelivery, want [3]string) error {
	names := [3]string{"text", "active", "count"}
	if len(d.ActualFields) != 3 {
		return fmt.Errorf("scalar identity native field count differs")
	}
	for _, raw := range []json.RawMessage{d.Actual, d.Expected} {
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil || len(fields) != 3 {
			return fmt.Errorf("scalar identity native record differs")
		}
		for i, name := range names {
			field := d.ActualFields[i]
			if string(fields[name]) != want[i] || field.Name != name || field.ID != "urn:gooo:scalar-identity:label:"+name ||
				string(field.Value) != want[i] {
				return fmt.Errorf("scalar identity exact field value or identity differs")
			}
		}
	}
	return nil
}
