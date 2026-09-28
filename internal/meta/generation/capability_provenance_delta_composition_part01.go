package generation

import (
	"errors"
	"sort"
)

const capabilityProvenanceDeltaCompositionSchema = CapabilityProvenanceDeltaSchema

// ComposeCapabilityProvenanceDelta composes two adjacent delta witnesses.
// It requires the first observation to end exactly where the second begins.
// The result remains a non-authorizing witness and never infers changes across
// UNKNOWN or REJECTED observations.
func ComposeCapabilityProvenanceDelta(first, second CapabilityProvenanceDelta) (CapabilityProvenanceDelta, error) {
	if err := validateCapabilityProvenanceDeltaWitness(first); err != nil {
		return CapabilityProvenanceDelta{}, errors.New("first capability provenance delta: " + err.Error())
	}
	if err := validateCapabilityProvenanceDeltaWitness(second); err != nil {
		return CapabilityProvenanceDelta{}, errors.New("second capability provenance delta: " + err.Error())
	}
	if first.CurrentChainDigest != second.PreviousChainDigest {
		return CapabilityProvenanceDelta{}, errors.New("capability provenance delta composition boundary mismatch")
	}

	composed := CapabilityProvenanceDelta{
		Schema:                 capabilityProvenanceDeltaCompositionSchema,
		PreviousChainDigest:    first.PreviousChainDigest,
		CurrentChainDigest:     second.CurrentChainDigest,
		PreviousSourceRevision: first.PreviousSourceRevision,
		CurrentSourceRevision:  second.CurrentSourceRevision,
		NonAuthorizing:         true,
	}
	if verdict, terminal := capabilityProvenanceDeltaTerminalVerdict(first.Verdict, second.Verdict); terminal {
		composed.Verdict = verdict
		composed.DeltaDigest = composed.StableHash()
		return composed, nil
	}

	changes, err := composeCapabilityProvenanceDeltaChanges(first.Changes, second.Changes)
	if err != nil {
		return CapabilityProvenanceDelta{}, err
	}
	composed.Changes = changes
	if len(changes) == 0 {
		composed.Verdict = CapabilityDeltaUnchanged
	} else {
		composed.Verdict = CapabilityDeltaChanged
	}
	composed.DeltaDigest = composed.StableHash()
	return composed, nil
}

func validateCapabilityProvenanceDeltaWitness(delta CapabilityProvenanceDelta) error {
	if delta.Schema != CapabilityProvenanceDeltaSchema || !delta.NonAuthorizing {
		return errors.New("capability provenance delta is not a non-authorizing witness")
	}
	switch delta.Verdict {
	case CapabilityDeltaUnchanged:
		if len(delta.Changes) != 0 {
			return errors.New("unchanged capability provenance delta contains changes")
		}
	case CapabilityDeltaChanged:
		if len(delta.Changes) == 0 {
			return errors.New("changed capability provenance delta has no changes")
		}
	case CapabilityDeltaUnknown, CapabilityDeltaRejected:
		if len(delta.Changes) != 0 {
			return errors.New("terminal capability provenance delta contains inferred changes")
		}
	default:
		return errors.New("capability provenance delta has an unsupported verdict")
	}
	if delta.DeltaDigest != delta.StableHash() {
		return errors.New("capability provenance delta digest mismatch")
	}
	return nil
}

func capabilityProvenanceDeltaTerminalVerdict(first, second string) (string, bool) {
	if first == CapabilityDeltaUnknown || second == CapabilityDeltaUnknown {
		return CapabilityDeltaUnknown, true
	}
	if first == CapabilityDeltaRejected || second == CapabilityDeltaRejected {
		return CapabilityDeltaRejected, true
	}
	return "", false
}

func composeCapabilityProvenanceDeltaChanges(first, second []CapabilityProvenanceDeltaChange) ([]CapabilityProvenanceDeltaChange, error) {
	bySequence := make(map[int]CapabilityProvenanceDeltaChange, len(first)+len(second))
	for _, change := range first {
		bySequence[change.Sequence] = change
	}
	for _, change := range second {
		previous, exists := bySequence[change.Sequence]
		if !exists {
			bySequence[change.Sequence] = change
			continue
		}
		if previous.Phase != change.Phase ||
			previous.Kind != change.Kind ||
			previous.AfterID != change.BeforeID ||
			previous.AfterDigest != change.BeforeDigest {
			return nil, errors.New("capability provenance delta composition phase boundary mismatch")
		}
		previous.AfterID = change.AfterID
		previous.AfterDigest = change.AfterDigest
		bySequence[change.Sequence] = previous
	}

	changes := make([]CapabilityProvenanceDeltaChange, 0, len(bySequence))
	for _, change := range bySequence {
		changes = append(changes, change)
	}
	sort.Slice(changes, func(left, right int) bool {
		if changes[left].Sequence != changes[right].Sequence {
			return changes[left].Sequence < changes[right].Sequence
		}
		return changes[left].Phase < changes[right].Phase
	})
	return changes, nil
}
