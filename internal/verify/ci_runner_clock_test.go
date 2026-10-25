package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestRunnerClockJobNamesFollowWorkflow(t *testing.T) {
	raw, err := os.ReadFile("../../.github/actions/work-domain/jobs.json")
	if err != nil {
		t.Fatal(err)
	}
	var profile map[string]string
	if err = json.Unmarshal(raw, &profile); err != nil {
		t.Fatal(err)
	}
	workflow, err := os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	pattern := regexp.MustCompile(`(?m)^  ([a-z][a-z0-9-]*):\n    name: ([^\n]+)\n`)
	jobs := pattern.FindAllStringSubmatch(string(workflow), -1)
	if len(jobs) != len(profile) || len(jobs) == 0 {
		t.Fatal("clock job population differs", len(jobs), len(profile))
	}
	for _, job := range jobs {
		if profile[job[1]] != job[2] {
			t.Fatal("clock job name differs", job[1], job[2], profile[job[1]])
		}
	}
}

func TestActualRunnerClockBinderPreservesRunnerFreeGaps(t *testing.T) {
	tool := os.Getenv("GOOO_CI_DOMAIN")
	if tool == "" || runtime.GOOS != "linux" {
		t.Skip("native CI domain/Linux binder available in CI")
	}
	action, err := os.ReadFile("../../.github/actions/work-domain/action.yml")
	if err != nil {
		t.Fatal(err)
	}
	_, block, ok := strings.Cut(string(action), "      run: |\n")
	if !ok {
		t.Fatal("binder body unavailable")
	}
	block, _, _ = strings.Cut(block, "    - name: Restore matching command cost origins\n")
	lines := strings.Split(block, "\n")
	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], "        ")
	}
	script := strings.Join(lines, "\n")
	root := t.TempDir()
	put := func(name string, raw []byte, mode os.FileMode) string {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, raw, mode); err != nil {
			t.Fatal(err)
		}
		return path
	}
	putJSON := func(name string, value any) string {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return put(name, raw, 0600)
	}
	toolBytes, err := os.ReadFile(tool)
	if err != nil {
		t.Fatal(err)
	}
	put("ci-domain-tool/gooo-ci-domain", toolBytes, 0700)
	digest := sha256.Sum256(toolBytes)
	head, revision := strings.Repeat("a", 40), strings.Repeat("b", 40)
	now := time.Now().UTC().Truncate(time.Second)
	created, prepared := now.Add(-time.Hour), now.Add(-56*time.Minute)
	putJSON("ci-domain-tool/origin.json", map[string]any{
		"schema": "gooo/work-domain-tool-origin/v1", "source_head": head, "run_id": "19725", "run_attempt": "1",
		"workflow_created_at": created, "started_at": prepared, "clock_scope": "admitted_source_work",
		"tool_sha256": hex.EncodeToString(digest[:]), "tool_revision": revision,
	})
	runPath := putJSON("original-run.json", map[string]any{"id": 19725, "run_attempt": 1, "head_sha": head, "created_at": created})
	job := func(id int, name, status string, start time.Time, end any) map[string]any {
		return map[string]any{"id": id, "run_id": 19725, "run_attempt": 1, "head_sha": head, "name": name,
			"status": status, "conclusion": "success", "started_at": start, "completed_at": end,
			"runner_id": id + 100, "steps": []any{}}
	}
	queued := job(3, "Queued timestamp", "queued", prepared, nil)
	queued["runner_id"], queued["conclusion"] = 0, nil
	cancelled := job(4, "Unassigned cancellation", "completed", prepared, prepared.Add(time.Minute))
	cancelled["runner_id"], cancelled["conclusion"] = 0, "cancelled"
	skipped := job(5, "Unassigned skip", "completed", now, now.Add(-time.Second))
	skipped["runner_id"], skipped["conclusion"] = nil, "skipped"
	pages := putJSON("original-pages.json", []any{map[string]any{"total_count": 5, "jobs": []any{
		job(1, "Work domain preparation", "completed", prepared, prepared.Add(10*time.Second)),
		job(2, "CI proof bundle", "in_progress", now.Add(-time.Second), nil),
		queued, cancelled, skipped,
	}}})
	put("bin/gh", []byte("#!/bin/bash\ncase \"$*\" in *attempts*) cat \"$CLOCK_TEST_PAGES\";; *) cat \"$CLOCK_TEST_RUN\";; esac\n"), 0700)
	body := put("bind.sh", []byte(script), 0600)
	actionPath, err := filepath.Abs("../../.github/actions/work-domain")
	if err != nil {
		t.Fatal(err)
	}
	envPath := put("environment", nil, 0600)
	cmd := exec.Command("/bin/bash", body)
	cmd.Env = append(os.Environ(), "PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"),
		"RUNNER_TEMP="+root, "GITHUB_ACTION_PATH="+actionPath, "EXPECTED_SOURCE_HEAD="+head,
		"GOOO_CI_DOMAIN_REVISION="+revision, "GOOO_CI_BUDGET_MS=3000000", "GITHUB_RUN_ID=19725",
		"GITHUB_RUN_ATTEMPT=1", "GITHUB_JOB=proof", "GITHUB_REPOSITORY=clock/fixture", "GITHUB_ENV="+envPath,
		"CLOCK_TEST_RUN="+runPath, "CLOCK_TEST_PAGES="+pages)
	if logs, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(string(logs), err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "command-origins-proof/runner-clock.json"))
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Decision   string
		Active     int64   `json:"active_elapsed_ms"`
		Wall       int64   `json:"wall_elapsed_ms"`
		Unassigned []int64 `json:"unassigned_job_ids"`
	}
	if err = json.Unmarshal(raw, &receipt); err != nil || receipt.Decision != "READY" || receipt.Active >= 60000 || receipt.Wall < 3600000 {
		t.Fatal("runner-free time spent the command allowance", string(raw), err)
	}
	if !slices.Equal(receipt.Unassigned, []int64{3, 4, 5}) {
		t.Fatal("unassigned observations were lost", receipt.Unassigned)
	}
	environment, err := os.ReadFile(envPath)
	if err != nil || !strings.Contains(string(environment), "GOOO_CI_CLOCK_SCOPE=runner_active_union\n") {
		t.Fatal(string(environment), err)
	}
}
