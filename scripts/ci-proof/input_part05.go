package main

import (
	"strings"
)

func validEventRef(event, ref string) bool {
	if ref == "" || strings.ContainsAny(ref, "\r\n") {
		return false
	}
	if event == "pull_request" {
		return strings.HasPrefix(ref, "refs/pull/") && strings.HasSuffix(ref, "/merge")
	}
	if event == "push" {
		return strings.HasPrefix(ref, "refs/heads/")
	}
	return false
}
func validDigest(value string) bool {
	if len(value) != 64 || value == strings.Repeat("0", 64) {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
func validArtifactDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:") && validDigest(strings.TrimPrefix(value, "sha256:"))
}
