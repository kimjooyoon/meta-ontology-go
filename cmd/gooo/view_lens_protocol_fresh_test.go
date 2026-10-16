package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/viewlens"
)

func TestDetachedViewCLIProtocolFresh1811(t *testing.T) {
	directory := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := runViewLens([]string{"load", "--state-dir", directory, "--actor", "cli1811"}, &stdout, &stderr)
	var loaded viewlens.StoreResult
	if err := json.Unmarshal(stdout.Bytes(), &loaded); err != nil || code != exitOK || loaded.State != "ABSENT" {
		t.Fatalf("load: %v %d %s %s", err, code, &stdout, &stderr)
	}
	hint, _ := json.Marshal(loaded.Snapshot)
	stdout.Reset()
	code = runBodyView([]string{"--activity", "Balance", "--actor", "cli1811", "--prefer", "inputs",
		"--hint", string(hint), "cli1811.gooo"}, fixtureReader{source: advisorySource1807}, &stdout, &stderr)
	var rendered bodyViewOutput
	if err := json.Unmarshal(stdout.Bytes(), &rendered); err != nil || code != exitOK || rendered.Selection.Record == nil {
		t.Fatalf("render: %v %d %s %s", err, code, &stdout, &stderr)
	}
	if got := viewlens.Load(directory, "cli1811", "agent"); got.Snapshot.Revision != 0 {
		t.Fatal("render performed implicit persistence")
	}
	event, _ := json.Marshal(rendered.Selection.Record)
	stdout.Reset()
	code = runViewLens([]string{"record", "--state-dir", directory, "--event", string(event)}, &stdout, &stderr)
	var stored viewlens.StoreResult
	if err := json.Unmarshal(stdout.Bytes(), &stored); err != nil || code != exitOK || !stored.Persisted ||
		stored.Snapshot.Preferred != viewlens.NamedInputs || stored.Snapshot.Revision != 1 {
		t.Fatalf("detached record: %v %d %s %s", err, code, &stdout, &stderr)
	}
	stdout.Reset()
	code = runViewLens([]string{"record", "--state-dir", directory, "--event", "{"}, &stdout, &stderr)
	if err := json.Unmarshal(stdout.Bytes(), &stored); err != nil || code != exitOK || stored.State != "UNAVAILABLE" ||
		stored.Persisted || stored.Issue != "HINT_SYNTAX_INVALID" {
		t.Fatalf("advisory storage error: %v %d %s", err, code, &stdout)
	}
}
