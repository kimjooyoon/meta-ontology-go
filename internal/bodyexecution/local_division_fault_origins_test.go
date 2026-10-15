package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

const localDivisionSource = `package localdivision
namespace localdivision
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
activity Divide(원금 Integer, 분모 Integer, 보정 Integer) -> Integer id "urn:localdivision:divide" computes ` + "`" + `
let amount = 원금
let divisor = 분모
amount /= divisor / 보정
amount /= 2
return amount
` + "`" + `
activity Main(원금 Integer, 분모 Integer, 보정 Integer, 건너뛰기 Boolean) -> Integer id "urn:localdivision:main" computes ` + "`" + `
if 건너뛰기 { return 원금 }
return Divide(원금, 분모, 보정)
` + "`\n"

func TestLocalDivisionPreservesFaultOrigins(t *testing.T) {
	suite := localDivisionFaultCases()
	observation := map[string]any{"source": localDivisionSource, "cases": suite}
	defer saveReadObservation(t, "local-division-faults.json", observation)
	program, err := GenerateCompositionWithOptions(context.Background(), "local-division.gooo", []byte(localDivisionSource), suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	observation["composition"] = program
	r, err := ExecuteComposition(context.Background(), "local-division.gooo", []byte(localDivisionSource), program, suite, "")
	observation["runtime"] = r
	if err != nil {
		observation["error"] = err.Error()
		t.Fatal(err)
	}
	if r.FinitePassed != 3 || r.Outcomes == nil || r.Outcomes.Faulted != 2 || r.ModelCalls != 0 || len(r.FaultSites) != 3 {
		t.Fatal("local division outcomes", r.Outcomes, len(r.FaultSites))
	}
	for i, left := range []int64{83, 5} {
		fault := r.Traces[i+2].Deliveries[0].Fault
		if fault == nil || fault.Left != left || fault.Right != 0 || fault.Site.RootActivityID != "urn:localdivision:main" || fault.Site.ActivityID != "urn:localdivision:divide" || (fault.Site.Assignment == "/=") != (i == 0) {
			t.Fatal("source fault identity", i, fault)
		}
		assertLocalDivisionInterrupted(t, r, i+2)
	}
	if len(r.Traces[4].LocalWrites) != 0 {
		t.Fatal("short circuit invoked the helper")
	}
}

func localDivisionFaultCases() CompositionCases {
	r := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range [][5]string{
		{"91", "12", "3", "false", "11"}, {"-97", "12", "3", "false", "-12"},
		{"83", "1", "3", "false", "83"}, {"89", "5", "0", "false", "89"}, {"101", "0", "0", "true", "101"},
	} {
		r.Cases = append(r.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Main.input0": json.RawMessage(row[0]),
			"Main.input1": json.RawMessage(row[1]), "Main.input2": json.RawMessage(row[2]), "Main.input3": json.RawMessage(row[3])},
			Expected: map[string]json.RawMessage{"Main": json.RawMessage(row[4])}})
	}
	return r
}

func assertLocalDivisionInterrupted(t *testing.T, r CompositionRuntime, c int) {
	t.Helper()
	count := 0
	for _, write := range r.Traces[c].LocalWrites {
		site := r.LocalSites[write.SiteIndex]
		if site.Operation != "/=" {
			continue
		}
		count++
		if write.Completed || write.UseEnd-write.UseStart != 2 {
			t.Fatal("division should retain old amount and divisor reads without completing", write)
		}
	}
	if count != 1 {
		t.Fatal("interrupted local division population", count)
	}
}
