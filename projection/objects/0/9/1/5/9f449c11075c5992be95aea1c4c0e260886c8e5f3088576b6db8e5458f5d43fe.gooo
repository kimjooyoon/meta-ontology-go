package main

import (
	"strings"
	"testing"
)

func TestProofDigestMismatchNamesEveryComparedField(t *testing.T) {
	evidence := evidenceInput{Digests: evidenceDigests{Source: "source-expected", IR: "semantic-expected", Generated: "projection-expected", Policy: "policy-expected", Toolchain: "toolchain-expected"}}
	mismatches := digestMismatchFields("source-actual", "semantic-actual", "projection-actual", "policy-actual", "toolchain-actual", evidence)
	if got := strings.Join(mismatches, ","); got != "source,semantic,projection,policy,toolchain" {
		t.Fatalf("digest mismatch fields = %q", got)
	}
}
