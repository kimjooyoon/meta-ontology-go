package workspaceexecution

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestScopedCounterFieldBridge19711(t *testing.T) {
	const source = `package counterfields
namespace counterfields19711
entity State id "urn:counterfields19711:state" fields {
 field value id "urn:counterfields19711:value" type integer required one
 field enabled id "urn:counterfields19711:enabled" type boolean required one
}
activity Move(State)->State id "urn:counterfields19711:move" computes ` + "`" + `let state=input
for let i=0;i<5;i++ {if state.enabled {state.value=state.value+i}}
return state` + "`" + "\n"
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "origin/counterfields19711", Activity: "Move"}, Packages: []packageruntime.PackageSpec{{Path: "origin/counterfields19711", Name: "counterfields", Sources: []packageruntime.Source{{Filename: "counter-fields19711.gooo", Content: source}}}}}
	cases, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/counterfields19711:Move":{"value":211,"enabled":true}},"expected":{"origin/counterfields19711:Move":{"value":221,"enabled":true}}},{"inputs":{"origin/counterfields19711:Move":{"value":-223,"enabled":false}},"expected":{"origin/counterfields19711:Move":{"value":-223,"enabled":false}}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, cases, "", "")
	if err != nil || result.Runtime.FinitePassed != 2 {
		t.Fatal("fresh counter/field native", err, result.Runtime.FinitePassed)
	}
	r := result.Runtime
	if root := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); root != "" {
		f, e := os.Create(filepath.Join(root, "counter-fields19711.json.gz"))
		if e != nil {
			t.Fatal(e)
		}
		z := gzip.NewWriter(f)
		if e = json.NewEncoder(z).Encode(result); e != nil {
			t.Fatal(e)
		}
		if e = z.Close(); e != nil {
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
	}
	if r.BridgeObservation == nil || r.BridgeObservation.Status != "COMPLETE" || r.BridgeObservation.CrossReads < 5 || !r.RuntimeReplayed || r.ModelCalls != 0 {
		t.Fatal("counter/field shared observations", r.BridgeObservation)
	}
	updates := 0
	for _, trace := range r.Traces {
		if trace.Bridge == nil {
			t.Fatal("bridge trace missing")
		}
		for _, w := range trace.LocalWrites {
			if r.LocalSites[w.SiteIndex].Operation == "++" {
				updates++
				if w.UseEnd-w.UseStart != 1 {
					t.Fatal("counter update read window", w)
				}
			}
		}
	}
	if updates != 10 {
		t.Fatal("counter update population", updates)
	}
	for _, b := range r.FieldFileSources.Bindings {
		sites := r.FieldSites
		if b.Kind == "use" {
			sites = r.FieldUseSites
		}
		site := sites[b.SiteIndex]
		if site.Kind == "field_assignment" && (b.Value.Status != "MAPPED" || b.Value.Expression != "state.value+i" || source[b.Value.Start:b.Value.End] != "state.value+i") {
			t.Fatal("field source after scoped counter lowering", b)
		}
	}
	t.Logf("new record cases2/2;true/false field branch;10 scoped ++ updates;cross-journal counter reads%d;field RHS original location;complete shared observations;two-run replay;model0", r.BridgeObservation.CrossReads)
}
