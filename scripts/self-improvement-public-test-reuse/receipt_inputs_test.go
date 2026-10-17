package main

import (
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
)

func TestUnobservedLegacyReceiptProducesUnresolvedReport(t *testing.T) {
	output := t.TempDir()
	initial := publictestreuse.Report{Schema: publictestreuse.ReportSchema, Operation: publictestreuse.ReplayOperation}
	initial.Binding.PackageInputs = publictestreuse.ExecutionPackageInputsFromBytes(
		[]byte("module fixture\n"), []byte("package fixture\n"), []byte("package fixture\n// test\n"))
	if err := writeInvalidReuseReceipt(output, initial, publictestreuse.ErrUnobservedPackageInputs); err != nil {
		t.Fatal(err)
	}
	report, err := readPublicReport(filepath.Join(output, "reuse-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != publictestreuse.DecisionUnknown || report.Reason != "EXECUTION_PACKAGE_INPUTS_UNOBSERVED" ||
		report.CaseID != publictestreuse.CaseStaleEvidence || report.ReceiptMisses != 1 ||
		report.TestExecutions != 0 || report.ReusedTestExecutions != 0 || report.BuildExecutions != 0 || report.ReceiptHits != 0 {
		t.Fatal("missing observations authorized execution or reuse")
	}
	if report.Unknown == nil || report.Unknown.NextOperation != publictestreuse.UnknownNextEvidence ||
		len(report.Unknown.BlockedBy) != 1 || report.Unknown.BlockedBy[0] != "execution_package_inputs" {
		t.Fatal("unresolved report lost its next operation or missing input domain")
	}
}
