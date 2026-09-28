package packageexecution

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteMultiFilePackageAndReplay(t *testing.T) {
	sources, err := LoadDirectory(filepath.Join("..", "..", "..", "examples", "billing-package"))
	if err != nil {
		t.Fatal(err)
	}
	request := Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: sources}
	first := Execute(request)
	second := Execute(request)
	if first.Scope != "DECLARATION_RESOLUTION_ONLY" || first.Execution == nil || first.Execution.Scope != first.Scope || first.Decision != "PASS" || first.Reason != "PACKAGE_EXECUTED" {
		t.Fatalf("decision=%s reason=%s", first.Decision, first.Reason)
	}
	if len(first.Sources) != 2 || first.Digest != second.Digest {
		t.Fatalf("sources=%d replay_equal=%t", len(first.Sources), first.Digest == second.Digest)
	}
	if _, err := Marshal(first); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteRejectsHeaderMismatch(t *testing.T) {
	sources, err := LoadDirectory(filepath.Join("..", "..", "..", "examples", "billing-package"))
	if err != nil {
		t.Fatal(err)
	}
	sources[0].Content = strings.Replace(sources[0].Content, "package billing", "package other", 1)
	receipt := Execute(Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: sources})
	if receipt.Scope != "DECLARATION_RESOLUTION_ONLY" || receipt.Decision != "FAIL_CLOSED" || receipt.Reason != "PACKAGE_HEADER_MISMATCH" || receipt.Resolution != "EXACT" {
		t.Fatalf("decision=%s reason=%s resolution=%s", receipt.Decision, receipt.Reason, receipt.Resolution)
	}
}

func TestValidateRejectsNestedScopeDisagreement(t *testing.T) {
	cases := []struct {
		name  string
		scope string
	}{
		{name: "registered-value-operation", scope: "REGISTERED_VALUE_OPERATION"},
	}
	const declaredCaseCount = 1
	if len(cases) != declaredCaseCount {
		t.Fatalf("declared package scope regression cases = %d, want %d", len(cases), declaredCaseCount)
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			sources, err := LoadDirectory(filepath.Join("..", "..", "..", "examples", "billing-package"))
			if err != nil {
				t.Fatal(err)
			}
			receipt := Execute(Request{PackagePath: "billing-package", Entry: "PayOrder", Sources: sources})
			if receipt.Execution == nil {
				t.Fatal("positive package receipt has no nested execution")
			}
			receipt.Execution.Scope = test.scope
			if err := Validate(receipt); err == nil || err.Error() != "packageexecution: nested execution scope mismatch" {
				t.Fatalf("nested scope validation error = %v, want packageexecution: nested execution scope mismatch", err)
			}
		})
	}
}
