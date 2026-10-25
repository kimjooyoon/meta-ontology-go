package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func sharedApplicationFixture(t *testing.T) (workspaceSourceSplitOptions, workspaceSourceSplitReceipt, map[string][]byte) {
	t.Helper()
	options, _, inputs := sourceSplitApplicationFixture(t)
	definition, _, err := loadWorkspaceRenameManifest(OSFileReader{}, options.manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	other := definition.Packages[0]
	other.Path = "named/other"
	definition.Packages = append(definition.Packages, other)
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
	return options, proposal, inputs
}

func TestSharedWorkspaceSplitAppliesCompleteScopeAndRejectsOldPartialScope(t *testing.T) {
	for _, mode := range []string{"shared", "legacy partial"} {
		t.Run(mode, func(t *testing.T) {
			options, proposal, inputs := sharedApplicationFixture(t)
			if mode == "legacy partial" {
				proposal = legacySharedProposal(t, options, proposal)
			}
			raw, err := json.Marshal(sourceSplitPublicReceipt(proposal))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(options.applyReceipt, raw, 0o600); err != nil {
				t.Fatal(err)
			}
			applied, err := applyWorkspaceSourceSplit(options, defaultAtomicFileOps())
			if mode == "legacy partial" {
				if err == nil || applied.Decision == "APPLIED" {
					t.Fatal("old incomplete scope was applied")
				}
				checkSourceSplitApplicationInputs(t, options, inputs)
				return
			}
			if err != nil || applied.Decision != "APPLIED" || len(applied.Result.AffectedPackages) != 2 {
				t.Fatal("shared scope was not installed", err, applied)
			}
			if err := verifyAppliedSourceSplit(applied); err != nil {
				t.Fatal("installed shared image mismatch", err)
			}
		})
	}
}

func legacySharedProposal(t *testing.T, options workspaceSourceSplitOptions, proposal workspaceSourceSplitReceipt) workspaceSourceSplitReceipt {
	t.Helper()
	definition, _, err := loadWorkspaceRenameManifest(OSFileReader{}, options.manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := loadPackageSources(OSFileReader{}, options.manifestPath, definition)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := packageruntime.SplitSource(loaded, "named/inputs", "main.gooo", "urn:gooo:example:named-inputs:net", "net.gooo")
	if err != nil {
		t.Fatal(err)
	}
	proposal.Result = &legacy
	proposal.UpdatedWorkspace.Packages[1].Sources = definition.Packages[1].Sources
	return proposal
}
