package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func discoveredSplitFixture(t *testing.T) (workspaceSourceSplitOptions, workspaceSourceSplitReceipt, map[string][]byte) {
	t.Helper()
	options, original, inputs := sharedConsumerApplicationFixture(t)
	root := filepath.Dir(options.manifestPath)
	args := []string{"package", "split-source", "--package", original.Result.PackagePath,
		"--source", original.Result.InputFilename, "--before", original.Result.BeforeID, "--into", "net.gooo",
		"--discover-consumers", root, "--json", options.manifestPath}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal("discovered proposal failed", code, stdout.String(), stderr.String())
	}
	proposal, err := decodeWorkspaceSourceSplitProposal(stdout.Bytes())
	if err != nil || proposal.Discovery == nil || len(proposal.Discovery.Definitions) != 2 || len(proposal.Consumers) != 1 {
		t.Fatal("discovery scope missing from proposal", err, proposal)
	}
	if err := os.WriteFile(options.applyReceipt, stdout.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	checkSourceSplitApplicationInputs(t, options, inputs)
	return options, proposal, inputs
}

func TestDiscoveredSplitAppliesThroughTheSavedConsumerScope(t *testing.T) {
	options, _, _ := discoveredSplitFixture(t)
	applied, err := applyWorkspaceSourceSplit(options, defaultAtomicFileOps())
	if err != nil || applied.Application.State != "OUTPUT_VERIFIED" || len(applied.Consumers) != 1 {
		t.Fatal("discovered source group was not applied", err, applied)
	}
	var stdout, stderr bytes.Buffer
	root := filepath.Dir(options.manifestPath)
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"),
		options.manifestPath}, &stdout, &stderr); code != exitOK {
		t.Fatal("discovered split changed execution", code, stdout.String(), stderr.String())
	}
	var execution packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &execution); err != nil || execution.Result == nil ||
		execution.Result.Runtime.FinitePassed != 4 || execution.Result.Runtime.FiniteTotal != 4 ||
		!execution.Result.Runtime.RuntimeReplayed {
		t.Fatal("independent output cases did not pass", err)
	}
	if err := verifyAppliedSourceSplit(applied); err != nil {
		t.Fatal("another discovered owner lost meaning", err)
	}
}

func TestDiscoveredSplitRechecksNewConsumersBeforeAnyPublication(t *testing.T) {
	for _, atPublication := range []bool{false, true} {
		options, _, inputs := discoveredSplitFixture(t)
		root := filepath.Dir(options.manifestPath)
		addConsumer := func() error {
			inputs["new-consumer.json"] = inputs["workspace.json"]
			return os.WriteFile(filepath.Join(root, "new-consumer.json"), inputs["new-consumer.json"], 0o600)
		}
		ops, renames := defaultAtomicFileOps(), 0
		ops.rename = func(from, to string) error { renames++; return os.Rename(from, to) }
		if atPublication {
			ops.beforePublish = addConsumer
		} else if err := addConsumer(); err != nil {
			t.Fatal(err)
		}
		applied, err := applyWorkspaceSourceSplit(options, ops)
		if err == nil || applied.Decision != "FAIL_CLOSED" || renames != 0 {
			t.Fatal("a newly discovered consumer was omitted", atPublication, err, applied)
		}
		checkSourceSplitApplicationInputs(t, options, inputs)
	}
}

func TestDiscoveredSplitRejectsMissingOrForgedScope(t *testing.T) {
	for _, mutation := range []string{"missing-definition", "wrong-candidates", "removed-owner", "broken-json"} {
		options, proposal, inputs := discoveredSplitFixture(t)
		switch mutation {
		case "missing-definition":
			proposal.Discovery.Definitions = proposal.Discovery.Definitions[:1]
		case "wrong-candidates":
			proposal.Discovery.Candidates = "every-reader"
		case "removed-owner":
			proposal.Consumers = nil
		case "broken-json":
			if err := os.WriteFile(filepath.Join(filepath.Dir(options.manifestPath), "broken.json"), []byte("{"), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		raw, _ := json.Marshal(proposal)
		if err := os.WriteFile(options.applyReceipt, raw, 0o600); err != nil {
			t.Fatal(err)
		}
		if applied, err := applyWorkspaceSourceSplit(options, defaultAtomicFileOps()); err == nil || applied.Decision == "APPLIED" {
			t.Fatal("invalid discovery accepted", mutation, err)
		}
		checkSourceSplitApplicationInputs(t, options, inputs)
	}
}

func TestDiscoveredSplitNormalizesTheSelectedScopeParentAlias(t *testing.T) {
	options, original, _ := sharedConsumerApplicationFixture(t)
	alias := filepath.Join(t.TempDir(), "project")
	if err := os.Symlink(filepath.Dir(options.manifestPath), alias); err != nil {
		t.Fatal(err)
	}
	proposal, err := prepareWorkspaceSourceSplit(workspaceSourceSplitOptions{
		packagePath: original.Result.PackagePath, source: original.Result.InputFilename,
		before: original.Result.BeforeID, into: "net.gooo", manifestPath: filepath.Join(alias, "workspace.json"),
		discoveryRoot: alias}, OSFileReader{})
	if err != nil || proposal.Discovery == nil || len(proposal.Consumers) != 1 {
		t.Fatal("parent alias split one source into different lexical roots", err, proposal)
	}
	raw, err := json.Marshal(sourceSplitPublicReceipt(proposal))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(options.applyReceipt, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	options.manifestPath = filepath.Join(alias, "workspace.json")
	if applied, err := applyWorkspaceSourceSplit(options, defaultAtomicFileOps()); err != nil || applied.Decision != "APPLIED" {
		t.Fatal("saved aliased scope did not replay", err)
	}
}

func TestDiscoveredSplitRequiresOneUnambiguousProposalScope(t *testing.T) {
	for _, args := range [][]string{
		{"--discover-consumers", ""},
		{"--discover-consumers", ".", "--discover-consumers", "nested"},
		{"--discover-consumers", ".", "--consumer-workspace", "other.json"},
		{"--discover-consumers", ".", "--apply", "proposal.json"},
	} {
		args = append(args, "--package", "example", "--source", "main.gooo", "--before", "urn:example",
			"--into", "part.gooo", "workspace.json")
		if _, err := parseWorkspaceSourceSplit(args, io.Discard); err == nil {
			t.Fatal("ambiguous discovery scope accepted", args)
		}
	}
}
