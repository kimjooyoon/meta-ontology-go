package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

const relayGraphFixture = "../../examples/relay-game-contract/main.gooo"

func TestPublicGraphPreservesRelayFields(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"check", "--json", relayGraphFixture}, &stdout, &stderr); code != exitOK {
		t.Fatalf("check=%d stderr=%s stdout=%s", code, &stderr, &stdout)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"graph", "dump", relayGraphFixture}, &stdout, &stderr); code != exitOK {
		t.Fatalf("graph=%d stderr=%s", code, &stderr)
	}
	var dump graphDump
	if err := json.Unmarshal(stdout.Bytes(), &dump); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"RunID": "run-id", "Turn": "turn", "Action": "action", "Direction": "direction"}
	fields := 0
	for _, node := range dump.Nodes {
		for _, field := range node.Fields {
			id, ok := want[field.Name]
			if !ok || node.ID != "gooo://relay-game/request" || field.Parent != node.ID ||
				field.ID != node.ID+"/"+id || field.TypeRefID == "" || field.Presence != "required" || field.Cardinality != "one" {
				t.Fatalf("field structure changed: %+v", field)
			}
			delete(want, field.Name)
			fields++
		}
	}
	if fields != 4 || len(want) != 0 || dump.Authorities.Graph != "derived" || dump.Projection.Status != "deferred" {
		t.Fatalf("field count or authority changed: %+v", dump)
	}
}

func TestGenericGraphStillRejectsDeferredFields(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runGraph([]string{"dump", relayGraphFixture}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr); code != exitFailure || stdout.Len() != 0 {
		t.Fatalf("generic seam unexpectedly activated: exit=%d stdout=%s", code, &stdout)
	}
}
