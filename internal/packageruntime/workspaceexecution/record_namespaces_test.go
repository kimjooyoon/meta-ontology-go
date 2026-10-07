package workspaceexecution

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func recordNamespaceFixture(t *testing.T) (packageruntime.Manifest, bodyexecution.CompositionCases) {
	t.Helper()
	root := "../../../examples/package-record-namespaces/"
	read, err := os.ReadFile(root + "read.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	write, err := os.ReadFile(root + "write.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(root + "cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := bodyexecution.DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "example/write", Activity: "Render"},
		Packages: []packageruntime.PackageSpec{
			{Path: "example/read", Name: "read", Sources: []packageruntime.Source{{Filename: "read.gooo", Content: string(read)}}},
			{Path: "example/write", Name: "write", Imports: []string{"example/read"}, Sources: []packageruntime.Source{{Filename: "write.gooo", Content: string(write)}}},
		}}
	return manifest, suite
}

func TestPackageRecordNamespacesPreserveIDsAndNativeValues(t *testing.T) {
	manifest, suite := recordNamespaceFixture(t)
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FinitePassed != 12 || result.Runtime.FiniteTotal != 12 || !result.Runtime.RuntimeReplayed || len(result.Program.EntityAliases) != 2 {
		t.Fatal("scoped records did not execute", result.Runtime.FinitePassed, result.Program.EntityAliases)
	}
	if len(result.Composition.Plan.Records) != 2 || result.Composition.Plan.Records[0].ID == result.Composition.Plan.Records[1].ID {
		t.Fatal("different stable records were merged")
	}
	for _, alias := range result.Program.EntityAliases {
		if alias.Name != "Result" || !strings.HasPrefix(alias.LoweredName, "GoooEntity") || alias.EntityID == "" {
			t.Fatal("package display-name mapping missing", alias)
		}
	}
	if !strings.Contains(result.Program.Source, "Result belongs to read") || !strings.Contains(manifest.Packages[0].Sources[0].Content, "entity Result id") {
		t.Fatal("scoping rewrote original source or string values")
	}
	manifest.Packages[0], manifest.Packages[1] = manifest.Packages[1], manifest.Packages[0]
	again, err := Prepare(manifest)
	if err != nil || again.Source != result.Program.Source || !reflect.DeepEqual(again.EntityAliases, result.Program.EntityAliases) {
		t.Fatal("manifest order changed record scoping", err)
	}
}

func TestRecordConstructorRewritingPreservesValuesAndLocalNames(t *testing.T) {
	n := workspaceRecordNames{canonical: map[string]string{"id": "Scoped"}, known: map[string]bool{"Result": true}, visible: map[string]map[string]string{"pkg": {"Result": "Scoped"}}}
	body := "let Result = true\n// Result{value: 3}\nif Result { return Result{value: \"한글 Result{value: 4}\"} }; return Result{value: \"let Result\"}"
	rewritten, err := n.rewriteConstructors("pkg", body, true)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(body, "return Result{", "return Scoped{", -1)
	if rewritten != want {
		t.Fatal("non-type occurrence changed", rewritten)
	}
	if _, err := n.rewriteConstructors("other", "Result{value: 1}", false); err == nil {
		t.Fatal("unimported record constructor was visible")
	}
	n.visible["pkg"]["Result"] = ""
	if _, err := n.rewriteConstructors("pkg", "Result{value: 1}", false); err == nil {
		t.Fatal("ambiguous constructor was chosen")
	}
}

func TestRecordNamespacesRejectSharedIDShapeDriftAndIgnoreUnusedBodies(t *testing.T) {
	manifest, _ := recordNamespaceFixture(t)
	manifest.Packages[0].Sources[0].Content += "\nactivity Unused(Integer) -> Result computes \"not a supported body\"\n"
	if _, err := Prepare(manifest); err != nil {
		t.Fatal("unreachable body was parsed", err)
	}
	manifest.Packages[1].Sources[0].Content = strings.ReplaceAll(manifest.Packages[1].Sources[0].Content, "record-namespaces://write/result", "record-namespaces://read/result")
	if _, err := Prepare(manifest); err == nil || !strings.Contains(err.Error(), "conflicting field") {
		t.Fatal("same ID accepted two record shapes", err)
	}
}

func TestPackageRecordNamespacesScopeExternalPlanConstructors(t *testing.T) {
	manifest, suite := recordNamespaceFixture(t)
	manifest.Packages[0].Sources[0].Content = strings.Replace(manifest.Packages[0].Sources[0].Content,
		`return Result{value: input + 1, note: "Result belongs to read"}`, `return __GOOO_BODY_HOLE_result__`, 1)
	plan := bodycodegen.IRBodyFillPlan{Schema: "gooo/body-codegen-ir-fill-plan/v3-record", Intent: "Construct the local read Result.",
		Holes: []bodycodegen.IRBodyFillHole{{ID: "result"}}, Candidates: []bodycodegen.IRBodyFillCandidate{
			{ID: "increment", Fills: map[string]string{"result": `Result{value: input + 1, note: "Result belongs to read"}`}},
			{ID: "identity", Fills: map[string]string{"result": `Result{value: input, note: "Result belongs to read"}`}},
		}, ValueCases: []assemblyspec.ValueCase{{Inputs: `[2]`, Expected: `{"note":"Result belongs to read","value":3}`}}}
	result, err := ExecuteWorkspaceWithOptions(context.Background(), manifest, suite, ExecuteOptions{BodyFillPlans: map[string]bodycodegen.IRBodyFillPlan{"example/read:Parse": plan}})
	if err != nil || result.Runtime.FinitePassed != 12 || len(result.BodyFills) != 1 {
		t.Fatal("external record constructor was not scoped", err)
	}
	if !strings.HasPrefix(plan.Candidates[0].Fills["result"], "Result{") {
		t.Fatal("lowering mutated the caller's external plan")
	}
}

func TestPackageRecordNamespacesScopeSourceFillConstructors(t *testing.T) {
	manifest, suite := recordNamespaceFixture(t)
	original := "activity Parse(Integer) -> Result computes `return Result{value: input + 1, note: \"Result belongs to read\"}`"
	replacement := "activity Parse(Integer) -> Result computes `if __GOOO_BODY_HOLE_condition__ { return __GOOO_BODY_HOLE_result__ }; return Result{value: input, note: \"Result belongs to read\"}` assembling {\n" + `
    source_fill intent "Construct the local read Result." {
        hole "condition"
        hole "result"
        candidate "increment" {
            fill "condition" "true"
            fill "result" "Result{value: input + 1, note: \"Result belongs to read\"}"
        }
        candidate "identity" {
            fill "condition" "false"
            fill "result" "Result{value: input, note: \"Result belongs to read\"}"
        }
    }
    value_case "[2]" -> "{\"note\":\"Result belongs to read\",\"value\":3}"
}`
	manifest.Packages[0].Sources[0].Content = strings.Replace(manifest.Packages[0].Sources[0].Content, original, replacement, 1)
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil || result.Runtime.FinitePassed != 12 || len(result.BodyFills) != 1 {
		t.Fatal("source-fill record constructor was not scoped", err)
	}
}

func TestPackageRecordNamespacesUnifyAliasesOfTheSameStableRecord(t *testing.T) {
	manifest, _ := recordNamespaceFixture(t)
	manifest.Packages[1].Sources[0].Content = `package write
namespace write
import read "example/read"
entity Renamed id "record-namespaces://read/result" fields {
    field value id "record-namespaces://read/value" type integer required one
    field note id "record-namespaces://read/note" type string required one
}
activity Render(Integer) -> Renamed computes "return Renamed{value: input, note: \"shared identity\"}"
bind read.Unwrap.result -> Render.input
`
	program, err := Prepare(manifest)
	if err != nil || strings.Count(program.Source, "entity Renamed id") != 1 || strings.Contains(program.Source, "entity Result id") {
		t.Fatal("record aliases lost stable identity", err, program.Source)
	}
}
