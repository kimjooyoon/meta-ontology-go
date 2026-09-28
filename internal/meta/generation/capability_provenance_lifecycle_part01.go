package generation

import (
	"errors"
	"slices"
	"strconv"
	"strings"
)

const CapabilityProvenanceLifecycleSchema = "gooo/capability-provenance-lifecycle/v1"

const (
	CapabilityLifecycleBorrowed = "BORROWED"
	CapabilityLifecycleOwned    = "OWNED"
	CapabilityLifecycleDropped  = "DROPPED"
	CapabilityLifecycleOpen     = "OPEN"
	CapabilityLifecycleUnknown  = "UNKNOWN"
	CapabilityLifecycleRejected = "REJECTED"
)

// CapabilityProvenanceLifecycle is a deterministic observation of capability
// handle ownership and borrowing transitions. It never grants or transfers a
// capability.
type CapabilityProvenanceLifecycle struct {
	Schema             string                               `json:"schema"`
	LineageDigest      string                               `json:"lineage_digest"`
	SourceRevision     SourceRevision                       `json:"source_revision"`
	OperationID        string                               `json:"operation_id"`
	CapabilityHandleID string                               `json:"capability_handle_id"`
	Events             []CapabilityProvenanceLifecycleEvent `json:"events"`
	FinalStatus        string                               `json:"final_status"`
	NonAuthorizing     bool                                 `json:"non_authorizing"`
	LifecycleDigest    string                               `json:"lifecycle_digest"`
}

// CapabilityProvenanceLifecycleEvent is one observed handle transition.
type CapabilityProvenanceLifecycleEvent struct {
	Sequence       int    `json:"sequence"`
	Mode           string `json:"mode"`
	EvidenceDigest string `json:"evidence_digest"`
}

// ObserveCapabilityProvenanceLifecycle computes a reversible lifecycle from
// the exact lineage. Unknown and rejected lineages remain terminally explicit;
// no unobserved transition is inferred for them.
func ObserveCapabilityProvenanceLifecycle(lineage CapabilityProvenanceLineage, modes []string) (CapabilityProvenanceLifecycle, error) {
	if err := validateCapabilityLineageForLifecycle(lineage); err != nil {
		return CapabilityProvenanceLifecycle{}, err
	}
	if len(lineage.Steps) == 0 {
		return CapabilityProvenanceLifecycle{}, errors.New("capability provenance lifecycle requires a capability step")
	}
	lifecycle := CapabilityProvenanceLifecycle{
		Schema:             CapabilityProvenanceLifecycleSchema,
		LineageDigest:      lineage.ChainDigest,
		SourceRevision:     lineage.SourceRevision,
		OperationID:        lineage.OperationID,
		CapabilityHandleID: lineage.Steps[len(lineage.Steps)-1].ID,
		NonAuthorizing:     true,
	}
	switch lineage.Verdict {
	case CapabilityLineageUnknown:
		if len(modes) != 0 {
			return CapabilityProvenanceLifecycle{}, errors.New("unknown capability lineage cannot infer lifecycle transitions")
		}
		lifecycle.FinalStatus = CapabilityLifecycleUnknown
	case CapabilityLineageRejected:
		if len(modes) != 0 {
			return CapabilityProvenanceLifecycle{}, errors.New("rejected capability lineage cannot infer lifecycle transitions")
		}
		lifecycle.FinalStatus = CapabilityLifecycleRejected
	case CapabilityLineageComplete:
		if err := validateCapabilityLifecycleModes(modes); err != nil {
			return CapabilityProvenanceLifecycle{}, err
		}
		lifecycle.Events = make([]CapabilityProvenanceLifecycleEvent, 0, len(modes))
		for sequence, mode := range modes {
			lifecycle.Events = append(lifecycle.Events, CapabilityProvenanceLifecycleEvent{
				Sequence:       sequence,
				Mode:           mode,
				EvidenceDigest: capabilityLifecycleEvidenceDigest(lineage.ChainDigest, sequence, mode),
			})
		}
		if modes[len(modes)-1] == CapabilityLifecycleDropped {
			lifecycle.FinalStatus = CapabilityLifecycleDropped
		} else {
			lifecycle.FinalStatus = CapabilityLifecycleOpen
		}
	default:
		return CapabilityProvenanceLifecycle{}, errors.New("capability provenance lifecycle requires a known lineage verdict")
	}
	lifecycle.LifecycleDigest = lifecycle.StableHash()
	return lifecycle, nil
}

