package generation

import (
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
)

const AuthorityBoundaryObservationSchema = "gooo/authority-boundary-observation/v1"

// AuthorityBoundaryObservation is an evidence-only projection of a semantic
// operation's effect boundary. It never authorizes execution or promotion.
type AuthorityBoundaryObservation struct {
	Schema                 string
	AuthorityDigest        string
	SourceRevision         SourceRevision
	OperationID            string
	RequestDigest          string
	GrantDigest            string
	ResultDigest           string
	RequestedEffects       []string
	GrantedEffects         []string
	ObservedResultEffects  []string
	RequestWithinGrant     bool
	ResultWithinGrant      bool
	ResultSourceRevisionOK bool
	BoundaryStatus         string
}

// ObserveAuthorityBoundary calculates the declared effect boundary directly
// from semantic IR. It does not consult a workload identity or mint authority.
func ObserveAuthorityBoundary(ir SemanticOperationIR) (AuthorityBoundaryObservation, error) {
	if ir.Schema != SemanticOperationEnvelopeSchema {
		return AuthorityBoundaryObservation{}, errors.New("authority boundary observation requires semantic operation IR")
	}
	if !validEnvelopeDigest(ir.AuthorityDigest) {
		return AuthorityBoundaryObservation{}, errors.New("authority boundary observation requires an authority digest")
	}
	if strings.TrimSpace(ir.Source.ID) == "" || !validEnvelopeDigest(ir.Source.Digest) {
		return AuthorityBoundaryObservation{}, errors.New("authority boundary observation requires a source revision")
	}
	if strings.TrimSpace(ir.Intent.OperationID) == "" {
		return AuthorityBoundaryObservation{}, errors.New("authority boundary observation requires an operation ID")
	}
	if strings.TrimSpace(ir.Grant.GrantID) == "" || strings.TrimSpace(ir.Grant.ParentID) == "" {
		return AuthorityBoundaryObservation{}, errors.New("authority boundary observation requires a grant lineage")
	}

	requested := []string{}
	requestDigest := ""
	if ir.Request != nil {
		requested = normalizedEffects(ir.Request.Effects)
		requestDigest = envelopeDigestJSON(*ir.Request)
	}
	granted := normalizedEffects(ir.Grant.Effects)
	grantDigest := envelopeDigestJSON(ir.Grant)
	observed := []string{}
	resultDigest := ""
	resultSourceRevisionOK := true
	if ir.Result != nil {
		observed = normalizedEffects(ir.Result.Effects)
		resultDigest = envelopeDigestJSON(*ir.Result)
		resultSourceRevisionOK = ir.Result.SourceRevision == ir.Source.ID
	}
	requestWithinGrant := ir.Request == nil || envelopeSubset(requested, granted)
	resultWithinGrant := ir.Result == nil || envelopeSubset(observed, granted)
	status := "WITHIN_DECLARED_GRANT"
	switch {
	case !requestWithinGrant || !resultWithinGrant:
		status = "EFFECT_ESCALATION"
	case !resultSourceRevisionOK:
		status = "STALE_RESULT"
	case ir.Request == nil && ir.Result == nil:
		status = "NO_EFFECT"
	case ir.Result == nil:
		status = "RESULT_UNOBSERVED"
	}
	return AuthorityBoundaryObservation{
		Schema:                 AuthorityBoundaryObservationSchema,
		AuthorityDigest:        ir.AuthorityDigest,
		SourceRevision:         ir.Source,
		OperationID:            ir.Intent.OperationID,
		RequestDigest:          requestDigest,
		GrantDigest:            grantDigest,
		ResultDigest:           resultDigest,
		RequestedEffects:       requested,
		GrantedEffects:         granted,
		ObservedResultEffects:  observed,
		RequestWithinGrant:     requestWithinGrant,
		ResultWithinGrant:      resultWithinGrant,
		ResultSourceRevisionOK: resultSourceRevisionOK,
		BoundaryStatus:         status,
	}, nil
}

// Validate recomputes the observation against the exact semantic IR.
func (o AuthorityBoundaryObservation) Validate(ir SemanticOperationIR) error {
	if o.Schema != AuthorityBoundaryObservationSchema {
		return errors.New("invalid authority boundary observation schema")
	}
	expected, err := ObserveAuthorityBoundary(ir)
	if err != nil {
		return err
	}
	if o.AuthorityDigest != expected.AuthorityDigest || o.SourceRevision != expected.SourceRevision || o.OperationID != expected.OperationID ||
		o.RequestDigest != expected.RequestDigest || o.GrantDigest != expected.GrantDigest || o.ResultDigest != expected.ResultDigest ||
		!slices.Equal(o.RequestedEffects, expected.RequestedEffects) || !slices.Equal(o.GrantedEffects, expected.GrantedEffects) ||
		!slices.Equal(o.ObservedResultEffects, expected.ObservedResultEffects) || o.RequestWithinGrant != expected.RequestWithinGrant ||
		o.ResultWithinGrant != expected.ResultWithinGrant || o.ResultSourceRevisionOK != expected.ResultSourceRevisionOK || o.BoundaryStatus != expected.BoundaryStatus {
		return errors.New("authority boundary observation does not match the exact semantic IR")
	}
	return nil
}

// Canonical is display-independent and suitable for a provenance digest.
func (o AuthorityBoundaryObservation) Canonical() string {
	return strings.Join([]string{
		"authority-boundary-observation", o.Schema, o.AuthorityDigest, o.SourceRevision.ID, o.SourceRevision.Digest,
		o.OperationID, o.RequestDigest, o.GrantDigest, o.ResultDigest,
		strings.Join(o.RequestedEffects, "\x00"), strings.Join(o.GrantedEffects, "\x00"), strings.Join(o.ObservedResultEffects, "\x00"),
		fmt.Sprint(o.RequestWithinGrant), fmt.Sprint(o.ResultWithinGrant), fmt.Sprint(o.ResultSourceRevisionOK), o.BoundaryStatus,
	}, "\t")
}

func (o AuthorityBoundaryObservation) StableHash() string { return envelopeDigestString(o.Canonical()) }

func normalizedEffects(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !slices.Contains(out, value) {
			out = append(out, value)
		}
	}
	slices.Sort(out)
	return out
}

func validEnvelopeDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
