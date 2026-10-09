package workspaceexecution

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestWorkspaceConstructionImportedTypedBranchReplay(t *testing.T) {
	read := func(name string) []byte {
		raw, err := os.ReadFile("../../../examples/caller-typed-paths/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	helper, _, _ := strings.Cut(string(read("main.gooo.fixture")), "activity Main(")
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "app/paths", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{
			{Path: "app/paths", Name: "app", Imports: []string{"tools/branch"}, Sources: []packageruntime.Source{{
				Filename: "app.gooo.fixture", Content: `package app
namespace app
import paths "tools/branch"
activity Main(Integer) -> Integer computes "return paths.Choose(input) + input"
`}}},
			{Path: "tools/branch", Name: "callerpaths", Sources: []packageruntime.Source{{
				Filename: "branch.gooo.fixture", Content: helper}}},
		}}
	qualified := func(name string) []byte {
		return []byte(strings.ReplaceAll(string(read(name)), `"Main"`, `"app/paths:Main"`))
	}
	feedback, err := bodyexecution.DecodeCompositionCases(qualified("construction-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err := bodyexecution.DecodeCompositionCases(qualified("evaluation-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	prior, err := ConstructWorkspace(context.Background(), manifest, feedback, evaluation, ConstructOptions{ProgramBudget: 2})
	if err != nil || prior.Construction.Schema != "gooo/joint-construction/v7" || prior.Evaluation.Runtime.FinitePassed != 3 {
		t.Fatal("imported branch was not reconsidered", err)
	}
	replay, err := ReplayWorkspaceConstruction(context.Background(), manifest, prior, evaluation, "")
	if err != nil || !replay.Evaluation.ConstructionReplayed || replay.Evaluation.NewModelCalls != 0 || replay.Evaluation.Runtime.FinitePassed != 3 {
		t.Fatal("imported selected body did not replay", err)
	}
}
