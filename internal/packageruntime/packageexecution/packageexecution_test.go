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
	if first.Decision != "PASS" || first.Reason != "PACKAGE_EXECUTED" {
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
	if receipt.Decision != "FAIL_CLOSED" || receipt.Reason != "PACKAGE_HEADER_MISMATCH" || receipt.Resolution != "EXACT" {
		t.Fatalf("decision=%s reason=%s resolution=%s", receipt.Decision, receipt.Reason, receipt.Resolution)
	}
}
