package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const linuxNamespaceReplacementContract = "same-directory-temp-over-destination/linux-v1"

type backupCleanupObservation struct {
	Status    string `json:"status"`
	Attempted int    `json:"attempted"`
	Removed   int    `json:"removed"`
	Failures  int    `json:"failures"`
}

type namespaceReplacementError struct {
	reason string
}

func (err *namespaceReplacementError) Error() string {
	return err.reason
}

type namespaceReplacementReceipt struct {
	LogicalPath           string `json:"logical_path"`
	Primitive             string `json:"primitive"`
	Contract              string `json:"contract"`
	GOOS                  string `json:"goos"`
	SameDirectory         bool   `json:"same_directory"`
	DestinationPreexisted bool   `json:"destination_preexisted"`
	TempDigest            string `json:"temp_digest"`
	ReplacementSuccess    bool   `json:"replacement_success"`
	FinalDigest           string `json:"final_digest"`
}

func validateNamespaceReplacements(root string, observed extractorSubject, replacements []namespaceReplacementReceipt) (bool, error) {
	if len(replacements) == 0 {
		return false, nil
	}
	expected := make(map[string]bool, len(observed.Files))
	for _, logical := range observed.Files {
		expected[logical] = true
	}
	seen := make(map[string]bool, len(replacements))
	for _, replacement := range replacements {
		if !expected[replacement.LogicalPath] {
			return false, &namespaceReplacementError{reason: "NAMESPACE_REPLACEMENT_CROSS_SUBJECT"}
		}
		if seen[replacement.LogicalPath] {
			return false, &namespaceReplacementError{reason: "NAMESPACE_REPLACEMENT_DUPLICATE"}
		}
		seen[replacement.LogicalPath] = true
		if !validNamespaceReplacement(replacement, observed) {
			return false, &namespaceReplacementError{reason: "NAMESPACE_REPLACEMENT_MALFORMED"}
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(replacement.LogicalPath)))
		if err != nil || digestBytes(data) != replacement.FinalDigest || replacement.TempDigest != replacement.FinalDigest {
			return false, &namespaceReplacementError{reason: "NAMESPACE_REPLACEMENT_DIGEST_MISMATCH"}
		}
	}
	if len(seen) != len(expected) || !seen[observed.Logical] {
		return false, nil
	}
	return true, nil
}

func validNamespaceReplacement(replacement namespaceReplacementReceipt, observed extractorSubject) bool {
	if runtime.GOOS != "linux" || replacement.GOOS != runtime.GOOS || replacement.Primitive != "os.Rename" || replacement.Contract != linuxNamespaceReplacementContract || !replacement.SameDirectory || !replacement.ReplacementSuccess {
		return false
	}
	if !strings.HasPrefix(replacement.TempDigest, "sha256:") || !strings.HasPrefix(replacement.FinalDigest, "sha256:") {
		return false
	}
	created := containsString(observed.CreatedFiles, replacement.LogicalPath)
	return replacement.DestinationPreexisted == !created
}
