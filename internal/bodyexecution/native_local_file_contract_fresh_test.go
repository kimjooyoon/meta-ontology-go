package bodyexecution

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRecordedFileAnchors12591(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	f, err := os.Open(filepath.Join(root, "internal/bodyexecution/testdata/local-files12421-runtime.json.gz"))
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
	decode := func() CompositionRuntime {
		var r CompositionRuntime
		if err := json.Unmarshal(raw, &r); err != nil {
			t.Fatal(err)
		}
		return r
	}
	query := NativeOriginQuery{Kind: "local-input-locations", Limit: 32}
	if rows, _, err := SelectRecordedNativeOrigins(decode(), query); err != nil || len(rows) != 24 {
		t.Fatal("retained record", err)
	}
	mutations := []func(*CompositionRuntime){
		func(r *CompositionRuntime) { r.LocalFileSources.SelectedSourceSHA256 = "different-source" },
		func(r *CompositionRuntime) { r.LocalFileSources.Bindings[0].SiteIndex = 1000 },
		func(r *CompositionRuntime) { r.LocalFileSources.Bindings[1] = r.LocalFileSources.Bindings[0] },
		func(r *CompositionRuntime) { r.LocalFileSources.Bindings[0].Value.Filename = "" },
		func(r *CompositionRuntime) {
			r.LocalFileSources.Bindings[0].Value.End = r.LocalFileSources.Bindings[0].Value.Start
		},
		func(r *CompositionRuntime) { r.LocalFileSources.Bindings[0].Value.Status = "COMPLETE" },
	}
	for index, mutate := range mutations {
		r := decode()
		mutate(&r)
		r.LocalFileSources.SHA256 = localFileSourcesDigest(*r.LocalFileSources)
		if _, _, err := SelectRecordedNativeOrigins(r, query); err == nil {
			t.Fatal("contradictory explicit file annotation accepted", index)
		}
	}
	legacy := decode()
	legacy.LocalFileSources = nil
	rows, _, err := SelectRecordedNativeOrigins(legacy, query)
	if err != nil || len(rows) != 24 {
		t.Fatal("absent historical annotation", err)
	}
	for _, row := range rows {
		var fact CompositionLocalFileFact
		if err := json.Unmarshal(row, &fact); err != nil {
			t.Fatal(err)
		}
		if fact.Mapping != "UNMAPPED" || fact.Reason != "ORIGINAL_FILE_ANNOTATION_ABSENT" || fact.Filename != "" || fact.SourceSHA256 != "" || fact.SourceStart != 0 || fact.SourceEnd != 0 {
			t.Fatal("absent history acquired a physical claim", fact)
		}
	}
	t.Log("saved record read only;6 contradictory annotations rejected even with recomputed digest;24 absent historical file claims stay UNMAPPED;native/model executions0")
}
