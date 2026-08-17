package coupling

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateEnvelope(envelope Envelope) error {
	if envelope.Schema != SchemaVersion {
		return fmt.Errorf("schema is %q, want %q", envelope.Schema, SchemaVersion)
	}
	for name, digest := range map[string]string{
		"snapshot": envelope.SnapshotDigest, "registry": envelope.RegistryDigest,
		"toolchain": envelope.ToolchainDigest, "profile": envelope.ProfileDigest,
		"detector result": envelope.DetectorResultDigest, "oracle result": envelope.OracleResultDigest,
		"evidence": envelope.EvidenceDigest,
	} {
		if err := validateDigest(digest); err != nil {
			return fmt.Errorf("%s digest: %w", name, err)
		}
	}
	if !exactText(envelope.Document.URI) || envelope.Document.Version <= 0 {
		return fmt.Errorf("document URI and positive version are required")
	}
	if !envelope.Status.valid() {
		return fmt.Errorf("invalid envelope status %q", envelope.Status)
	}
	if envelope.Status == OutcomePass && envelope.Reason != "" {
		return fmt.Errorf("PASS envelope cannot have a reason")
	}
	if envelope.Status != OutcomePass && !envelope.Reason.valid() {
		return fmt.Errorf("non-PASS envelope requires a known reason")
	}
	for index, explanation := range envelope.Explanations {
		if err := validateExplanation(envelope, explanation); err != nil {
			return fmt.Errorf("explanation %d: %w", index, err)
		}
	}
	expected, err := ComputeEvidenceDigest(envelope)
	if err != nil {
		return fmt.Errorf("evidence digest: %w", err)
	}
	if envelope.EvidenceDigest != expected {
		return fmt.Errorf("evidence digest mismatch")
	}
	return nil
}

func validateDigest(value string) error {
	if len(value) != hex.EncodedLen(32) || strings.TrimSpace(value) != value || strings.ToLower(value) != value {
		return fmt.Errorf("must be a canonical SHA-256 hex digest")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("must be a canonical SHA-256 hex digest")
	}
	return nil
}

func validateExplanation(envelope Envelope, explanation Explanation) error {
	if err := validateIdentity(explanation.CodeSymbolID, "code symbol"); err != nil {
		return fmt.Errorf("code symbol ID: %w", err)
	}
	if err := validateIdentity(explanation.SemanticOwnerID, "semantic owner"); err != nil {
		return fmt.Errorf("semantic owner ID: %w", err)
	}
	if err := validateBoundLocation(explanation.Origin); err != nil {
		return fmt.Errorf("origin: %w", err)
	}
	if explanation.Origin.URI != envelope.Document.URI {
		return fmt.Errorf("origin URI does not match document URI")
	}
	if err := validateBoundLocation(explanation.Target); err != nil {
		return fmt.Errorf("target: %w", err)
	}
	if !explanation.Claim.valid() {
		return fmt.Errorf("invalid change claim %q", explanation.Claim)
	}
	if !explanation.Status.valid() {
		return fmt.Errorf("invalid explanation status %q", explanation.Status)
	}
	if explanation.Status == OutcomePass && explanation.Reason != "" {
		return fmt.Errorf("PASS explanation cannot have a reason")
	}
	if explanation.Status != OutcomePass && !explanation.Reason.valid() {
		return fmt.Errorf("non-PASS explanation requires a known reason")
	}
	if explanation.Status == OutcomePass && len(explanation.CausalSpans) == 0 {
		return fmt.Errorf("PASS explanation requires causal spans")
	}
	seen := make(map[string]struct{}, len(explanation.CausalSpans))
	for index, span := range explanation.CausalSpans {
		if err := validateIdentity(span.StableID, "causal span"); err != nil {
			return fmt.Errorf("causal span %d ID: %w", index, err)
		}
		if err := validateSourceMapBinding(span.SourceMapID, span.SourceMapDigest); err != nil {
			return fmt.Errorf("causal span %d source map: %w", index, err)
		}
		if !exactText(span.URI) {
			return fmt.Errorf("causal span %d URI is required", index)
		}
		if err := validateRange(span.Range); err != nil {
			return fmt.Errorf("causal span %d: %w", index, err)
		}
		if span.Ordinal < 0 {
			return fmt.Errorf("causal span %d has negative ordinal", index)
		}
		if _, exists := seen[span.StableID]; exists {
			return fmt.Errorf("duplicate causal span %q", span.StableID)
		}
		seen[span.StableID] = struct{}{}
	}
	return nil
}

func validateBoundLocation(location BoundLocation) error {
	if err := validateIdentity(location.StableID, "location"); err != nil {
		return fmt.Errorf("stable ID: %w", err)
	}
	if err := validateSourceMapBinding(location.SourceMapID, location.SourceMapDigest); err != nil {
		return err
	}
	if !exactText(location.URI) {
		return fmt.Errorf("URI is required")
	}
	return validateRange(location.Range)
}

func validateSourceMapBinding(id, digest string) error {
	if err := validateIdentity(id, "source-map"); err != nil {
		return fmt.Errorf("source-map ID: %w", err)
	}
	if err := validateDigest(digest); err != nil {
		return fmt.Errorf("source-map digest: %w", err)
	}
	return nil
}

func validateIdentity(raw, kind string) error {
	parsed, err := semantic.ParseIdentity(raw)
	if err != nil {
		return fmt.Errorf("%s ID: %w", kind, err)
	}
	if parsed.String() != raw {
		return fmt.Errorf("%s ID is not canonical", kind)
	}
	return nil
}

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

func validateRange(value Range) error {
	if value.Start.Line < 0 || value.Start.Character < 0 || value.End.Line < 0 || value.End.Character < 0 {
		return fmt.Errorf("range coordinates must be non-negative")
	}
	if value.End.Line < value.Start.Line || (value.End.Line == value.Start.Line && value.End.Character < value.Start.Character) {
		return fmt.Errorf("range end precedes start")
	}
	return nil
}

func exactText(value string) bool { return value != "" && strings.TrimSpace(value) == value }
