package externalecosystemexecution

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strings"
)

func runGoTest(ctx context.Context, root string, index int) RunObservation {
	cmd := exec.CommandContext(ctx, "go", "test", "-json", "-count=1", "./...")
	cmd.Dir, cmd.Env = root, controlledEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = 1
		if e, ok := err.(*exec.ExitError); ok {
			exitCode = e.ExitCode()
		}
	}
	outcomes, unknown, diagnostics, events := parseEvents(stdout.Bytes())
	passed := exitCode == 0 && len(outcomes) > 0 && len(unknown) == 0 && !hasFailure(outcomes)
	return RunObservation{
		Index: index, ExitCode: exitCode, Passed: passed, EventCount: events,
		RawSHA256: Digest(stdout.Bytes()), StderrSHA256: Digest(stderr.Bytes()),
		StderrLineCount:  bytes.Count(stderr.Bytes(), []byte{'\n'}),
		NormalizedSHA256: Digest(outcomes), Outcomes: outcomes,
		UnknownEvents: unknown, Diagnostics: diagnostics,
	}
}

func parseEvents(data []byte) ([]Outcome, []string, []string, int) {
	final := map[string]Outcome{}
	unknown := map[string]bool{}
	diagnostics := make([]string, 0, 64)

	scanner, count := bufio.NewScanner(bytes.NewReader(data)), 0
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		count++
		var event goEvent
		if json.Unmarshal(scanner.Bytes(), &event) != nil {
			unknown["invalid-json"] = true
			continue
		}
		if !knownEventActions[event.Action] {
			unknown[event.Action] = true
			continue
		}
		if len(diagnostics) < 64 && diagnosticEvent(event) {
			diagnostics = append(diagnostics, strings.TrimSpace(event.Output))
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
	outcomes, unknowns := normalizedEventResults(final, unknown)
	return outcomes, unknowns, diagnostics, count
}
