package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestBodyPlanHintViewFresh1883(t *testing.T) {
	source, err := os.ReadFile("../../internal/bodyexecution/testdata/wire-capacity.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"view", "broken_hint"} {
		t.Run(mode, func(t *testing.T) {
			var out, diagnostic bytes.Buffer
			args := []string{"--source", "wire.gooo.fixture", "--entry", "Second", "--json"}
			if code := runBodyPlan(args, mapSourceReader{"wire.gooo.fixture": source}, &out, &diagnostic); code != exitOK {
				t.Fatal(code, diagnostic.String())
			}
			var view map[string]json.RawMessage
			if err := json.Unmarshal(out.Bytes(), &view); err != nil {
				t.Fatal(err)
			}
			var capacity bodyexecution.CompositionCapacity
			if err := json.Unmarshal(view["transport_capacity"], &capacity); err != nil ||
				capacity.InputJSONBytes != 1048576 || capacity.NativeStdoutBytes != 1048576 {
				t.Fatal("CLI lost the bounded operational hint", err)
			}
			if mode == "broken_hint" {
				view["transport_capacity"] = json.RawMessage(`{"input_json_bytes":"unreadable"}`)
			}
			candidate, err := json.Marshal(view)
			if err != nil {
				t.Fatal(err)
			}
			var core bodyexecution.CompositionInspection
			if err := json.Unmarshal(candidate, &core); err != nil {
				t.Fatal("advisory content blocked core decoding", err)
			}
			expected, err := bodyexecution.InspectComposition(context.Background(), "wire.gooo.fixture", source, "Second")
			if err != nil || !reflect.DeepEqual(core, expected) {
				t.Fatal("advisory content changed the structural source plan", err)
			}
			encoded, err := json.Marshal(core)
			if err != nil || bytes.Contains(encoded, []byte("transport_capacity")) {
				t.Fatal("view metadata became structural identity", err)
			}
			if _, err := bodyexecution.CompositionInputTemplate(core); err != nil {
				t.Fatal("view metadata blocked typed input template", err)
			}
		})
	}
}
