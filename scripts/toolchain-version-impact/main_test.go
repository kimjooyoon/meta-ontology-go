package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestClassifyOnlyExactSelectors(t *testing.T) {
	for _, test := range []struct {
		line       string
		classified bool
	}{
		{`go-version: '1.27.3'`, true},
		{`GOTOOLCHAIN: go1.27.3 # selected`, true},
		{`go-version: '1.27.2' # previous 1.27.3`, false},
		{`run: test "$(go env GOVERSION)" = "go1.27.3"`, true},
		{`test "$(go env GOVERSION)" = "go1.27.3"`, true},
	} {
		_, got := classify(".github/workflows/ci.yml", test.line, "1.27.3", 1)
		if got != test.classified {
			t.Fatalf("line %q: classified=%t, want %t", test.line, got, test.classified)
		}
	}
}

func TestGoooToolchainDeclarationsJoinDetachedProjection(t *testing.T) {
	clearGitTestEnvironment(t)
	root := t.TempDir()
	write := func(path, source string) {
		t.Helper()
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	contract := `activity Receipt computes "role=compatibility;compatibility-test-contract=go1.27.3|go test ./...|go test -race ./..."`
	badge := `activity Badge computes "role=public;workflow-go-version=1.27.3;badge-row=go-toolchain|Language|Go 1.27.3 toolchain|Uses Go 1.27.3|https://example.org/Go-1.27.3-blue|CLOSED"`
	write("go.mod", "module example.org/impact\n\ngo 1.27.3\n")
	write("contract.gooo", contract)
	write("badge.gooo", badge)
	paths := []string{"badge.gooo", "contract.gooo", "go.mod"}
	for _, args := range [][]string{{"init", "-q"}, {"add", "go.mod", "contract.gooo", "badge.gooo"}} {
		command := exec.Command("git", args...)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	observed, err := scanPaths(root, paths)
	if err != nil {
		t.Fatal(err)
	}
	if observed.ActiveOccurrences != 6 || observed.ClassifiedOccurrences != 6 || observed.UnresolvedOccurrences != 0 {
		t.Fatalf("Gooo declarations did not join the source-bound set: %+v", observed)
	}
	plan, err := buildProjectionPlan(observed, "1.27.4")
	if err != nil || len(plan.CandidateEdits) != 6 {
		t.Fatalf("missing Gooo projection edits: %+v, %v", plan, err)
	}
	decision := projectionDecision{ModuleSHA256: plan.ModuleSHA256, PlanCoreSHA256: plan.PlanCoreSHA256,
		CurrentVersion: plan.CurrentVersion, NextVersion: plan.NextVersion,
		CandidateEdits: len(plan.CandidateEdits), Status: "READY_FOR_DETACHED_PROJECTION"}
	parent := t.TempDir()
	receipt, err := materializeCandidate(root, filepath.Join(parent, "candidate"), plan, decision)
	if err != nil || len(receipt.Files) != 3 {
		t.Fatalf("Gooo sources were not projected together: %+v, %v", receipt, err)
	}
	for _, path := range paths {
		data, err := os.ReadFile(filepath.Join(parent, "candidate", path))
		if err != nil || strings.Contains(string(data), "1.27.3") || !strings.Contains(string(data), "1.27.4") {
			t.Fatalf("candidate did not update %s: %v", path, err)
		}
		original, err := os.ReadFile(filepath.Join(root, path))
		if err != nil || !strings.Contains(string(original), "1.27.3") {
			t.Fatalf("source was changed at %s: %v", path, err)
		}
	}
	write("badge.gooo", strings.Replace(badge, "Go-1.27.3-blue", "Go-1.27.2-blue", 1))
	stale, err := scanPaths(root, paths)
	if err != nil || stale.StaleSelectors != 1 || stale.UnresolvedOccurrences == 0 {
		t.Fatalf("mixed Gooo versions were not left unresolved: %+v, %v", stale, err)
	}
}

func TestDetachedProjectionRequiresClosedSourceAndExactPreimages(t *testing.T) {
	clearGitTestEnvironment(t)
	root := t.TempDir()
	write := func(path, value string) {
		t.Helper()
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	module := "module example.org/impact\n\ngo 1.27.3\n"
	workflow := "name: CI\nsteps:\n  - uses: actions/setup-go@v7\n    with:\n      go-version: '1.27.3'\n"
	write("go.mod", module)
	write(".github/workflows/ci.yml", workflow)
	for _, args := range [][]string{{"init", "-q"}, {"add", "go.mod", ".github/workflows/ci.yml"}} {
		command := exec.Command("git", args...)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	observed, err := scan(root)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := buildProjectionPlan(observed, "1.27.4")
	if err != nil {
		t.Fatal(err)
	}
	decision := projectionDecision{ModuleSHA256: plan.ModuleSHA256, PlanCoreSHA256: plan.PlanCoreSHA256,
		CurrentVersion: plan.CurrentVersion, NextVersion: plan.NextVersion,
		CandidateEdits: len(plan.CandidateEdits), Status: "READY_FOR_DETACHED_PROJECTION"}
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		t.Fatal(err)
	}
	actualDecision, actualErr := executeProjectionDecision(projectionInputs(plan))
	if err := os.Chdir(previousDir); err != nil {
		t.Fatal(err)
	}
	if actualErr != nil || actualDecision != decision {
		t.Fatalf("Gooo decision disagrees with closed plan: %+v, %v", actualDecision, actualErr)
	}
	dest := filepath.Join(t.TempDir(), "candidate")
	if _, err := materializeCandidate(root, filepath.Join(root, "candidate"), plan, decision); err == nil {
		t.Fatal("source-tree destination accepted")
	}
	linkedParent := filepath.Join(t.TempDir(), "source-link")
	if err := os.Symlink(root, linkedParent); err == nil {
		if _, err := materializeCandidate(root, filepath.Join(linkedParent, "candidate"), plan, decision); err == nil {
			t.Fatal("symlinked source-tree destination accepted")
		}
	} else {
		t.Logf("symlink unavailable on this host: %v", err)
	}
	forged := decision
	forged.PlanCoreSHA256 = "sha256:forged"
	if _, err := materializeCandidate(root, dest, plan, forged); err == nil {
		t.Fatal("unbound Gooo decision accepted")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatal("rejected decision created candidate directory")
	}
	write(".github/workflows/ci.yml", workflow+"# concurrent edit\n")
	if _, err := materializeCandidate(root, dest, plan, decision); err == nil {
		t.Fatal("changed source accepted")
	}
	write(".github/workflows/ci.yml", workflow)
	receipt, err := materializeCandidate(root, dest, plan, decision)
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Files) != 2 || receipt.Provenance["@graph"] == nil {
		t.Fatalf("missing projected files or actual provenance: %+v", receipt)
	}
	for _, path := range []string{"go.mod", ".github/workflows/ci.yml"} {
		candidate, err := os.ReadFile(filepath.Join(dest, path))
		if err != nil || !strings.Contains(string(candidate), "1.27.4") {
			t.Fatalf("no candidate at %s: %v", path, err)
		}
		original, err := os.ReadFile(filepath.Join(root, path))
		if err != nil || !strings.Contains(string(original), "1.27.3") {
			t.Fatalf("source changed at %s: %v", path, err)
		}
	}
	if _, err := materializeCandidate(root, dest, plan, decision); err == nil {
		t.Fatal("existing candidate directory overwritten")
	}
	write(".github/workflows/ci.yml", workflow+"run: echo go1.27.3\n")
	openObserved, err := scan(root)
	if err != nil {
		t.Fatal(err)
	}
	openPlan, err := buildProjectionPlan(openObserved, "1.27.4")
	if err != nil {
		t.Fatal(err)
	}
	openDecision := decision
	openDecision.PlanCoreSHA256 = openPlan.PlanCoreSHA256
	openDecision.CandidateEdits = len(openPlan.CandidateEdits)
	if _, err := materializeCandidate(root, filepath.Join(t.TempDir(), "open"), openPlan, openDecision); err == nil {
		t.Fatal("unresolved consumer accepted despite ready-looking decision")
	}
}

func clearGitTestEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE"} {
		previous, existed := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if existed {
				_ = os.Setenv(key, previous)
			} else {
				_ = os.Unsetenv(key)
			}
		})
	}
}

