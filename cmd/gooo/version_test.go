package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestRunVersionIsDeterministic(t *testing.T) {
	var first, second, firstErr, secondErr bytes.Buffer
	if code := runVersion(nil, &first, &firstErr); code != exitOK {
		t.Fatalf("first version code = %d, stderr=%q", code, firstErr.String())
	}
	if code := runVersion(nil, &second, &secondErr); code != exitOK {
		t.Fatalf("second version code = %d, stderr=%q", code, secondErr.String())
	}
	if first.String() != "gooo 0.1.0-dev (development)\n" || first.String() != second.String() || firstErr.Len() != 0 || secondErr.Len() != 0 {
		t.Fatalf("version output = %q/%q, stderr=%q/%q", first.String(), second.String(), firstErr.String(), secondErr.String())
	}
}

func TestRunVersionJSONBindsVersionedContracts(t *testing.T) {
	var output, stderr bytes.Buffer
	if code := runVersion([]string{"--json"}, &output, &stderr); code != exitOK || stderr.Len() != 0 {
		t.Fatalf("version json code = %d, stdout=%q, stderr=%q", code, output.String(), stderr.String())
	}
	var info versionInfo
	if err := json.Unmarshal(output.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	want := versionInfo{
		SchemaVersion: versionSchema, Language: "gooo", Version: goooVersion,
		Status: versionStatus, SemanticIR: semantic.CurrentIRVersion,
		Graph: graphDumpSchemaVersion, FixPlan: fixPlanSchemaVersion,
	}
	if info != want {
		t.Fatalf("version info = %#v, want %#v", info, want)
	}
}

func TestRunVersionRejectsUnknownArguments(t *testing.T) {
	var stderr bytes.Buffer
	if code := runVersion([]string{"--verbose"}, &bytes.Buffer{}, &stderr); code != exitUsage || stderr.String() != versionUsage+"\n" {
		t.Fatalf("version usage = code %d, stderr=%q", code, stderr.String())
	}
}
