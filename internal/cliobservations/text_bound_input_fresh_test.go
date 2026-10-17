package cliobservations

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestBoundedRecordIngressAndRenameNativeFresh1905(t *testing.T) {
	original, err := os.ReadFile("../../examples/bounded-text/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	original = append(original, []byte("\nactivity Identity(Expanded) -> Expanded id \"bound1905://identity\" computes `return input`\n")...)
	for _, name := range []string{"value", "canonical"} {
		t.Run(name, func(t *testing.T) {
			source := strings.ReplaceAll(string(original), "field value id", "field "+name+" id")
			source = strings.ReplaceAll(source, "Expanded{value:", "Expanded{"+name+":")
			object, _ := json.Marshal(map[string]string{name: strings.Repeat("x", 1030)})
			suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{{
				Inputs:   map[string]json.RawMessage{"app/bound1905:Identity": object},
				Expected: map[string]json.RawMessage{"app/bound1905:Identity": object},
			}}}
			receipt, code := executeTextBound1905(t, "record-input-"+name, []byte(source), "Identity", suite, false)
			assertTextBoundNative1905(t, receipt, code, 1)
			fields := receipt.Result.Runtime.Traces[0].Deliveries[0].ActualFields
			if len(fields) != 1 || fields[0].ID != "gooo://examples/bounded-text/expanded/value" || fields[0].Name != name {
				t.Fatal("source rename lost the bound field identity", fields)
			}
		})
	}
}
