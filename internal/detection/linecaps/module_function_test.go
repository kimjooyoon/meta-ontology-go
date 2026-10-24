package linecaps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestModuleOwnershipReachesFunctionAndDuplicateIndicators(t *testing.T) {
	for _, version := range []string{"1.27.3", "1.99.0"} {
		t.Run(version, func(t *testing.T) {
			root := t.TempDir()
			writeMetricFile(t, root, "go.mod", "module example.test/root\n")
			writeMetricFile(t, root, "nested/go.mod", "module example.test/nested\ngo "+version+"\n")
			source := "package nested\n" +
				"func First(x int) int {\n y := x+1\n y *= 2\n return y\n}\n" +
				"func Second(x int) int {\n y := x+1\n y *= 2\n return y\n}\n" +
				"func Assign(x int) int {\n y := x+1\n return y\n}\n" +
				"func Long(x int) int {\n" + strings.Repeat(" x++\n", 80) + " return x\n}\n"
			writeMetricFile(t, root, "nested/source.go", source)
			report, err := AnalyzeLineMetrics(root)
			if err != nil {
				t.Fatal(err)
			}
			found := map[sourcepolicy.Dimension]bool{}
			for _, indicator := range report.Meta.Indicators {
				switch indicator.MetricID {
				case sourcepolicy.DimensionGoFileLines, sourcepolicy.DimensionFunctionLines,
					sourcepolicy.DimensionRefactorAssign, sourcepolicy.DimensionRefactorDuplicate:
				default:
					continue
				}
				found[indicator.MetricID] = true
				if indicator.ModuleRoot != "nested" || indicator.Value <= 0 {
					t.Fatalf("function/group origin or measurement lost: %+v", indicator)
				}
				if version == "1.99.0" {
					if indicator.Applicability != sourcepolicy.ApplicabilityNotApplicable || indicator.Blocking ||
						indicator.ApplicabilityReason != sourcepolicy.ApplicabilityReasonModuleToolchainUnavailable {
						t.Fatalf("unavailable module still selected a rewrite: %+v", indicator)
					}
				} else if indicator.Applicability != sourcepolicy.ApplicabilityApplicable {
					t.Fatalf("supported nested module was excluded: %+v", indicator)
				}
			}
			if len(found) != 4 {
				t.Fatalf("missing rewrite dimensions: %v", found)
			}
		})
	}
}

func TestDuplicateGroupRetainsCrossModuleSourceReason(t *testing.T) {
	root := t.TempDir()
	writeMetricFile(t, root, "go.mod", "module example.test/root\n")
	writeMetricFile(t, root, "nested/go.mod", "module example.test/nested\n")
	writeMetricFile(t, root, "root.go", "package p\nfunc First(x int) int {\n y := x+1\n y *= 2\n return y\n}\n")
	writeMetricFile(t, root, "nested/own.go", "package p\nfunc Second(x int) int {\n y := x+1\n y *= 2\n return y\n}\n")
	if err := os.Symlink("../root.go", filepath.Join(root, "nested", "linked.go")); err != nil {
		t.Fatal(err)
	}
	report, err := AnalyzeLineMetrics(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, indicator := range report.Meta.Indicators {
		if indicator.MetricID == sourcepolicy.DimensionRefactorDuplicate {
			if indicator.ModuleRoot != "nested" || indicator.Value != 1 ||
				indicator.Applicability != sourcepolicy.ApplicabilityNotApplicable || indicator.Blocking ||
				indicator.ApplicabilityReason != sourcepolicy.ApplicabilityReasonCrossModuleSource {
				t.Fatalf("aggregate lost member's module boundary: %+v", indicator)
			}
			return
		}
	}
	t.Fatal("duplicate group disappeared")
}
