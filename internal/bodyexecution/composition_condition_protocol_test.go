package bodyexecution

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNativeConditionProtocolBindsPresenceAndOrder991(t *testing.T) {
	graph := compositionGraph{count: 2}
	graph.nodes[0].ID, graph.nodes[1].ID = "events991://first", "events991://second"
	sites := []CompositionConditionSite{{RootActivityID: graph.nodes[0].ID}, {RootActivityID: graph.nodes[1].ID}}
	valid := `{"schema":"gooo/native-condition-observation/v1","result":[[11,22],[33,44]],"conditions":[{"case_index":0,"site_index":0,"value":false},{"case_index":0,"site_index":1,"value":true},{"case_index":1,"site_index":0,"value":true}]}`
	native, rows, err := graph.nativeConditionRows([]byte(valid), 2, sites)
	if err != nil || string(native) != "[[11,22],[33,44]]" || len(rows[0]) != 2 || rows[0][0].Value || len(rows[1]) != 1 {
		t.Fatal("exact false observation unavailable", err, rows)
	}
	for name, invalid := range map[string]string{
		"missing_truth":  strings.Replace(valid, `,"value":false`, "", 1),
		"null_truth":     strings.Replace(valid, `"value":false`, `"value":null`, 1),
		"missing_site":   strings.Replace(valid, `"site_index":0,`, "", 1),
		"case_alias":     strings.Replace(valid, `"case_index":0`, `"Case_Index":0`, 1),
		"site_outside":   strings.Replace(valid, `"site_index":1`, `"site_index":2`, 1),
		"case_outside":   strings.Replace(valid, `"case_index":1`, `"case_index":2`, 1),
		"case_backwards": strings.Replace(valid, `"case_index":0`, `"case_index":1`, 1),
		"root_backwards": strings.Replace(strings.Replace(valid, `"site_index":0,"value":false`, `"site_index":1,"value":false`, 1),
			`"site_index":1,"value":true`, `"site_index":0,"value":true`, 1),
		"wrong_truth":     strings.Replace(valid, `"value":false`, `"value":"false"`, 1),
		"duplicate_truth": strings.Replace(valid, `"value":false`, `"value":false,"value":true`, 1),
		"missing_events":  `{"schema":"gooo/native-condition-observation/v1","result":[[11,22]]}`,
		"null_result":     `{"schema":"gooo/native-condition-observation/v1","result":null,"conditions":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := graph.nativeConditionRows([]byte(invalid), 2, sites); err == nil {
				t.Fatal("ambiguous native condition event accepted", invalid)
			}
		})
	}
	traces := []CompositionTrace{{Deliveries: make([]CompositionDelivery, 2)}, {Deliveries: make([]CompositionDelivery, 2)}}
	traces[0].Deliveries[1].BlockedBy = []string{graph.nodes[0].ID}
	if graph.bindNativeConditions(traces, rows, sites) == nil {
		t.Fatal("blocked activity claimed a condition evaluation")
	}
	var event json.RawMessage = []byte(`{"case_index":0,"site_index":0,"value":true}`)
	overflow, _ := json.Marshal(map[string]any{"schema": conditionObservationSchema, "result": [][]int{{11, 22}},
		"conditions": repeatedConditionEvents(event, conditionObservationLimit+1)})
	if _, _, err := graph.nativeConditionRows(overflow, 1, sites); err == nil {
		t.Fatal("bounded condition protocol accepted excess events")
	}
}

func repeatedConditionEvents(event json.RawMessage, count int) []json.RawMessage {
	events := make([]json.RawMessage, count)
	for i := range events {
		events[i] = event
	}
	return events
}
