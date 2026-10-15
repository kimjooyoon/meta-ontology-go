package bodycodegen

import (
	"encoding/json"
	"strings"
	"testing"
)

func declarationTransportReport(t *testing.T, activity string) TypedValueOrigins {
	t.Helper()
	r, err := InspectTypedValueOrigins("transport.gooo", []byte(inputDeclarationSourceFixture+"# declaration transport\n"), activity)
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	return r
}

func TestDeclarationTransportPreservesPositions(t *testing.T) {
	for _, activity := range []string{"Difference", "Read"} {
		r := declarationTransportReport(t, activity)
		raw, err := TypedInputDeclarationInputs(r, "source/control:InspectInputs")
		if err != nil {
			t.Fatal(err)
		}
		var envelope struct {
			Schema string                               `json:"schema"`
			Inputs []map[string]InputDeclarationColumns `json:"inputs"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil || len(envelope.Inputs) != 1 {
			t.Fatal(err)
		}
		c := envelope.Inputs[0]["source/control:InspectInputs"]
		want := 2
		if activity == "Read" {
			want = 1
		}
		if envelope.Schema != "gooo/body-composition-inputs/v1" || c.Count != want ||
			c.Schema != "gooo/input-declaration-columns/v1" || c.Scope != "activity_signature" ||
			c.SourceSHA256 != r.SourceSHA256 || c.ActivityID != r.ActivityID {
			t.Fatal("declaration projection", c)
		}
		for i := range want {
			d := r.Inputs[i].Declaration
			if c.Indices[i] != i || c.Ports[i] != d.Port || c.TypeNames[i] != d.TypeName ||
				c.TypeIDs[i] != d.TypeID || c.Starts[i] != d.Start || c.Ends[i] != d.End {
				t.Fatal("declaration changed during transport", i)
			}
		}
	}
}

func TestDeclarationTransportRejectsBrokenBindings(t *testing.T) {
	mutations := map[string]func(*TypedValueOrigins){
		"legacy omission": func(r *TypedValueOrigins) { r.Inputs[0].Declaration = nil },
		"source":          func(r *TypedValueOrigins) { r.Inputs[0].Declaration.SourceSHA256 = "other" },
		"activity":        func(r *TypedValueOrigins) { r.Flow.ActivityID = "other" },
		"port":            func(r *TypedValueOrigins) { r.Inputs[0].Declaration.Port = "other" },
		"position":        func(r *TypedValueOrigins) { r.Inputs[0].Declaration.InputIndex = 1 },
		"node id":         func(r *TypedValueOrigins) { r.Flow.Nodes[0].ID = 2 },
		"field":           func(r *TypedValueOrigins) { r.Inputs[0].FieldID = "other" },
		"node activity":   func(r *TypedValueOrigins) { r.Flow.Nodes[0].Span.ActivityID = "other" },
		"missing first":   func(r *TypedValueOrigins) { r.Inputs = r.Inputs[1:] },
		"bound":           func(r *TypedValueOrigins) { r.Inputs[0].Declaration.InputIndex = 16 },
		"bytes": func(r *TypedValueOrigins) {
			r.Inputs[0].Declaration.TypeName = strings.Repeat("x", 33<<10)
			r.Inputs[0].Declaration.Expression = r.Inputs[0].Declaration.TypeName
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			r := declarationTransportReport(t, "Difference")
			mutate(&r)
			if _, err := TypedInputDeclarationInputs(r, "source/control:InspectInputs"); err == nil {
				t.Fatal("inconsistent declaration accepted")
			}
		})
	}
	r := declarationTransportReport(t, "Read")
	r.Inputs[1].Declaration.TypeID = "urn:other:type"
	if _, err := TypedInputDeclarationInputs(r, "source/control:InspectInputs"); err == nil {
		t.Fatal("conflicting record positions accepted")
	}
	for _, target := range []string{"", "reader", ":Inspect", "reader:", "reader:Inspect:Other", "reader:Inspect Other"} {
		if _, err := TypedInputDeclarationInputs(declarationTransportReport(t, "Difference"), target); err == nil {
			t.Fatal("invalid consumer target", target)
		}
	}
}
