package bodyexecution

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestSavedFieldClaims13649(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	path := "testdata/fields13517-runtime.json.gz"
	if root != "" {
		path = filepath.Join(root, "internal/bodyexecution", path)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	z, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
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
	if err := validateRecordedFieldObservations(decode()); err != nil {
		t.Fatal(err)
	}
	changes := []func(*CompositionRuntime){
		func(r *CompositionRuntime) { r.Traces[0].FieldUses[len(r.Traces[0].FieldUses)-1].Assignment-- },
		func(r *CompositionRuntime) { r.Traces[0].FieldUses[len(r.Traces[0].FieldUses)-1].CompletionClock = 0 },
		func(r *CompositionRuntime) {
			r.Traces[0].FieldWrites[1].Completion = r.Traces[0].FieldWrites[0].Completion
		},
		func(r *CompositionRuntime) { r.FieldObservation = nil },
		func(r *CompositionRuntime) { r.Traces[0].FieldUses[0].Reason = "UNKNOWN" },
		func(r *CompositionRuntime) { r.Traces[0].FieldUseDropped = -1 },
		func(r *CompositionRuntime) {
			for _, write := range r.Traces[0].FieldWrites {
				if write.ParentAssignment < 0 {
					continue
				}
				target := r.FieldSites[write.SiteIndex]
				parent := r.FieldSites[r.Traces[0].FieldWrites[write.ParentAssignment].SiteIndex]
				if target.Binding != parent.Binding {
					r.FieldBindings[target.Binding].Record.Fields[target.Field].ID = "urn:changed13649:field"
					break
				}
			}
			p := runtimeFieldPlan(*r)
			r.FieldObservation.MetadataSHA256 = compositionDigest([]any{p.sourceSHA256, p.inventory.bindings, p.sites, p.uses})
		},
	}
	for i, change := range changes {
		r := decode()
		change(&r)
		if err := validateRecordedFieldObservations(r); err == nil {
			t.Fatal("inconsistent field claim accepted", i)
		}
	}
	old := decode()
	old.FieldObservation, old.FieldBindings, old.FieldSites, old.FieldUseSites, old.FieldProjectionSHA256 = nil, nil, nil, nil, ""
	for i := range old.Traces {
		old.Traces[i].FieldWrites, old.Traces[i].FieldUses = nil, nil
		old.Traces[i].FieldFrames, old.Traces[i].FieldFrameDropped, old.Traces[i].FieldWriteDropped, old.Traces[i].FieldUseDropped = 0, 0, 0, 0
	}
	if err := validateRecordedFieldObservations(old); err != nil {
		t.Fatal("absent historical claim", err)
	}
	if _, _, err := SelectRecordedNativeOrigins(old, NativeOriginQuery{Kind: "field-uses", Limit: 1}); err == nil {
		t.Fatal("absent historical field journal became a claim")
	}
	t.Log("saved13517 record only;7 contradictory identity/order/population claims rejected,including changed field ID with resealed metadata;absent field history remains absent;native/model executions0")
}
