package linecaps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestNestedModuleLineMetricsRetainOwningBoundary(t *testing.T) {
	root := t.TempDir()
	writeMetricFile(t, root, "go.mod", "module example.test/root\n")
	writeMetricFile(t, root, "root.go", "package root\n")
	writeMetricFile(t, root, "nested/go.mod", "module example.test/nested\n")
	writeMetricFile(t, root, "nested/source.go", "package nested\n"+strings.Repeat("// line\n", 75))
	report, err := AnalyzeLineMetrics(root)
	if err != nil {
		t.Fatal(err)
	}
	if total := report.Total(); total.GoFiles != 2 || total.GoLines != 77 {
		t.Fatalf("nested module disappeared from repository totals: %#v", total)
	}
	for _, file := range report.Files {
		switch file.Path {
		case "root.go":
			if file.ModuleRoot != "" {
				t.Fatalf("root file owner changed: %#v", file)
			}
		case "nested/source.go":
			if file.ModuleRoot != "nested" || file.Lines != 76 {
				t.Fatalf("nested file owner or lines lost: %#v", file)
			}
		}
	}
	found := false
	for _, indicator := range report.Meta.Indicators {
		if indicator.Subject == "nested/source.go" && indicator.MetricID == sourcepolicy.DimensionGoFileLines {
			found = true
			if indicator.ModuleRoot != "nested" || indicator.Applicability != sourcepolicy.ApplicabilityApplicable ||
				indicator.Operation != sourcepolicy.OperationSplitGo {
				t.Fatalf("source subject lost module identity: %#v", indicator)
			}
		}
	}
	if !found {
		t.Fatal("nested file has no line indicator")
	}
	raw, err := report.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"module_root": "nested"`) ||
		strings.Contains(string(raw), `"module_root": "."`) {
		t.Fatalf("module identity was not compactly serialized: %s", raw)
	}
}

func TestCrossModuleSourceLinkKeepsCountsButNotRewriteAction(t *testing.T) {
	root := t.TempDir()
	writeMetricFile(t, root, "go.mod", "module example.test/root\n")
	writeMetricFile(t, root, "root.go", "package root\n"+strings.Repeat("// line\n", 75))
	writeMetricFile(t, root, "nested/go.mod", "module example.test/nested\n")
	if err := os.Symlink(filepath.Join("..", "root.go"), filepath.Join(root, "nested", "source.go")); err != nil {
		t.Fatal(err)
	}
	report, err := AnalyzeLineMetrics(root)
	if err != nil {
		t.Fatal(err)
	}
	if total := report.Total(); total.GoFiles != 2 || total.GoLines != 152 {
		t.Fatalf("cross-module source disappeared from totals: %#v", total)
	}
	for _, indicator := range report.Meta.Indicators {
		if indicator.Subject != "nested/source.go" || indicator.MetricID != sourcepolicy.DimensionGoFileLines {
			continue
		}
		if indicator.ModuleRoot != "nested" || indicator.Applicability != sourcepolicy.ApplicabilityNotApplicable ||
			indicator.ApplicabilityReason != sourcepolicy.ApplicabilityReasonCrossModuleSource ||
			indicator.ApplicabilityRule != sourcepolicy.ApplicabilityRuleGoModuleBoundary {
			t.Fatalf("unsafe rewrite remained actionable: %#v", indicator)
		}
		return
	}
	t.Fatal("cross-module source has no line indicator")
}

func TestInvalidNestedManifestMarksRewriteNotApplicable(t *testing.T) {
	root := t.TempDir()
	writeMetricFile(t, root, "go.mod", "module example.test/root\n")
	ensureMetricDir(t, root, "nested")
	if err := os.Symlink(filepath.Join("..", "go.mod"), filepath.Join(root, "nested", "go.mod")); err != nil {
		t.Fatal(err)
	}
	writeMetricFile(t, root, "nested/source.go", "package nested\n"+strings.Repeat("// line\n", 75))
	report, err := AnalyzeLineMetrics(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, indicator := range report.Meta.Indicators {
		if indicator.Subject != "nested/source.go" || indicator.MetricID != sourcepolicy.DimensionGoFileLines {
			continue
		}
		if indicator.Applicability != sourcepolicy.ApplicabilityNotApplicable ||
			indicator.ApplicabilityReason != sourcepolicy.ApplicabilityReasonModuleOwnerUnavailable {
			t.Fatalf("invalid module manifest remained actionable: %#v", indicator)
		}
		return
	}
	t.Fatal("nested source has no line indicator")
}

func TestNewerNestedGoRequirementMarksRewriteNotApplicable(t *testing.T) {
	root := t.TempDir()
	writeMetricFile(t, root, "go.mod", "module example.test/root\n\ngo 1.27.3\n")
	writeMetricFile(t, root, "nested/go.mod", "module example.test/nested\n\ngo 1.99.0\n")
	writeMetricFile(t, root, "nested/source.go", "package nested\n"+strings.Repeat("// line\n", 75))
	report, err := AnalyzeLineMetrics(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, indicator := range report.Meta.Indicators {
		if indicator.Subject != "nested/source.go" || indicator.MetricID != sourcepolicy.DimensionGoFileLines {
			continue
		}
		if indicator.ModuleRoot != "nested" || indicator.Applicability != sourcepolicy.ApplicabilityNotApplicable ||
			indicator.ApplicabilityReason != sourcepolicy.ApplicabilityReasonModuleToolchainUnavailable {
			t.Fatalf("newer nested Go requirement remained actionable: %#v", indicator)
		}
		return
	}
	t.Fatal("nested source has no line indicator")
}
