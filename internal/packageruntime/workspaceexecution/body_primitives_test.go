package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func primitiveWorkspace(body, helpers string) packageruntime.Manifest {
	return packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry:    packageruntime.EntrySpec{PackagePath: "text", Activity: "Size"},
		Packages: []packageruntime.PackageSpec{{Path: "text", Name: "text", Sources: []packageruntime.Source{{Filename: "text.gooo", Content: "package text\nnamespace text\nentity Text id \"text://text\"\nentity Integer id \"text://integer\"\nactivity Size(Text) -> Integer computes `" + body + "`\n" + helpers}}}}}
}

func TestWorkspaceBodyPrimitivesAndNestedSourceCalls(t *testing.T) {
	for _, nested := range []bool{false, true} {
		body, helpers := "return int64(len(input))", ""
		want := []int64{0, 3, 7, 1}
		if nested {
			body = "return int64(len(Append(input)))"
			helpers = "activity Append(Text) -> Text computes `return input + \"!\"`\n"
			want = []int64{1, 4, 8, 2}
		}
		m := primitiveWorkspace(body, helpers)
		suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1"}
		for i, input := range []string{"", "abc", "한🙂", "\n"} {
			x, _ := json.Marshal(input)
			y, _ := json.Marshal(want[i])
			suite.Cases = append(suite.Cases, bodyexecution.CompositionCase{Inputs: map[string]json.RawMessage{"text:Size": x}, Expected: map[string]json.RawMessage{"text:Size": y}})
		}
		r, err := ExecuteWorkspace(context.Background(), m, suite, "", "")
		if err != nil || r.Runtime.FinitePassed != 4 || !r.Runtime.RuntimeReplayed {
			t.Fatal("primitive native execution", nested, err)
		}
		if !nested && r.Program.PureCalls != nil {
			t.Fatal("primitive became a source dependency")
		}
		if nested && (r.Program.PureCalls == nil || len(r.Program.PureCalls.Sites) != 1) {
			t.Fatal("nested activity dependency was lost")
		}
		again, err := ReplayWorkspace(context.Background(), m, r, suite, "")
		if err != nil || again.Replay.ModelCalls != 0 || again.Runtime.FinitePassed != 4 {
			t.Fatal("primitive replay", err)
		}
	}
}

func TestWorkspacePrimitiveNamesPreserveSourceAndLocalShadows(t *testing.T) {
	for _, name := range []string{"len", "int64"} {
		m := primitiveWorkspace("return "+name+"(input)", "activity "+name+"(Text) -> Integer computes `return 41`\n")
		suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"text:Size":"abc"},"expected":{"text:Size":41}}]}`))
		if err != nil {
			t.Fatal(err)
		}
		r, err := ExecuteWorkspace(context.Background(), m, suite, "", "")
		if err != nil || r.Runtime.FinitePassed != 1 || r.Program.PureCalls == nil || len(r.Program.PureCalls.Sites) != 1 {
			t.Fatal("source activity shadow lost", name, err)
		}
		local := primitiveWorkspace("let "+name+" = input; return "+name+"(input)", "")
		if _, err := Prepare(local); err == nil {
			t.Fatal("local value treated as primitive", name)
		}
		abstract := primitiveWorkspace("return "+name+"(input)", "activity "+name+"(Text) -> Integer\n")
		if _, err := Prepare(abstract); err == nil {
			t.Fatal("declaration without body became a primitive", name)
		}
	}
}

func TestWorkspaceTextAssemblyUsesPrimitivesInAllSurfaces(t *testing.T) {
	raw, err := os.ReadFile("../../../examples/text-operations/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	m := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "files", Activity: "Classify"}, Packages: []packageruntime.PackageSpec{{Path: "files", Name: "filenames", Sources: []packageruntime.Source{{Filename: "files.gooo", Content: string(raw)}}}}}
	// Put primitives directly in a field alternative as well as in helpers.
	m.Packages[0].Sources[0].Content = strings.Replace(string(raw), `alternative "bytes"`, `alternative "int64(len(input))"`, 1)
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"files:Classify":"한🙂.gooo"},"expected":{"files:Classify":{"source":true,"stem":"한🙂","bytes":12}}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteWorkspace(context.Background(), m, suite, "", "")
	if err != nil || r.Runtime.FinitePassed != 1 {
		t.Fatal("text field assembly", err)
	}
	again, err := ReplayWorkspace(context.Background(), m, r, suite, "")
	if err != nil || again.Runtime.FinitePassed != 1 || again.Replay.ModelCalls != 0 {
		t.Fatal("text assembly replay", err)
	}
}

func TestWorkspacePrimitiveCallsStillTypecheck(t *testing.T) {
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"text:Size":"abc"},"expected":{"text:Size":3}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{"return int64(len())", "return int64(len(input,input))", "return int64(input)", "return int64(cap(input))"} {
		if _, err := ExecuteWorkspace(context.Background(), primitiveWorkspace(body, ""), suite, "", ""); err == nil {
			t.Fatal("invalid primitive call executed", body)
		}
	}
}

func TestWorkspacePrimitiveFieldUpdateUsesUpdatedText(t *testing.T) {
	source := "package text\nnamespace text\nentity Text id \"text://text\"\n" +
		"entity Result id \"text://result\" fields {\n" +
		"field text id \"text://result/text\" type string required one\n" +
		"field bytes id \"text://result/bytes\" type integer required one\n}\n" +
		"activity Size(Text) -> Result computes `let output = Result{text: input, bytes: 0}\n" +
		"output.text = input + \"!\"\noutput.bytes = 0\nreturn output` assembling {\n" +
		"choice \"size\" field_update at \"1\" alternative \"int64(len(output.text))\" intent \"Count updated bytes.\"\n" +
		"value_case \"[\\\"x\\\"]\" -> \"{\\\"text\\\":\\\"x!\\\",\\\"bytes\\\":2}\"\nattempts \"2\"\n}\n"
	m := primitiveWorkspace("return 0", "")
	m.Packages[0].Sources[0].Content = source
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"text:Size":"한🙂"},"expected":{"text:Size":{"text":"한🙂!","bytes":8}}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteWorkspace(context.Background(), m, suite, "", "")
	if err != nil || r.Runtime.FinitePassed != 1 {
		t.Fatal("primitive field update", err)
	}
	again, err := ReplayWorkspace(context.Background(), m, r, suite, "")
	if err != nil || again.Runtime.FinitePassed != 1 || again.Replay.ModelCalls != 0 {
		t.Fatal("field update replay", err)
	}
}
