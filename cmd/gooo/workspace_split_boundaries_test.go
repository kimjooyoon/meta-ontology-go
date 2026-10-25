package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestWorkspaceSplitBoundaryReachesProposalApplyAndExecution(t *testing.T) {
	root := t.TempDir()
	inputs := map[string][]byte{}
	for _, name := range []string{"main.gooo", "gooo.workspace.json", "cases.json"} {
		raw, err := os.ReadFile(filepath.Join("../../examples/agent-tool-inventory", name))
		if err != nil {
			t.Fatal(err)
		}
		inputs[name] = raw
		if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	manifest := filepath.Join(root, "gooo.workspace.json")
	base := []string{"package", "split-source", "--package", "example/agent-tool-inventory", "--source", "./main.gooo"}
	var stdout, stderr bytes.Buffer
	if code := run(append(append([]string{}, base...), "--boundaries", "--json", manifest), &stdout, &stderr); code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	var observed workspaceSourceSplitReceipt
	if err := json.Unmarshal(stdout.Bytes(), &observed); err != nil || observed.Boundaries == nil ||
		observed.Decision != "OBSERVED" || observed.Boundaries.Input.Lines != 82 || len(observed.Boundaries.Candidates) != 5 {
		t.Fatal("incomplete original source observation", observed, err)
	}
	chosen := observed.Boundaries.Candidates[4]
	if chosen.BeforeID != "urn:gooo:agent-tool-inventory:inspect" || chosen.Left.Lines != 44 || chosen.Right.Lines != 41 {
		t.Fatal("physical sizes or declaration order changed", chosen)
	}
	proposalPath := filepath.Join(root, "proposal.json")
	if err := os.WriteFile(proposalPath, stdout.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"package", "split-source", "--apply", proposalPath, manifest}, io.Discard, io.Discard); code != exitFailure {
		t.Fatal("observation was treated as a writable proposal", code)
	}
	for name, want := range inputs {
		got, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || !bytes.Equal(want, got) {
			t.Fatal("inspection or refused apply changed input", name, err)
		}
	}
	stdout.Reset()
	args := append(append([]string{}, base...), "--before", chosen.BeforeID, "--into", "inspect.gooo",
		"--discover-consumers", root, "--json", manifest)
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal("selected boundary did not prepare through scoped consumers", code, stdout.String(), stderr.String())
	}
	if err := os.WriteFile(proposalPath, stdout.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"package", "split-source", "--apply", proposalPath, manifest}, io.Discard, &stderr); code != exitOK {
		t.Fatal("selected proposal did not apply", code, stderr.String())
	}
	stdout.Reset()
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), manifest},
		&stdout, &stderr); code != exitOK {
		t.Fatal("partitioned program failed its independent cases", code, stdout.String(), stderr.String())
	}
	var execution packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &execution); err != nil || execution.Result == nil ||
		execution.Result.Runtime.FinitePassed != 8 || execution.Result.Runtime.FiniteTotal != 8 ||
		!execution.Result.Runtime.RuntimeReplayed {
		t.Fatal("partitioned program lost expected output or replay", err)
	}
}

func TestWorkspaceSplitBoundaryModeCannotSmuggleProposalOptions(t *testing.T) {
	base := []string{"--boundaries", "--package", "p", "--source", "main.gooo"}
	for _, extra := range [][]string{{"--before", "urn:id"}, {"--into", "new.gooo"}, {"--apply", "receipt.json"},
		{"--consumer-workspace", "other.json"}, {"--discover-consumers", "."}} {
		args := append(append(append([]string{}, base...), extra...), "workspace.json")
		if _, err := parseWorkspaceSourceSplit(args, io.Discard); err == nil {
			t.Fatal("mixed observation and proposal scope accepted", extra)
		}
	}
	_, proposal, _ := sharedConsumerApplicationFixture(t)
	proposal.Boundaries = &packageruntime.SourceSplitBoundaries{}
	raw, _ := json.Marshal(proposal)
	if _, err := decodeWorkspaceSourceSplitProposal(raw); err == nil {
		t.Fatal("proposal carrying an unverified candidate observation was accepted")
	}
}
