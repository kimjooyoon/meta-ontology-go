package main

import (
	"strings"
	"testing"
)

func TestArtifactDigestSpellingFresh1683(t *testing.T) {
	hexadecimal := strings.Repeat("a", 64)
	for _, row := range []struct {
		name, value string
		allowed     bool
	}{
		{"upload-output", hexadecimal, true},
		{"api-metadata", "sha256:" + hexadecimal, true},
		{"upper-case", strings.ToUpper(hexadecimal), false},
		{"short", hexadecimal[:62], false},
		{"malformed", strings.Repeat("z", 64), false},
		{"empty", "", false},
	} {
		t.Run(row.name, func(t *testing.T) {
			value, err := canonicalDigest(row.value)
			if (err == nil) != row.allowed || row.allowed && value != "sha256:"+hexadecimal {
				t.Fatalf("archive digest spelling: %q, %v", value, err)
			}
		})
	}
}
