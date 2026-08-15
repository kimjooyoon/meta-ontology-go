package verify

import "testing"

func couplingControlFixtures(t *testing.T) []couplingFixtureCase {
	return []couplingFixtureCase{
		{"controls mismatch", "DELTA", CouplingDecisionFailClosed, "controls-mismatch", func(in *CouplingInput) {
			wrongPolicy := couplingDigest("wrong-policy")
			in.Receipts[0].Path.Edges[2].Controls.PolicyDigest = wrongPolicy
			in.Receipts[0].Path.Evidence[2].Controls.PolicyDigest = wrongPolicy
			refreshCouplingReceipt(t, &in.Receipts[0])
		}},
	}
}
