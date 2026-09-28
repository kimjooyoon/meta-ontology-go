package generation

import (
	"errors"
	"strings"
)

const (
	WorkloadDeclarationProvenanceSchema = "gooo/workload-declaration-provenance/v1"
	WorkloadDeclarationBindingBound     = "BOUND"
	WorkloadDeclarationBindingUnknown   = "UNKNOWN"
	WorkloadDeclarationBindingRejected  = "REJECTED"
)

type WorkloadDeclarationProvenance struct {
	Schema                     string
	WorkloadIdentityDigest     string
	DeclarationDigest          string
	AuthorityDigest            string
	AuthorityObservationDigest string
	SourceRevisionDigest       string
	BindingStatus              string
	NonAuthorizing             bool
	ObservationDigest          string
}

func ObserveWorkloadDeclarationProvenance(workload WorkloadAuthorityProvenance, declarationDigest string) (WorkloadDeclarationProvenance, error) {
	if workload.Schema != WorkloadAuthorityProvenanceSchema ||
		!workload.NonAuthorizing ||
		!validEnvelopeDigest(workload.IdentityDigest) ||
		!validEnvelopeDigest(workload.AuthorityDigest) ||
		!validEnvelopeDigest(workload.AuthorityObservationDigest) ||
		!validEnvelopeDigest(workload.SourceRevision.Digest) ||
		!validEnvelopeDigest(declarationDigest) {
		return WorkloadDeclarationProvenance{}, errors.New("workload declaration provenance requires valid workload and declaration evidence")
	}
	status := WorkloadDeclarationBindingUnknown
	switch workload.BindingStatus {
	case WorkloadBindingBound:
		status = WorkloadDeclarationBindingBound
	case WorkloadBindingUnknown:
		status = WorkloadDeclarationBindingUnknown
	case WorkloadBindingRejected:
		status = WorkloadDeclarationBindingRejected
	default:
		return WorkloadDeclarationProvenance{}, errors.New("workload declaration provenance requires a known workload binding status")
	}
	observation := WorkloadDeclarationProvenance{
		Schema:                     WorkloadDeclarationProvenanceSchema,
		WorkloadIdentityDigest:     workload.IdentityDigest,
		DeclarationDigest:          declarationDigest,
		AuthorityDigest:            workload.AuthorityDigest,
		AuthorityObservationDigest: workload.AuthorityObservationDigest,
		SourceRevisionDigest:       workload.SourceRevision.Digest,
		BindingStatus:              status,
		NonAuthorizing:             true,
	}
	observation.ObservationDigest = observation.StableHash()
	return observation, nil
}

func (p WorkloadDeclarationProvenance) Canonical() string {
	return strings.Join([]string{
		"workload-declaration-provenance",
		p.Schema,
		p.WorkloadIdentityDigest,
		p.DeclarationDigest,
		p.AuthorityDigest,
		p.AuthorityObservationDigest,
		p.SourceRevisionDigest,
		p.BindingStatus,
	}, "\t")
}

func (p WorkloadDeclarationProvenance) StableHash() string {
	return envelopeDigestString(p.Canonical())
}

func (p WorkloadDeclarationProvenance) Validate() error {
	if p.Schema != WorkloadDeclarationProvenanceSchema ||
		!p.NonAuthorizing ||
		p.BindingStatus == "" ||
		!validEnvelopeDigest(p.WorkloadIdentityDigest) ||
		!validEnvelopeDigest(p.DeclarationDigest) ||
		!validEnvelopeDigest(p.AuthorityDigest) ||
		!validEnvelopeDigest(p.AuthorityObservationDigest) ||
		!validEnvelopeDigest(p.SourceRevisionDigest) {
		return errors.New("workload declaration provenance is invalid")
	}
	if p.ObservationDigest != p.StableHash() {
		return errors.New("workload declaration provenance digest mismatch")
	}
	return nil
}
