package main

import (
	"runtime"
	"testing"
)

func namespaceReasonForHost(reason string) string {
	if runtime.GOOS != "linux" {
		return "NAMESPACE_REPLACEMENT_MALFORMED"
	}
	return reason
}

func TestNamespaceReplacementRequiresSupportedHost(t *testing.T) {
	root, observed, replacement := namespaceReplacementFixture(t)
	pass, err := validateNamespaceReplacements(root, observed, []namespaceReplacementReceipt{replacement})
	if runtime.GOOS == "linux" {
		if err != nil || !pass {
			t.Fatalf("valid Linux receipt: pass=%v err=%v", pass, err)
		}
		return
	}
	if pass || err == nil || err.Error() != "NAMESPACE_REPLACEMENT_MALFORMED" {
		t.Fatalf("unsupported host: pass=%v err=%v", pass, err)
	}
}

func TestDuplicateNamespaceReplacementIsRefuted(t *testing.T) {
	root, observed, replacement := namespaceReplacementFixture(t)
	assertNamespaceReplacementReason(t, root, observed, []namespaceReplacementReceipt{replacement, replacement}, namespaceReasonForHost("NAMESPACE_REPLACEMENT_DUPLICATE"))
}

func TestCrossSubjectNamespaceReplacementIsRefuted(t *testing.T) {
	root, observed, replacement := namespaceReplacementFixture(t)
	replacement.LogicalPath = "other.go"
	assertNamespaceReplacementReason(t, root, observed, []namespaceReplacementReceipt{replacement}, "NAMESPACE_REPLACEMENT_CROSS_SUBJECT")
}

func TestDigestMismatchNamespaceReplacementIsRefuted(t *testing.T) {
	root, observed, replacement := namespaceReplacementFixture(t)
	replacement.FinalDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	assertNamespaceReplacementReason(t, root, observed, []namespaceReplacementReceipt{replacement}, namespaceReasonForHost("NAMESPACE_REPLACEMENT_DIGEST_MISMATCH"))
}

func TestUnsupportedGOOSNamespaceReplacementIsRefuted(t *testing.T) {
	root, observed, replacement := namespaceReplacementFixture(t)
	replacement.GOOS = "darwin"
	assertNamespaceReplacementReason(t, root, observed, []namespaceReplacementReceipt{replacement}, "NAMESPACE_REPLACEMENT_MALFORMED")
}
