package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
	"github.com/kimjooyoon/meta-ontology-go/scripts/source-splitter/completion"
)

func TestBatchRetainsUnprojectableInputAndItsOrigin(t *testing.T) {
	root := t.TempDir()
	plan := []byte(`{"schema":"gooo.logical-split-plan.v1","source_sha":"source",` +
		`"subjects":[{"logical":"large.gooo","reason":"unclassified","consumer":"inspect","meta_operation":""}]}`)
	planPath, output := filepath.Join(root, "plan.json"), filepath.Join(root, "report.json")
	if err := os.WriteFile(planPath, plan, 0600); err != nil {
		t.Fatal(err)
	}
	indicators := []metricevidence.Indicator{{Subject: "large.gooo"}}
	if err := runBatch(config{root: root, sha: "source", plan: planPath, output: output},
		metricevidence.Report{}, indicators); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var report splitBatchReport
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != "PROGRESS" || report.NextOperation != "resolve-unhandled-source-subjects" ||
		!reflect.DeepEqual(report.Unhandled, []string{"large.gooo"}) || report.Coordinates.InputSubjects != 1 ||
		report.Coordinates.UnhandledSubjects != 1 || report.PlanSourceDigest != fmt.Sprintf("sha256:%x", sha256.Sum256(plan)) ||
		report.PolicySourceDigest != completion.SourceDigest() {
		t.Fatal("unprojectable source or its origins disappeared from the actual batch receipt", report)
	}
}
