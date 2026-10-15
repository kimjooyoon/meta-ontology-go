package workspaceexecution

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestFieldInputLocations16301(t *testing.T) {
	const records = `package fieldfiles
namespace fieldfiles16301
entity Input id "urn:fieldfiles16301:input" fields {
 field amount id "urn:fieldfiles16301:amount" type integer required one
 field active id "urn:fieldfiles16301:active" type boolean required one
 field title id "urn:fieldfiles16301:title" type string required one
}
entity State id "urn:fieldfiles16301:state" fields {
 field score id "urn:fieldfiles16301:score" type integer required one
 field live id "urn:fieldfiles16301:live" type boolean required one
 field tag id "urn:fieldfiles16301:tag" type string required one
}
activity Retain(State)->State id "urn:fieldfiles16301:retain" computes ` + "`" + `let adjusted=State{score:input.score+23,live:!input.live,tag:"한글:"+input.tag}
adjusted.score=adjusted.score+5
return adjusted` + "`" + "\n"
	const body = `let seed=State{score:input.amount+16301,live:input.active,tag:input.title}
let i=0;for i=0;i<3;i=i+1 {seed.score=seed.score+2}
let result=Retain(seed)
return result`
	main := `package fieldfiles
namespace fieldfiles16301
activity Build(Input)->State id "urn:fieldfiles16301:build" computes ` + strconv.Quote(body) + "\n"
	files := map[string]string{"records16301.gooo": records, "build16301.gooo": main}
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "origin/fieldfiles16301", Activity: "Build"}, Packages: []packageruntime.PackageSpec{{Path: "origin/fieldfiles16301", Name: "fieldfiles", Sources: []packageruntime.Source{{Filename: "records16301.gooo", Content: records}, {Filename: "build16301.gooo", Content: main}}}}}
	cases, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/fieldfiles16301:Build":{"amount":97,"active":true,"title":"다"}},"expected":{"origin/fieldfiles16301:Build":{"score":16432,"live":false,"tag":"한글:다"}}},{"inputs":{"origin/fieldfiles16301:Build":{"amount":-101,"active":false,"title":"B"}},"expected":{"origin/fieldfiles16301:Build":{"score":16234,"live":true,"tag":"한글:B"}}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, cases, "", "")
	if err != nil || result.Runtime.FinitePassed != 2 {
		t.Fatal("new field source producer", err, result.Runtime)
	}
	r := result.Runtime
	if r.FieldSources == nil || r.FieldFileSources == nil {
		t.Fatal("static field source tables missing")
	}
	mapped := map[string]int{}
	writes := map[int]int{}
	for _, b := range r.FieldFileSources.Bindings {
		sites := r.FieldSites
		if b.Kind == "use" {
			sites = r.FieldUseSites
		}
		site := sites[b.SiteIndex]
		for _, loc := range []bodyexecution.CompositionLocalFileLocation{b.Target, b.Value} {
			switch site.Kind {
			case "field_assignment", "record_constructor", "field_read":
				if loc.Status != "MAPPED" || loc.SourceSHA256 != sourceSHA256([]byte(files[loc.Filename])) || loc.Start < 0 || loc.End > len(files[loc.Filename]) || loc.End <= loc.Start {
					t.Fatalf("exact authored field span: site=%+v location=%+v", site, loc)
				}
				mapped[site.Kind]++
			default:
				if loc.Status != "UNMAPPED" {
					t.Fatal("invented original body span for unsupported site kind", site, loc)
				}
			}
		}
	}
	for _, w := range r.Traces[0].FieldWrites {
		writes[w.SiteIndex]++
	}
	repeated := 0
	for _, n := range writes {
		if n >= 3 {
			repeated++
		}
	}
	if repeated == 0 || mapped["field_assignment"] == 0 || mapped["record_constructor"] == 0 || mapped["field_read"] == 0 {
		t.Fatal("missing populations", mapped, writes)
	}
	rows, report, err := bodyexecution.SelectRecordedNativeOrigins(r, bodyexecution.NativeOriginQuery{Kind: "field-input-locations", Limit: 128})
	if err != nil || len(rows) == 0 || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
		t.Fatal("recorded location read", err)
	}
	if root := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); root != "" {
		save := func(name string, value any) {
			t.Helper()
			f, e := os.Create(filepath.Join(root, name))
			if e != nil {
				t.Fatal(e)
			}
			gz := gzip.NewWriter(f)
			e = json.NewEncoder(gz).Encode(value)
			if e != nil {
				t.Fatal(e)
			}
			if e = gz.Close(); e != nil {
				t.Fatal(e)
			}
			if e = f.Close(); e != nil {
				t.Fatal(e)
			}
		}
		save("field-files16301-runtime.json.gz", r)
		save("field-files16301-workspace.json.gz", result)
		save("field-files16301-manifest.json.gz", manifest)
	}
	t.Logf("new native business cases2/2;authored location counts%v;repeat sites%d;field/local journals retained;source files2;model0;saved read producer0", mapped, repeated)
}
