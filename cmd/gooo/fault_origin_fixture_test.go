package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestFaultOriginPublicFixtureRetainsArchivedRuntime(t *testing.T) {
	raw, err := os.ReadFile("../../examples/fault-origins/recorded-runtime.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := decodePackageOriginRuntime(decoded)
	if err != nil || !reflect.DeepEqual(runtime, savedFaultOriginRuntime(t, "repair-outputs/local-division-faults.json")) {
		t.Fatal("public fixture changed producer history", err)
	}
	origin, err := os.ReadFile("../../examples/fault-origins/recorded-origin.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(origin, &manifest); err != nil || manifest["fixture_sha256"] != namedInputObservationSHA(raw) || manifest["selected_source_sha256"] != runtime.SelectedSourceSHA256 {
		t.Fatal("fixture source identity", err)
	}
}
