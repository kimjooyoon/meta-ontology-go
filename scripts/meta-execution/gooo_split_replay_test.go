package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestGoooSplitReplayRetainsContentAcrossDisposableAddresses(t *testing.T) {
	var first []byte
	var firstRawDigest string
	for _, root := range []string{"/tmp/first-copy", "/tmp/second-copy"} {
		proposal, _ := json.Marshal(map[string]string{"manifest_path": root + "/workspace.json", "source": "same input"})
		rawDigest := digestBytes(proposal)
		receipt, _ := json.Marshal(map[string]any{"manifest_path": root + "/workspace.json", "source": "same output",
			"application": map[string]string{"state": "OUTPUT_VERIFIED", "proposal_sha256": rawDigest}})
		canonical, err := canonicalGoooSplitReceipt(root, receipt, digestBytes(canonicalProcessBytes(root, proposal)))
		if err != nil {
			t.Fatal(err)
		}
		if first == nil {
			first, firstRawDigest = canonical, rawDigest
			continue
		}
		if rawDigest == firstRawDigest || !bytes.Equal(first, canonical) {
			t.Fatal("runtime address affected semantic replay or raw audit identity was lost")
		}
		changed := bytes.Replace(receipt, []byte("same output"), []byte("changed output"), 1)
		next, err := canonicalGoooSplitReceipt(root, changed, digestBytes(canonicalProcessBytes(root, proposal)))
		if err != nil || bytes.Equal(first, next) {
			t.Fatal("normalization erased a real output change", err)
		}
	}
}
