package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func chosenRegisteredSplit(t *testing.T, limit int) (workspaceSourceSplitOptions, workspaceSourceSplitReceipt, []byte) {
	t.Helper()
	options := registeredSplitFixture(t)
	var stdout, stderr bytes.Buffer
	args := []string{"package", "split-source", "--package", options.packagePath, "--source", options.source,
		"--line-limit", strconv.Itoa(limit), "--into", options.into, "--syntax-registry-root", options.registryRoot,
		"--discover-consumers", options.registryRoot, "--json", options.manifestPath}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	var receipt workspaceSourceSplitReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	return options, receipt, bytes.Clone(stdout.Bytes())
}

func TestWorkspaceChosenSplitPublishesAndExecutes(t *testing.T) {
	for _, limit := range []int{75, 30} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			options, receipt, raw := chosenRegisteredSplit(t, limit)
			selection := receipt.Selection
			if receipt.Decision != "PROPOSED" || selection == nil || selection.State != "SELECTED" ||
				selection.Candidates != 5 || selection.Boundary.BeforeID != options.before ||
				selection.Boundary.Left.Lines != 44 || selection.Boundary.Right.Lines != 41 ||
				selection.LimitSatisfied != (limit >= 44) || receipt.Registration == nil || receipt.Discovery == nil {
				t.Fatal("selection lost physical result, remaining pressure, or consumers", selection)
			}
			name := filepath.Join(options.registryRoot, "proposal.json")
			if err := os.WriteFile(name, raw, 0600); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			args := []string{"package", "split-source", "--apply", name, "--json", options.manifestPath}
			if code := run(args, &stdout, &stderr); code != exitOK {
				t.Fatal(code, stdout.String(), stderr.String())
			}
			var applied workspaceSourceSplitReceipt
			if err := json.Unmarshal(stdout.Bytes(), &applied); err != nil || applied.Decision != "APPLIED" ||
				applied.Selection == nil || applied.Selection.LimitSatisfied != (limit >= 44) {
				t.Fatal("application lost selection or remaining pressure", err)
			}
			assertChosenSplitRuntime(t, options)
		})
	}
}

func assertChosenSplitRuntime(t *testing.T, options workspaceSourceSplitOptions) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cases := filepath.Join(filepath.Dir(options.manifestPath), "cases.json")
	if code := run([]string{"package", "execute", "--json", "--cases", cases, options.manifestPath}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	var executed packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &executed); err != nil || executed.Result == nil ||
		executed.Result.Runtime.FinitePassed != 8 || executed.Result.Runtime.FiniteTotal != 8 ||
		!executed.Result.Runtime.RuntimeReplayed || executed.Result.Runtime.ModelCalls != 0 {
		t.Fatal("independent eight-case runtime oracle failed", err)
	}
}

func TestWorkspaceSplitChoiceModesAreExclusive(t *testing.T) {
	base := []string{"--package", "p", "--source", "main.gooo", "--into", "other.gooo", "--line-limit", "75"}
	for _, extra := range [][]string{{"--line-limit", "80"}, {"--before", "urn:manual"}, {"--boundaries"}, {"--apply", "plan.json"}} {
		args := append(append(append([]string{}, base...), extra...), "workspace.json")
		if _, err := parseWorkspaceSourceSplit(args, io.Discard); err == nil {
			t.Fatal("ambiguous selection mode accepted", extra)
		}
	}
	for _, limit := range []string{"0", "-1", "NaN", "9223372036854775808"} {
		if _, err := parseWorkspaceSourceSplit([]string{"--line-limit", limit, "workspace.json"}, io.Discard); err == nil {
			t.Fatal("invalid limit accepted", limit)
		}
	}
}
