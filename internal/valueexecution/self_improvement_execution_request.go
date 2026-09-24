package valueexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementExecutionRequestObservationSchema = "gooo.self-improvement.execution-request.v1"

type SelfImprovementExecutionRequestStatus string

const SelfImprovementExecutionRequestStatusReady SelfImprovementExecutionRequestStatus = "READY"
const SelfImprovementExecutionRequestStatusRejected SelfImprovementExecutionRequestStatus = "REJECTED"
const SelfImprovementExecutionRequestStatusUnknown SelfImprovementExecutionRequestStatus = "UNKNOWN"

type SelfImprovementExecutionRequestContext struct {
	RequestID              string `json:"request_id"`
	SourceURI              string `json:"source_uri"`
	EnvironmentDigest      string `json:"environment_digest"`
	WorkloadIdentityDigest string `json:"workload_identity_digest"`
	GatewayPolicyDigest    string `json:"gateway_policy_digest"`
}

type SelfImprovementExecutionRequestObservation struct {
	Schema          string                                 `json:"schema"`
	SelectionDigest string                                 `json:"selection_digest"`
	Request         SelfImprovementExecutionRequestContext `json:"request"`
	Status          SelfImprovementExecutionRequestStatus  `json:"status"`
	Reason          string                                 `json:"reason"`
	NonAuthorizing  bool                                   `json:"non_authorizing"`
	Digest          string                                 `json:"digest"`
}

// ObserveSelfImprovementExecutionRequest validates the evidence envelope for
// a selected candidate. READY is not an execution grant or authorization.
func ObserveSelfImprovementExecutionRequest(
	selection SelfImprovementCandidateSelectionObservation,
	request SelfImprovementExecutionRequestContext,
) SelfImprovementExecutionRequestObservation {
	observation := SelfImprovementExecutionRequestObservation{
		Schema:          SelfImprovementExecutionRequestObservationSchema,
		SelectionDigest: selection.Digest,
		Request:         request,
		Status:          SelfImprovementExecutionRequestStatusUnknown,
		Reason:          "EXECUTION_REQUEST_UNKNOWN",
		NonAuthorizing:  true,
	}
	switch {
	case !selection.NonAuthorizing || selection.Digest == "":
		observation.Reason = "EXECUTION_REQUEST_SELECTION_UNTRUSTED"
	case selection.Status == SelfImprovementCandidateSelectionStatusUnknown:
		observation.Reason = "EXECUTION_REQUEST_SELECTION_UNKNOWN"
	case selection.Status != SelfImprovementCandidateSelectionStatusSelected:
		observation.Status = SelfImprovementExecutionRequestStatusRejected
		observation.Reason = "EXECUTION_REQUEST_CANDIDATE_NOT_SELECTED"
	case request.RequestID == "" || request.SourceURI == "":
		observation.Reason = "EXECUTION_REQUEST_IDENTITY_MISSING"
	case request.EnvironmentDigest == "" || request.WorkloadIdentityDigest == "" || request.GatewayPolicyDigest == "":
		observation.Reason = "EXECUTION_REQUEST_SECURITY_DIGEST_MISSING"
	case !validDigest(request.EnvironmentDigest) || !validDigest(request.WorkloadIdentityDigest) || !validDigest(request.GatewayPolicyDigest):
		observation.Reason = "EXECUTION_REQUEST_SECURITY_DIGEST_INVALID"
	default:
		observation.Status = SelfImprovementExecutionRequestStatusReady
		observation.Reason = "EXECUTION_REQUEST_EVIDENCE_READY"
	}
	observation.Digest = selfImprovementExecutionRequestObservationDigest(observation)
	return observation
}

func selfImprovementExecutionRequestObservationDigest(observation SelfImprovementExecutionRequestObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}