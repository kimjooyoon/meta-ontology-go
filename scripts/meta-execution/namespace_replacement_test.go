package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func namespaceReplacementFixture(t *testing.T) (string, extractorSubject, namespaceReplacementReceipt) {
	t.Helper()
	root := t.TempDir()
	logical := "subject.go"
	data := []byte("package p\n")
	if err := os.WriteFile(filepath.Join(root, logical), data, 0o644); err != nil {
		t.Fatal(err)
	}
	observed := extractorSubject{Logical: logical, Files: []string{logical}}
	digest := digestBytes(data)
	replacement := namespaceReplacementReceipt{
		LogicalPath: logical, Primitive: "os.Rename",
		Contract: linuxNamespaceReplacementContract, GOOS: runtime.GOOS,
		SameDirectory: true, DestinationPreexisted: true,
		TempDigest: digest, ReplacementSuccess: true, FinalDigest: digest,
	}
	return root, observed, replacement
}

func TestMissingNamespaceReplacementStaysUnknown(t *testing.T) {
	root, observed, _ := namespaceReplacementFixture(t)
	pass, err := validateNamespaceReplacements(root, observed, nil)
	if err != nil || pass {
		t.Fatalf("missing receipt was not left unknown: pass=%v err=%v", pass, err)
	}
}

func TestMalformedNamespaceReplacementIsRefuted(t *testing.T) {
	root, observed, replacement := namespaceReplacementFixture(t)
	replacement.ReplacementSuccess = false
	assertNamespaceReplacementReason(t, root, observed, []namespaceReplacementReceipt{replacement}, "NAMESPACE_REPLACEMENT_MALFORMED")
}

func assertNamespaceReplacementReason(t *testing.T, root string, observed extractorSubject, replacements []namespaceReplacementReceipt, want string) {
	t.Helper()
	_, err := validateNamespaceReplacements(root, observed, replacements)
	if err == nil {
		t.Fatalf("replacement unexpectedly passed; want %s", want)
	}
	replacementErr, ok := err.(*namespaceReplacementError)
	if !ok || replacementErr.reason != want {
		t.Fatalf("replacement reason=%v, want %s", err, want)
	}
}
