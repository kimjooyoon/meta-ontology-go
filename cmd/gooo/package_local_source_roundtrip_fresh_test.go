package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageLocalSourceRoundtrip10773(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := t.TempDir()
	source := `package origin
namespace origin10773
entity Input id "urn:origin10773:input" fields {
 field values id "urn:origin10773:values" type integer required many
 field flag id "urn:origin10773:flag" type boolean required one
}
entity Result id "urn:origin10773:result" fields {
 field text id "urn:origin10773:text" type string required one
 field total id "urn:origin10773:total" type integer required one
 field ok id "urn:origin10773:ok" type boolean required one
}
activity Main(Input) -> Result id "urn:origin10773:main" computes "let total=277;for item in input.values {let shifted=item+281;total=total+shifted};let label=\"steady\";if input.flag {label=\"changed\"};let accepted=input.flag && total>0;return Result{text:label,total:total,ok:accepted}"
`
	files := map[string]string{
		"main.gooo":           source,
		"gooo.workspace.json": `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"origin/10773","activity":"Main"},"packages":[{"path":"origin/10773","name":"origin","imports":[],"sources":["main.gooo"]}]}`,
		"first.json": `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"origin/10773:Main":{"values":[],"flag":true}},"expected":{"origin/10773:Main":{"text":"changed","total":277,"ok":true}}},
{"inputs":{"origin/10773:Main":{"values":[7,11],"flag":false}},"expected":{"origin/10773:Main":{"text":"steady","total":857,"ok":false}}},
{"inputs":{"origin/10773:Main":{"values":[-900],"flag":true}},"expected":{"origin/10773:Main":{"text":"changed","total":-342,"ok":false}}}]}`,
		"next.json": `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"origin/10773:Main":{"values":[13],"flag":true}},"expected":{"origin/10773:Main":{"text":"changed","total":571,"ok":true}}},
{"inputs":{"origin/10773:Main":{"values":[-15,17],"flag":false}},"expected":{"origin/10773:Main":{"text":"steady","total":841,"ok":false}}},
{"inputs":{"origin/10773:Main":{"values":[],"flag":false}},"expected":{"origin/10773:Main":{"text":"steady","total":277,"ok":false}}}]}`,
	}
	for name, value := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var first, next, diagnostics bytes.Buffer
	manifest, receipt := filepath.Join(root, "gooo.workspace.json"), filepath.Join(root, "saved.json")
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "first.json"), manifest}, &first, &diagnostics); code != exitOK {
		t.Fatal("first execution", code, diagnostics.String(), first.String())
	}
	if err := os.WriteFile(receipt, first.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"package", "replay", "--json", "--receipt", receipt, "--cases", filepath.Join(root, "next.json"), manifest}, &next, &diagnostics); code != exitOK {
		t.Fatal("saved local-origin envelope", code, diagnostics.String(), next.String())
	}
	for stage, raw := range map[string][]byte{"first": first.Bytes(), "replay": next.Bytes()} {
		var result packageExecutionReceipt
		if err := json.Unmarshal(raw, &result); err != nil || result.Result == nil || result.Decision != "PASS" {
			t.Fatal(stage, err, string(raw))
		}
		runtime := result.Result.Runtime
		if runtime.FinitePassed != 3 || runtime.FiniteTotal != 3 || runtime.ModelCalls != 0 || !runtime.RuntimeReplayed {
			t.Fatal(stage, "native results")
		}
		types := map[string]bool{}
		for _, site := range runtime.LocalUseSites {
			if site.Source == nil || site.Source.Status != "MAPPED" || site.Source.Expression != site.Name || site.Source.SourceSHA256 != runtime.SelectedSourceSHA256 {
				t.Fatal(stage, "source use coordinates", site)
			}
			types[site.ScalarType] = true
		}
		if len(types) != 3 || !types["int64"] || !types["bool"] || !types["string"] {
			t.Fatal(stage, "scalar type population", types)
		}
		if stage == "replay" && (result.ReplayedFrom == "" || result.Result.Replay == nil || result.Result.Replay.ModelCalls != 0) {
			t.Fatal("saved source lineage", result.ReplayedFrom)
		}
	}
	if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
		for name, data := range map[string][]byte{"first10773.json": first.Bytes(), "replay10773.json": next.Bytes()} {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Log("new source CLI3/3;stored-envelope replay with three disjoint inputs3/3;range/short-circuit/three scalar types have selected Gooo use coordinates;models0")
}
