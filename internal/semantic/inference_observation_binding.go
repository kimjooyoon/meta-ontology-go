package semantic

import (
	"errors"
	"fmt"
	"strings"
)

// InferenceObservationBinding is the non-authorizing identity of one
// declaration-to-observation path. It records where a semantic observation
// came from without granting permission to execute, mutate, or promote it.
type InferenceObservationBinding struct {
	Schema         string
	DeclarationID  ID
	SemanticID     ID
	ObservationID  ID
	SourceDigest   string
	SemanticDigest string
	ChainDigest    string
	EvidenceDigest string
}

const InferenceObservationBindingSchema = "gooo/inference-observation-binding/v1"

// BindInferenceObservation derives a stable reverse-observation identity from
// a finite inference chain. The chain must include a source declaration, a
// semantic derivation, and a candidate observation or independent verification.
func BindInferenceObservation(chain InferencePathChain) (InferenceObservationBinding, error) {
	normalized, err := NewInferencePathChain(chain.Edges...)
	if err != nil {
		return InferenceObservationBinding{}, err
	}
	if len(normalized.Edges) < 3 {
		return InferenceObservationBinding{}, fmt.Errorf("%w: observation binding requires declaration, derivation, and observation", ErrInferencePath)
	}
	declaration := normalized.Edges[0]
	if declaration.Kind != InferenceAuthoritativeDeclaration {
		return InferenceObservationBinding{}, fmt.Errorf("%w: observation binding must begin with authoritative declaration", ErrInferencePath)
	}
	if declaration.After.Source == "" {
		return InferenceObservationBinding{}, fmt.Errorf("%w: declaration must carry a source snapshot", ErrInferencePath)
	}
	hasDerivation := false
	for _, edge := range normalized.Edges[1 : len(normalized.Edges)-1] {
		if edge.Kind == InferenceDeterministicDerivation || edge.Kind == InferenceDerivedProjection {
			hasDerivation = true
			break
		}
	}
	if !hasDerivation {
		return InferenceObservationBinding{}, fmt.Errorf("%w: observation binding requires a semantic derivation", ErrInferencePath)
	}
	observation := normalized.Edges[len(normalized.Edges)-1]
	if observation.Kind != InferenceObservationCandidate && observation.Kind != InferenceIndependentVerification {
		return InferenceObservationBinding{}, fmt.Errorf("%w: observation binding must end with candidate observation or independent verification", ErrInferencePath)
	}
	if observation.After.Semantic == "" {
		return InferenceObservationBinding{}, fmt.Errorf("%w: observation must carry a semantic snapshot", ErrInferencePath)
	}
	if len(observation.Evidence) == 0 {
		return InferenceObservationBinding{}, fmt.Errorf("%w: observation must carry evidence references", ErrInferencePath)
	}
	return InferenceObservationBinding{
		Schema:         InferenceObservationBindingSchema,
		DeclarationID:  declaration.SubjectID,
		SemanticID:     declaration.ObjectID,
		ObservationID:  observation.ObjectID,
		SourceDigest:   declaration.After.Source,
		SemanticDigest: observation.After.Semantic,
		ChainDigest:    StableHashString(normalized.Canonical()),
		EvidenceDigest: StableHashString("inference-observation-evidence\n" + observation.Canonical()),
	}, nil
}

// Validate recomputes the binding from the supplied chain and rejects stale or
// tampered identities. It never treats the binding as execution authority.
func (b InferenceObservationBinding) Validate(chain InferencePathChain) error {
	if b.Schema != InferenceObservationBindingSchema {
		return errors.New("invalid inference observation binding schema")
	}
	sourceDigest, err := normalizeDigest(b.SourceDigest)
	if err != nil {
		return fmt.Errorf("source digest: %w", err)
	}
	semanticDigest, err := normalizeDigest(b.SemanticDigest)
	if err != nil {
		return fmt.Errorf("semantic digest: %w", err)
	}
	chainDigest, err := normalizeDigest(b.ChainDigest)
	if err != nil {
		return fmt.Errorf("chain digest: %w", err)
	}
	evidenceDigest, err := normalizeDigest(b.EvidenceDigest)
	if err != nil {
		return fmt.Errorf("evidence digest: %w", err)
	}
	expected, err := BindInferenceObservation(chain)
	if err != nil {
		return err
	}
	if b.SourceDigest != sourceDigest || b.SemanticDigest != semanticDigest || b.ChainDigest != chainDigest || b.EvidenceDigest != evidenceDigest {
		return errors.New("inference observation binding digests are not normalized")
	}
	if b.DeclarationID != expected.DeclarationID || b.SemanticID != expected.SemanticID || b.ObservationID != expected.ObservationID ||
		b.SourceDigest != expected.SourceDigest || b.SemanticDigest != expected.SemanticDigest || b.ChainDigest != expected.ChainDigest || b.EvidenceDigest != expected.EvidenceDigest {
		return errors.New("inference observation binding does not match the exact chain")
	}
	return nil
}

// Canonical is the stable, display-independent representation of the binding.
func (b InferenceObservationBinding) Canonical() string {
	var out strings.Builder
	out.WriteString("inference-observation-binding\t")
	for _, value := range []string{
		b.Schema, b.DeclarationID.String(), b.SemanticID.String(), b.ObservationID.String(),
		b.SourceDigest, b.SemanticDigest, b.ChainDigest, b.EvidenceDigest,
	} {
		writeCanonicalField(&out, value)
	}
	return out.String()
}

func (b InferenceObservationBinding) StableHash() string { return StableHashString(b.Canonical()) }
