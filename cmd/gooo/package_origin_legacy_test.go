package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestRecordedLegacyOrigins9767(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "observations", "native-read-origins", "native9059.json"))
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	runtime, err := decodePackageOriginRuntime(envelope["runtime"])
	if err != nil {
		t.Fatal(err)
	}
	rows, origin, err := bodyexecution.SelectRecordedNativeOrigins(runtime, bodyexecution.NativeOriginQuery{Kind: "reads", Offset: 18, Limit: 2})
	if err != nil || len(rows) != 2 || origin.TotalRows != 20 || origin.Observation.Status != "PARTIAL" {
		t.Fatal(origin, err)
	}
	if _, _, err := bodyexecution.SelectRecordedNativeOrigins(runtime, bodyexecution.NativeOriginQuery{Kind: "elements", Limit: 1}); err == nil {
		t.Fatal("legacy summary invented constructions")
	}
	origin.Observation.Status = "CHANGED_COPY"
	if runtime.ReadObservation.Status != "PARTIAL" {
		t.Fatal("new input metadata rewrote its parent record")
	}
	t.Log("saved v1: 2/20 read rows selected, PARTIAL retained, construction query rejected, detached summary; producer execution 0")
}
