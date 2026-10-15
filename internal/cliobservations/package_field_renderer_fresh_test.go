package cliobservations

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const fieldRendererSource14321 = `package renderer14321
namespace renderer14321
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:renderer14321:input" fields {
field left id "urn:renderer14321:input-left" type integer required one
field right id "urn:renderer14321:input-right" type integer required one
field flip id "urn:renderer14321:input-flip" type boolean required one
field enabled id "urn:renderer14321:input-enabled" type boolean required one
field tag id "urn:renderer14321:input-tag" type string required one
}
entity State id "urn:renderer14321:state" fields {
field score id "urn:renderer14321:score" type integer required one
field live id "urn:renderer14321:live" type boolean required one
field tag id "urn:renderer14321:tag" type string required one
}
activity Adjust(State)->State id "urn:renderer14321:adjust" computes ` + "`let saved=input;saved.score=saved.score+19;saved.live=!saved.live;saved.tag=saved.tag+\":linked\";return saved`" + `
activity Read(State)->Integer id "urn:renderer14321:read" computes ` + "`return input.score`" + `
activity Build(Input)->State id "urn:renderer14321:build" computes ` + "`let original=State{score:input.left,live:input.enabled,tag:input.tag};let adjusted=Adjust(original);if input.flip {adjusted=State{score:input.right,live:adjusted.live,tag:adjusted.tag}};return State{score:Read(adjusted),live:adjusted.live,tag:adjusted.tag}`" + `
`

func TestGoooFreshFieldRenderer14321(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	sourceRoot := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if sourceRoot == "" {
		sourceRoot = "../.."
	}
	root := t.TempDir()
	files := map[string]string{
		"source.gooo":         fieldRendererSource14321,
		"gooo.workspace.json": `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"fresh/renderer","activity":"Build"},"packages":[{"path":"fresh/renderer","name":"renderer14321","imports":[],"sources":["source.gooo"]}]}`,
		"cases.json": `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"fresh/renderer:Build":{"left":11,"right":31,"flip":false,"enabled":true,"tag":"first"}},"expected":{"fresh/renderer:Build":{"score":30,"live":false,"tag":"first:linked"}}},
{"inputs":{"fresh/renderer:Build":{"left":11,"right":31,"flip":true,"enabled":true,"tag":"second"}},"expected":{"fresh/renderer:Build":{"score":31,"live":false,"tag":"second:linked"}}},
{"inputs":{"fresh/renderer:Build":{"left":-19,"right":0,"flip":false,"enabled":false,"tag":"zero-left"}},"expected":{"fresh/renderer:Build":{"score":0,"live":true,"tag":"zero-left:linked"}}},
{"inputs":{"fresh/renderer:Build":{"left":-19,"right":0,"flip":true,"enabled":false,"tag":"zero-right"}},"expected":{"fresh/renderer:Build":{"score":0,"live":true,"tag":"zero-right:linked"}}}]}`,
	}
	for name, value := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	call := func(name string, args ...string) packageExecutionReceipt {
		t.Helper()
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK {
			t.Fatal(name, code, diagnostics.String(), out.String())
		}
		var receipt packageExecutionReceipt
		if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil {
			t.Fatal(name, err, receipt.Error)
		}
		if dir := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); dir != "" {
			var compressed bytes.Buffer
			z := gzip.NewWriter(&compressed)
			if _, err := z.Write(out.Bytes()); err != nil {
				t.Fatal(err)
			}
			if err := z.Close(); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, name+".json.gz"), compressed.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
		}
		return receipt
	}
	producer := call("producer14321", "package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "gooo.workspace.json"))
	runtime := producer.Result.Runtime
	if producer.Decision != "PASS" || runtime.FinitePassed != 4 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 || runtime.FieldObservation == nil || runtime.FieldObservation.Status != "COMPLETE" {
		t.Fatal("fresh mixed-scalar native outcome", runtime.FinitePassed, runtime.FieldObservation)
	}
	offset := 0
	var inputs []map[string]bodyexecution.FieldLineageColumns
	for c, trace := range runtime.Traces {
		selected := -1
		for i, use := range trace.FieldUses {
			site := runtime.FieldUseSites[use.SiteIndex]
			if runtime.FieldBindings[site.Binding].ActivityID == "urn:renderer14321:read" {
				selected = i
			}
		}
		if selected < 0 {
			t.Fatal("fresh final score read absent", c)
		}
		columns, report, err := bodyexecution.SelectRecordedFieldLineage(runtime, bodyexecution.NativeOriginQuery{Kind: "field-uses", Offset: offset + selected, Limit: 1})
		if err != nil || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
			t.Fatal(err, report)
		}
		inputs = append(inputs, map[string]bodyexecution.FieldLineageColumns{"field/lineage:Trace": columns})
		offset += len(trace.FieldUses)
	}
	raw, err := json.Marshal(map[string]any{"schema": bodyexecution.CompositionInputsSchema, "inputs": inputs})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "field-inputs14321.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	consumer := call("consumer14321", "package", "execute", "--json", "--inputs", path, filepath.Join(sourceRoot, "examples/native-field-lineage/gooo.workspace.json"))
	if consumer.Decision != "OBSERVED" || consumer.Result.Runtime.FiniteTotal != 0 || consumer.Result.Runtime.ModelCalls != 0 || len(consumer.Result.Runtime.Traces) != 4 {
		t.Fatal("fresh input-only Gooo traversal scope")
	}
	for c, trace := range consumer.Result.Runtime.Traces {
		var actual fieldLineageAnswer14203
		field := "left"
		if c%2 == 1 {
			field = "right"
		}
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil || actual.Status != "SCOPED_FIELD_ORIGINS" || actual.Roots != "urn:renderer14321:build|input|urn:renderer14321:input-"+field+";" || actual.RootCount != 1 || actual.Unresolved != 0 {
			t.Fatal("fresh native field origins", strconv.Itoa(c), actual)
		}
	}
	t.Log("new mixed integer/bool/string producer4/4 and exact native replay;decomposed renderer preserves constructor/copy/parameter/return semantics;Gooo resolves4 native origins including equal-zero branch alternatives;models0;producer replay during ancestry reads0")
}
