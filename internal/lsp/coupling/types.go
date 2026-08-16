// Package coupling adapts immutable semantic-coupling explanations to
// standard Language Server Protocol values. The input is a read-only,
// snapshot-bound projection; it is never a source of semantic authority.
package coupling

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const SchemaVersion = "gooo/lsp-coupling-explanation/v1"

const diagnosticSource = "gooo/coupling"

// Outcome is the closed result set of the upstream explanation projection.
type Outcome string

const (
	OutcomePass       Outcome = "PASS"
	OutcomeUnknown    Outcome = "UNKNOWN"
	OutcomeFailClosed Outcome = "FAIL_CLOSED"
)

func (o Outcome) valid() bool {
	return o == OutcomePass || o == OutcomeUnknown || o == OutcomeFailClosed
}

// ChangeClaim is deliberately separate from inference edge kinds.
type ChangeClaim string

const (
	ClaimDelta   ChangeClaim = "DELTA"
	ClaimNoDelta ChangeClaim = "NO_DELTA"
)

func (c ChangeClaim) valid() bool { return c == ClaimDelta || c == ClaimNoDelta }

// Reason values are upstream failure partitions. The adapter does not
// collapse them into a successful no-op.
type Reason string

const (
	ReasonAmbiguous       Reason = "AMBIGUOUS"
	ReasonStaleSnapshot   Reason = "STALE_SNAPSHOT"
	ReasonUnregistered    Reason = "UNREGISTERED"
	ReasonMissing         Reason = "MISSING"
	ReasonUpstreamUnknown Reason = "UPSTREAM_UNKNOWN"
	ReasonUpstreamFail    Reason = "UPSTREAM_FAIL"
)

func (r Reason) valid() bool {
	switch r {
	case ReasonAmbiguous, ReasonStaleSnapshot, ReasonUnregistered, ReasonMissing,
		ReasonUpstreamUnknown, ReasonUpstreamFail:
		return true
	default:
		return false
	}
}

// Position and Range use the standard LSP zero-based UTF-16 representation.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// BoundLocation is input-only metadata. Stable IDs are required to validate
// the immutable explanation, but are intentionally not present in output LSP
// locations.
type BoundLocation struct {
	StableID        string `json:"stable_id"`
	SourceMapID     string `json:"source_map_id"`
	SourceMapDigest string `json:"source_map_digest"`
	URI             string `json:"uri"`
	Range           Range  `json:"range"`
	Label           string `json:"label,omitempty"`
}

// CausalSpan is an exact contributing source span from the immutable query
// result. Ordinal preserves the query's causal ordering; StableID binds the
// span without making it a custom LSP wire field.
type CausalSpan struct {
	StableID        string `json:"stable_id"`
	SourceMapID     string `json:"source_map_id"`
	SourceMapDigest string `json:"source_map_digest"`
	URI             string `json:"uri"`
	Range           Range  `json:"range"`
	Ordinal         int    `json:"ordinal"`
	Message         string `json:"message,omitempty"`
}

type Document struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// Explanation is one source-bound coupling explanation. Names and labels are
// presentation only; selection is performed by its exact origin range.
type Explanation struct {
	CodeSymbolID    string        `json:"code_symbol_id"`
	SemanticOwnerID string        `json:"semantic_owner_id"`
	Label           string        `json:"label,omitempty"`
	Origin          BoundLocation `json:"origin"`
	Target          BoundLocation `json:"target"`
	CausalSpans     []CausalSpan  `json:"causal_spans"`
	Claim           ChangeClaim   `json:"claim"`
	Status          Outcome       `json:"status"`
	Reason          Reason        `json:"reason,omitempty"`
}

// Envelope is the future immutable query/coupling explanation byte contract
// consumed by this package. It is intentionally a projection envelope, not a
// semantic authority or a writable document model.
type Envelope struct {
	Schema               string        `json:"schema"`
	SnapshotDigest       string        `json:"snapshot_digest"`
	RegistryDigest       string        `json:"registry_digest"`
	ToolchainDigest      string        `json:"toolchain_digest"`
	ProfileDigest        string        `json:"profile_digest"`
	DetectorResultDigest string        `json:"detector_result_digest"`
	OracleResultDigest   string        `json:"oracle_result_digest"`
	EvidenceDigest       string        `json:"evidence_digest"`
	Document             Document      `json:"document"`
	Status               Outcome       `json:"status"`
	Reason               Reason        `json:"reason,omitempty"`
	Explanations         []Explanation `json:"explanations"`
}

// Request carries every freshness and cancellation input required for a
// read. A nil Context, empty snapshot digest, or non-positive version is not
// treated as an implicit current value.
type Request struct {
	Context         context.Context
	DocumentURI     string
	DocumentVersion int
	Position        Position
	SnapshotDigest  string
}

