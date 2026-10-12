package workspaceexecution

import (
	"context"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func observedWorkspace() packageruntime.Manifest {
	return packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "example/observed", Activity: "Fee"},
		Packages: []packageruntime.PackageSpec{{Path: "example/observed", Name: "observed", Sources: []packageruntime.Source{
			{Filename: "fee.gooo", Content: `package observed
namespace observed
entity Integer id "workspace-observed://integer"
activity Fee(Integer) -> Integer computes "return input - 19" assembling {
 choice "direction" operand_order at "0" intent "19에서 입력을 뺀다."
 case "19" -> "0"
 observe { input "23" oracle "Reference" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"
}`},
			{Filename: "reference.gooo", Content: `package observed
namespace observed
activity Reference(Integer) -> Integer computes "return 19 - input"`},
		}}}}
}

func TestSourceObserveWorkspaceRetainsAndRenamesReference(t *testing.T) {
	program, err := Prepare(observedWorkspace())
	if err != nil {
		t.Fatal(err)
	}
	if program.PureCalls == nil || len(program.PureCalls.Sites) != 1 ||
		program.PureCalls.Sites[0].Surface != "observe:oracle" || len(program.PureCalls.Activities) != 2 {
		t.Fatal("workspace dropped the source oracle dependency", program.PureCalls)
	}
	source := []byte(program.Source)
	ctx := context.Background()
	spec, err := bodycodegen.SourceAssembly(ctx, "workspace.gooo", source, program.Entry.LoweredName)
	if err != nil || spec.Observation == nil || spec.Observation.OracleActivity == "Reference" {
		t.Fatal("source reference was not lowered with its package", err)
	}
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(ctx, "workspace.gooo", source, program.Entry.LoweredName)
	if err != nil || result.Report.BodyPaths == nil || result.Report.BodyPaths.Observation == nil ||
		result.Report.BodyPaths.Observation.Status != "ONE_SURVIVING_CANDIDATE" {
		t.Fatal("workspace reference could not answer its source question", err)
	}
	for _, invalid := range []string{"Missing", "Fee"} {
		manifest := observedWorkspace()
		manifest.Packages[0].Sources[0].Content = strings.Replace(manifest.Packages[0].Sources[0].Content,
			`oracle "Reference"`, `oracle "`+invalid+`"`, 1)
		if _, err := Prepare(manifest); err == nil {
			t.Fatal("invalid workspace observation reference accepted", invalid)
		}
	}
}
