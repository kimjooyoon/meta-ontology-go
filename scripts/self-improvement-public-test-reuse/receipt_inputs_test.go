package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
)

func TestUnobservedLegacyReceiptProducesUnresolvedReport(t *testing.T) {
	output := t.TempDir()
	initial := publictestreuse.Report{Schema: publictestreuse.ReportSchema, Operation: publictestreuse.ReplayOperation}
	initial.Binding.PackageInputs = publictestreuse.ExecutionPackageInputsFromBytes(
		[]byte("module fixture\n"), []byte("package fixture\n"), []byte("package fixture\n// test\n"))
	policy := readGoooPolicyFixture(t)
	initial.Binding.PolicySourceDigest = policy.SourceDigest
	initial.Binding.PolicyEvaluatorDigest = policy.EvaluatorDigest
	observed := publictestreuse.ReuseObservation{Requested: true, ReceiptState: "LEGACY_INPUTS_UNOBSERVED",
		InputProfile: initial.Binding.PackageInputs.Profile}
	if _, err := assessReplayReport(policy, observed, &initial); err != nil {
		t.Fatal(err)
	}
	initial.ReceiptMisses = 1
	if err := writeReport(output, initial, "Historical input observations remain unresolved."); err != nil {
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
	raw, err := os.ReadFile(filepath.Join(output, "reuse-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"trailing":  append(append([]byte(nil), raw...), []byte("\n{}")...),
		"duplicate": bytes.Replace(raw, []byte("{"), []byte(`{"schema":"duplicate",`), 1),
	} {
		t.Run(name, func(t *testing.T) {
			filename := filepath.Join(output, name+".json")
			if err := os.WriteFile(filename, data, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := readPublicReport(filename); err == nil {
				t.Fatal("ambiguous report was accepted")
			}
		})
	}
}

func readGoooPolicyFixture(t *testing.T) publictestreuse.Policy {
	t.Helper()
	filename := "../../examples/self-improvement-discovery/project.gooo"
	source, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := publictestreuse.Load(filename, source)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}
