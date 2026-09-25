package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const WorkloadIdentityProvenanceBindingSchema = "gooo/workload-identity-provenance-binding/v1"

type WorkloadIdentityProvenanceBindingStatus string

const (
	WorkloadIdentityProvenanceBindingUnknown  WorkloadIdentityProvenanceBindingStatus = "UNKNOWN"
	WorkloadIdentityProvenanceBindingObserved WorkloadIdentityProvenanceBindingStatus = "OBSERVED"
)

// WorkloadIdentityProvenanceBinding records a non-authorizing relationship
// between workload identity evidence and the ordered self-improvement chain.
// It never turns identity, SVID, or trust-bundle observations into adoption
// or merge authority.
type WorkloadIdentityProvenanceBinding struct {
	Schema                 string                                  `json:"schema"`
	ProvenanceChainDigest  string                                  `json:"provenance_chain_digest"`
	IdentityEvidenceDigest string                                  `json:"identity_evidence_digest"`
	Status                 WorkloadIdentityProvenanceBindingStatus `json:"status"`
	Reason                 string                                  `json:"reason"`
	NonAuthorizing         bool                                    `json:"non_authorizing"`
	BindingDigest          string                                  `json:"binding_digest"`
}

// BindWorkloadIdentityEvidenceToProvenancePart01 joins already-observed
// workload identity evidence to a validated provenance chain. Invalid or
// incomplete evidence remains UNKNOWN and is never promoted by this function.
func BindWorkloadIdentityEvidenceToProvenancePart01(
	chain SelfImprovementProvenanceChainPart01,
	evidence WorkloadIdentityEvidence,
) WorkloadIdentityProvenanceBinding {
	binding := WorkloadIdentityProvenanceBinding{
		Schema:                 WorkloadIdentityProvenanceBindingSchema,
		ProvenanceChainDigest:  chain.ChainDigest,
		IdentityEvidenceDigest: evidence.EvidenceDigest,
		Status:                 WorkloadIdentityProvenanceBindingUnknown,
		Reason:                 "PROVENANCE_IDENTITY_BINDING_INCOMPLETE",
		NonAuthorizing:         true,
	}
	switch {
	case !evidence.NonAuthorizing:
		binding.Reason = "IDENTITY_EVIDENCE_INVALID_AUTHORITY"
	case evidence.EvidenceDigest == "" || workloadIdentityEvidenceDigest(evidence) != evidence.EvidenceDigest:
		binding.Reason = "IDENTITY_EVIDENCE_DIGEST_INVALID"
	case chain.Validate() != nil:
		binding.Reason = "PROVENANCE_CHAIN_INVALID"
	case chain.Status != SelfImprovementProvenanceChainBoundPart01:
		binding.Reason = "PROVENANCE_CHAIN_INCOMPLETE"
	case evidence.Status != WorkloadIdentityEvidenceObserved:
		binding.Reason = "IDENTITY_EVIDENCE_NOT_OBSERVED"
	default:
		binding.Status = WorkloadIdentityProvenanceBindingObserved
		binding.Reason = "PROVENANCE_AND_IDENTITY_EVIDENCE_BOUND"
	}
	binding.BindingDigest = workloadIdentityProvenanceBindingDigest(binding)
	return binding
}

func (binding WorkloadIdentityProvenanceBinding) Validate() error {
	if binding.Schema != WorkloadIdentityProvenanceBindingSchema {
		return fmt.Errorf("unexpected workload identity provenance binding schema %q", binding.Schema)
	}
	if !binding.NonAuthorizing {
		return fmt.Errorf("workload identity provenance binding must remain non-authorizing")
	}
	if binding.Status != WorkloadIdentityProvenanceBindingUnknown && binding.Status != WorkloadIdentityProvenanceBindingObserved {
		return fmt.Errorf("unknown workload identity provenance binding status %q", binding.Status)
	}
	if binding.Status == WorkloadIdentityProvenanceBindingObserved && binding.Reason != "PROVENANCE_AND_IDENTITY_EVIDENCE_BOUND" {
		return fmt.Errorf("observed binding has unexpected reason %q", binding.Reason)
	}
	if binding.BindingDigest == "" || binding.BindingDigest != workloadIdentityProvenanceBindingDigest(binding) {
		return fmt.Errorf("workload identity provenance binding digest does not match its evidence")
	}
	return nil
}

func workloadIdentityProvenanceBindingDigest(value WorkloadIdentityProvenanceBinding) string {
	value.BindingDigest = ""
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
