package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func workspaceReviewRenameFiles() map[string]string {
	return map[string]string{
		"rules.gooo": `package rules
namespace review_rules
entity Integer id "urn:review:integer"
entity Boolean id "urn:review:boolean"
activity Review(Integer,Integer) -> Boolean id "urn:review:rule" computes "return input1 == 12-input0"
activity Local(Integer,Integer) -> Boolean computes "return Review(input0,input1)"
`,
		"peer.gooo": "package peer\nnamespace review_peer\nentity Integer id \"urn:review:integer\"\nentity Boolean id \"urn:review:boolean\"\nactivity Review(Integer,Integer) -> Boolean id \"urn:review:peer\" computes `return input1 == input0+12`\n",
		"app.gooo": `package app
namespace review_app
import rules "review/rules"
activity Main(Integer) -> Integer computes "let text=\"Review()\"; if text==\"\" { return input }; if rules.\u0052eview(input,12-input) { return Draft(input) }; return 0"
activity Draft(Integer) -> Integer computes "return input-12" assembling {
 choice "direction" operand_order at "0" intent "deduct input from twelve"
 criteria { predicate "reference" activity "rules.Review" input "[5]" holdout_input "[2]" }
 attempts "2"
}
activity Accept(Boolean) -> Boolean computes "return input"
bind rules.Review.result -> Accept.input
`,
		"unused.gooo": "package app\nnamespace review_app\nimport rules \"review/rules\"\nactivity Unused(Integer) -> Boolean computes `return rules.Review(input,0)`\n",
		"other.gooo":  "package app\nnamespace review_app\nimport rules \"review/peer\"\nactivity Peer(Integer) -> Boolean computes `return rules.Review(input,0)`\n",
	}
}

func writeWorkspaceReviewRenameFixture(t *testing.T) (string, map[string]string) {
	t.Helper()
	root := t.TempDir()
	sources := workspaceReviewRenameFiles()
	manifest := workspaceManifest{Schema: workspaceSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "review/app", Activity: "Main"},
		Packages: []workspacePackage{
			{Path: "review/rules", Name: "rules", Sources: []string{"rules.gooo"}},
			{Path: "review/peer", Name: "peer", Sources: []string{"peer.gooo"}},
			{Path: "review/app", Name: "app", Imports: []string{"review/rules", "review/peer"}, Sources: []string{"app.gooo", "unused.gooo", "other.gooo"}},
		}}
	raw, _ := json.Marshal(manifest)
	sources["workspace.json"] = string(raw)
	for name, source := range sources {
		if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(root, "workspace.json"), sources
}

func TestWorkspaceRenameCLIIncludesDormantFixedCriteriaAndBinding(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	path, sources := writeWorkspaceReviewRenameFixture(t)
	var stdout, stderr bytes.Buffer
	args := []string{"package", "rename-activity", "--package", "review/rules", "--id", "urn:review:rule", "--to", "검토", "--json", path}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String(), stdout.String())
	}
	var receipt workspaceActivityRenameReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "PROPOSED" || receipt.Result == nil || receipt.UpdatedWorkspace == nil || receipt.ManifestDigest == "" {
		t.Fatal("CLI did not return the complete source proposal")
	}
	checkWorkspaceReviewRenameReceipt(t, receipt, sources)
	for name, original := range sources {
		raw, err := os.ReadFile(filepath.Join(filepath.Dir(path), name))
		if err != nil || string(raw) != original {
			t.Fatal("proposal changed source or workspace on disk", name, err)
		}
	}
	r := receipt.Result
	raw, _ := json.Marshal(map[string]any{"activity_id": r.ActivityID, "references": r.UpdatedReferences,
		"reference_total": r.ResolvedReferences, "changed_files": len(r.Files), "source_files": 5, "packages": 3,
		"entry_updated": r.EntryUpdated, "identity_pinned": r.IdentityPinned, "dormant_updated": true,
		"criterion_updated": true, "other_package_preserved": true, "ordinary_literal_preserved": true,
		"program_executions": r.ProgramExecutions, "model_predictions": r.ModelPredictions})
	t.Log("WORKSPACE_RENAME_CLI_OBSERVATION", string(raw))
}

func checkWorkspaceReviewRenameReceipt(t *testing.T, receipt workspaceActivityRenameReceipt, sources map[string]string) {
	t.Helper()
	r := receipt.Result
	if r.ActivityID != "urn:review:rule" || r.IdentityPinned || r.EntryUpdated ||
		r.UpdatedReferences != 5 || r.ResolvedReferences != 5 || len(r.Files) != 3 ||
		r.ProgramExecutions != 0 || r.ModelPredictions != 0 || receipt.UpdatedWorkspace.Entry.Activity != "Main" {
		t.Fatal("target, entry or reference counts changed")
	}
	app := r.Manifest.Packages[2].Sources[0].Content
	unused := r.Manifest.Packages[2].Sources[1].Content
	if !strings.Contains(app, "rules.검토(input,12-input)") || !strings.Contains(app, `activity "rules.검토"`) ||
		!strings.Contains(app, "bind rules.검토.result -> Accept.input") || !strings.Contains(app, `\"Review()\"`) ||
		!strings.Contains(unused, "rules.검토(input,0)") ||
		r.Manifest.Packages[1].Sources[0].Content != sources["peer.gooo"] || r.Manifest.Packages[2].Sources[2].Content != sources["other.gooo"] {
		t.Fatal("classified reference or same-name peer locality changed")
	}
}
