package coupling

import (
	"context"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/query/couplingexplain"
	"strings"
	"testing"
)

// This is an independent production-query transcript. It is intentionally
// literal: the LSP tests do not construct or recompute detector, manifest,
// path, receipt, verifier, or envelope authority.
const literalVerifiedQueryEnvelope = `{"schema":"gooo-coupling-explanation/v1","binding":{"snapshot_digest":"16a0eeb0791b6c92451fd284dd9f599e0a7dbe7f6ebea6e2d2d06c7f74aec112","registry_digest":"872491a30d60d598962de6e7b834ab76b2aa65fbab102c6ebaaae6acdc238822","source_map_digest":"fd2e37dce67af09e686328c4633012b79c7677a63f0d4f91fa14cc54a0c1a28b","manifest_digest":"05b3abf2579a5eb66403cd78be557fd860633a1fe2103c7642030defe32c657f","toolchain_digest":"0db3de82a739e43a2b560d166d037c3c0061601bb194866eb79b2c87045d00f2","profile_digest":"1900eab6c028483d7126599ee6f50de0d27907b5c65fa90524580b4b0f9852b0","detector_input_digest":"decafd9d993a5e8078da48cd17041100b8be6775c0e5cdea3bbdb23215a22ff1","detector_result_digest":"d397b20d9fd27aaa9aebe25229a68563e3c77064c386670d8e5f069ccfafedc6","verifier_result_digest":"1678e16fce2d4a9fa3b9b6baf97207cefd6fb141ed5908c13e546fafe06f0745","envelope_digest":"","control":{"request_version":7,"observed_version":7,"request_cancellation_version":11,"observed_cancellation_version":11}},"code_binding":{"code_symbol_id":"code://billing/pay-order","semantic_owner_id":"owner://billing/pay-order","registered_surface_id":"surface://pay-order","source_map_id":"sourcemap://pay-order","binding_digest":"80f70afeef3caa57646fd20afb95be9c3f2c03d38e091906de31838813dcc22e","code_binding_digest":"16d0d29bf279a96f342aeff1e936a367889ae1fa54d83c376734fe679d148907"},"semantic_owner":"owner://billing/pay-order","term":{"term_id":"term://pay-order","semantic_owner_id":"owner://billing/pay-order","version":"v1","definition_digest":"9efc05819657dc7d15c80fbbfc904f59e45f8ff9b4df78be0a6677f0f75598e2"},"origin_path":{"path_id":"path://pay-order","start_id":"code://billing/pay-order","end_id":"evidence://coupling","step_count":3,"path_digest":"289791d1fab43470f301f2a023e1d6dad2a003a8077ddb71ebde9864a744a2dc","steps":[{"from_id":"code://billing/pay-order","to_id":"owner://billing/pay-order","kind":"DERIVED_PROJECTION","phase":"PROJECTION","phase_ordinal":1,"input_digest":"80f70afeef3caa57646fd20afb95be9c3f2c03d38e091906de31838813dcc22e","output_digest":"4c1029697ee358715d3a14a2add817c4b01651440de808371f78165ac90dc581"},{"from_id":"owner://billing/pay-order","to_id":"term://pay-order","kind":"AUTHORITATIVE_DECLARATION","phase":"DECLARATION","phase_ordinal":2,"input_digest":"4c1029697ee358715d3a14a2add817c4b01651440de808371f78165ac90dc581","output_digest":"9efc05819657dc7d15c80fbbfc904f59e45f8ff9b4df78be0a6677f0f75598e2"},{"from_id":"term://pay-order","to_id":"evidence://coupling","kind":"INDEPENDENT_VERIFICATION","phase":"VERIFICATION","phase_ordinal":3,"input_digest":"9efc05819657dc7d15c80fbbfc904f59e45f8ff9b4df78be0a6677f0f75598e2","output_digest":"ee8250fb76e094b34b471f13a73dbbe51d1ae142e9df59d7c0d31ec20f0a0a8e","evidence_ref":"evidence://coupling"}]},"receipt":{"receipt_id":"receipt://pay-order","surface_id":"surface://pay-order","change_claim":"DELTA","receipt_kind":"SEMANTIC_DELTA","before_ir_digest":"6db7d803e74f1ffa7d8f5adc0bf95b3e15bf4c8373fffadf546227cc6c6742cb","after_ir_digest":"f39592393ef0859cb196a52693d2cea00fb2df784b3c04ae54aa7cadb8e562f8","canonical_delta":"owner=billing://pay-order relation=used","delta_digest":"f083f2025689f6b547605d267145bd5469caed9e88a3a54fc2de5af495fda744","receipt_digest":"67389bf778c30affeb56f6c691c889711e3c8f4d280a1cc7f971aae544865757","origin_path_id":"path://pay-order","evidence_refs":["evidence://coupling"]},"verifier":{"evidence_id":"evidence://coupling","receipt_id":"receipt://pay-order","state":"PASS","independent":true,"evidence_digest":"ee8250fb76e094b34b471f13a73dbbe51d1ae142e9df59d7c0d31ec20f0a0a8e","verifier_digest":"96254158ba10b70351e9959fa5782148e3e1990e49f9ec3d2174404b60b3065e","evidence_refs":["path://pay-order"]},"verdict":"VERIFIED","evidence_digest":"ee8250fb76e094b34b471f13a73dbbe51d1ae142e9df59d7c0d31ec20f0a0a8e","envelope_digest":"114cf323cd5eaabe2834ce8277fb04ed275da3db5054fe27f5b98ea1e7e4b33f"}`
const (
	liveSnapshotDigest  = "16a0eeb0791b6c92451fd284dd9f599e0a7dbe7f6ebea6e2d2d06c7f74aec112"
	liveSourceMapDigest = "fd2e37dce67af09e686328c4633012b79c7677a63f0d4f91fa14cc54a0c1a28b"
	liveManifestDigest  = "05b3abf2579a5eb66403cd78be557fd860633a1fe2103c7642030defe32c657f"
	liveToolchainDigest = "0db3de82a739e43a2b560d166d037c3c0061601bb194866eb79b2c87045d00f2"
	liveProfileDigest   = "1900eab6c028483d7126599ee6f50de0d27907b5c65fa90524580b4b0f9852b0"
	liveDetectorInput   = "decafd9d993a5e8078da48cd17041100b8be6775c0e5cdea3bbdb23215a22ff1"
	liveDetectorResult  = "d397b20d9fd27aaa9aebe25229a68563e3c77064c386670d8e5f069ccfafedc6"
	liveVerifierResult  = "1678e16fce2d4a9fa3b9b6baf97207cefd6fb141ed5908c13e546fafe06f0745"
	liveEnvelopeDigest  = "114cf323cd5eaabe2834ce8277fb04ed275da3db5054fe27f5b98ea1e7e4b33f"
)

