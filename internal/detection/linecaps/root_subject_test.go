package linecaps

import (
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestRootWithoutReadmeRemainsAnExplicitMetricException(t *testing.T) {
	root := t.TempDir()
	writeMetricFile(t, root, "main.go", "package main\n")
	writeMetricFile(t, root, filepath.Join("nested", "value.gooo"), "intent: nested\n")

	report, err := AnalyzeLineMetrics(root)
	if err != nil {
		t.Fatal(err)
	}
	total := report.Total()
	if total.SubjectKind != sourcepolicy.SubjectKindProjectRoot ||
		total.RecursiveFiles != 2 || total.GoFiles != 1 || total.GoooFiles != 1 {
		t.Fatalf("invalid project-root metric: %#v", total)
	}
	nested := directoryForPath(report, "nested")
	if nested.SubjectKind != sourcepolicy.SubjectKindDirectory {
		t.Fatalf("nested directory is not classified: %#v", nested)
	}
	exemptions := 0
	for _, indicator := range report.Meta.Indicators {
		if indicator.Subject != "." ||
			(indicator.MetricID != sourcepolicy.DimensionDirectEntries &&
				indicator.MetricID != sourcepolicy.DimensionDirectoryKinds) {
			continue
		}
		exemptions++
		if indicator.Applicability != sourcepolicy.ApplicabilityNotApplicable ||
			indicator.ApplicabilityReason != sourcepolicy.ApplicabilityReasonRootTopologyExempt ||
			indicator.Operation != sourcepolicy.OperationExemptRoot {
			t.Fatalf("root exception is not bound to meta code: %#v", indicator)
		}
	}
	if exemptions != 2 {
		t.Fatalf("root topology exemptions = %d", exemptions)
	}
}