func TestScanKeepsHistoricalAndUnknownSeparate(t *testing.T) {
	root := t.TempDir()
	write := func(path, data string) {
		t.Helper()
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.org/impact\n\ngo 1.27.3\n")
	write(".github/workflows/ci.yml", "go-version: '1.27.3'\nrun: test go1.27.3\n")
	write(".github/workflows/stale.yml", "go-version: '1.27.2'\n")
	write("docs/releases/old.md", "Built with Go 1.27.2; SDK 0.2.20.\n")
	write("untracked.md", "Go 1.27.3\n")
	paths := []string{"go.mod", ".github/workflows/ci.yml", ".github/workflows/stale.yml", "docs/releases/old.md"}
	result, err := scanPaths(root, paths)
	if err != nil {
		t.Fatal(err)
	}
	if result.ActiveOccurrences != 4 || result.ClassifiedOccurrences != 2 ||
		result.UnresolvedOccurrences != 2 || result.StaleSelectors != 1 ||
		result.HistoricalOccurrences != 1 || result.OutsideScope != 0 {
		t.Fatalf("unexpected impact: %+v", result)
	}
	if err := os.Remove(filepath.Join(root, ".github/workflows/ci.yml")); err != nil {
		t.Fatal(err)
	}
	if _, err := scanPaths(root, paths); err == nil {
		t.Fatal("missing tracked workflow was silently omitted")
	}
}

func TestProjectionPlanBindsExactSpansAndLeavesUnknownOpen(t *testing.T) {
	root := t.TempDir()
	for name, data := range map[string]string{
		"go.mod":                   "module example.org/impact\n\ngo 1.27.3\n",
		".github/workflows/ci.yml": "go-version: '1.27.3'\nrun: echo go1.27.3\n",
	} {
		full := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	observed, err := scanPaths(root, []string{"go.mod", ".github/workflows/ci.yml"})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := buildProjectionPlan(observed, "1.27.4")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.CandidateEdits) != 2 || len(plan.Unresolved) != 1 || plan.HistoricalWrites != 0 {
		t.Fatalf("unexpected plan coverage: %+v", plan)
	}
	if plan.ObservationSHA256 == "" || plan.PlanCoreSHA256 == "" || plan.Provenance["@graph"] == nil {
		t.Fatalf("plan lacks bound observation and provenance: %+v", plan)
	}
	graph := plan.Provenance["@graph"].([]any)
	moduleID := "urn:gooo:toolchain-projection:module:" + plan.ModuleSHA256
	moduleFound, moduleUsed := false, false
	for _, node := range graph {
		entity := node.(map[string]any)
		if entity["@id"] == moduleID {
			moduleFound = entity["gooo:path"] == "go.mod" && entity["gooo:content_sha256"] == plan.ModuleSHA256
		}
		if used, ok := entity["prov:used"].([]any); ok {
			for _, ref := range used {
				moduleUsed = moduleUsed || ref.(map[string]string)["@id"] == moduleID
			}
		}
	}
	if !moduleFound || !moduleUsed {
		t.Fatal("plan did not identify and use the exact module source")
	}
	for _, edit := range plan.CandidateEdits {
		data, err := os.ReadFile(filepath.Join(root, edit.Path))
		if err != nil {
			t.Fatal(err)
		}
		if string(data[edit.StartByte:edit.EndByte]) != edit.Before || digest(data) != edit.PreimageSHA256 {
			t.Fatalf("candidate lacks exact preimage: %+v", edit)
		}
		if !strings.HasPrefix(edit.ClassifierID, "urn:gooo:toolchain-impact:classifier:") {
			t.Fatalf("candidate lacks classifier identity: %+v", edit)
		}
	}
	if _, err := buildProjectionPlan(observed, "next"); err == nil {
		t.Fatal("unsupported target version accepted")
	}
}
