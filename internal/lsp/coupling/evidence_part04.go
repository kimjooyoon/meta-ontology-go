package coupling

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

// ComputeEvidenceDigest canonically binds the non-presentation fields of the
// explanation to all upstream result and toolchain digests. URIs, labels, and
// messages are deliberately excluded: relocation and presentation renames do
// not change immutable evidence identity.
func ComputeEvidenceDigest(envelope Envelope) (string, error) {
	type digestTuple struct {
		Snapshot, Registry, Toolchain, Profile, Detector, Oracle string
	}
	type locationTuple struct {
		StableID, SourceMapID, SourceMapDigest string
		Range                                  Range
	}
	type spanTuple struct {
		StableID, SourceMapID, SourceMapDigest string
		Range                                  Range
		Ordinal                                int
	}
	type explanationTuple struct {
		CodeSymbolID, SemanticOwnerID string
		Origin, Target                locationTuple
		CausalSpans                   []spanTuple
		Claim                         ChangeClaim
		Status                        Outcome
		Reason                        Reason
	}
	type evidenceTuple struct {
		Schema          string
		Digests         digestTuple
		DocumentVersion int
		Status          Outcome
		Reason          Reason
		Explanations    []explanationTuple
	}
	result := evidenceTuple{Schema: envelope.Schema, Digests: digestTuple{
		Snapshot: envelope.SnapshotDigest, Registry: envelope.RegistryDigest,
		Toolchain: envelope.ToolchainDigest, Profile: envelope.ProfileDigest,
		Detector: envelope.DetectorResultDigest, Oracle: envelope.OracleResultDigest,
	}, DocumentVersion: envelope.Document.Version, Status: envelope.Status, Reason: envelope.Reason}
	for _, explanation := range envelope.Explanations {
		value := explanationTuple{CodeSymbolID: explanation.CodeSymbolID, SemanticOwnerID: explanation.SemanticOwnerID,
			Origin: locationTuple{StableID: explanation.Origin.StableID, SourceMapID: explanation.Origin.SourceMapID, SourceMapDigest: explanation.Origin.SourceMapDigest, Range: explanation.Origin.Range},
			Target: locationTuple{StableID: explanation.Target.StableID, SourceMapID: explanation.Target.SourceMapID, SourceMapDigest: explanation.Target.SourceMapDigest, Range: explanation.Target.Range},
			Claim:  explanation.Claim, Status: explanation.Status, Reason: explanation.Reason}
		for _, span := range explanation.CausalSpans {
			value.CausalSpans = append(value.CausalSpans, spanTuple{StableID: span.StableID, SourceMapID: span.SourceMapID, SourceMapDigest: span.SourceMapDigest, Range: span.Range, Ordinal: span.Ordinal})
		}
		sort.Slice(value.CausalSpans, func(i, j int) bool {
			if value.CausalSpans[i].Ordinal != value.CausalSpans[j].Ordinal {
				return value.CausalSpans[i].Ordinal < value.CausalSpans[j].Ordinal
			}
			return value.CausalSpans[i].StableID < value.CausalSpans[j].StableID
		})
		result.Explanations = append(result.Explanations, value)
	}
	sort.Slice(result.Explanations, func(i, j int) bool { return result.Explanations[i].CodeSymbolID < result.Explanations[j].CodeSymbolID })
	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return semantic.StableHashString(string(data)), nil
}
