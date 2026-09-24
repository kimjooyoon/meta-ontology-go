package valueexecution

import "testing"

func TestCompareReplayDoesNotReuseFailedLifecycleReceipt(t *testing.T) {
	plan, err := CompilePlan("replay-lifecycle.gooo", []byte(lifecycleFixture))
	if err != nil {
		t.Fatal(err)
	}
	failed, err := plan.Execute(map[string]int64{})
	if err == nil {
		t.Fatal("missing root input unexpectedly succeeded")
	}
	completed, err := plan.Execute(map[string]int64{"Produce": 41})
	if err != nil {
		t.Fatal(err)
	}
	comparison := CompareReplay(failed, completed)
	if comparison.State != ReplayUnknown || comparison.Reason != "REPLAY_RECEIPT_INCOMPLETE" {
		t.Fatalf("comparison = %#v, want failed lifecycle to remain unknown", comparison)
	}
}
