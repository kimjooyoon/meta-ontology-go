package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const SPIFFEExecutionBoundarySchemaPart01 = "gooo/spiffe-execution-boundary/v1"

type SPIFFEExecutionBoundaryStatusPart01 string

const (
	SPIFFEExecutionBoundaryUnknownPart01  SPIFFEExecutionBoundaryStatusPart01 = "UNKNOWN"
	SPIFFEExecutionBoundaryObservedPart01 SPIFFEExecutionBoundaryStatusPart01 = "OBSERVED"
)

// SPIFFEExecutionBoundaryMetricsPart01 records evidence counts only. It never
// grants workload authority or changes the decision made by the envelope.
type SPIFFEExecutionBoundaryMetricsPart01 struct {
	GeneratedArtifactCount  int `json:"generated_artifact_count"`
	ReverseObservationCount int `json:"reverse_observation_count"`
	EvidenceStageCount      int `json:"evidence_stage_count"`
}

// SPIFFEExecutionBoundaryInputPart01 is the input to the generated envelope.
// Every digest is an identity of an already-observed stage, not an authority.
type SPIFFEExecutionBoundaryInputPart01 struct {
	WorkloadIdentity          string
	WorkloadAttestationDigest string
	TrustBundleDigest         string
	NonAuthorizing            bool
	GoooDeclarationDigest     string
	IRDigest                  string
	GeneratedArtifactDigest   string
	ReverseObservationDigest  string
	ExecutionPlanDigest       string
	Metrics                   SPIFFEExecutionBoundaryMetricsPart01
}

// SPIFFEExecutionBoundaryEnvelopePart01 binds declaration, IR, generation,
// reverse observation, execution planning, and workload evidence while
// remaining explicitly non-authorizing.
type SPIFFEExecutionBoundaryEnvelopePart01 struct {
	Schema                    string                               `json:"schema"`
	WorkloadIdentity          string                               `json:"workload_identity"`
	WorkloadAttestationDigest string                               `json:"workload_attestation_digest"`
	TrustBundleDigest         string                               `json:"trust_bundle_digest"`
	GoooDeclarationDigest     string                               `json:"gooo_declaration_digest"`
	IRDigest                  string                               `json:"ir_digest"`
	GeneratedArtifactDigest   string                               `json:"generated_artifact_digest"`
	ReverseObservationDigest  string                               `json:"reverse_observation_digest"`
	ExecutionPlanDigest       string                               `json:"execution_plan_digest"`
	Status                    SPIFFEExecutionBoundaryStatusPart01  `json:"status"`
	Reason                    string                               `json:"reason"`
	MissingStageIndex         int                                  `json:"missing_stage_index"`
	NonAuthorizing            bool                                 `json:"non_authorizing"`
	Metrics                   SPIFFEExecutionBoundaryMetricsPart01 `json:"metrics"`
	ObservationDigest         string                               `json:"observation_digest"`
}

// SPIFFEExecutionBoundaryReverseObservationPart01 is the independently
// observed result of replaying the generated envelope. It copies evidence and
// metrics without turning them into authorization.
type SPIFFEExecutionBoundaryReverseObservationPart01 struct {
	Schema                   string                               `json:"schema"`
	Status                   SPIFFEExecutionBoundaryStatusPart01  `json:"status"`
	Reason                   string                               `json:"reason"`
	MissingStageIndex        int                                  `json:"missing_stage_index"`
	NonAuthorizing           bool                                 `json:"non_authorizing"`
	GoooDeclarationDigest    string                               `json:"gooo_declaration_digest"`
	IRDigest                 string                               `json:"ir_digest"`
	GeneratedArtifactDigest  string                               `json:"generated_artifact_digest"`
	ReverseObservationDigest string                               `json:"reverse_observation_digest"`
	ExecutionPlanDigest      string                               `json:"execution_plan_digest"`
	Metrics                  SPIFFEExecutionBoundaryMetricsPart01 `json:"metrics"`
	EnvelopeDigest           string                               `json:"envelope_digest"`
}

