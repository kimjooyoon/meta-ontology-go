package toolchainlsp

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/lsp/coupling"
)

func observeClosedCoupling(
	resolve func(coupling.Request) coupling.Result,
	envelope coupling.Envelope,
	request coupling.Request,
	result map[string]observation,
	stats *runtimeStats,
) error {
	unknown, err := couplingOutcome(envelope, coupling.OutcomeUnknown, coupling.ReasonUpstreamUnknown, request)
	if err != nil {
		return err
	}
	unknownOK := unknown.Outcome == coupling.OutcomeUnknown && len(unknown.Links) == 0 && unknown.Hover == nil
	result["coupling-upstream-unknown"] = observation{"UNKNOWN_NO_NAVIGATION", unknownOK}
	if !unknownOK {
		stats.UnknownLeaks++
	} else {
		stats.FailClosedPaths++
	}
	failure, err := couplingOutcome(envelope, coupling.OutcomeFailClosed, coupling.ReasonUpstreamFail, request)
	if err != nil {
		return err
	}
	failureOK := failure.Outcome == coupling.OutcomeFailClosed && len(failure.Links) == 0 && failure.Hover == nil
	result["coupling-upstream-fail-closed"] = observation{"FAIL_CLOSED_NO_NAVIGATION", failureOK}
	if !failureOK {
		stats.FailClosedLeaks++
	} else {
		stats.FailClosedPaths++
	}
	staleRequest := request
	staleRequest.SnapshotDigest = couplingDigest("stale")
	stale := resolve(staleRequest)
	staleOK := stale.Outcome == coupling.OutcomeUnknown && len(stale.Links) == 0 && stale.Hover == nil
	result["coupling-stale-snapshot"] = observation{"STALE_NO_NAVIGATION", staleOK}
	if !staleOK {
		stats.StaleLeaks++
	} else {
		stats.FailClosedPaths++
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	cancelRequest := request
	cancelRequest.Context = cancelled
	cancelResult := resolve(cancelRequest)
	cancelOK := cancelResult.Outcome == coupling.OutcomeUnknown && len(cancelResult.Links) == 0 && cancelResult.Hover == nil
	result["coupling-cancelled"] = observation{"CANCELLED_NO_NAVIGATION", cancelOK}
	if !cancelOK {
		stats.UnknownLeaks++
	} else {
		stats.FailClosedPaths++
	}
	return nil
}
