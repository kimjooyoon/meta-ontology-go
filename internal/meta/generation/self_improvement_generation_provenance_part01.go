package generation

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementGenerationProvenanceSchema = "gooo/self-improvement-generation-provenance/v1"

const (
	SelfImprovementGenerationProvenanceBound   = "BOUND"
	SelfImprovementGenerationProvenanceUnknown = "UNKNOWN"
)

type SelfImprovementGenerationProvenanceInput struct {
	DeclarationDigest        string
	SourceDigest             string
	SemanticDigest           string
	GraphDigest              string
	GeneratedDigest          string
	ReverseObservationDigest string
}

type SelfImprovementGenerationProvenance struct {
	Schema                   string
	DeclarationDigest        string
	SourceDigest             string
	SemanticDigest           string
	GraphDigest              string
	GeneratedDigest          string
	ReverseObservationDigest string
	BindingStatus            string
	Reason                   string
	NonAuthorizing           bool
	ObservationDigest        string
}

func ObserveSelfImprovementGenerationProvenance(input SelfImprovementGenerationProvenanceInput) SelfImprovementGenerationProvenance {
	observation := SelfImprovementGenerationProvenance{
		Schema:                   SelfImprovementGenerationProvenanceSchema,
		DeclarationDigest:        input.DeclarationDigest,
		SourceDigest:             input.SourceDigest,
		SemanticDigest:           input.SemanticDigest,
		GraphDigest:              input.GraphDigest,
		GeneratedDigest:          input.GeneratedDigest,
		ReverseObservationDigest: input.ReverseObservationDigest,
		BindingStatus:            SelfImprovementGenerationProvenanceUnknown,
		NonAuthorizing:           true,
	}

	switch {
	case !validEnvelopeDigest(input.DeclarationDigest):
		observation.Reason = "DECLARATION_DIGEST_MISSING"
	case !validEnvelopeDigest(input.SourceDigest):
		observation.Reason = "SOURCE_DIGEST_MISSING"
	case !validEnvelopeDigest(input.SemanticDigest):
		observation.Reason = "SEMANTIC_DIGEST_MISSING"
	case !validEnvelopeDigest(input.GraphDigest):
		observation.Reason = "GRAPH_DIGEST_MISSING"
	case !validEnvelopeDigest(input.GeneratedDigest):
		observation.Reason = "GENERATED_DIGEST_MISSING"
	case !validEnvelopeDigest(input.ReverseObservationDigest):
		observation.Reason = "REVERSE_OBSERVATION_DIGEST_MISSING"
	default:
		observation.BindingStatus = SelfImprovementGenerationProvenanceBound
		observation.Reason = "GENERATION_EVIDENCE_BOUND"
	}
	observation.ObservationDigest = observation.StableHash()
	return observation
}

func (p SelfImprovementGenerationProvenance) Canonical() string {
	return strings.Join([]string{
		"self-improvement-generation-provenance",
		p.Schema,
		p.DeclarationDigest,
		p.SourceDigest,
		p.SemanticDigest,
		p.GraphDigest,
		p.GeneratedDigest,
		p.ReverseObservationDigest,
		p.BindingStatus,
		p.Reason,
	}, "\t")
}

func (p SelfImprovementGenerationProvenance) StableHash() string {
	digest, _ := cache.DigestOf(p.Canonical())
	return digest.String()
}

func (p SelfImprovementGenerationProvenance) Validate() error {
	if p.Schema != SelfImprovementGenerationProvenanceSchema ||
		!p.NonAuthorizing ||
		(p.BindingStatus != SelfImprovementGenerationProvenanceBound && p.BindingStatus != SelfImprovementGenerationProvenanceUnknown) ||
		p.Reason == "" ||
		p.ObservationDigest != p.StableHash() {
		return fmt.Errorf("self-improvement generation provenance is invalid")
	}
	return nil
}