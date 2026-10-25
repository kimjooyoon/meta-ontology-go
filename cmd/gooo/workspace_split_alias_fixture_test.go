package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func saveSplitAliasProposal(t *testing.T, options workspaceSourceSplitOptions, inputs map[string][]byte, source string) {
	t.Helper()
	definition, _, err := loadWorkspaceRenameManifest(OSFileReader{}, options.manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	definition.Packages[1].Sources[0] = source
	raw, err := workspaceinput.Encode(definition)
	if err != nil {
		t.Fatal(err)
	}
	inputs["workspace.json"] = raw
	if err := os.WriteFile(options.manifestPath, raw, 0o640); err != nil {
		t.Fatal(err)
	}
	proposal, err := prepareWorkspaceSourceSplit(workspaceSourceSplitOptions{
		packagePath: "named/inputs", source: "main.gooo", before: "urn:gooo:example:named-inputs:net",
		into: "net.gooo", manifestPath: options.manifestPath}, OSFileReader{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(sourceSplitPublicReceipt(proposal))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(options.applyReceipt, raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

func createSplitAlias(t *testing.T, root, mode string, original []byte) string {
	t.Helper()
	source := "alias.gooo"
	var err error
	switch mode {
	case "directory":
		source = "alias/main.gooo"
		err = os.Symlink(root, filepath.Join(root, "alias"))
		if err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	case "case":
		source = "MAIN.gooo"
		if _, err := os.Stat(filepath.Join(root, source)); os.IsNotExist(err) {
			t.Skip("case-sensitive filesystem")
		}
	case "hardlink":
		err = os.Link(filepath.Join(root, "main.gooo"), filepath.Join(root, source))
		if err != nil {
			t.Skipf("hard links unavailable: %v", err)
		}
	case "copy":
		err = os.WriteFile(filepath.Join(root, source), original, 0o640)
	default:
		t.Fatal("unknown alias fixture", mode)
	}
	if err != nil {
		t.Fatal(err)
	}
	return source
}
