package verify

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestActualSharedWorkShell124073(t *testing.T) {
	tool := os.Getenv("GOOO_CI_DOMAIN")
	if tool == "" {
		t.Skip("actual work-domain producer unavailable outside CI")
	}
	root := t.TempDir()
	write := func(name, value string) {
		if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module work124073.example/program\n\ngo 1.27.3\n")
	write("source.go", "package work124073\nconst Value=191\n")
	write("intent.txt", "source-owned work124073\n")
	git := func(args ...string) string {
		command := exec.Command("git", args...)
		command.Dir = root
		command.Env = append(os.Environ(), "GIT_DIR=", "GIT_WORK_TREE=", "GIT_INDEX_FILE=", "GIT_CONFIG_COUNT=2", "GIT_CONFIG_KEY_0=user.name", "GIT_CONFIG_VALUE_0=work124073", "GIT_CONFIG_KEY_1=user.email", "GIT_CONFIG_VALUE_1=work124073@example.invalid")
		// Empty Git control variables are removed, rather than interpreted as paths.
		command.Env = workEnvironment124073(command.Env)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatal(string(output), err)
		}
		return string(bytes.TrimSpace(output))
	}
	git("init", "-q")
	git("add", ".")
	git("commit", "-qm", "new actual command input124073")
	head := git("rev-parse", "HEAD")
	body := filepath.Join(t.TempDir(), "step.sh")
	if err := os.WriteFile(body, []byte("cat intent.txt\nprintf 'answer=%s\\n' 191 >> \"$GITHUB_OUTPUT\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	outputFile := filepath.Join(t.TempDir(), "step-output")
	origins := t.TempDir()
	adapter, err := filepath.Abs("../../scripts/ci-work-shell.sh")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("/bin/bash", adapter, body)
	command.Dir = root
	command.Env = append(workEnvironment124073(os.Environ()),
		"GOOO_CI_DOMAIN="+tool, "GOOO_CI_STARTED_AT="+time.Now().UTC().Format(time.RFC3339Nano),
		"GOOO_CI_SOURCE_HEAD="+head, "GOOO_CI_ORIGIN_ROOT="+origins,
		"GOOO_CI_ACTIVATION_ORIGIN=", "GOOO_CI_BUDGET_MS=600000", "GITHUB_JOB=fresh124073", "GITHUB_ACTION=fresh_step124073",
		"GITHUB_RUN_ID=124073", "GITHUB_RUN_ATTEMPT=2",
		"GITHUB_EVENT_NAME=pull_request", "GITHUB_BASE_REF=dev", "GITHUB_REF_NAME=fresh124073", "GITHUB_OUTPUT="+outputFile)
	logs, err := command.CombinedOutput()
	if err != nil {
		t.Fatal(string(logs), err)
	}
	receipt, err := os.ReadFile(filepath.Join(origins, "fresh_step124073.json"))
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Decision                   string
		ScopeElapsedMS, DeadlineMS int64
		Spec                       struct{ ExecutionID string }
	}
	if err = json.Unmarshal(receipt, &result); err != nil || result.Decision != "COMMAND_COMPLETED" ||
		result.ScopeElapsedMS >= result.DeadlineMS || result.Spec.ExecutionID != "124073/2" {
		t.Fatal(string(receipt), err)
	}
	if string(logs) != "source-owned work124073\n" {
		t.Fatal(string(logs))
	}
	actual, err := os.ReadFile(outputFile)
	if err != nil || string(actual) != "answer=191\n" {
		t.Fatal(string(actual), err)
	}
}

func workEnvironment124073(parent []string) []string {
	result := []string{}
	for _, entry := range parent {
		if strings.HasPrefix(entry, "GIT_DIR=") || strings.HasPrefix(entry, "GIT_WORK_TREE=") || strings.HasPrefix(entry, "GIT_INDEX_FILE=") {
			continue
		}
		result = append(result, entry)
	}
	return result
}
