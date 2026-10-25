package workspaceinput_test

import (
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func reloadNativeSplitFixture(t *testing.T, files fstest.MapFS, definition workspaceinput.Definition) packageruntime.Manifest {
	t.Helper()
	proposal, err := workspaceinput.PrepareSplit(files, "workspace.json", definition, workspaceinput.SplitRequest{
		PackagePath: "shared/app", Source: "app.gooo", BeforeID: "urn:shared:run", Into: "run.gooo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(definition.Packages[1].Sources) != 1 || proposal.Definition.Packages[1].Name != "" ||
		proposal.Definition.Packages[1].Imports != nil {
		t.Fatal("native source plan changed the authored definition")
	}
	for _, file := range proposal.Result.Files {
		files[file.Filename] = &fstest.MapFile{Data: []byte(file.Source)}
	}
	afterBytes, err := workspaceinput.Encode(proposal.Definition)
	if err != nil {
		t.Fatal(err)
	}
	afterDefinition, err := workspaceinput.Decode(afterBytes)
	if err != nil {
		t.Fatal(err)
	}
	after, err := workspaceinput.Load(files, "workspace.json", afterDefinition)
	if err != nil {
		t.Fatal(err)
	}
	view, err := packageruntime.Describe(after)
	if err != nil || view.ImageDigest != proposal.Result.OutputImageDigest {
		t.Fatal("authored output did not reconstruct the planned image", view.ImageDigest, err)
	}
	return after
}
