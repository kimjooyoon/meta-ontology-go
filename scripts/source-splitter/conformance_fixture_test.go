package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	conformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/operationconformance"
)

const splitConformanceSource = `//go:build linux && amd64

package fixture

import "fmt"

func first() string {
	return fmt.Sprint("a")
}

func second() int {
	return 2
}
`

func conformanceEvidence(t *testing.T) ([]byte, conformance.SplitGoEvidence) {
	t.Helper()
	root, subject := t.TempDir(), "fixture_linux_amd64_test.go"
	before := []byte(splitConformanceSource)
	if err := os.WriteFile(filepath.Join(root, subject), before, 0o600); err != nil {
		t.Fatal(err)
	}
	plan, err := planSource(root, subject, 10)
	if err != nil {
		t.Fatal(err)
	}
	var observed []splitEvent
	if err := applySplitObserved(plan, func(event splitEvent) { observed = append(observed, event) }); err != nil {
		t.Fatal(err)
	}
	candidates := make([]conformance.FileEvidence, len(plan.Parts))
	for index, part := range plan.Parts {
		data, readErr := os.ReadFile(part.Path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		candidates[index] = conformance.FileEvidence{Path: part.Subject, Data: data}
	}
	events := normalizeSplitEvents(t, root, observed)
	targets := make([]string, len(candidates))
	for index := range candidates {
		targets[index] = candidates[index].Path
	}
	contract, err := os.ReadFile("../../examples/source-splitter-conformance/contract.json")
	if err != nil {
		t.Fatal(err)
	}
	evidence := conformance.SplitGoEvidence{ExpectedHeadSHA: strings.Repeat("a", 40),
		OperationID: conformance.OperationID, EvidenceComplete: true,
		Source: conformance.FileEvidence{Path: subject, Data: before}, Candidates: candidates,
		BuildContexts: []conformance.BuildContext{{GOOS: "linux", GOARCH: "amd64"}, {GOOS: "windows", GOARCH: "amd64"}},
		Write: conformance.WriteReceipt{Complete: true, ExecutionSucceeded: true,
			DeclaredTargets: targets, Events: events}}
	return contract, evidence
}
