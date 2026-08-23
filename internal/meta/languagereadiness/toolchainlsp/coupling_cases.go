package toolchainlsp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/lsp/coupling"
)

func observeCoupling() (map[string]observation, runtimeStats, error) {
	result := map[string]observation{}
	stats := runtimeStats{}
	envelope := couplingEnvelope()
	raw, err := couplingBytes(envelope)
	if err != nil { return nil, stats, err }
	adapter, err := coupling.New(raw)
	if err != nil { return nil, stats, err }
	request := coupling.Request{Context: context.Background(), DocumentURI: envelope.Document.URI,
		DocumentVersion: envelope.Document.Version, Position: coupling.Position{Line: 4, Character: 10}, SnapshotDigest: envelope.SnapshotDigest}
	pass := adapter.Resolve(request)
	passOK := pass.Outcome == coupling.OutcomePass && len(pass.Links) == 1 && pass.Hover != nil
	result["coupling-pass"] = observation{"ONE_STANDARD_LINK", passOK}
	if passOK { stats.NavigationPaths++ }
	wire, _ := json.Marshal(pass)
	for _, field := range []string{"stable_id", "code_symbol_id", "semantic_owner_id"} { if strings.Contains(string(wire), field) { stats.NonstandardWireFields++ } }
	unknown, err := couplingOutcome(envelope, coupling.OutcomeUnknown, coupling.ReasonUpstreamUnknown, request)
	if err != nil { return nil, stats, err }
	unknownOK := unknown.Outcome == coupling.OutcomeUnknown && len(unknown.Links) == 0 && unknown.Hover == nil
	result["coupling-upstream-unknown"] = observation{"UNKNOWN_NO_NAVIGATION", unknownOK}
	if !unknownOK { stats.UnknownLeaks++ } else { stats.FailClosedPaths++ }
	failure, err := couplingOutcome(envelope, coupling.OutcomeFailClosed, coupling.ReasonUpstreamFail, request)
	if err != nil { return nil, stats, err }
	failureOK := failure.Outcome == coupling.OutcomeFailClosed && len(failure.Links) == 0 && failure.Hover == nil
	result["coupling-upstream-fail-closed"] = observation{"FAIL_CLOSED_NO_NAVIGATION", failureOK}
	if !failureOK { stats.FailClosedLeaks++ } else { stats.FailClosedPaths++ }
	staleRequest := request; staleRequest.SnapshotDigest = couplingDigest("stale")
	stale := adapter.Resolve(staleRequest)
	staleOK := stale.Outcome == coupling.OutcomeUnknown && len(stale.Links) == 0 && stale.Hover == nil
	result["coupling-stale-snapshot"] = observation{"STALE_NO_NAVIGATION", staleOK}
	if !staleOK { stats.StaleLeaks++ } else { stats.FailClosedPaths++ }
	cancelled, cancel := context.WithCancel(context.Background()); cancel()
	cancelRequest := request; cancelRequest.Context = cancelled
	cancelResult := adapter.Resolve(cancelRequest)
	cancelOK := cancelResult.Outcome == coupling.OutcomeUnknown && len(cancelResult.Links) == 0 && cancelResult.Hover == nil
	result["coupling-cancelled"] = observation{"CANCELLED_NO_NAVIGATION", cancelOK}
	if !cancelOK { stats.UnknownLeaks++ } else { stats.FailClosedPaths++ }
	original := append([]byte(nil), raw...); raw[0] ^= 1
	immutable := bytes.Equal(adapter.RawBytes(), original) && adapter.Resolve(request).Outcome == coupling.OutcomePass
	result["coupling-input-immutability"] = observation{"CALLER_BYTES_ISOLATED", immutable}
	return result, stats, nil
}

func couplingOutcome(envelope coupling.Envelope, status coupling.Outcome, reason coupling.Reason, request coupling.Request) (coupling.Result, error) {
	envelope.Status, envelope.Reason = status, reason
	raw, err := couplingBytes(envelope)
	if err != nil { return coupling.Result{}, err }
	adapter, err := coupling.New(raw)
	if err != nil { return coupling.Result{}, err }
	return adapter.Resolve(request), nil
}
