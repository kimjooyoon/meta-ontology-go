package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSharedConsumerSplitUsesEachManifestRelativeAddress(t *testing.T) {
	options, old, inputs := sourceSplitApplicationFixture(t)
	root := filepath.Dir(options.manifestPath)
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"main.gooo", "peer.gooo", "workspace.json"} {
		if err := os.Rename(filepath.Join(root, name), filepath.Join(nested, name)); err != nil {
			t.Fatal(err)
		}
	}
	definition, err := decodeWorkspaceManifest(inputs["workspace.json"])
	if err != nil {
		t.Fatal(err)
	}
	for i := range definition.Packages[0].Sources {
		definition.Packages[0].Sources[i] = "nested/" + definition.Packages[0].Sources[i]
	}
	raw, _ := json.Marshal(definition)
	if err := os.WriteFile(options.manifestPath, raw, 0o640); err != nil {
		t.Fatal(err)
	}
	proposal, err := prepareWorkspaceSourceSplit(workspaceSourceSplitOptions{
		packagePath: old.Result.PackagePath, source: "nested/main.gooo", into: "nested/net.gooo",
		before: old.Result.BeforeID, manifestPath: options.manifestPath,
		consumerWorkspaces: []string{filepath.Join(nested, "workspace.json")}}, OSFileReader{})
	if err != nil || len(proposal.Consumers) != 1 || proposal.Consumers[0].Result.InputFilename != "main.gooo" ||
		proposal.Consumers[0].Result.Files[1].Filename != "net.gooo" {
		t.Fatal("consumer address was copied from the primary instead of resolved", err, proposal)
	}
	raw, _ = json.Marshal(sourceSplitPublicReceipt(proposal))
	if err := os.WriteFile(options.applyReceipt, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	applied, err := applyWorkspaceSourceSplit(options, defaultAtomicFileOps())
	if err != nil || applied.Decision != "APPLIED" {
		t.Fatal("same physical publication did not accept distinct manifest-relative addresses", err)
	}
	if err := verifyAppliedSourceSplit(applied); err != nil {
		t.Fatal(err)
	}
}

func TestSharedConsumerSplitRefusesUnrelatedOwnerWithoutPartialProposal(t *testing.T) {
	options, old, inputs := sharedConsumerApplicationFixture(t)
	root := filepath.Dir(options.manifestPath)
	definition, err := decodeWorkspaceManifest(inputs["second.workspace.json"])
	if err != nil {
		t.Fatal(err)
	}
	definition.Packages[0].Sources[0] = "copy.gooo"
	raw, _ := json.Marshal(definition)
	inputs["second.workspace.json"], inputs["copy.gooo"] = raw, inputs["main.gooo"]
	for _, name := range []string{"second.workspace.json", "copy.gooo"} {
		if err := os.WriteFile(filepath.Join(root, name), inputs[name], 0o640); err != nil {
			t.Fatal(err)
		}
	}
	proposal, err := prepareWorkspaceSourceSplit(workspaceSourceSplitOptions{
		packagePath: old.Result.PackagePath, source: "main.gooo", into: "net.gooo", before: old.Result.BeforeID,
		manifestPath: options.manifestPath, consumerWorkspaces: []string{filepath.Join(root, "second.workspace.json")}},
		OSFileReader{})
	if err == nil || proposal.Decision != "FAIL_CLOSED" || proposal.Result != nil || len(proposal.Consumers) != 0 {
		t.Fatal("same bytes at another source address became a consumer or leaked partial scope", err, proposal)
	}
	checkSourceSplitApplicationInputs(t, options, inputs)
}
