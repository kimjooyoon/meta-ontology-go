package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestAuthoredLocalDivisionExample(t *testing.T) {
	const path = "../../examples/compound-updates/main.gooo"
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range [][3]string{{"137", "7", "19"}, {"-139", "7", "-19"}, {"149", "0", "149"}} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{
			"Divide.input0": json.RawMessage(row[0]), "Divide.input1": json.RawMessage(row[1])},
			Expected: map[string]json.RawMessage{"Divide": json.RawMessage(row[2])}})
	}
	observation := map[string]any{"source": string(source), "cases": suite}
	defer saveReadObservation(t, "public-local-division.json", observation)
	program, err := GenerateCompositionWithOptions(context.Background(), path, source, suite, CompositionOptions{EntryActivity: "Divide"})
	if err != nil {
		t.Fatal(err)
	}
	observation["composition"] = program
	r, err := ExecuteComposition(context.Background(), path, source, program, suite, "")
	observation["runtime"] = r
	if err != nil || r.FinitePassed != 2 || r.Outcomes == nil || r.Outcomes.Faulted != 1 || len(r.FaultSites) != 1 || r.ModelCalls != 0 {
		t.Fatal("public division results", err, r.Outcomes)
	}
	fault := r.Traces[2].Deliveries[0].Fault
	if fault == nil || fault.Left != 149 || fault.Right != 0 || fault.Site.Variable != "amount" || fault.Site.Assignment != "/=" {
		t.Fatal("public fault origin", fault)
	}
}
