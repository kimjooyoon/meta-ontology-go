package externalecosystemexecution

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"sort"
)

type goEvent struct {
	Action     string `json:"Action"`
	Package    string `json:"Package"`
	Test       string `json:"Test"`
	ImportPath string `json:"ImportPath"`
}

func runGoTest(ctx context.Context, root string, index int) RunObservation {
	cmd := exec.CommandContext(ctx, "go", "test", "-json", "-count=1", "./...")
	cmd.Dir, cmd.Env = root, controlledEnv()
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = 1
		if e, ok := err.(*exec.ExitError); ok {
			exitCode = e.ExitCode()
		}
	}
	outcomes, unknown, events := parseEvents(output.Bytes())
	passed := exitCode == 0 && len(outcomes) > 0 && len(unknown) == 0 && !hasFailure(outcomes)
	return RunObservation{index, exitCode, passed, events, Digest(output.Bytes()), Digest(outcomes), outcomes, unknown}
}

func parseEvents(data []byte) ([]Outcome, []string, int) {
	final := map[string]Outcome{}
	unknown := map[string]bool{}
	known := map[string]bool{"start": true, "run": true, "pause": true, "cont": true, "pass": true,
		"bench": true, "fail": true, "output": true, "skip": true, "build-output": true, "build-fail": true}
	scanner, count := bufio.NewScanner(bytes.NewReader(data)), 0
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		count++
		var event goEvent
		if json.Unmarshal(scanner.Bytes(), &event) != nil {
			unknown["invalid-json"] = true
			continue
		}
		if !known[event.Action] {
			unknown[event.Action] = true
			continue
		}
		if event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
			if event.Package != "" {
				final[event.Package+"\x00"+event.Test] = Outcome{event.Package, event.Test, event.Action}
			}
		}
		if event.Action == "build-fail" {
			final[event.ImportPath+"\x00"] = Outcome{event.ImportPath, "", event.Action}
		}
	}
	if scanner.Err() != nil {
		unknown["scanner-error"] = true
	}
	outcomes := make([]Outcome, 0, len(final))
	for _, item := range final {
		outcomes = append(outcomes, item)
	}
	sort.Slice(outcomes, func(i, j int) bool {
		if outcomes[i].Package == outcomes[j].Package {
			return outcomes[i].Test < outcomes[j].Test
		}
		return outcomes[i].Package < outcomes[j].Package
	})
	unknowns := make([]string, 0, len(unknown))
	for item := range unknown {
		unknowns = append(unknowns, item)
	}
	sort.Strings(unknowns)
	return outcomes, unknowns, count
}

func hasFailure(outcomes []Outcome) bool {
	for _, item := range outcomes {
		if item.Action == "fail" || item.Action == "build-fail" {
			return true
		}
	}
	return false
}
