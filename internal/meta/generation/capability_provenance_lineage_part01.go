package generation

import (
	"errors"
	"slices"
	"strings"
)

const CapabilityProvenanceLineageSchema = "gooo/capability-provenance-lineage/v1"

const (
	CapabilityLineageComplete = "COMPLETE"
	CapabilityLineageUnknown  = "UNKNOWN"
	CapabilityLineageRejected = "REJECTED"
)

// CapabilityProvenanceLineage is the deterministic origin chain for one
// observed capability. It is an evidence projection and never authorizes work.
type CapabilityProvenanceLineage struct {
	Schema                     string                            `json:"schema"`
	SourceRevision             SourceRevision                    `json:"source_revision"`
	OperationID                string                            `json:"operation_id"`
	Steps                      []CapabilityProvenanceLineageStep `json:"steps"`
	AuthorityObservationDigest string                            `json:"authority_observation_digest"`
	WorkloadProvenanceDigest   string                            `json:"workload_provenance_digest"`
	CapabilityHandleDigest     string                            `json:"capability_handle_digest"`
	Verdict                    string                            `json:"verdict"`
	NonAuthorizing             bool                              `json:"non_authorizing"`
	ChainDigest                string                            `json:"chain_digest"`
}

// CapabilityProvenanceLineageStep is one exact hop in the computed origin.
type CapabilityProvenanceLineageStep struct {
	Phase  string `json:"phase"`
	ID     string `json:"id"`
	Digest string `json:"digest"`
}

// ObserveCapabilityProvenanceLineage composes the exact workload and handle
// witnesses into a stable source-to-capability chain.
func ObserveCapabilityProvenanceLineage(workload WorkloadAuthorityProvenance, boundary AuthorityBoundaryObservation, capability CapabilityHandleProvenance) (CapabilityProvenanceLineage, error) {
	if err := workload.Validate(boundary); err != nil {
		return CapabilityProvenanceLineage{}, err
	}
	if err := capability.Validate(workload, boundary); err != nil {
		return CapabilityProvenanceLineage{}, err
	}
	verdict := CapabilityLineageUnknown
	switch capability.ObservationStatus {
	case CapabilityHandleObserved:
		verdict = CapabilityLineageComplete
	case CapabilityHandleUnknown:
		verdict = CapabilityLineageUnknown
	case CapabilityHandleRejected:
		verdict = CapabilityLineageRejected
	default:
		return CapabilityProvenanceLineage{}, errors.New("capability provenance lineage requires a known handle status")
	}
	lineage := CapabilityProvenanceLineage{
		Schema:         CapabilityProvenanceLineageSchema,
		SourceRevision: boundary.SourceRevision,
		OperationID:    boundary.OperationID,
		Steps: []CapabilityProvenanceLineageStep{
			{Phase: "SOURCE_REVISION", ID: boundary.SourceRevision.ID, Digest: boundary.SourceRevision.Digest},
			{Phase: "AUTHORITY_BOUNDARY", ID: boundary.OperationID, Digest: boundary.StableHash()},
			{Phase: "WORKLOAD_IDENTITY", ID: workload.WorkloadURI, Digest: workload.StableHash()},
			{Phase: "CAPABILITY_HANDLE", ID: capability.HandleID, Digest: capability.StableHash()},
		},
		AuthorityObservationDigest: workload.AuthorityObservationDigest,
		WorkloadProvenanceDigest:   workload.StableHash(),
		CapabilityHandleDigest:     capability.StableHash(),
		Verdict:                    verdict,
		NonAuthorizing:             true,
	}
	lineage.ChainDigest = lineage.StableHash()
	return lineage, nil
}

// Validate recomputes the complete origin chain against its exact witnesses.
func (l CapabilityProvenanceLineage) Validate(workload WorkloadAuthorityProvenance, boundary AuthorityBoundaryObservation, capability CapabilityHandleProvenance) error {
	if l.Schema != CapabilityProvenanceLineageSchema || !l.NonAuthorizing {
		return errors.New("capability provenance lineage is not a non-authorizing witness")
	}
	expected, err := ObserveCapabilityProvenanceLineage(workload, boundary, capability)
	if err != nil {
		return err
	}
	if l.SourceRevision != expected.SourceRevision || l.OperationID != expected.OperationID ||
		!slices.Equal(l.Steps, expected.Steps) || l.AuthorityObservationDigest != expected.AuthorityObservationDigest ||
		l.WorkloadProvenanceDigest != expected.WorkloadProvenanceDigest || l.CapabilityHandleDigest != expected.CapabilityHandleDigest ||
		l.Verdict != expected.Verdict || l.ChainDigest != expected.ChainDigest {
		return errors.New("capability provenance lineage does not match exact witnesses")
	}
	return nil
}

// ReverseOrigin returns a copy ordered from the observed capability back to
// the source revision. It never mutates the forward lineage.
func (l CapabilityProvenanceLineage) ReverseOrigin() []CapabilityProvenanceLineageStep {
	reversed := make([]CapabilityProvenanceLineageStep, len(l.Steps))
	for index, step := range l.Steps {
		reversed[len(l.Steps)-index-1] = step
	}
	return reversed
}

// Canonical is display-independent and excludes ChainDigest to avoid recursion.
func (l CapabilityProvenanceLineage) Canonical() string {
	steps := make([]string, 0, len(l.Steps))
	for _, step := range l.Steps {
		steps = append(steps, strings.Join([]string{step.Phase, step.ID, step.Digest}, "\x00"))
	}
	return strings.Join([]string{
		"capability-provenance-lineage", l.Schema, l.SourceRevision.ID, l.SourceRevision.Digest,
		l.OperationID, strings.Join(steps, "\x1e"), l.AuthorityObservationDigest,
		l.WorkloadProvenanceDigest, l.CapabilityHandleDigest, l.Verdict, boolString(l.NonAuthorizing),
	}, "\t")
}

func (l CapabilityProvenanceLineage) StableHash() string {
	return envelopeDigestString(l.Canonical())
}
