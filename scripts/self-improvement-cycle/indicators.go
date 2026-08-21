package main

import "fmt"

func cycleIndicators(check validation, envelope Envelope, replay bool) []Indicator {
	return []Indicator{
		{ID: "foundation.artifact-schemas", Route: "FOUNDATION",
			Verdict: verdict(check.Schemas), Relation: "=", Value: fmt.Sprint(check.Schemas), Limit: "true"},
		{ID: "foundation.ci-run-binding", Route: "FOUNDATION",
			Verdict: verdict(check.Context), Relation: "=", Value: fmt.Sprint(check.Context), Limit: "true"},
		{ID: "foundation.exact-head-binding", Route: "FOUNDATION",
			Verdict: verdict(check.Heads), Relation: "=", Value: fmt.Sprint(check.Heads), Limit: "true"},
		{ID: "foundation.gooo-contract-binding", Route: "FOUNDATION",
			Verdict: verdict(check.Contract), Relation: "=", Value: fmt.Sprint(check.Contract), Limit: "true"},
		{ID: "coherence.artifact-state", Route: "COHERENCE",
			Verdict: verdict(check.States), Relation: "=", Value: fmt.Sprint(check.States), Limit: "true"},
		{ID: "coherence.plan-execution-receipt-provenance", Route: "COHERENCE",
			Verdict: verdict(check.Links), Relation: "=", Value: fmt.Sprint(check.Links), Limit: "true"},
		{ID: "coherence.indicator-ledger-binding", Route: "COHERENCE",
			Verdict: verdict(check.Ledger), Relation: "=", Value: fmt.Sprint(check.Ledger), Limit: "true"},
		{ID: "coherence.content-addressed-cycle", Route: "COHERENCE",
			Verdict: verdict(check.Digests), Relation: "sha256",
			Value: envelope.ArtifactSetDigest, Limit: "bound"},
		{ID: "regression.canonical-replay", Route: "REGRESSION",
			Verdict: verdict(replay), Relation: "=", Value: fmt.Sprint(replay), Limit: "true"},
	}
}

func finishEnvelope(envelope *Envelope) {
	envelope.Status, envelope.Reason = "BOUND", "SELF_IMPROVEMENT_CYCLE_BOUND"
	for _, indicator := range envelope.Indicators {
		if indicator.Verdict != "PASS" {
			envelope.Status, envelope.Reason = "OPEN", "SELF_IMPROVEMENT_CYCLE_OPEN"
			return
		}
	}
}

func verdict(pass bool) string {
	if pass {
		return "PASS"
	}
	return "FAIL"
}
