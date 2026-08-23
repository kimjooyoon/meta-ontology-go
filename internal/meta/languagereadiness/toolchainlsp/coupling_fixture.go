package toolchainlsp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/lsp/coupling"
)

func couplingDigest(value string) string {
	digest := sha256.Sum256([]byte("toolchain-lsp/" + value))
	return hex.EncodeToString(digest[:])
}

func couplingEnvelope() coupling.Envelope {
	uri := "file:///workspace/main.go"
	return coupling.Envelope{
		Schema: coupling.SchemaVersion, SnapshotDigest: couplingDigest("snapshot"),
		RegistryDigest: couplingDigest("registry"), ToolchainDigest: couplingDigest("toolchain"),
		ProfileDigest: couplingDigest("profile"), DetectorResultDigest: couplingDigest("detector"),
		OracleResultDigest: couplingDigest("oracle"), Document: coupling.Document{URI: uri, Version: 7},
		Status: coupling.OutcomePass, Explanations: []coupling.Explanation{{
			CodeSymbolID: "stable://code/pay-order", SemanticOwnerID: "stable://activity/pay-order", Label: "PayOrder",
			Origin: coupling.BoundLocation{StableID: "stable://span/code", SourceMapID: "stable://map/code", SourceMapDigest: couplingDigest("map-code"), URI: uri, Range: coupling.Range{Start: coupling.Position{Line: 4, Character: 8}, End: coupling.Position{Line: 4, Character: 17}}},
			Target: coupling.BoundLocation{StableID: "stable://span/model", SourceMapID: "stable://map/model", SourceMapDigest: couplingDigest("map-model"), URI: "file:///workspace/model.gooo", Range: coupling.Range{Start: coupling.Position{Line: 8, Character: 9}, End: coupling.Position{Line: 8, Character: 17}}},
			CausalSpans: []coupling.CausalSpan{{StableID: "stable://span/source", SourceMapID: "stable://map/source", SourceMapDigest: couplingDigest("map-source"), URI: uri, Range: coupling.Range{Start: coupling.Position{Line: 1}, End: coupling.Position{Line: 1, Character: 10}}, Ordinal: 1}},
			Claim: coupling.ClaimDelta, Status: coupling.OutcomePass,
		}},
	}
}

func couplingBytes(envelope coupling.Envelope) ([]byte, error) {
	digest, err := coupling.ComputeEvidenceDigest(envelope)
	if err != nil { return nil, err }
	envelope.EvidenceDigest = digest
	return json.Marshal(envelope)
}
