package candidatedelivery

import (
	"encoding/json"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport/originpolicy"
)

type Eligibility struct {
	Schema       string          `json:"schema"`
	State        string          `json:"state"`
	Reason       string          `json:"reason"`
	Continue     bool            `json:"continue"`
	SubjectSHA   string          `json:"subject_sha,omitempty"`
	Unknown      json.RawMessage `json:"unknown,omitempty"`
	PolicyDigest string          `json:"policy_digest"`
	Digest       string          `json:"digest"`
}

// ObserveEligibility preserves unresolved authorization; it grants no execution.
func ObserveEligibility(requestRaw, resolutionRaw, verificationRaw []byte) Eligibility {
	r := Eligibility{Schema: "gooo/candidate-continuation-eligibility/v1", State: "REFUTED",
		Reason: "ELIGIBILITY_BINDING_MISMATCH", PolicyDigest: originpolicy.SourceDigest()}
	var request selfimprovementcandidate.AuthorizationRequest
	var resolution selfimprovementcandidate.AuthorizationResolution
	var verification selfimprovementcandidate.AuthorizationVerification
	if json.Unmarshal(resolutionRaw, &resolution) != nil ||
		selfimprovementcandidate.ValidateAuthorizationResolution(resolution) != nil {
		return sealEligibility(r)
	}
	if resolution.Decision == "UNKNOWN" {
		var header struct {
			Unknown json.RawMessage `json:"unknown"`
		}
		_ = json.Unmarshal(resolutionRaw, &header)
		r.State, r.Reason, r.Unknown = "UNKNOWN", resolution.Reason, header.Unknown
		return sealEligibility(r)
	}
	if json.Unmarshal(requestRaw, &request) != nil || json.Unmarshal(verificationRaw, &verification) != nil ||
		selfimprovementcandidate.ValidateAuthorizationRequest(request) != nil ||
		selfimprovementcandidate.VerifyAuthorizationResolution(request, resolution) != nil {
		return sealEligibility(r)
	}
	expected, err := selfimprovementcandidate.BuildAuthorizationVerification(request, resolution)
	if err != nil || !reflect.DeepEqual(expected, verification) || !expected.DecisionVerified {
		return sealEligibility(r)
	}
	r.State, r.Reason, r.SubjectSHA = resolution.Decision, resolution.Reason, request.Candidate.SubjectSHA
	r.Continue = originpolicy.ContinuationEligible(resolution.Decision, resolution.Outcome, expected.DecisionVerified)
	return sealEligibility(r)
}

func sealEligibility(r Eligibility) Eligibility { r.Digest = ""; r.Digest = digestJSON(r); return r }
