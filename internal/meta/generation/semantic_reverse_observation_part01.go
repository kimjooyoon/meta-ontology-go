package generation

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// SemanticReverseObservationSchema identifies the independent read-back of a
// generated semantic operation envelope.
const SemanticReverseObservationSchema = "gooo/semantic-reverse-observation/v1"

var semanticOperationArtifactNames = [...]string{
	"operation-manifest.json",
	"effect-requests.ndjson",
	"effect-results.ndjson",
	"semantic-patch.json",
	"operation-receipt.json",
	"operation-report.md",
}

// SemanticReverseObservation records only evidence obtained by rereading the
// caller-owned generated artifacts and checking them against the .gooo source.
type SemanticReverseObservation struct {
	Schema                string            `json:"schema"`
	AuthorityDigest       string            `json:"authority_digest"`
	ScenarioID            string            `json:"scenario_id"`
	ReceiptDigest         string            `json:"receipt_digest"`
	ArtifactDigests       map[string]string `json:"artifact_digests"`
	ArtifactCount         int               `json:"artifact_count"`
	Decision              string            `json:"decision"`
	Reason                string            `json:"reason"`
	ObservationDecision   string            `json:"observation_decision"`
	ObservationReason     string            `json:"observation_reason"`
	ObservationCandidates int               `json:"observation_candidates"`
	Metrics               EnvelopeMetrics   `json:"metrics"`
}

// SemanticOperationRoundTrip keeps the generated IR beside its independent
// reverse observation so callers cannot mistake generation for verification.
type SemanticOperationRoundTrip struct {
	Generated SemanticOperationRun
	Reverse   SemanticReverseObservation
}

// GenerateAndReverseObserveSemanticOperationEnvelope performs generation from
// the .gooo authority and then independently rereads the generated envelope.
func GenerateAndReverseObserveSemanticOperationEnvelope(source []byte, scenarioID, outputDir string, observation *SemanticObservation) (SemanticOperationRoundTrip, error) {
	var roundTrip SemanticOperationRoundTrip
	var err error
	if observation == nil {
		roundTrip.Generated, err = GenerateSemanticOperationEnvelope(source, scenarioID, outputDir)
	} else {
		roundTrip.Generated, err = GenerateSemanticOperationEnvelopeWithObservation(source, scenarioID, outputDir, *observation)
	}
	if err != nil {
		return roundTrip, err
	}
	roundTrip.Reverse, err = ReverseObserveSemanticOperationEnvelope(source, outputDir)
	if err != nil {
		return SemanticOperationRoundTrip{}, err
	}
	return roundTrip, nil
}

// ReverseObserveSemanticOperationEnvelope derives evidence from the exact
// authority bytes and the six generated artifacts, never from generator state.
func ReverseObserveSemanticOperationEnvelope(source []byte, outputDir string) (SemanticReverseObservation, error) {
	var observation SemanticReverseObservation
	if len(source) == 0 {
		return observation, errors.New(".gooo authority is empty")
	}
	if err := validateSemanticOperationAuthority(source); err != nil {
		return observation, err
	}
	if outputDir == "" {
		return observation, errors.New("caller-owned output directory is required")
	}
	verified, err := VerifySemanticOperationEnvelope(outputDir)
	if err != nil {
		return observation, err
	}
	authorityDigest := envelopeDigestBytes(source)
	manifestBytes, err := os.ReadFile(filepath.Join(outputDir, semanticOperationArtifactNames[0]))
	if err != nil {
		return observation, fmt.Errorf("read operation manifest: %w", err)
	}
	var manifest semanticOperationManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return observation, fmt.Errorf("decode operation manifest: %w", err)
	}
	if manifest.AuthorityDigest != authorityDigest {
		return observation, errors.New("reverse observation authority digest mismatch")
	}
	if manifest.ScenarioID != verified.ScenarioID {
		return observation, errors.New("reverse observation scenario mismatch")
	}
	artifactDigests := make(map[string]string, len(semanticOperationArtifactNames))
	for _, name := range semanticOperationArtifactNames {
		contents, err := os.ReadFile(filepath.Join(outputDir, name))
		if err != nil {
			return observation, fmt.Errorf("read %s: %w", name, err)
		}
		artifactDigests[name] = envelopeDigestBytes(contents)
	}
	observation = SemanticReverseObservation{
		Schema:                SemanticReverseObservationSchema,
		AuthorityDigest:       authorityDigest,
		ScenarioID:            verified.ScenarioID,
		ReceiptDigest:         verified.ReceiptDigest,
		ArtifactDigests:       artifactDigests,
		ArtifactCount:         len(semanticOperationArtifactNames),
		Decision:              verified.Decision,
		Reason:                verified.Reason,
		ObservationDecision:   verified.ObservationDecision,
		ObservationReason:     verified.ObservationReason,
		ObservationCandidates: verified.ObservationCandidates,
		Metrics:               verified.Metrics,
	}
	if err := ValidateSemanticReverseObservation(observation); err != nil {
		return SemanticReverseObservation{}, err
	}
	return observation, nil
}

// ValidateSemanticReverseObservation rejects incomplete or unbound read-back
// evidence without asserting that the observed result is useful.
func ValidateSemanticReverseObservation(observation SemanticReverseObservation) error {
	if observation.Schema != SemanticReverseObservationSchema ||
		!knownEnvelopeDigest(observation.AuthorityDigest) ||
		observation.ScenarioID == "" ||
		!knownEnvelopeDigest(observation.ReceiptDigest) ||
		observation.ArtifactCount != len(semanticOperationArtifactNames) ||
		len(observation.ArtifactDigests) != len(semanticOperationArtifactNames) ||
		observation.Metrics.OutputArtifactFiles != len(semanticOperationArtifactNames) {
		return errors.New("semantic reverse observation is incomplete")
	}
	for _, name := range semanticOperationArtifactNames {
		if !knownEnvelopeDigest(observation.ArtifactDigests[name]) {
			return errors.New("semantic reverse observation contains an unbound artifact")
		}
	}
	if observation.Decision != "CLOSED" && observation.Decision != "UNKNOWN" && observation.Decision != "REFUTED" {
		return errors.New("semantic reverse observation decision is invalid")
	}
	return nil
}
