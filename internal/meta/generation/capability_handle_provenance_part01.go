package generation

import (
	"errors"
	"net/url"
	"slices"
	"strings"
)

const CapabilityHandleProvenanceSchema = "gooo/capability-handle-provenance/v1"

const (
	CapabilityHandleObserved = "OBSERVED_WITHIN_GRANT"
	CapabilityHandleUnknown  = "UNKNOWN"
	CapabilityHandleRejected = "REJECTED"
)

// CapabilityHandleProvenance binds one observed capability handle to the
// workload witness and exact semantic authority boundary. It never creates,
// transfers, or authorizes a capability.
type CapabilityHandleProvenance struct {
	Schema                     string         `json:"schema"`
	HandleID                   string         `json:"handle_id"`
	CapabilityURI              string         `json:"capability_uri"`
	Effect                     string         `json:"effect"`
	AuthorityDigest            string         `json:"authority_digest"`
	AuthorityObservationDigest string         `json:"authority_observation_digest"`
	WorkloadProvenanceDigest   string         `json:"workload_provenance_digest"`
	SourceRevision             SourceRevision `json:"source_revision"`
	OperationID                string         `json:"operation_id"`
	ObservationStatus          string         `json:"observation_status"`
	EffectWithinGrant          bool           `json:"effect_within_grant"`
	NonAuthorizing             bool           `json:"non_authorizing"`
}

// ObserveCapabilityHandleProvenance records an observed handle without
// turning its identity into ambient authority. Missing results remain unknown.
func ObserveCapabilityHandleProvenance(workload WorkloadAuthorityProvenance, boundary AuthorityBoundaryObservation, handleID, capabilityURI, effect string) (CapabilityHandleProvenance, error) {
	if err := workload.Validate(boundary); err != nil {
		return CapabilityHandleProvenance{}, err
	}
	if strings.TrimSpace(handleID) == "" || strings.TrimSpace(effect) == "" {
		return CapabilityHandleProvenance{}, errors.New("capability handle provenance requires a handle ID and effect")
	}
	if err := validateCapabilityURI(capabilityURI); err != nil {
		return CapabilityHandleProvenance{}, err
	}
	withinGrant := slices.Contains(boundary.GrantedEffects, effect)
	status := CapabilityHandleUnknown
	switch {
	case boundary.BoundaryStatus == "RESULT_UNOBSERVED":
		status = CapabilityHandleUnknown
	case !withinGrant || !boundary.RequestWithinGrant || !boundary.ResultWithinGrant || !boundary.ResultSourceRevisionOK:
		status = CapabilityHandleRejected
	case boundary.BoundaryStatus == "WITHIN_DECLARED_GRANT" || boundary.BoundaryStatus == "NO_EFFECT":
		status = CapabilityHandleObserved
	default:
		status = CapabilityHandleRejected
	}
	return CapabilityHandleProvenance{
		Schema:                     CapabilityHandleProvenanceSchema,
		HandleID:                   handleID,
		CapabilityURI:              capabilityURI,
		Effect:                     effect,
		AuthorityDigest:            boundary.AuthorityDigest,
		AuthorityObservationDigest: boundary.StableHash(),
		WorkloadProvenanceDigest:   workload.StableHash(),
		SourceRevision:             boundary.SourceRevision,
		OperationID:                boundary.OperationID,
		ObservationStatus:          status,
		EffectWithinGrant:          withinGrant,
		NonAuthorizing:             true,
	}, nil
}

// Validate recomputes the handle binding against the exact source and boundary.
func (p CapabilityHandleProvenance) Validate(workload WorkloadAuthorityProvenance, boundary AuthorityBoundaryObservation) error {
	if p.Schema != CapabilityHandleProvenanceSchema || !p.NonAuthorizing {
		return errors.New("capability handle provenance is not a non-authorizing witness")
	}
	expected, err := ObserveCapabilityHandleProvenance(workload, boundary, p.HandleID, p.CapabilityURI, p.Effect)
	if err != nil {
		return err
	}
	if p != expected {
		return errors.New("capability handle provenance does not match the exact boundary")
	}
	return nil
}

// Canonical is display-independent and suitable for a provenance digest.
func (p CapabilityHandleProvenance) Canonical() string {
	return strings.Join([]string{
		"capability-handle-provenance", p.Schema, p.HandleID, p.CapabilityURI, p.Effect,
		p.AuthorityDigest, p.AuthorityObservationDigest, p.WorkloadProvenanceDigest,
		p.SourceRevision.ID, p.SourceRevision.Digest, p.OperationID, p.ObservationStatus,
		strings.TrimSpace(strings.Join([]string{boolString(p.EffectWithinGrant), boolString(p.NonAuthorizing)}, "\x00")),
	}, "\t")
}

func (p CapabilityHandleProvenance) StableHash() string {
	return envelopeDigestString(p.Canonical())
}

func validateCapabilityURI(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "gooo" || parsed.Host == "" || parsed.Host != strings.ToLower(parsed.Host) || parsed.User != nil || parsed.Path == "" || parsed.Path == "/" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("capability URI must be canonical")
	}
	return nil
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
