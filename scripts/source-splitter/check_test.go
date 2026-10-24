package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const checkSource = `package fixture

func first() int {
	return 1
}

func second() int {
	return 2
}
`

func TestCheckSubjectReturnsPlanningBlockerWithoutWriting(t *testing.T) {
	cases := []struct {
		name, source, want string
		files, folders     int
		bulk, omitDir      bool
	}{
		{name: "ready", source: checkSource, files: 1},
		{name: "declaration-capacity", source: strings.Replace(checkSource, "return 1", "_ = 1\n_ = 2\nreturn 1", 1),
			files: 1, want: "declarations cannot preserve source order"},
		{name: "directory-capacity", source: checkSource, files: 10, want: "projects 11 entries"},
		{name: "mixed-topology", source: checkSource, files: 1, folders: 1, want: "contains folders"},
		{name: "already-split", source: "package fixture\n\nfunc first() int { return 1 }\n",
			files: 1, want: "does not require declaration splitting"},
		{name: "missing-directory", source: checkSource, omitDir: true, want: "metric evidence omits directory"},
		{name: "bulk-inventory-keeps-blockers", source: checkSource, files: 10, bulk: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, sha := checkRepository(t, tc.source)
			report := checkMetrics(sha)
			report.Directories[0].DirectFiles, report.Directories[0].DirectFolders = tc.files, tc.folders
			if tc.omitDir {
				report.Directories = nil
			}
			args := checkArguments(t, root, sha, report)
			if !tc.bulk {
				args = append(args, "-subject", "fixture.go")
			}
			err := checkWithoutWrites(t, root, args)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("check error = %v; want %q", err, tc.want)
			} else if !tc.omitDir && !errors.Is(err, errSplitBlocked) {
				t.Fatalf("planning blocker identity lost: %v", err)
			}
		})
	}
}

func TestCheckSubjectKeepsEvidenceBinding(t *testing.T) {
	root, sha := checkRepository(t, checkSource)
	for _, name := range []string{"unmatched-subject", "stale-metrics", "different-head"} {
		t.Run(name, func(t *testing.T) {
			report, expected := checkMetrics(sha), sha
			subject, want := "fixture.go", "does not match"
			switch name {
			case "unmatched-subject":
				subject, want = "other.go", "not an actionable split indicator"
			case "stale-metrics":
				report.CommitSHA = strings.Repeat("0", 40)
			case "different-head":
				expected, report.CommitSHA = strings.Repeat("0", 40), strings.Repeat("0", 40)
			}
			args := append(checkArguments(t, root, expected, report), "-subject", subject)
			if err := checkWithoutWrites(t, root, args); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("check error = %v; want %q", err, want)
			}
		})
	}
}

func checkRepository(t *testing.T, source string) (string, string) {
	t.Helper()
	root := t.TempDir()
	for name, data := range map[string]string{"fixture.go": source, "go.mod": "module fixture\n\ngo 1.20\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var output []byte
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"commit", "-qm", "fixture"}, {"rev-parse", "HEAD"}} {
		command := exec.Command("git", append([]string{"-C", root, "-c", "user.name=Fixture",
			"-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false"}, args...)...)
		var err error
		output, err = command.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	return root, strings.TrimSpace(string(output))
}

func checkMetrics(sha string) metricevidence.Report {
	report := metricevidence.Report{CommitSHA: sha,
		Directories: []metricevidence.Directory{{Path: ".", DirectFiles: 1}}}
	report.Meta.Policy.MaxFileLines, report.Meta.Policy.MaxDirectEntries = 5, 10
	report.Meta.Indicators = []metricevidence.Indicator{{MetricID: sourcepolicy.DimensionGoFileLines,
		Subject: "fixture.go", MetaOperation: sourcepolicy.OperationSplitGo}}
	return report
}

func checkArguments(t *testing.T, root, sha string, report metricevidence.Report) []string {
	t.Helper()
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	metrics := filepath.Join(t.TempDir(), "metrics.json")
	if err := os.WriteFile(metrics, data, 0600); err != nil {
		t.Fatal(err)
	}
	return []string{"-root", root, "-metrics", metrics, "-sha", sha, "-check"}
}

func checkWithoutWrites(t *testing.T, root string, args []string) error {
	t.Helper()
	before := checkTree(t, root)
	err := run(args)
	if after := checkTree(t, root); !maps.Equal(before, after) {
		t.Fatal("read-only check changed the source tree")
	}
	return err
}

func checkTree(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == ".git" {
			return filepath.SkipDir
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		files[path] = info.Mode().String()
		if !entry.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files[path] += string(data)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}