// Validate recomputes the lifecycle against the exact source lineage and
// rejects changed modes, ordering, lineage, or terminal status.
func (l CapabilityProvenanceLifecycle) Validate(lineage CapabilityProvenanceLineage) error {
	if l.Schema != CapabilityProvenanceLifecycleSchema || !l.NonAuthorizing {
		return errors.New("capability provenance lifecycle is not a non-authorizing witness")
	}
	modes := make([]string, len(l.Events))
	for sequence, event := range l.Events {
		if event.Sequence != sequence {
			return errors.New("capability provenance lifecycle sequence is not canonical")
		}
		modes[sequence] = event.Mode
	}
	expected, err := ObserveCapabilityProvenanceLifecycle(lineage, modes)
	if err != nil {
		return err
	}
	if l.LineageDigest != expected.LineageDigest ||
		l.SourceRevision != expected.SourceRevision ||
		l.OperationID != expected.OperationID ||
		l.CapabilityHandleID != expected.CapabilityHandleID ||
		!slices.Equal(l.Events, expected.Events) ||
		l.FinalStatus != expected.FinalStatus ||
		l.LifecycleDigest != expected.LifecycleDigest {
		return errors.New("capability provenance lifecycle does not match the exact lineage")
	}
	return nil
}

// ReverseLifecycle returns the observed transitions from terminal observation
// back toward acquisition without mutating the forward lifecycle.
func (l CapabilityProvenanceLifecycle) ReverseLifecycle() []CapabilityProvenanceLifecycleEvent {
	reversed := make([]CapabilityProvenanceLifecycleEvent, len(l.Events))
	for index, event := range l.Events {
		reversed[len(l.Events)-index-1] = event
	}
	return reversed
}

func validateCapabilityLineageForLifecycle(lineage CapabilityProvenanceLineage) error {
	if lineage.Schema != CapabilityProvenanceLineageSchema || !lineage.NonAuthorizing ||
		strings.TrimSpace(lineage.OperationID) == "" ||
		strings.TrimSpace(lineage.ChainDigest) == "" ||
		lineage.ChainDigest != lineage.StableHash() {
		return errors.New("capability provenance lifecycle requires an untampered non-authorizing lineage")
	}
	expectedPhases := []string{"SOURCE_REVISION", "AUTHORITY_BOUNDARY", "WORKLOAD_IDENTITY", "CAPABILITY_HANDLE"}
	if len(lineage.Steps) != len(expectedPhases) {
		return errors.New("capability provenance lifecycle requires the canonical lineage phases")
	}
	for index, phase := range expectedPhases {
		if lineage.Steps[index].Phase != phase || strings.TrimSpace(lineage.Steps[index].ID) == "" || strings.TrimSpace(lineage.Steps[index].Digest) == "" {
			return errors.New("capability provenance lifecycle requires complete lineage steps")
		}
	}
	return nil
}

func validateCapabilityLifecycleModes(modes []string) error {
	if len(modes) == 0 {
		return errors.New("complete capability lineage requires at least one lifecycle transition")
	}
	for sequence, mode := range modes {
		switch mode {
		case CapabilityLifecycleBorrowed, CapabilityLifecycleOwned:
			if sequence == 0 {
				continue
			}
		case CapabilityLifecycleDropped:
			if sequence == 0 || sequence != len(modes)-1 {
				return errors.New("dropped capability lifecycle must be terminal")
			}
		default:
			return errors.New("capability provenance lifecycle contains an unknown transition")
		}
		if sequence > 0 && modes[sequence-1] == CapabilityLifecycleDropped {
			return errors.New("capability provenance lifecycle cannot continue after drop")
		}
	}
	return nil
}

func capabilityLifecycleEvidenceDigest(lineageDigest string, sequence int, mode string) string {
	return envelopeDigestString(strings.Join([]string{
		"capability-provenance-lifecycle-event",
		CapabilityProvenanceLifecycleSchema,
		lineageDigest,
		strconv.Itoa(sequence),
		mode,
	}, "\t"))
}

// Canonical is display-independent and excludes LifecycleDigest to avoid
// recursion.
func (l CapabilityProvenanceLifecycle) Canonical() string {
	events := make([]string, 0, len(l.Events))
	for _, event := range l.Events {
		events = append(events, strings.Join([]string{
			strconv.Itoa(event.Sequence), event.Mode, event.EvidenceDigest,
		}, "\x00"))
	}
	return strings.Join([]string{
		"capability-provenance-lifecycle",
		l.Schema,
		l.LineageDigest,
		l.SourceRevision.ID,
		l.SourceRevision.Digest,
		l.OperationID,
		l.CapabilityHandleID,
		strings.Join(events, "\x1e"),
		l.FinalStatus,
		boolString(l.NonAuthorizing),
	}, "\t")
}

func (l CapabilityProvenanceLifecycle) StableHash() string {
	return envelopeDigestString(l.Canonical())
}
