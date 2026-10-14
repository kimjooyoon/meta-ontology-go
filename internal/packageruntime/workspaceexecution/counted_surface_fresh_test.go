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

func TestScopedCounterNative18217(t *testing.T) {
	const source = `package scopedsum
namespace scopedsum18217
entity Integer id "urn:gooo:type:integer"
activity Sum(Integer)->Integer id "urn:scopedsum18217:sum" computes ` + "`" + `let total=input+18217
for let i=0;i<4;i++ {total=total+i}
for let i=0;i<3;i++ {total=total+i*2}
let i=43
return total+i` + "`" + "\n"
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "origin/scopedsum18217", Activity: "Sum"}, Packages: []packageruntime.PackageSpec{{Path: "origin/scopedsum18217", Name: "scopedsum", Sources: []packageruntime.Source{{Filename: "반복18217.gooo", Content: source}}}}}
	cases, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/scopedsum18217:Sum":83},"expected":{"origin/scopedsum18217:Sum":18355}},{"inputs":{"origin/scopedsum18217:Sum":-97},"expected":{"origin/scopedsum18217:Sum":18175}},{"inputs":{"origin/scopedsum18217:Sum":0},"expected":{"origin/scopedsum18217:Sum":18272}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, cases, "", "")
	if err != nil || result.Runtime.FinitePassed != 3 {
		t.Fatal("new scoped counters native", err, result.Runtime.FinitePassed)
	}
	r := result.Runtime
	ids := map[int]bool{}
	updateSites := map[int]bool{}
	for j, site := range r.LocalSites {
		if site.Name == "i" {
			ids[site.VariableOrdinal] = true
		}
		if site.Operation == "++" {
			updateSites[j] = true
			if site.Kind != "update" || site.Source.Value == nil || site.Source.Value.Expression != "i++" || site.Source.Value.Status != "MAPPED" {
				t.Fatal("authored post source", site)
			}
		}
	}
	updates := 0
	for _, trace := range r.Traces {
		for index, w := range trace.LocalWrites {
			if !updateSites[w.SiteIndex] {
				continue
			}
			updates++
			if !w.Completed || w.UseEnd-w.UseStart != 1 {
				t.Fatal("increment must read its previous value once", index, w)
			}
			use := trace.LocalUses[w.UseStart]
			if use.Assignment < 0 || use.Assignment >= index || r.LocalUseSites[use.SiteIndex].VariableOrdinal != r.LocalSites[w.SiteIndex].VariableOrdinal {
				t.Fatal("wrong counter history", index, use)
			}
		}
	}
	if len(ids) != 3 || len(updateSites) != 2 || updates != 21 || !r.RuntimeReplayed || r.ModelCalls != 0 {
		t.Fatal("scope/read/write populations", ids, updateSites, updates)
	}
	for _, b := range r.LocalFileSources.Bindings {
		if b.Value.Expression == "i++" && (b.Value.Status != "MAPPED" || b.Value.Filename != "반복18217.gooo" || source[b.Value.Start:b.Value.End] != "i++") {
			t.Fatal("original increment file", b)
		}
	}
	if root := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); root != "" {
		f, e := os.Create(filepath.Join(root, "counter18217-runtime.json.gz"))
		if e != nil {
			t.Fatal(e)
		}
		z := gzip.NewWriter(f)
		if e = json.NewEncoder(z).Encode(r); e != nil {
			t.Fatal(e)
		}
		if e = z.Close(); e != nil {
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
	}
	t.Log("new native scoped-counter cases3/3;two loop counters and later same-name variable distinct;21 completed ++ updates each read own previous assignment once;original Korean file coordinates;two-run replay;model0")
}
