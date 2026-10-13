package bodyexecution

import (
	"encoding/json"
	"testing"
)

func TestReadProtocol9113(t *testing.T) {
	graph := compositionGraph{count: 1}
	graph.nodes[0] = CompositionActivity{ID: "urn:protocol9113:root", Name: "Read"}
	sites := []CompositionReadSite{{RootActivityID: "urn:protocol9113:root", ActivityID: "urn:protocol9113:root", ElementType: "int64"}}
	inputs := []CompositionReadInput{{ActivityID: "urn:protocol9113:root", InputKey: "Read", FieldName: "values", ElementType: "int64"}}
	suite := CompositionCases{Cases: []CompositionCase{{Inputs: map[string]json.RawMessage{"Read": json.RawMessage(`{"values":[227,227,229]}`)}}}}
	event := func() map[string]any {
		return map[string]any{"case_index": 0, "site_index": 0, "index": 1, "length": 2, "origin": map[string]any{"input_collection": 0, "element_index": 2}}
	}
	encode := func(events []map[string]any, dropped int) []byte {
		raw, _ := json.Marshal(map[string]any{"schema": "gooo/native-read-observation/v1", "reads": events, "dropped": []int{dropped}, "registration_dropped": []int{0}})
		return raw
	}
	rows, _, _, summary, err := graph.nativeReadFile(encode([]map[string]any{event()}, 0), suite, sites, inputs)
	if err != nil || summary.ResolvedReads != 1 || rows[0][0].Origin.Index != 2 {
		t.Fatal("explicit original index", err, summary)
	}
	for name, mutate := range map[string]func(map[string]any){
		"missing_index":        func(e map[string]any) { delete(e, "index") },
		"negative_index":       func(e map[string]any) { e["index"] = -1 },
		"length_boundary":      func(e map[string]any) { e["index"] = 2 },
		"unknown_site":         func(e map[string]any) { e["site_index"] = 1 },
		"unknown_case":         func(e map[string]any) { e["case_index"] = 1 },
		"missing_origin_input": func(e map[string]any) { delete(e["origin"].(map[string]any), "input_collection") },
		"origin_outside_input": func(e map[string]any) { e["origin"].(map[string]any)["element_index"] = 3 },
		"origin_with_reason":   func(e map[string]any) { e["reason"] = "NO_REGISTERED_INPUT_ELEMENT" },
		"unexplained_unknown":  func(e map[string]any) { delete(e, "origin") },
	} {
		t.Run(name, func(t *testing.T) {
			e := event()
			mutate(e)
			if _, _, _, _, err := graph.nativeReadFile(encode([]map[string]any{e}, 0), suite, sites, inputs); err == nil {
				t.Fatal("accepted malformed origin")
			}
		})
	}
	if _, _, _, _, err := graph.nativeReadFile(encode([]map[string]any{event()}, 1), suite, sites, inputs); err == nil {
		t.Fatal("drop before full event bound")
	}
	events := make([]map[string]any, readObservationLimit)
	for i := range events {
		events[i] = event()
	}
	_, drops, _, summary, err := graph.nativeReadFile(encode(events, 3), suite, sites, inputs)
	if err != nil || drops[0] != 3 || summary.Status != "PARTIAL" || summary.RecordedReads != 65536 || summary.DroppedReads != 3 {
		t.Fatal("bounded partial observations", err, summary)
	}
	read := compositionReadRun{rows: rows, dropped: []int{0}, registration: []int{0}}
	traces := []CompositionTrace{{Deliveries: []CompositionDelivery{{BlockedBy: []string{"earlier-fault"}}}}}
	if err := read.bind(graph, traces, &CompositionRuntime{ReadSites: sites}); err == nil {
		t.Fatal("blocked activity claimed a read")
	}
	t.Log("protocol9113: required zero-valued coordinates, bounds, explicit unknowns, blocked roots and65536-event drops checked")
}
