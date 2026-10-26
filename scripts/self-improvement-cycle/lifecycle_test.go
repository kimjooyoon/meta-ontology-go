package main

import "testing"

func TestCancelledSourceCycleRemainsInterrupted(t *testing.T) {
	opts, in := cycleFixture()
	opts.conclusion = "cancelled"
	envelope := buildEnvelope(in, opts)
	if envelope.Status != "INTERRUPTED" || envelope.Reason != "SOURCE_CI_CANCELLED" ||
		envelope.CIConclusion != "cancelled" || envelope.PromotionAuthorized || !envelope.recordValid {
		t.Fatalf("valid interrupted observations were treated as a broken cycle: status=%s reason=%s", envelope.Status, envelope.Reason)
	}
	for _, indicator := range envelope.Indicators {
		if indicator.Verdict != "PASS" {
			t.Fatal("interrupted observation has an unresolved binding", indicator)
		}
	}
	if next := buildEnvelope(in, opts); next.EnvelopeDigest != envelope.EnvelopeDigest || next.ReplayDigest != envelope.ReplayDigest {
		t.Fatal("interrupted observation did not replay deterministically")
	}
}

func TestCancellationCannotHideBrokenCycleBindings(t *testing.T) {
	for _, change := range []struct {
		name string
		edit func(*options, *inputs)
	}{
		{"source-head", func(o *options, _ *inputs) { o.headSHA = "unbound" }},
		{"run", func(o *options, _ *inputs) { o.runID = 0 }},
		{"branch", func(o *options, _ *inputs) { o.branch = "feature" }},
		{"plan-link", func(_ *options, i *inputs) { i.Execution.Value.PlanDigest = "changed" }},
		{"contract", func(_ *options, i *inputs) { i.Contract.Value.Status = "FAIL" }},
		{"authority", func(_ *options, i *inputs) { i.Contract.Value.PromotionAuthorized = true }},
		{"metric", func(_ *options, i *inputs) { i.Metrics.Value.Meta.Policy.ExemptProjectRootTopology = false }},
		{"ledger", func(_ *options, i *inputs) { i.Provenance.Value.IndicatorDecisionLedgerCount++ }},
	} {
		t.Run(change.name, func(t *testing.T) {
			opts, in := cycleFixture()
			opts.conclusion = "cancelled"
			change.edit(&opts, &in)
			got := buildEnvelope(in, opts)
			if got.Status != "OPEN" || got.recordValid || got.PromotionAuthorized || got.CIConclusion != "cancelled" {
				t.Fatal("cancellation hid an invalid observation", got.Status, got.Reason)
			}
		})
	}
}
