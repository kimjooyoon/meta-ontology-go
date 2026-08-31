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
	if err != nil {
		return nil, stats, err
	}
	adapter, err := coupling.New(raw)
	if err != nil {
		return nil, stats, err
	}
	request := coupling.Request{Context: context.Background(), DocumentURI: envelope.Document.URI,
		DocumentVersion: envelope.Document.Version, Position: coupling.Position{Line: 4, Character: 10}, SnapshotDigest: envelope.SnapshotDigest}
	pass := adapter.Resolve(request)
	passOK := pass.Outcome == coupling.OutcomePass && len(pass.Links) == 1 && pass.Hover != nil
	result["coupling-pass"] = observation{"ONE_STANDARD_LINK", passOK}
	if passOK {
		stats.NavigationPaths++
	}
	wire, _ := json.Marshal(pass)
	for _, field := range []string{"stable_id", "code_symbol_id", "semantic_owner_id"} {
		if strings.Contains(string(wire), field) {
			stats.NonstandardWireFields++
		}
	}
	if err := observeClosedCoupling(adapter.Resolve, envelope, request, result, &stats); err != nil {
		return nil, stats, err
	}
	original := append([]byte(nil), raw...)
	raw[0] ^= 1
	immutable := bytes.Equal(adapter.RawBytes(), original) && adapter.Resolve(request).Outcome == coupling.OutcomePass
	result["coupling-input-immutability"] = observation{"CALLER_BYTES_ISOLATED", immutable}
	return result, stats, nil
}
