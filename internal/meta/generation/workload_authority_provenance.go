package generation

import (
	"errors"
	"net/url"
	"strings"
)

const WorkloadAuthorityProvenanceSchema = "gooo/workload-authority-provenance/v1"

const (
	WorkloadBindingBound    = "BOUND"
	WorkloadBindingUnknown  = "UNKNOWN"
	WorkloadBindingRejected = "REJECTED"
)

// WorkloadAuthorityProvenance binds a SPIFFE-shaped workload identity to an
// already observed semantic authority boundary. It is provenance only: it
// never verifies workload credentials, grants effects, or authorizes work.
type WorkloadAuthorityProvenance struct {
	Schema                     string         `json:"schema"`
	WorkloadURI                string         `json:"workload_uri"`
	TrustDomain                string         `json:"trust_domain"`
	WorkloadPath               string         `json:"workload_path"`
	IdentityDigest             string         `json:"identity_digest"`
	AuthorityDigest            string         `json:"authority_digest"`
	AuthorityObservationDigest string         `json:"authority_observation_digest"`
	SourceRevision             SourceRevision `json:"source_revision"`
	OperationID                string         `json:"operation_id"`
	BoundaryStatus             string         `json:"boundary_status"`
	BindingStatus              string         `json:"binding_status"`
	NonAuthorizing             bool           `json:"non_authorizing"`
}

// ObserveWorkloadAuthorityProvenance records the workload identity observed
// alongside one exact semantic boundary. The boundary classification controls
// provenance status, but does not become a credential or an execution grant.
func ObserveWorkloadAuthorityProvenance(workloadURI string, boundary AuthorityBoundaryObservation) (WorkloadAuthorityProvenance, error) {
	identity, err := parseWorkloadURI(workloadURI)
	if err != nil {
		return WorkloadAuthorityProvenance{}, err
	}
	if boundary.Schema != AuthorityBoundaryObservationSchema || !validEnvelopeDigest(boundary.AuthorityDigest) ||
		strings.TrimSpace(boundary.SourceRevision.ID) == "" || !validEnvelopeDigest(boundary.SourceRevision.Digest) ||
		strings.TrimSpace(boundary.OperationID) == "" {
		return WorkloadAuthorityProvenance{}, errors.New("workload provenance requires an authority boundary observation")
	}
	status := WorkloadBindingUnknown
	switch {
	case !boundary.RequestWithinGrant || !boundary.ResultWithinGrant || !boundary.ResultSourceRevisionOK:
		status = WorkloadBindingRejected
	case boundary.BoundaryStatus == "WITHIN_DECLARED_GRANT", boundary.BoundaryStatus == "NO_EFFECT":
		status = WorkloadBindingBound
	case boundary.BoundaryStatus == "EFFECT_ESCALATION", boundary.BoundaryStatus == "STALE_RESULT":
		status = WorkloadBindingRejected
	case boundary.BoundaryStatus == "RESULT_UNOBSERVED":
		status = WorkloadBindingUnknown
	default:
		return WorkloadAuthorityProvenance{}, errors.New("workload provenance requires a known boundary status")
	}
	return WorkloadAuthorityProvenance{
		Schema:                     WorkloadAuthorityProvenanceSchema,
		WorkloadURI:                workloadURI,
		TrustDomain:                identity.trustDomain,
		WorkloadPath:               identity.workloadPath,
		IdentityDigest:             workloadIdentityDigest(workloadURI),
		AuthorityDigest:            boundary.AuthorityDigest,
		AuthorityObservationDigest: boundary.StableHash(),
		SourceRevision:             boundary.SourceRevision,
		OperationID:                boundary.OperationID,
		BoundaryStatus:             boundary.BoundaryStatus,
		BindingStatus:              status,
		NonAuthorizing:             true,
	}, nil
}

// Validate recomputes the identity-to-boundary binding and rejects tampering.
func (p WorkloadAuthorityProvenance) Validate(boundary AuthorityBoundaryObservation) error {
	if p.Schema != WorkloadAuthorityProvenanceSchema || !p.NonAuthorizing {
		return errors.New("workload provenance is not a non-authorizing witness")
	}
	expected, err := ObserveWorkloadAuthorityProvenance(p.WorkloadURI, boundary)
	if err != nil {
		return err
	}
	if p != expected {
		return errors.New("workload provenance does not match the exact authority boundary")
	}
	return nil
}

// Canonical is display-independent and suitable for a provenance digest.
func (p WorkloadAuthorityProvenance) Canonical() string {
	return strings.Join([]string{
		"workload-authority-provenance", p.Schema, p.WorkloadURI, p.TrustDomain, p.WorkloadPath,
		p.AuthorityDigest, p.AuthorityObservationDigest, p.SourceRevision.ID, p.SourceRevision.Digest,
		p.OperationID, p.BoundaryStatus, p.BindingStatus,
	}, "\t")
}

func (p WorkloadAuthorityProvenance) StableHash() string {
	return envelopeDigestString(p.Canonical())
}

type workloadIdentity struct {
	trustDomain  string
	workloadPath string
}

func workloadIdentityDigest(workloadURI string) string {
	return envelopeDigestString(strings.Join([]string{
		"spiffe-workload-identity/v1",
		workloadURI,
	}, "\t"))
}

func parseWorkloadURI(value string) (workloadIdentity, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "spiffe" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Host != strings.ToLower(parsed.Host) || parsed.Path == "" || parsed.Path == "/" {
		return workloadIdentity{}, errors.New("workload identity must be a canonical spiffe URI")
	}
	return workloadIdentity{trustDomain: parsed.Host, workloadPath: parsed.Path}, nil
}
