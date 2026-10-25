package workspacesplit_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func TestJointSplitReloadsAndExecutesEverySuppliedWorkspace(t *testing.T) {
	for _, shared := range []bool{false, true} {
		checkJointPlan(t, shared)
	}
}

func checkJointPlan(t *testing.T, shared bool) {
	t.Helper()
	files, owners := sharedOwners(t)
	if shared {
		extra := owners[0].Definition.Packages[0]
		extra.Path = "shared/other"
		owners[0].Definition.Packages = append(owners[0].Definition.Packages, extra)
	}
	before, _ := json.Marshal(owners)
	plan, err := workspacesplit.Prepare(files, owners)
	if err != nil {
		t.Fatal(err)
	}
	if shared && len(plan.Owners[0].Split.Result.AffectedPackages) != 2 {
		t.Fatal("shared package inside one workspace was dropped")
	}
	after, _ := json.Marshal(owners)
	if string(before) != string(after) || len(plan.Owners) != 2 || len(plan.Files) != 2 || len(files) != 1 {
		t.Fatal("joint preparation changed inputs or duplicated shared writes")
	}
	for _, file := range plan.Files {
		files[filepath.Base(file.Path)] = &fstest.MapFile{Data: []byte(file.Source.Source)}
	}
	for _, owner := range plan.Owners {
		t.Run(owner.ManifestPath, func(t *testing.T) { checkJointWorkspace(t, files, owner) })
	}
	plan.Owners[0].Split.Definition.Packages[0].Sources[0] = "changed.gooo"
	if owners[0].Definition.Packages[0].Sources[0] != "shared.gooo" {
		t.Fatal("joint proposal aliases authored source lists")
	}
}

func checkJointWorkspace(t *testing.T, files fstest.MapFS, owner workspacesplit.OwnedProposal) {
	t.Helper()
	definition, err := workspaceinput.Decode(owner.DefinitionJSON)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := workspaceinput.Load(files, owner.ManifestPath, definition)
	if err != nil {
		t.Fatal(err)
	}
	view, err := packageruntime.Describe(manifest)
	if err != nil || view.ImageDigest != owner.Split.Result.OutputImageDigest {
		t.Fatal(view, err)
	}
	suite, total := jointCases(t, definition.Entry.Activity)
	result, err := workspaceexecution.ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Runtime
	if r.FinitePassed != total || r.FiniteTotal != total || !r.RuntimeReplayed || r.ModelCalls != 0 {
		t.Fatal("joint source outputs/replay differ from the independent oracle", r)
	}
}
