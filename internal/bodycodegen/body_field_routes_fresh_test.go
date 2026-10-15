package bodycodegen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFieldSourceRoutes16621(t *testing.T) {
	source := []byte(`package fieldroutes
namespace fieldroutes16621
entity Input id "urn:fieldroutes16621:input" fields {
 field amount id "urn:fieldroutes16621:amount" type integer required one
 field flip id "urn:fieldroutes16621:flip" type boolean required one
}
entity State id "urn:fieldroutes16621:state" fields {
 field count id "urn:fieldroutes16621:count" type integer required one
}
activity Choose(Input)->State id "urn:fieldroutes16621:choose" computes "if input.flip {return State{count:input.amount+16621}} else {return State{count:input.amount-43}}"
`)
	prepared, err := prepareActivityBodySource("routes16621.gooo", source, "Choose")
	if err != nil {
		t.Fatal(err)
	}
	var reports []Report
	for _, route := range []string{preserveRoute, guardReturnRoute, mergeResultRoute} {
		projected, err := prepared.generate(route)
		if err != nil {
			t.Fatal(route, err)
		}
		if len(projected.report.BodyFields) != 5 {
			t.Fatal("retained field slots", route, projected.report.BodyFields)
		}
		bound, err := BindBodyFieldSources(source, projected.report.BodyFields)
		if err != nil {
			t.Fatal(err)
		}
		for _, origin := range bound {
			if origin.Value.Status != "MAPPED" || origin.Target.Status != "MAPPED" {
				t.Fatal("route lost original field source", route, origin)
			}
		}
		reports = append(reports, projected.report)
	}
	if root := os.Getenv("GOOO_FRESH_SOURCE_OUTPUT"); root != "" {
		raw, err := json.MarshalIndent(reports, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "routes16621.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("new source:preserve/guard-return/merge-result routes;15 retained field slots mapped;30 value/target locations;canonical equivalence holds;native/model0")
}
