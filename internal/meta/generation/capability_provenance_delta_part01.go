package generation

import (
	"errors"
	"slices"
	"strconv"
	"strings"
)

const CapabilityProvenanceDeltaSchema = "gooo/capability-provenance-delta/v1"

const (
	CapabilityDeltaUnchanged = "UNCHANGED"
	CapabilityDeltaChanged   = "CHANGED"
	CapabilityDeltaUnknown   = "UNKNOWN"
	CapabilityDeltaRejected  = "REJECTED"
	CapabilityDeltaUpdated   = "UPDATED"
)

// CapabilityProvenanceDelta is a deterministic comparison of two complete
// capability lineage observations. It is evidence only and never authorizes
// the transition it describes.
type CapabilityProvenanceDelta struct {
	Schema                 string                            `json:"schema"`
	PreviousChainDigest    string                            `json:"previous_chain_digest"`
	CurrentChainDigest     string                            `json:"current_chain_digest"`
	PreviousSourceRevision SourceRevision                    `json:"previous_source_revision"`
	CurrentSourceRevision  SourceRevision                    `json:"current_source_revision"`
	Changes                []CapabilityProvenanceDeltaChange `json:"changes"`
	Verdict                string                            `json:"verdict"`
	NonAuthorizing         bool                              `json:"non_authorizing"`
	DeltaDigest            string                            `json:"delta_digest"`
}

// CapabilityProvenanceDeltaChange identifies one changed lineage phase.
type CapabilityProvenanceDeltaChange struct {
	Sequence     int    `json:"sequence"`
	Phase        string `json:"phase"`
	Kind         string `json:"kind"`
	BeforeID     string `json:"before_id"`
	BeforeDigest string `json:"before_digest"`
	AfterID      string `json:"after_id"`
	AfterDigest  string `json:"after_digest"`
}

// CompareCapabilityProvenanceLineage computes a phase-by-phase delta while
// preserving UNKNOWN and REJECTED evidence without inventing changes.
func CompareCapabilityProvenanceLineage(before, after CapabilityProvenanceLineage) (CapabilityProvenanceDelta, error) {
	if err := validateCapabilityLineageForLifecycle(before); err != nil {
		return CapabilityProvenanceDelta{}, err
	}
	if err := validateCapabilityLineageForLifecycle(after); err != nil {
		return CapabilityProvenanceDelta{}, err
	}
	delta := CapabilityProvenanceDelta{
		Schema:                 CapabilityProvenanceDeltaSchema,
		PreviousChainDigest:    before.ChainDigest,
		CurrentChainDigest:     after.ChainDigest,
		PreviousSourceRevision: before.SourceRevision,
		CurrentSourceRevision:  after.SourceRevision,
		NonAuthorizing:         true,
	}
	switch {
	case before.Verdict == CapabilityLineageUnknown || after.Verdict == CapabilityLineageUnknown:
		delta.Verdict = CapabilityDeltaUnknown
	case before.Verdict == CapabilityLineageRejected || after.Verdict == CapabilityLineageRejected:
		delta.Verdict = CapabilityDeltaRejected
	case before.Verdict == CapabilityLineageComplete && after.Verdict == CapabilityLineageComplete:
		delta.Changes = capabilityLineageChanges(before.Steps, after.Steps)
		if len(delta.Changes) == 0 {
			delta.Verdict = CapabilityDeltaUnchanged
		} else {
			delta.Verdict = CapabilityDeltaChanged
		}
	default:
		return CapabilityProvenanceDelta{}, errors.New("capability provenance delta requires known lineage verdicts")
	}
	delta.DeltaDigest = delta.StableHash()
	return delta, nil
}

// Validate recomputes the phase delta against the exact before and after
// lineages and rejects changed evidence or ordering.
func (d CapabilityProvenanceDelta) Validate(before, after CapabilityProvenanceLineage) error {
	if d.Schema != CapabilityProvenanceDeltaSchema || !d.NonAuthorizing {
		return errors.New("capability provenance delta is not a non-authorizing witness")
	}
	expected, err := CompareCapabilityProvenanceLineage(before, after)
	if err != nil {
		return err
	}
	if d.PreviousChainDigest != expected.PreviousChainDigest ||
		d.CurrentChainDigest != expected.CurrentChainDigest ||
		d.PreviousSourceRevision != expected.PreviousSourceRevision ||
		d.CurrentSourceRevision != expected.CurrentSourceRevision ||
		!slices.Equal(d.Changes, expected.Changes) ||
		d.Verdict != expected.Verdict ||
		d.DeltaDigest != expected.DeltaDigest {
		return errors.New("capability provenance delta does not match exact lineages")
	}
	return nil
}

// ReverseChanges returns the changed phases from the newest observation back
// toward the previous observation without mutating the delta.
func (d CapabilityProvenanceDelta) ReverseChanges() []CapabilityProvenanceDeltaChange {
	reversed := make([]CapabilityProvenanceDeltaChange, len(d.Changes))
	for index, change := range d.Changes {
		reversed[len(d.Changes)-index-1] = change
	}
	return reversed
}

func capabilityLineageChanges(before, after []CapabilityProvenanceLineageStep) []CapabilityProvenanceDeltaChange {
	changes := make([]CapabilityProvenanceDeltaChange, 0, len(before))
	for sequence := range before {
		if before[sequence] == after[sequence] {
			continue
		}
		changes = append(changes, CapabilityProvenanceDeltaChange{
			Sequence:     sequence,
			Phase:        after[sequence].Phase,
			Kind:         CapabilityDeltaUpdated,
			BeforeID:     before[sequence].ID,
			BeforeDigest: before[sequence].Digest,
			AfterID:      after[sequence].ID,
			AfterDigest:  after[sequence].Digest,
		})
	}
	return changes
}

// Canonical is display-independent and excludes DeltaDigest to avoid
// recursion.
func (d CapabilityProvenanceDelta) Canonical() string {
	changes := make([]string, 0, len(d.Changes))
	for _, change := range d.Changes {
		changes = append(changes, strings.Join([]string{
			strconv.Itoa(change.Sequence), change.Phase, change.Kind,
			change.BeforeID, change.BeforeDigest, change.AfterID, change.AfterDigest,
		}, "\x00"))
	}
	return strings.Join([]string{
		"capability-provenance-delta",
		d.Schema,
		d.PreviousChainDigest,
		d.CurrentChainDigest,
		d.PreviousSourceRevision.ID,
		d.PreviousSourceRevision.Digest,
		d.CurrentSourceRevision.ID,
		d.CurrentSourceRevision.Digest,
		strings.Join(changes, "\x1e"),
		d.Verdict,
		boolString(d.NonAuthorizing),
	}, "\t")
}

func (d CapabilityProvenanceDelta) StableHash() string {
	return envelopeDigestString(d.Canonical())
}
