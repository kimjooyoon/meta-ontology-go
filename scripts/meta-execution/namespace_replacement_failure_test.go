package main

import "testing"

func TestDuplicateNamespaceReplacementIsRefuted(t *testing.T) {
	root, observed, replacement := namespaceReplacementFixture(t)
	assertNamespaceReplacementReason(t, root, observed, []namespaceReplacementReceipt{replacement, replacement}, "NAMESPACE_REPLACEMENT_DUPLICATE")
}

func TestCrossSubjectNamespaceReplacementIsRefuted(t *testing.T) {
	root, observed, replacement := namespaceReplacementFixture(t)
	replacement.LogicalPath = "other.go"
	assertNamespaceReplacementReason(t, root, observed, []namespaceReplacementReceipt{replacement}, "NAMESPACE_REPLACEMENT_CROSS_SUBJECT")
}

func TestDigestMismatchNamespaceReplacementIsRefuted(t *testing.T) {
	root, observed, replacement := namespaceReplacementFixture(t)
	replacement.FinalDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	assertNamespaceReplacementReason(t, root, observed, []namespaceReplacementReceipt{replacement}, "NAMESPACE_REPLACEMENT_DIGEST_MISMATCH")
}

func TestUnsupportedGOOSNamespaceReplacementIsRefuted(t *testing.T) {
	root, observed, replacement := namespaceReplacementFixture(t)
	replacement.GOOS = "darwin"
	assertNamespaceReplacementReason(t, root, observed, []namespaceReplacementReceipt{replacement}, "NAMESPACE_REPLACEMENT_MALFORMED")
}
