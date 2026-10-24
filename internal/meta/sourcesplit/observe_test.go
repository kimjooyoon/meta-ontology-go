package sourcesplit

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestObservedSplitSelectsLaterNativeCandidate(t *testing.T) {
	root, head := observedRepository(t, map[string]string{
		"a.go":          "package fixture\n" + observedFunction("Oversized", 85),
		"b.go":          "package fixture\n" + observedFunction("First", 40) + observedFunction("Second", 40),
		"expression.go": "package fixture\nfunc Value() int {\n value := 1\n return value\n}\n",
	})
	report, raw := observedMetrics(t, root, head)
	observe, err := NewObserver(root, head, raw)
	if err != nil {
		t.Fatal(err)
	}
	before := observedGit(t, root, "status", "--porcelain", "--untracked-files=all")
	first := generation.BuildWithPrerequisites(head, head, report.Meta, nil, observe)
	second := generation.BuildWithPrerequisites(head, head, report.Meta, nil, observe)
	if first.Decision != generation.DecisionPlan || !reflect.DeepEqual(first, second) || len(first.Selected) != 2 {
		t.Fatalf("native prerequisites did not yield a replayable plan: %+v", first)
	}
	if first.Selected[1].Subject != "b.go" || len(first.Prerequisites) != 2 || first.IndicatorDecisionLedger == nil ||
		first.IndicatorDecisionLedger.IndicatorCount != len(report.Meta.Indicators) {
		t.Fatalf("blocked first candidate hid the later candidate or denominator: %+v", first)
	}
	if after := observedGit(t, root, "status", "--porcelain", "--untracked-files=all"); after != before {
		t.Fatal("planning changed the source tree", after)
	}
	if manifest := generation.BuildExecutionManifest(first); manifest.Decision != generation.ExecutionDecisionProposed {
		t.Fatal("observed plan cannot reach the execution consumer", manifest)
	}
}

func TestNativeSplitPrerequisiteDomains(t *testing.T) {
	root, head := observedRepository(t, map[string]string{
		"fixture.go": "package fixture\n" + observedFunction("First", 40) + observedFunction("Second", 40),
	})
	report, _ := observedMetrics(t, root, head)
	contract, err := generation.SplitGoInputContractEvidence()
	if err != nil {
		t.Fatal(err)
	}
	action := generation.Action{Subject: "fixture.go", IndicatorID: "sha256:" + strings.Repeat("1", 64),
		Operation: sourcepolicy.OperationSplitGo, InputContractSourceDigest: contract.SourceDigest,
		InputContractSemanticDigest: contract.SemanticDigest}
	cases := []struct {
		name, state, reason string
		mutate              func(*linecaps.LineMetricsReport)
	}{
		{"ready", "READY", "SPLIT_PLAN_READY", func(*linecaps.LineMetricsReport) {}},
		{"full", "BLOCKED", "LOGICAL_DIRECTORY_CAPACITY", func(r *linecaps.LineMetricsReport) { r.Directories[0].DirectFiles = 10 }},
		{"mixed", "BLOCKED", "MIXED_LOGICAL_TOPOLOGY", func(r *linecaps.LineMetricsReport) { r.Directories[0].DirectFolders = 1 }},
		{"missing", "UNKNOWN", "DIRECTORY_OBSERVATION_MISSING", func(r *linecaps.LineMetricsReport) { r.Directories = nil }},
		{"projection", "DEFERRED", "SUCCESSOR_PROJECTION_REQUIRED", func(r *linecaps.LineMetricsReport) { r.StorageRoot = "physical" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			copy := report
			copy.Directories = append([]linecaps.DirectoryMetric{}, report.Directories...)
			tc.mutate(&copy)
			raw, _ := json.Marshal(copy)
			observe, err := NewObserver(root, head, raw)
			if err != nil {
				t.Fatal(err)
			}
			observed := observe(action)
			if observed.State != tc.state || observed.Reason != tc.reason || observed.SourceDigest == "" ||
				observed.MetricsDigest != sourceDigest(raw) || observed.ContractSourceDigest != contract.SourceDigest {
				t.Fatalf("prerequisite lost its domain or origin: %+v", observed)
			}
		})
	}
	if status := observedGit(t, root, "status", "--porcelain", "--untracked-files=all"); status != "" {
		t.Fatal("read-only observations changed source", status)
	}
}

func observedFunction(name string, statements int) string {
	var source strings.Builder
	fmt.Fprintf(&source, "\nfunc %s() int {\n total := 0\n", name)
	for index := range statements {
		fmt.Fprintf(&source, " total += %d\n", index)
	}
	source.WriteString(" return total\n}\n")
	return source.String()
}

func observedRepository(t testing.TB, files map[string]string) (string, string) {
	t.Helper()
	root := t.TempDir()
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	observedGit(t, root, "init", "-q")
	observedGit(t, root, "add", ".")
	observedGit(t, root, "commit", "-qm", "native observation fixture")
	return root, observedGit(t, root, "rev-parse", "HEAD")
}

func observedGit(t testing.TB, root string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root, "-c", "user.name=Fixture",
		"-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false"}, args...)...)
	raw, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, raw)
	}
	return strings.TrimSpace(string(raw))
}

func observedMetrics(t testing.TB, root, head string) (linecaps.LineMetricsReport, []byte) {
	t.Helper()
	report, err := linecaps.AnalyzeLineMetrics(root)
	if err != nil {
		t.Fatal(err)
	}
	report.CommitSHA = head
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	return report, raw
}

func BenchmarkSplitPrerequisite(b *testing.B) {
	for _, size := range []int{40, 85} {
		b.Run(fmt.Sprintf("statements-%d", size), func(b *testing.B) {
			root, head := observedRepository(b, map[string]string{
				"fixture.go": "package fixture\n" + observedFunction("First", size) + observedFunction("Second", size),
			})
			_, raw := observedMetrics(b, root, head)
			observe, err := NewObserver(root, head, raw)
			if err != nil {
				b.Fatal(err)
			}
			contract, err := generation.SplitGoInputContractEvidence()
			if err != nil {
				b.Fatal(err)
			}
			action := generation.Action{Subject: "fixture.go", Operation: sourcepolicy.OperationSplitGo,
				InputContractSourceDigest: contract.SourceDigest, InputContractSemanticDigest: contract.SemanticDigest}
			want := "READY"
			if size == 85 {
				want = "BLOCKED"
			}
			b.ReportAllocs()
			for b.Loop() {
				if result := observe(action); result.State != want {
					b.Fatal(result)
				}
			}
		})
	}
}
