package linecaps

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestAnalyzeProjectedLineMetricsSeparatesSourceAndStorage(t *testing.T) {
	source, storage := t.TempDir(), t.TempDir()
	for index := range 11 {
		name := filepath.Join(source, fmt.Sprintf("note-%02d.txt", index))
		if err := os.WriteFile(name, []byte("metric\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(source, "source.go"), []byte("package source\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"objects", "catalog"} {
		if err := os.Mkdir(filepath.Join(storage, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	report, err := AnalyzeProjectedLineMetrics(source, storage)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.StorageDirectories) != 3 {
		t.Fatalf("projection planes were not preserved: %#v", report)
	}
	foundSource := false
	for _, file := range report.Files {
		foundSource = foundSource || file.Path == "source.go"
	}
	if !foundSource {
		t.Fatal("logical source metric was not preserved")
	}
	for _, indicator := range report.Meta.Failed() {
		if indicator.MetricID == sourcepolicy.DimensionDirectEntries || indicator.MetricID == sourcepolicy.DimensionDirectoryKinds {
			t.Fatalf("logical topology leaked into storage policy: %#v", indicator)
		}
	}
	for _, indicator := range report.Meta.Indicators {
		if indicator.MetricID == sourcepolicy.DimensionDirectEntries && indicator.Subject == "." && indicator.Producer != "repository-projector.topology" {
			t.Fatalf("topology indicator lost its meta producer: %#v", indicator)
		}
	}
}
