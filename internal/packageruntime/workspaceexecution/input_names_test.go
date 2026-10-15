package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func namedInputWorkspace() packageruntime.Manifest {
	const core = `package core
namespace namedcore
entity Integer id "urn:gooo:type:integer"
activity Difference(minuend Integer, subtrahend Integer) -> Integer id "urn:named:subtract" computes "return minuend - subtrahend"
`
	const app = `package app
namespace namedapp
import core "named/core"
activity Credit(원금 Integer) -> Integer id "urn:named:credit" computes "return 원금 + 23"
activity Net(원금 Integer, 공제 Integer) -> Integer id "urn:named:net" computes "return core.Difference(원금, 공제)"
bind Credit.result -> Net.input0
`
	return packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "named/app", Activity: "Net"},
		Packages: []packageruntime.PackageSpec{
			{Path: "named/core", Name: "core", Sources: []packageruntime.Source{{Filename: "difference.gooo", Content: core}}},
			{Path: "named/app", Name: "app", Imports: []string{"named/core"}, Sources: []packageruntime.Source{{Filename: "금액.gooo", Content: app}}},
		}}
}

func TestAuthoredInputWorkspaceNativeFirst(t *testing.T) {
	manifest := namedInputWorkspace()
	checkNamedInputInterface(t, manifest)
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"named/app:Credit":79,"named/app:Net.input1":31},"expected":{"named/app:Credit":102,"named/app:Net":71}},
 {"inputs":{"named/app:Credit":-41,"named/app:Net.input1":7},"expected":{"named/app:Credit":-18,"named/app:Net":-25}},
 {"inputs":{"named/app:Credit":0,"named/app:Net.input1":23},"expected":{"named/app:Credit":23,"named/app:Net":0}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FinitePassed != 6 || result.Runtime.FiniteTotal != 6 || result.Runtime.ModelCalls != 0 {
		t.Fatal("named input output obligations", result.Runtime)
	}
	flow, err := InspectSelectedValueOrigins(context.Background(), result, OriginQuery{Limit: 128})
	if err != nil || flow.Analysis != "MAPPED" || flow.Truncated {
		t.Fatal(flow, err)
	}
	checkNamedInputPaths(t, flow)
	if output := os.Getenv("GOOO_NAMED_INPUT_OUTPUT"); output != "" {
		raw, err := json.MarshalIndent(map[string]any{"manifest": manifest, "cases": suite, "result": result, "flow": flow}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "workspace.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("3 distinct cases, 6/6 output obligations; named helper arguments and canonical bind; model0")
}

func checkNamedInputInterface(t *testing.T, manifest packageruntime.Manifest) {
	t.Helper()
	i, err := packageruntime.Describe(manifest)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string][]string{}
	for _, pkg := range i.Packages {
		for _, d := range pkg.Declarations {
			if d.Kind == "activity" {
				names[d.ID] = d.InputNames
			}
		}
	}
	if !reflect.DeepEqual(names["urn:named:subtract"], []string{"minuend", "subtrahend"}) ||
		!reflect.DeepEqual(names["urn:named:net"], []string{"원금", "공제"}) ||
		!reflect.DeepEqual(names["urn:named:credit"], []string{"원금"}) {
		t.Fatal(names)
	}
}

func checkNamedInputPaths(t *testing.T, flow ProgramValueOrigins) {
	t.Helper()
	seen := map[string]bool{}
	bound, called, arguments := false, false, false
	for _, row := range flow.Rows {
		if row.Kind == "unresolved" {
			t.Fatal("named value origin unresolved", row)
		}
		if row.Kind == "external_input" {
			last := row.Path[len(row.Path)-1]
			seen[row.ActivityID+":"+last.Expression] = true
		}
		for _, step := range row.Path {
			bound = bound || step.Via == "bind"
			called = called || step.Via == "call_return"
			arguments = arguments || step.Via == "call_argument"
		}
	}
	if !seen["urn:named:credit:원금"] || !seen["urn:named:net:공제"] || !bound || !called || !arguments {
		t.Fatal("source origins did not cross the intended call/bind positions", seen, bound, called, arguments)
	}
}

func TestAuthoredInputWorkspaceShadowFirst(t *testing.T) {
	for _, name := range []string{"core", "length"} {
		manifest := namedInputWorkspace()
		source := &manifest.Packages[1].Sources[0].Content
		*source = strings.Replace(*source, "Net(원금 Integer, 공제 Integer)", "Net("+name+" Integer, 공제 Integer)", 1)
		call := name + ".Difference(" + name + ", 공제)"
		if name == "length" {
			call = "length(공제)"
		}
		*source = strings.Replace(*source, "core.Difference(원금, 공제)", call, 1)
		if _, err := Prepare(manifest); err == nil || !strings.Contains(err.Error(), "shadowed") {
			t.Fatal("named input resolved as a package/primitive", name, err)
		}
	}
}
