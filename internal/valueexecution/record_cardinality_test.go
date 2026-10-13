package valueexecution

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSourceDeclaredCardinalityMatrix8017(t *testing.T) {
	types := []struct{ name, scalar, list, wrong string }{
		{"integer", "9007199254741107", "[9007199254741107,-43]", "true"},
		{"string", `"기원8017"`, `["기원8017","tail"]`, "81"},
		{"boolean", "false", "[false,true]", `"false"`},
	}
	var observations []map[string]any
	accepted, decoded := 0, 0
	for _, kind := range types {
		for _, presence := range []string{"required", "optional"} {
			for _, cardinality := range []string{"one", "many"} {
				identity := kind.name + "/" + presence + "/" + cardinality
				source := []byte(fmt.Sprintf(`package cardinality8017
namespace cardinality8017
entity Packet id "urn:cardinality8017:%s" fields {
 field value id "urn:cardinality8017:%s:value" type %s %s %s
}
activity Transfer(Packet) -> Packet id "urn:cardinality8017:%s:transfer" computes "record.forward:v1"
`, identity, identity, kind.name, presence, cardinality, identity))
				plan, err := CompileRecordPlan(identity+".gooo", source)
				if err != nil {
					t.Fatalf("%s source: %v", identity, err)
				}
				cases := []struct {
					name, raw       string
					decode, execute bool
				}{
					{"absent", "{}", true, presence == "optional"},
					{"empty", `{"value":[]}`, true, cardinality == "many"},
					{"scalar", `{"value":` + kind.scalar + `}`, true, cardinality == "one"},
					{"single", `{"value":[` + kind.scalar + `]}`, true, cardinality == "many"},
					{"ordered", `{"value":` + kind.list + `}`, true, cardinality == "many"},
					{"wrong-scalar", `{"value":` + kind.wrong + `}`, true, false},
					{"wrong-list", `{"value":[` + kind.wrong + `]}`, true, false},
					{"null", `{"value":null}`, false, false},
					{"nested", `{"value":[[]]}`, false, false},
					{"object", `{"value":{}}`, false, false},
				}
				for _, c := range cases {
					fields, err := DecodeRecordInput([]byte(c.raw))
					if (err == nil) != c.decode {
						t.Fatalf("%s/%s decode=%v", identity, c.name, err)
					}
					row := map[string]any{"source_sha256": fmt.Sprintf("%x", sha256.Sum256(source)),
						"source": string(source), "profile": identity, "case": c.name, "input": c.raw,
						"decoded": err == nil, "expected_execution": c.execute, "executed": false}
					if err == nil {
						decoded++
						result, executeErr := plan.Execute(map[string]RecordFields{"Transfer": fields})
						if (executeErr == nil) != c.execute {
							t.Fatalf("%s/%s execute=%v", identity, c.name, executeErr)
						}
						if c.execute {
							accepted++
							row["executed"] = true
							out := result.Results["Transfer"]
							if result.ApplyCalls != 1 || !reflect.DeepEqual(out.Fields, fields) ||
								out.SourceDigest != plan.SourceDigest || out.RootInputDigest != digestValue(fields) {
								t.Fatalf("%s/%s changed value, presence or origin: %+v", identity, c.name, out)
							}
							row["output"] = out.Fields
						} else if result.ApplyCalls != 0 || result.Deliveries != 0 || len(result.Results) != 0 {
							t.Fatalf("%s/%s invalid input partially executed", identity, c.name)
						}
					}
					observations = append(observations, row)
				}
			}
		}
	}
	if len(observations) != 120 || decoded != 84 || accepted != 30 {
		t.Fatalf("cases=%d decoded=%d accepted=%d", len(observations), decoded, accepted)
	}
	if directory := os.Getenv("GOOO_CARDINALITY_OBSERVATION"); directory != "" {
		data, err := json.MarshalIndent(map[string]any{"cases": observations, "expected_decisions_matched": 120,
			"decoded": decoded, "executed": accepted, "model_calls": 0}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "cardinality8017.json"), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("8017: 12 source declarations x 10 input forms; 120 expected decisions, 84 decoded, 30 executed; type, presence, cardinality and exact values preserved; model0")
}