func liveRequest() LiveRequest {
	control := couplingexplain.Control{RequestVersion: 7, ObservedVersion: 7, RequestCancellationVersion: 11, ObservedCancellationVersion: 11}
	query := couplingexplain.Request{CodeSymbolID: "code://billing/pay-order", SnapshotDigest: liveSnapshotDigest, RegistryDigest: "872491a30d60d598962de6e7b834ab76b2aa65fbab102c6ebaaae6acdc238822", SourceMapDigest: liveSourceMapDigest, ManifestDigest: liveManifestDigest, ToolchainDigest: liveToolchainDigest, ProfileDigest: liveProfileDigest, DetectorInputDigest: liveDetectorInput, DetectorResultDigest: liveDetectorResult, VerifierResultDigest: liveVerifierResult, EnvelopeDigest: liveEnvelopeDigest, Control: control}
	return LiveRequest{Context: context.Background(), DocumentURI: "file:///workspace/billing.gooo", DocumentVersion: 7, Position: Position{Line: 0, Character: 3}, SnapshotDigest: liveSnapshotDigest, Query: query, Locations: LocationSnapshot{SnapshotDigest: liveSnapshotDigest, DocumentURI: "file:///workspace/billing.gooo", DocumentVersion: 7, Locations: []SourceLocation{
		{StableID: "code://billing/pay-order", SourceMapID: "sourcemap://pay-order", SourceMapDigest: liveSourceMapDigest, URI: "file:///workspace/billing.gooo", Range: Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 0, Character: 8}}},
		{StableID: "owner://billing/pay-order", SourceMapID: "sourcemap://pay-order", SourceMapDigest: liveSourceMapDigest, URI: "file:///workspace/model.gooo", Range: Range{Start: Position{Line: 1, Character: 0}, End: Position{Line: 1, Character: 8}}, Message: "Verified semantic owner."},
		{StableID: "term://pay-order", SourceMapID: "sourcemap://pay-order", SourceMapDigest: liveSourceMapDigest, URI: "file:///workspace/model.gooo", Range: Range{Start: Position{Line: 2, Character: 0}, End: Position{Line: 2, Character: 8}}},
		{StableID: "evidence://coupling", SourceMapID: "sourcemap://pay-order", SourceMapDigest: liveSourceMapDigest, URI: "file:///workspace/evidence.json", Range: Range{Start: Position{Line: 3, Character: 0}, End: Position{Line: 3, Character: 8}}},
	}}}
}
func liveAdapter(t *testing.T) *LiveAdapter {
	t.Helper()
	adapter, err := NewLiveAdapter([]byte(literalVerifiedQueryEnvelope))
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}
func TestLiveQueryPassProducesOnlyStandardLSPValues(t *testing.T) {
	result := liveAdapter(t).Resolve(liveRequest())
	if result.Outcome != OutcomePass || len(result.Links) != 1 || result.Hover == nil || len(result.Diagnostics) != 1 {
		t.Fatalf("live result = %#v", result)
	}
	if len(result.Diagnostics[0].RelatedInformation) == 0 {
		t.Fatal("verified causal locations were not emitted as related information")
	}
	if result.Links[0].TargetURI != "file:///workspace/model.gooo" || result.Links[0].TargetRange.Start.Character != 0 {
		t.Fatalf("standard location link = %#v", result.Links[0])
	}
	for _, value := range []any{result.Links[0], *result.Hover, result.Diagnostics[0]} {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "stable_id") || strings.Contains(string(data), "source_map_digest") {
			t.Fatalf("custom identity leaked to LSP wire: %s", data)
		}
	}
}
func TestLiveQueryMissingProductionLocationsWithholdsLink(t *testing.T) {
	request := liveRequest()
	request.Locations.Locations = nil
	result := liveAdapter(t).Resolve(request)
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Diagnostics[0].Code != DiagnosticLiveMissingLocations {
		t.Fatalf("missing locations result = %#v", result)
	}
}
