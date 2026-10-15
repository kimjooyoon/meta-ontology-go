package bodycodegen

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestOriginTextConditionFacts(t *testing.T) {
	source := []byte("package conditionfacts\nnamespace conditionfacts\n" +
		"entity Integer id \"urn:gooo:type:integer\"\nentity Boolean id \"urn:gooo:type:boolean\"\n" +
		"activity Adjust(Integer, Boolean) -> Integer id \"urn:condition-facts:adjust\" computes `let amount=input0+4201; if input1 {amount*=3} else {amount-=7}; return amount`\n" +
		"activity Constant(Integer) -> Integer id \"urn:condition-facts:constant\" computes `if false {return input+4207};return input-4211`\n")
	for _, item := range []struct {
		activity string
		facts    []string
	}{
		{"Adjust", []string{`when="true"`, `when="false"`, `operator="+"`, `operator="*"`, `operator="-"`}},
		{"Constant", []string{`constant-condition="false"`, `when="true"`, `when="false"`}},
	} {
		t.Run(item.activity, func(t *testing.T) {
			report, err := InspectTypedValueOrigins("condition-facts.gooo", source, item.activity)
			if err != nil || report.Status != "RESOLVED" {
				t.Fatal(report.Status, report.Reason, err)
			}
			before, _ := json.Marshal(report)
			var output bytes.Buffer
			if err := WriteTypedOriginText(&output, report, source); err != nil {
				t.Fatal(err)
			}
			for _, fact := range append(item.facts, "Scope: possible source dependencies") {
				if !strings.Contains(output.String(), fact) {
					t.Fatalf("missing source fact %q", fact)
				}
			}
			after, _ := json.Marshal(report)
			if !bytes.Equal(before, after) {
				t.Fatal("text display changed the source graph")
			}
			if err := WriteTypedOriginText(rejectOriginOperators{io.Discard}, report, source); !errors.Is(err, io.ErrClosedPipe) {
				t.Fatal("operator output failure lost", err)
			}
		})
	}
}

type rejectOriginOperators struct{ io.Writer }

func (w rejectOriginOperators) Write(p []byte) (int, error) {
	if bytes.Contains(p, []byte("operator=")) {
		return 0, io.ErrClosedPipe
	}
	return w.Writer.Write(p)
}
