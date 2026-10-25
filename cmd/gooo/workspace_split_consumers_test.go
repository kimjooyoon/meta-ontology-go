package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func sharedConsumerApplicationFixture(t *testing.T) (workspaceSourceSplitOptions, workspaceSourceSplitReceipt, map[string][]byte) {
	t.Helper()
	options, original, inputs := sourceSplitApplicationFixture(t)
	root := filepath.Dir(options.manifestPath)
	definition, err := decodeWorkspaceManifest(inputs["workspace.json"])
	if err != nil {
		t.Fatal(err)
	}
	definition.Entry.Activity = "Credit"
	definition.Packages[0].Sources = append(definition.Packages[0].Sources, "other.gooo")
	inputs["other.gooo"] = []byte("package namedinputs\nnamespace namedinputs\n" +
		"activity Other(Integer) -> Integer id \"urn:split:other\" computes \"return input + 1\"\n")
	inputs["second.workspace.json"], err = json.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"other.gooo", "second.workspace.json"} {
		if err := os.WriteFile(filepath.Join(root, name), inputs[name], 0o640); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{"package", "split-source", "--package", original.Result.PackagePath,
		"--source", original.Result.InputFilename, "--before", original.Result.BeforeID, "--into", "net.gooo",
		"--consumer-workspace", filepath.Join(root, "second.workspace.json"), "--json", options.manifestPath}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal("joint proposal failed", code, stdout.String(), stderr.String())
	}
	var proposal workspaceSourceSplitReceipt
	if err := json.Unmarshal(stdout.Bytes(), &proposal); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(options.applyReceipt, stdout.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	checkSourceSplitApplicationInputs(t, options, inputs)
	return options, proposal, inputs
}

func TestSharedConsumerSplitAppliesAndExecutesBothDeclaredWorkspaces(t *testing.T) {
	options, proposal, _ := sharedConsumerApplicationFixture(t)
	if proposal.Decision != "PROPOSED" || len(proposal.Consumers) != 1 ||
		len(proposal.Consumers[0].Result.Manifest.Packages) != 0 {
		t.Fatal("proposal did not retain a compact independent consumer image", proposal)
	}
	var stdout, stderr bytes.Buffer
	args := []string{"package", "split-source", "--apply", options.applyReceipt, "--json", options.manifestPath}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal("joint application failed", code, stdout.String(), stderr.String())
	}
	var applied workspaceSourceSplitReceipt
	if err := json.Unmarshal(stdout.Bytes(), &applied); err != nil || applied.Decision != "APPLIED" ||
		len(applied.Consumers) != 1 || applied.Application.State != "OUTPUT_VERIFIED" {
		t.Fatal("incomplete joint application", err, applied)
	}
	if err := verifyAppliedSourceSplit(applied); err != nil {
		t.Fatal(err)
	}
	root := filepath.Dir(options.manifestPath)
	creditCases := `{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"named/inputs:Credit":101},"expected":{"named/inputs:Credit":124}},
 {"inputs":{"named/inputs:Credit":-53},"expected":{"named/inputs:Credit":-30}}]}`
	if err := os.WriteFile(filepath.Join(root, "credit-cases.json"), []byte(creditCases), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		manifest, cases string
		outputs         int
	}{{"workspace.json", "cases.json", 4}, {"second.workspace.json", "credit-cases.json", 2}} {
		stdout.Reset()
		if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, check.cases),
			filepath.Join(root, check.manifest)}, &stdout, &stderr); code != exitOK {
			t.Fatal("installed owner failed execution", check.manifest, code, stdout.String(), stderr.String())
		}
		var execution packageExecutionReceipt
		if err := json.Unmarshal(stdout.Bytes(), &execution); err != nil || execution.Result == nil ||
			execution.Result.Runtime.FinitePassed != check.outputs || execution.Result.Runtime.FiniteTotal != check.outputs ||
			!execution.Result.Runtime.RuntimeReplayed || execution.Result.Runtime.ModelCalls != 0 {
			t.Fatal("independent owner output/replay failed", check.manifest, err)
		}
	}
	stdout.Reset()
	if code := run(args, &stdout, &stderr); code != exitFailure {
		t.Fatal("joint proposal applied twice")
	}
	if err := verifyAppliedSourceSplit(applied); err != nil {
		t.Fatal("rejected replay damaged an owner", err)
	}
}
