package bodyexecution

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRecordedFieldSources16517(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = "../.."
	}
	f, err := os.Open(filepath.Join(root, "internal/bodyexecution/testdata/field-files16301-runtime.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	raw, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}
	read := func() CompositionRuntime {
		var r CompositionRuntime
		if err := json.Unmarshal(raw, &r); err != nil {
			t.Fatal(err)
		}
		return r
	}
	query := NativeOriginQuery{Kind: "field-input-locations", Limit: 7}
	r := read()
	if rows, _, err := SelectRecordedNativeOrigins(r, query); err != nil || len(rows) != 7 {
		t.Fatal("saved field source query", err)
	}
	index := 0
	for r.FieldSources.Bindings[index].Origin == nil {
		index++
	}
	changes := []func(*CompositionRuntime){
		func(r *CompositionRuntime) { r.FieldSources.SelectedSourceSHA256 = "different-source" },
		func(r *CompositionRuntime) { r.FieldSources.Bindings[index].Origin.FieldID = "different-field" },
		func(r *CompositionRuntime) {
			r.FieldSources.Bindings[index].Origin.NativeOrdinal = r.FieldSources.Bindings[index].Origin.NativePopulation
		},
		func(r *CompositionRuntime) { r.FieldSources.Bindings[index].Source.Value.BodySHA256 = "different-body" },
		func(r *CompositionRuntime) { r.FieldSources.Bindings[index].Origin = nil },
		func(r *CompositionRuntime) { r.FieldSources.Bindings[1] = r.FieldSources.Bindings[0] },
		func(r *CompositionRuntime) { r.FieldFileSources.Bindings[index].Value.Filename = "" },
		func(r *CompositionRuntime) {
			r.FieldFileSources.Bindings[index].Value.End = r.FieldFileSources.Bindings[index].Value.Start
		},
		func(r *CompositionRuntime) { r.FieldFileSources.Bindings[1] = r.FieldFileSources.Bindings[0] },
		func(r *CompositionRuntime) { r.FieldSources = nil },
	}
	for i, change := range changes {
		current := read()
		change(&current)
		if current.FieldSources != nil {
			current.FieldSources.SHA256 = fieldSourcesDigest(*current.FieldSources)
			current.FieldFileSources.FieldSourcesSHA256 = current.FieldSources.SHA256
		}
		current.FieldFileSources.SHA256 = fieldFileSourcesDigest(*current.FieldFileSources)
		if _, _, err := SelectRecordedNativeOrigins(current, query); err == nil {
			t.Fatal("contradictory explicit field source accepted", i)
		}
	}
	legacy := read()
	legacy.FieldSources = nil
	legacy.FieldFileSources = nil
	rows, _, err := SelectRecordedNativeOrigins(legacy, query)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		var fact CompositionFieldFileFact
		if json.Unmarshal(row, &fact) != nil || fact.Mapping != "UNMAPPED" || fact.Reason != "ORIGINAL_FILE_ANNOTATION_ABSENT" {
			t.Fatal("invented old origin", fact)
		}
	}
	if !sameJointRuntime(r, legacy) || sameJointRuntime(legacy, r) {
		t.Fatal("directional retained claim comparison")
	}
	ordinary, _, err := SelectRecordedNativeOrigins(legacy, NativeOriginQuery{Kind: "field-writes", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if json.Unmarshal(ordinary[0], &fields) != nil || len(fields) != 32 {
		t.Fatal("existing primitive field fact layout changed", len(fields))
	}
	t.Log("10 explicit source/file contradictions rejected even after self-checksum update;absent old annotation stays UNMAPPED;32-field wire unchanged;producer executions0;model0")
}
