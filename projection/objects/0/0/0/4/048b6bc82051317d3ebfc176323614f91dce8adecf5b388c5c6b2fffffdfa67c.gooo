package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSelectiveCIShadowProofAndLaneFailClosedPartitions(t *testing.T) {
	cases := []struct {
		name   string
		stage  string
		mutate func(*shadowFixture)
	}{
		{name: "proof fail closed", stage: "PROOF_FAIL_CLOSED", mutate: func(f *shadowFixture) {
			f.proofInput.CommandReceipts[0].Digest = shadowDigest("wrong-receipt")
			f.reencodeProof()
		}},
		{name: "proof unknown", stage: "PROOF_UNKNOWN", mutate: func(f *shadowFixture) {
			f.proofInput.InferencePath.Edges = f.proofInput.InferencePath.Edges[:2]
			f.reencodeProof()
		}},
		{name: "lane unknown", stage: "LANE_UNKNOWN", mutate: func(f *shadowFixture) { f.laneInput.BaseSHA = ""; f.reencodeLane() }},
		{name: "lane ineligible", stage: "LANE_INELIGIBLE", mutate: func(f *shadowFixture) { f.laneInput.ActiveLeaseCount = 1; f.reencodeLane() }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newShadowFixture(t)
			test.mutate(&fixture)
			output := runShadowFixture(t, fixture)
			if output.Status != "FULL_SUITE_FALLBACK" || output.Stage != test.stage || output.ExecutionAuthorized || !output.ShadowOnly {
				t.Fatalf("partition fallback = %#v", output)
			}
		})
	}
}
func runShadowFixture(t *testing.T, fixture shadowFixture) selectiveCIShadowOutput {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := runSelectiveCI(fixture.args(), fixture.reader(), &stdout, &stderr); code != exitOK {
		t.Fatalf("shadow code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
	var output selectiveCIShadowOutput
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &output); err != nil {
		t.Fatalf("decode shadow output: %v", err)
	}
	if output.CanonicalDigest == "" || output.CanonicalDigest != output.stableDigest() {
		t.Fatalf("invalid output digest: %#v", output)
	}
	return output
}
