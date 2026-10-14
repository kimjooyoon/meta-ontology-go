package syntaxregistration

import (
	"encoding/json"
	"go/ast"
	"strings"
	"testing"
	"testing/fstest"
)

func TestSameFileRegistryOrigins33503(t *testing.T) {
	data, request := promotedMetaFixture(t)
	prepared := map[string][]byte{}
	for name, file := range data {
		prepared[name] = file.Data
	}
	source, err := parseSourceUnits(prepared, syntaxRoot, "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	var fused strings.Builder
	fused.WriteString("package languagesyntax\n")
	source.activity = "gooo://physical-origins33503/layout"
	for _, name := range []string{"expectedRegistry", "expectedSyntaxCaseGroup5"} {
		function, err := source.function(name)
		if err != nil {
			t.Fatal(err)
		}
		fused.WriteString(source.text(function) + "\n")
		source.replace(function, "")
	}
	changed, err := source.finish()
	if err != nil {
		t.Fatal(err)
	}
	for name, raw := range changed {
		data[name] = &fstest.MapFile{Data: raw}
	}
	owner := syntaxRoot + "same_file_origins33503.go"
	data[owner] = &fstest.MapFile{Data: []byte(fused.String())}
	request.Case.ID = "same-file-origin33503"
	pin(t, data, &request)
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := plan.Generate(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.ValidateCandidate(data, candidate); err != nil {
		t.Fatal(err)
	}
	if candidate.Required != 9 || candidate.Emitted != 9 || len(candidate.Members) != 9 || len(candidate.Artifacts) != 9 {
		t.Fatal("same owner lost exact9-file/9-role write set")
	}
	native := candidate.Artifacts[1]
	if len(native.Paths) != 1 || native.Paths[0] != owner {
		t.Fatal("physical owner was guessed", native.Paths)
	}
	for _, member := range candidate.Members {
		data[member.Path] = &fstest.MapFile{Data: member.Content}
	}
	var registry struct {
		MetaSources []string `json:"meta_sources"`
	}
	if err := json.Unmarshal(data[corpusPath].Data, &registry); err != nil {
		t.Fatal(err)
	}
	for _, path := range registry.MetaSources {
		if path == request.Case.Path {
			t.Fatal("promoted membership retained")
		}
	}
	inputs := map[string][]byte{}
	for name, file := range data {
		inputs[name] = file.Data
	}
	updated, err := parseSourceUnits(inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	function, err := updated.function("expectedRegistry")
	if err != nil {
		t.Fatal(err)
	}
	metaCount := 0
	ast.Inspect(function, func(node ast.Node) bool {
		field, ok := node.(*ast.KeyValueExpr)
		if !ok || updated.text(field.Key) != "MetaSources" {
			return true
		}
		metaCount++
		values, err := nativeMetaSourceValues(updated, field.Value.(*ast.CompositeLit))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range values {
			if path == request.Case.Path {
				t.Fatal("native promoted membership retained")
			}
		}
		return false
	})
	if metaCount != 1 {
		t.Fatal("native membership owner missing or duplicated")
	}
	t.Log("new fused physical layout:9 exact members/9 source-owned activities;both promoted memberships removed;no guessed filename;model0")
}