// Location is the standard LSP location shape.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// LocationLink is the standard LSP definition response shape.
type LocationLink struct {
	OriginSelectionRange *Range `json:"originSelectionRange,omitempty"`
	TargetURI            string `json:"targetUri"`
	TargetRange          Range  `json:"targetRange"`
	TargetSelectionRange Range  `json:"targetSelectionRange"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           int                            `json:"severity,omitempty"`
	Code               string                         `json:"code,omitempty"`
	Source             string                         `json:"source,omitempty"`
	Message            string                         `json:"message"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

const (
	DiagnosticError       = 1
	DiagnosticWarning     = 2
	DiagnosticInformation = 3
)

const (
	DiagnosticExplanation         = "gooo.coupling.explanation"
	DiagnosticMissingCancellation = "gooo.coupling.missing-cancellation"
	DiagnosticMissingSnapshot     = "gooo.coupling.missing-snapshot"
	DiagnosticMissingVersion      = "gooo.coupling.missing-document-version"
	DiagnosticDocumentMismatch    = "gooo.coupling.document-mismatch"
	DiagnosticWrongVersion        = "gooo.coupling.wrong-document-version"
	DiagnosticStaleSnapshot       = "gooo.coupling.stale-snapshot"
	DiagnosticCancelled           = "gooo.coupling.cancelled"
	DiagnosticAmbiguous           = "gooo.coupling.ambiguous"
	DiagnosticUpstreamUnknown     = "gooo.coupling.upstream-unknown"
	DiagnosticUpstreamFail        = "gooo.coupling.upstream-fail-closed"
	DiagnosticNoBinding           = "gooo.coupling.no-binding"
	DiagnosticInvalidEnvelope     = "gooo.coupling.invalid-envelope"
	DiagnosticInvalidPosition     = "gooo.coupling.invalid-position"
)

// Result is an in-process aggregation. Its fields are independently
// serializable standard LSP values; Result itself is not a custom wire
// envelope.
type Result struct {
	Outcome     Outcome
	Links       []LocationLink
	Hover       *Hover
	Diagnostics []Diagnostic
}

type Adapter struct {
	envelope Envelope
	raw      []byte
}

// New decodes an immutable explanation byte slice and validates all
// identity-bearing locations before any protocol response can be produced.
// The input is copied so later caller mutation cannot change an adapter.
func New(data []byte) (*Adapter, error) {
	raw := append([]byte(nil), data...)
	if err := validateJSONDocument(raw); err != nil {
		return nil, fmt.Errorf("coupling explanation: decode: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var envelope Envelope
	if err := decoder.Decode(&envelope); err != nil {
		return nil, fmt.Errorf("coupling explanation: decode: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("coupling explanation: trailing JSON")
		}
		return nil, fmt.Errorf("coupling explanation: trailing JSON: %w", err)
	}
	if err := validateEnvelope(envelope); err != nil {
		return nil, fmt.Errorf("coupling explanation: %w", err)
	}
	return &Adapter{envelope: envelope, raw: raw}, nil
}

// RawBytes returns a copy for transcript tests and diagnostics. It cannot be
// used to mutate the adapter's snapshot.
func (a *Adapter) RawBytes() []byte { return append([]byte(nil), a.raw...) }

func validateJSONDocument(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON")
		}
		return fmt.Errorf("trailing JSON: %w", err)
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	switch delimiter := token.(type) {
	case json.Delim:
		switch delimiter {
		case '{':
			seen := make(map[string]struct{})
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("object key is not a string")
				}
				if _, exists := seen[key]; exists {
					return fmt.Errorf("duplicate object key %q", key)
				}
				seen[key] = struct{}{}
				if err := scanJSONValue(decoder); err != nil {
					return err
				}
			}
			end, err := decoder.Token()
			if err != nil {
				return err
			}
			if end != json.Delim('}') {
				return fmt.Errorf("object did not close")
			}
		case '[':
			for decoder.More() {
				if err := scanJSONValue(decoder); err != nil {
					return err
				}
			}
			end, err := decoder.Token()
			if err != nil {
				return err
			}
			if end != json.Delim(']') {
				return fmt.Errorf("array did not close")
			}
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
		}
	}
	return nil
}

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
	result := evidenceTuple{
		Schema: envelope.Schema,
		Digests: digestTuple{
			Snapshot: envelope.SnapshotDigest, Registry: envelope.RegistryDigest,
			Toolchain: envelope.ToolchainDigest, Profile: envelope.ProfileDigest,
			Detector: envelope.DetectorResultDigest, Oracle: envelope.OracleResultDigest,
		},
		DocumentVersion: envelope.Document.Version, Status: envelope.Status, Reason: envelope.Reason,
	}
	for _, explanation := range envelope.Explanations {
		value := explanationTuple{
			CodeSymbolID: explanation.CodeSymbolID, SemanticOwnerID: explanation.SemanticOwnerID,
			Origin: locationTuple{StableID: explanation.Origin.StableID, SourceMapID: explanation.Origin.SourceMapID, SourceMapDigest: explanation.Origin.SourceMapDigest, Range: explanation.Origin.Range},
			Target: locationTuple{StableID: explanation.Target.StableID, SourceMapID: explanation.Target.SourceMapID, SourceMapDigest: explanation.Target.SourceMapDigest, Range: explanation.Target.Range},
			Claim:  explanation.Claim, Status: explanation.Status, Reason: explanation.Reason,
		}
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
	sort.Slice(result.Explanations, func(i, j int) bool {
		return result.Explanations[i].CodeSymbolID < result.Explanations[j].CodeSymbolID
	})
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
