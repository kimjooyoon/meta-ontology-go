package bodyexecution

import (
	"encoding/json"
	"testing"
)

func TestConstructionProtocol9613(t *testing.T) {
	graph := compositionGraph{count: 1}
	graph.nodes[0] = CompositionActivity{ID: "urn:protocol9613:root", Name: "Build"}
	sites := []CompositionReadSite{{RootActivityID: "urn:protocol9613:root", ActivityID: "urn:protocol9613:root", ElementType: "int64"}}
	inputs := []CompositionReadInput{{ActivityID: "urn:protocol9613:root", InputKey: "Build", FieldName: "values", ElementType: "int64"}}
	suite := CompositionCases{Cases: []CompositionCase{{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":[631,641]}`)}}}}
	constructors := []CompositionConstructionSite{{RootActivityID: "urn:protocol9613:root", ActivityID: "urn:protocol9613:root", ElementType: "int64", Elements: make([]CompositionConstructionElementSite, 2)}}
	const envelope = `{"schema":"gooo/native-read-observation/v2","reads":[{"case_index":0,"site_index":0,"index":1,"length":2,"origin":{"input_collection":0,"element_index":1}},{"case_index":0,"site_index":0,"index":0,"length":2,"construction":{"instance_index":0,"element_index":0}}],"dropped":[0],"registration_dropped":[0],"constructions":[{"case_index":0,"site_index":0,"read_start":0,"read_end":1,"completed":true,"elements":[{"read_start":0,"read_end":1,"reads_dropped":0,"completed":true},{"read_start":1,"read_end":1,"reads_dropped":0,"completed":true}]}],"construction_dropped":[0]}`
	decodeEnvelope := func() map[string]any {
		var m map[string]any
		if err := json.Unmarshal([]byte(envelope), &m); err != nil {
			t.Fatal(err)
		}
		return m
	}
	parse := func(m map[string]any) ([][]CompositionReadObservation, *compositionConstructionContext, error) {
		raw, _ := json.Marshal(m)
		ctx := &compositionConstructionContext{sites: constructors}
		rows, _, _, _, err := graph.nativeReadFile(raw, suite, sites, inputs, ctx)
		return rows, ctx, err
	}
	if rows, _, err := parse(decodeEnvelope()); err != nil || rows[0][1].Construction == nil {
		t.Fatal("construction reference", err)
	}
	for _, name := range []string{"missing_case", "missing_end", "missing_completion", "missing_element_start", "future_origin", "ambiguous_origin", "unknown_instance", "unknown_element", "false_complete_count", "unfinished_reference", "unexpected_drop", "wrong_version"} {
		t.Run(name, func(t *testing.T) {
			m := decodeEnvelope()
			created := m["constructions"].([]any)[0].(map[string]any)
			read := m["reads"].([]any)[1].(map[string]any)
			elements := created["elements"].([]any)
			switch name {
			case "missing_case":
				delete(created, "case_index")
			case "missing_end":
				delete(created, "read_end")
			case "missing_completion":
				delete(created, "completed")
			case "missing_element_start":
				delete(elements[0].(map[string]any), "read_start")
			case "future_origin":
				m["reads"].([]any)[0] = read
			case "ambiguous_origin":
				read["origin"] = map[string]any{"input_collection": 0, "element_index": 1}
			case "unknown_instance":
				read["construction"].(map[string]any)["instance_index"] = 1
			case "unknown_element":
				read["construction"].(map[string]any)["element_index"] = 2
			case "false_complete_count":
				created["elements"] = elements[:1]
			case "unfinished_reference":
				created["completed"] = false
				elements[1].(map[string]any)["completed"] = false
			case "unexpected_drop":
				m["construction_dropped"] = []int{1}
			case "wrong_version":
				m["schema"] = "gooo/native-read-observation/v1"
			}
			if _, _, err := parse(m); err == nil {
				t.Fatal("accepted inconsistent construction")
			}
		})
	}
	rows, ctx, err := parse(decodeEnvelope())
	if err != nil {
		t.Fatal(err)
	}
	traces := []CompositionTrace{{Reads: rows[0], Deliveries: []CompositionDelivery{{BlockedBy: []string{"prior-fault"}}}}}
	if graph.bindConstructions(ctx, traces, sites) == nil {
		t.Fatal("blocked root claimed creation")
	}
}

func TestConstructionBounds9629(t *testing.T) {
	graph := compositionGraph{count: 1}
	graph.nodes[0] = CompositionActivity{ID: "urn:bound9629:root"}
	suite := CompositionCases{Cases: []CompositionCase{{}}}
	for _, probe := range []struct {
		name                string
		instances, elements int
	}{{"instances", 4096, 0}, {"elements", 256, 256}} {
		t.Run(probe.name, func(t *testing.T) {
			site := CompositionConstructionSite{RootActivityID: "urn:bound9629:root", ActivityID: "urn:bound9629:root", ElementType: "int64", Elements: make([]CompositionConstructionElementSite, probe.elements)}
			events := []map[string]any{}
			for range probe.instances {
				elements := []map[string]any{}
				for range probe.elements {
					elements = append(elements, map[string]any{"read_start": 0, "read_end": 0, "reads_dropped": 0, "completed": true})
				}
				events = append(events, map[string]any{"case_index": 0, "site_index": 0, "read_start": 0, "read_end": 0, "completed": true, "elements": elements})
			}
			raw, _ := json.Marshal(map[string]any{"schema": "gooo/native-read-observation/v2", "reads": []any{}, "dropped": []int{0}, "registration_dropped": []int{0}, "constructions": events, "construction_dropped": []int{2}})
			ctx := &compositionConstructionContext{sites: []CompositionConstructionSite{site}}
			_, _, _, summary, err := graph.nativeReadFile(raw, suite, nil, nil, ctx)
			if err != nil || summary.Status != "PARTIAL" || summary.RecordedConstructions != probe.instances || summary.ConstructionDropped != 2 {
				t.Fatal("bounded construction observations", err, summary)
			}
		})
	}
}
