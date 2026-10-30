package generation

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

const splitPlannerSource = `package fixture

func First() int {
	return 1
}

func Second() int {
	return 2
}
`

func TestGoSplitPlannerPrepareIsReadOnlyAndDeterministic(t *testing.T) {
	root, report := splitPlannerFixture(t, splitPlannerSource)
	planner := NewGoSplitPlanner()
	first, err := planner.Prepare(root, "fixture.go", report)
	if err != nil {
		t.Fatal(err)
	}
	second, err := planner.Prepare(root, "fixture.go", report)
	if err != nil || !reflect.DeepEqual(first, second) || len(first.Parts) != 2 {
		t.Fatalf("native plan did not replay: first=%+v second=%+v error=%v", first, second, err)
	}
	if first.Parts[0].Subject != "fixture.go" || first.Parts[1].Subject != "fixture_split02.go" {
		t.Fatalf("native target identity changed: %+v", first.Parts)
	}
	assertSplitPlannerSource(t, root, splitPlannerSource)
}

func TestGoSplitPlannerPrepareRefusalsReturnNoParts(t *testing.T) {
	for _, test := range []struct {
		name, source, reason string
		directories          []metricevidence.Directory
	}{
		{"oversized-declaration",
			"package fixture\nfunc One() int {\nvalue := 1\nvalue++\nvalue++\nvalue++\nreturn value\n}\n",
			"declarations cannot preserve source order", []metricevidence.Directory{{Path: ".", DirectFiles: 1}}},
		{"full-directory", splitPlannerSource, "projects 3 entries", []metricevidence.Directory{{Path: ".", DirectFiles: 2}}},
		{"mixed-directory", splitPlannerSource, "contains folders",
			[]metricevidence.Directory{{Path: ".", DirectFiles: 1, DirectFolders: 1}}},
		{"missing-directory-observation", splitPlannerSource, "metric evidence omits directory", nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, report := splitPlannerFixture(t, test.source)
			report.Directories = test.directories
			plan, err := NewGoSplitPlanner().Prepare(root, "fixture.go", report)
			if err == nil || !strings.Contains(err.Error(), test.reason) || len(plan.Parts) != 0 {
				t.Fatalf("refusal acquired parts or lost cause: plan=%+v error=%v", plan, err)
			}
			assertSplitPlannerSource(t, root, test.source)
		})
	}
}

func TestGoSplitPlannerPrepareRereadsCurrentSource(t *testing.T) {
	root, report := splitPlannerFixture(t, splitPlannerSource)
	planner := NewGoSplitPlanner()
	if _, err := planner.Prepare(root, "fixture.go", report); err != nil {
		t.Fatal(err)
	}
	const changed = "package fixture\nfunc One() int {\nvalue := 1\nvalue++\nvalue++\nvalue++\nreturn value\n}\n"
	if err := os.WriteFile(filepath.Join(root, "fixture.go"), []byte(changed), 0o640); err != nil {
		t.Fatal(err)
	}
	plan, err := planner.Prepare(root, "fixture.go", report)
	if err == nil || len(plan.Parts) != 0 {
		t.Fatalf("prior plan survived a changed declaration: plan=%+v error=%v", plan, err)
	}
	assertSplitPlannerSource(t, root, changed)
}

func splitPlannerFixture(t *testing.T, source string) (string, metricevidence.Report) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "fixture.go"), []byte(source), 0o640); err != nil {
		t.Fatal(err)
	}
	report := metricevidence.Report{Directories: []metricevidence.Directory{{Path: ".", DirectFiles: 1}}}
	report.Meta.Policy.MaxFileLines, report.Meta.Policy.MaxDirectEntries = 6, 2
	return root, report
}

func assertSplitPlannerSource(t *testing.T, root, expected string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, "fixture.go"))
	if err != nil || !bytes.Equal(raw, []byte(expected)) {
		t.Fatalf("planning changed original bytes: %q, %v", raw, err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 || entries[0].Name() != "fixture.go" {
		t.Fatalf("planning created a successor or temporary: %v, %v", entries, err)
	}
	info, err := os.Stat(filepath.Join(root, "fixture.go"))
	if err != nil || info.Mode().Perm() != 0o640 {
		t.Fatalf("planning changed source mode: %v, %v", info, err)
	}
}
