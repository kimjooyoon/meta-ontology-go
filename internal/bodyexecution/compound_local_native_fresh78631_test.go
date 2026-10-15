package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

const compoundSource78631 = `package compound78631
namespace compound78631
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
activity Adjust(Integer,Integer,Boolean)->Integer id "urn:compound78631:adjust" computes ` + "`" + `
let amount=input0
amount+=input1
if input2 {amount*=2}else{amount-=3}
amount/=2
return amount
` + "`\n"

func TestCompoundLocalNative78631(t *testing.T) {
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range [][4]string{
		{"5", "7", "true", "12"}, {"11", "-4", "false", "2"},
		{"-9", "2", "true", "-7"}, {"0", "3", "false", "0"},
	} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{
			"Adjust.input0": json.RawMessage(row[0]), "Adjust.input1": json.RawMessage(row[1]), "Adjust.input2": json.RawMessage(row[2])},
			Expected: map[string]json.RawMessage{"Adjust": json.RawMessage(row[3])}})
	}
	observation := map[string]any{"source": compoundSource78631, "cases": suite}
	defer saveReadObservation(t, "compound-native78631.json", observation)
	ctx := context.Background()
	program, err := GenerateCompositionWithOptions(ctx, "갱신78631.gooo", []byte(compoundSource78631), suite, CompositionOptions{EntryActivity: "Adjust"})
	observation["composition"] = program
	if err != nil {
		t.Fatal("source compound generation", err)
	}
	runtime, err := ExecuteComposition(ctx, "갱신78631.gooo", []byte(compoundSource78631), program, suite, "")
	observation["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 4 || runtime.ModelCalls != 0 || !runtime.RuntimeReplayed {
		t.Fatal("finite native update outcomes", err, runtime.FinitePassed)
	}
	if runtime.LocalObservation == nil || runtime.LocalObservation.Status != "COMPLETE" ||
		runtime.LocalObservation.RecordedWrites != 16 || runtime.LocalObservation.RecordedUses != 16 {
		t.Fatal("complete update populations", runtime.LocalObservation)
	}
	for _, trace := range runtime.Traces {
		assertCompoundTrace78631(t, trace, runtime.LocalSites, runtime.LocalUseSites)
	}
	if err := validateRecordedLocalOrigins(runtime); err != nil {
		t.Fatal("retained native origins", err)
	}
	t.Log("native78631:new four cases pass;16 writes/16 reads;each compound read resolves preceding write;model0")
}

func assertCompoundTrace78631(t *testing.T, trace CompositionTrace, sites []CompositionLocalSite, uses []CompositionLocalUseSite) {
	t.Helper()
	if len(trace.LocalWrites) != 4 || len(trace.LocalUses) != 4 {
		t.Fatal("case-relative local population", len(trace.LocalWrites), len(trace.LocalUses))
	}
	for index, write := range trace.LocalWrites {
		if index == 0 {
			continue
		}
		if write.UseEnd-write.UseStart != 1 {
			t.Fatal("compound update lost its implicit old-value read", write)
		}
		use := trace.LocalUses[write.UseStart]
		if use.Assignment != index-1 || sites[write.SiteIndex].VariableOrdinal != sites[trace.LocalWrites[use.Assignment].SiteIndex].VariableOrdinal {
			t.Fatal("prior lexical assignment differs", write, use)
		}
		origin := uses[use.SiteIndex].Source
		if origin == nil || origin.Status != "MAPPED" || origin.Expression != "amount" {
			t.Fatal("authored implicit-read source missing", origin)
		}
	}
	if trace.LocalUses[3].Assignment != 3 {
		t.Fatal("return did not use final update", trace.LocalUses[3])
	}
}
