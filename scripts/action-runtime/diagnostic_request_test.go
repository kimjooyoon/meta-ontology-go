package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiagnosticRequestOwnsHistoricalSubject(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".github/workflows/function-extractor-diagnostic.yml"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	diagnostic := strings.SplitN(source, "  diagnose:\n", 2)
	if len(diagnostic) != 2 || !strings.HasPrefix(diagnostic[1], "    name: Diagnose function extractor process boundary\n    if: github.event_name == 'workflow_dispatch'\n") {
		t.Fatal("historical execution needs an explicit diagnostic request")
	}
	request, _, _ := strings.Cut(source, "permissions:")
	if strings.Contains(request, "default:") || strings.Count(request, "required: true") != 4 {
		t.Fatal("the caller must supply the complete subject/run/artifact/digest tuple")
	}
	for key, input := range map[string]string{"SUBJECT_SHA": "subject_sha", "SOURCE_RUN_ID": "source_run_id", "SOURCE_ARTIFACT_ID": "source_artifact_id", "SOURCE_ARTIFACT_DIGEST": "source_artifact_digest"} {
		if !strings.Contains(source, "  "+key+": ${{ inputs."+input+" }}\n") {
			t.Fatal("diagnostic subject has an implicit or unrelated source", key)
		}
	}
	if !strings.Contains(diagnostic[0], "    if: github.event_name == 'pull_request'\n") || !strings.Contains(diagnostic[0], "ref: ${{ github.event.pull_request.head.sha }}") {
		t.Fatal("implementation checks must use the current PR source")
	}
	if bodies := checkDiagnosticShellSyntax(t, source); bodies < 1 {
		t.Fatal("no diagnostic shell body was checked")
	}
}

func checkDiagnosticShellSyntax(t *testing.T, source string) int {
	t.Helper()
	lines := strings.Split(source, "\n")
	bodies := 0
	for i, line := range lines {
		if strings.TrimSpace(line) != "run: |" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		var body strings.Builder
		for _, next := range lines[i+1:] {
			if strings.TrimSpace(next) != "" && len(next)-len(strings.TrimLeft(next, " ")) <= indent {
				break
			}
			body.WriteString(strings.TrimPrefix(next, strings.Repeat(" ", indent+2)) + "\n")
		}
		command := exec.Command("bash", "-n")
		command.Stdin = strings.NewReader(body.String())
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("shell at workflow line %d: %v: %s", i+1, err, output)
		}
		bodies++
	}
	return bodies
}