// DigestSPIFFEGoooDeclarationPart01 computes the canonical source identity
// used for a .gooo declaration before lowering to IR.
func DigestSPIFFEGoooDeclarationPart01(source []byte) string {
	sum := sha256.Sum256(source)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// GenerateSPIFFEExecutionBoundaryEnvelopePart01 creates an evidence envelope.
// The first missing or invalid stage remains UNKNOWN and is never promoted.
func GenerateSPIFFEExecutionBoundaryEnvelopePart01(input SPIFFEExecutionBoundaryInputPart01) SPIFFEExecutionBoundaryEnvelopePart01 {
	envelope := SPIFFEExecutionBoundaryEnvelopePart01{
		Schema:                    SPIFFEExecutionBoundarySchemaPart01,
		WorkloadIdentity:          input.WorkloadIdentity,
		WorkloadAttestationDigest: input.WorkloadAttestationDigest,
		TrustBundleDigest:         input.TrustBundleDigest,
		GoooDeclarationDigest:     input.GoooDeclarationDigest,
		IRDigest:                  input.IRDigest,
		GeneratedArtifactDigest:   input.GeneratedArtifactDigest,
		ReverseObservationDigest:  input.ReverseObservationDigest,
		ExecutionPlanDigest:       input.ExecutionPlanDigest,
		Status:                    SPIFFEExecutionBoundaryUnknownPart01,
		Reason:                    "SPIFFE_EXECUTION_BOUNDARY_INCOMPLETE",
		MissingStageIndex:         -1,
		NonAuthorizing:            input.NonAuthorizing,
		Metrics:                   input.Metrics,
	}
	switch {
	case !input.NonAuthorizing:
		envelope.Reason = "SPIFFE_EXECUTION_BOUNDARY_AUTHORITY_INVALID"
	case !validSPIFFEWorkloadIdentityPart01(input.WorkloadIdentity):
		envelope.MissingStageIndex = 0
		envelope.Reason = "SPIFFE_EXECUTION_BOUNDARY_WORKLOAD_IDENTITY_INVALID"
	case !validSHA256DigestPart01(input.WorkloadAttestationDigest):
		envelope.MissingStageIndex = 1
		envelope.Reason = "SPIFFE_EXECUTION_BOUNDARY_ATTESTATION_INVALID"
	case !validSHA256DigestPart01(input.TrustBundleDigest):
		envelope.MissingStageIndex = 2
		envelope.Reason = "SPIFFE_EXECUTION_BOUNDARY_TRUST_BUNDLE_INVALID"
	case !validSHA256DigestPart01(input.GoooDeclarationDigest):
		envelope.MissingStageIndex = 3
		envelope.Reason = "SPIFFE_EXECUTION_BOUNDARY_GOOO_DECLARATION_INVALID"
	case !validSHA256DigestPart01(input.IRDigest):
		envelope.MissingStageIndex = 4
		envelope.Reason = "SPIFFE_EXECUTION_BOUNDARY_IR_INVALID"
	case !validSHA256DigestPart01(input.GeneratedArtifactDigest):
		envelope.MissingStageIndex = 5
		envelope.Reason = "SPIFFE_EXECUTION_BOUNDARY_GENERATED_ARTIFACT_INVALID"
	case !validSHA256DigestPart01(input.ReverseObservationDigest):
		envelope.MissingStageIndex = 6
		envelope.Reason = "SPIFFE_EXECUTION_BOUNDARY_REVERSE_OBSERVATION_INVALID"
	case !validSHA256DigestPart01(input.ExecutionPlanDigest):
		envelope.MissingStageIndex = 7
		envelope.Reason = "SPIFFE_EXECUTION_BOUNDARY_EXECUTION_PLAN_INVALID"
	case input.Metrics.GeneratedArtifactCount < 0 || input.Metrics.ReverseObservationCount < 0 || input.Metrics.EvidenceStageCount < 0:
		envelope.Reason = "SPIFFE_EXECUTION_BOUNDARY_METRICS_INVALID"
	default:
		envelope.Status = SPIFFEExecutionBoundaryObservedPart01
		envelope.Reason = "SPIFFE_EXECUTION_BOUNDARY_EVIDENCE_OBSERVED"
	}
	envelope.ObservationDigest = spiffeExecutionBoundaryEnvelopeDigestPart01(envelope)
	return envelope
}

// ReverseObserveSPIFFEExecutionBoundaryEnvelopePart01 independently observes
// the generated envelope and preserves UNKNOWN on malformed or tampered data.
func ReverseObserveSPIFFEExecutionBoundaryEnvelopePart01(envelope SPIFFEExecutionBoundaryEnvelopePart01) SPIFFEExecutionBoundaryReverseObservationPart01 {
	observation := SPIFFEExecutionBoundaryReverseObservationPart01{
		Schema:                   SPIFFEExecutionBoundarySchemaPart01,
		Status:                   SPIFFEExecutionBoundaryUnknownPart01,
		Reason:                   "SPIFFE_EXECUTION_BOUNDARY_REVERSE_UNKNOWN",
		MissingStageIndex:        envelope.MissingStageIndex,
		NonAuthorizing:           envelope.NonAuthorizing,
		GoooDeclarationDigest:    envelope.GoooDeclarationDigest,
		IRDigest:                 envelope.IRDigest,
		GeneratedArtifactDigest:  envelope.GeneratedArtifactDigest,
		ReverseObservationDigest: envelope.ReverseObservationDigest,
		ExecutionPlanDigest:      envelope.ExecutionPlanDigest,
		Metrics:                  envelope.Metrics,
		EnvelopeDigest:           envelope.ObservationDigest,
	}
	if err := envelope.Validate(); err != nil {
		observation.Reason = "SPIFFE_EXECUTION_BOUNDARY_ENVELOPE_INVALID"
		return observation
	}
	observation.Status = envelope.Status
	observation.Reason = envelope.Reason
	return observation
}

func (envelope SPIFFEExecutionBoundaryEnvelopePart01) Validate() error {
	if envelope.Schema != SPIFFEExecutionBoundarySchemaPart01 {
		return fmt.Errorf("unexpected SPIFFE execution boundary schema %q", envelope.Schema)
	}
	if !envelope.NonAuthorizing {
		return fmt.Errorf("SPIFFE execution boundary must remain non-authorizing")
	}
	if envelope.MissingStageIndex < -1 || envelope.MissingStageIndex > 7 {
		return fmt.Errorf("invalid missing stage index %d", envelope.MissingStageIndex)
	}
	if envelope.Status != SPIFFEExecutionBoundaryUnknownPart01 && envelope.Status != SPIFFEExecutionBoundaryObservedPart01 {
		return fmt.Errorf("unknown SPIFFE execution boundary status %q", envelope.Status)
	}
	if envelope.Status == SPIFFEExecutionBoundaryObservedPart01 {
		if !validSPIFFEWorkloadIdentityPart01(envelope.WorkloadIdentity) || !validSHA256DigestPart01(envelope.WorkloadAttestationDigest) || !validSHA256DigestPart01(envelope.TrustBundleDigest) || !validSHA256DigestPart01(envelope.GoooDeclarationDigest) || !validSHA256DigestPart01(envelope.IRDigest) || !validSHA256DigestPart01(envelope.GeneratedArtifactDigest) || !validSHA256DigestPart01(envelope.ReverseObservationDigest) || !validSHA256DigestPart01(envelope.ExecutionPlanDigest) {
			return fmt.Errorf("observed SPIFFE execution boundary has invalid evidence")
		}
		if envelope.MissingStageIndex != -1 || envelope.Reason != "SPIFFE_EXECUTION_BOUNDARY_EVIDENCE_OBSERVED" {
			return fmt.Errorf("observed SPIFFE execution boundary has inconsistent status")
		}
	}
	if envelope.Metrics.GeneratedArtifactCount < 0 || envelope.Metrics.ReverseObservationCount < 0 || envelope.Metrics.EvidenceStageCount < 0 {
		return fmt.Errorf("SPIFFE execution boundary metrics must be non-negative")
	}
	if envelope.ObservationDigest == "" || envelope.ObservationDigest != spiffeExecutionBoundaryEnvelopeDigestPart01(envelope) {
		return fmt.Errorf("SPIFFE execution boundary observation digest does not match evidence")
	}
	return nil
}

func spiffeExecutionBoundaryEnvelopeDigestPart01(value SPIFFEExecutionBoundaryEnvelopePart01) string {
	value.ObservationDigest = ""
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validSHA256DigestPart01(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func validSPIFFEWorkloadIdentityPart01(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "spiffe" && parsed.Hostname() != "" && parsed.Opaque == "" && parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" && strings.HasPrefix(parsed.Path, "/")
}
