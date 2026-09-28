package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func lspSelfImprovementDigest(value string) string {
	return cache.HashBytes([]byte(value)).String()
}

func lspSelfImprovementObservation(status valueexecution.SelfImprovementStatus) valueexecution.ExecutedSelfImprovementObservation {
	candidateDigest := lspSelfImprovementDigest("candidate")
	return valueexecution.ExecutedSelfImprovementObservation{
		Schema:          valueexecution.ExecutedSelfImprovementObservationSchema,
		Candidate:       valueexecution.SelfImprovementObservation{AfterOriginDigest: lspSelfImprovementDigest("origin"), EnvironmentDigest: lspSelfImprovementDigest("environment"), Status: status, Digest: candidateDigest},
		CandidateDigest: candidateDigest,
		Status:          status,
		Reason:          "METRIC_REGRESSED",
		NonAuthorizing:  true,
		Digest:          lspSelfImprovementDigest("execution-observation"),
	}
}

func TestObserveSelfImprovementProvenanceKeepsRegressionVisible(t *testing.T) {
	observation := ObserveSelfImprovementProvenance(
		"examples/language-runtime-binding/typed-chain.gooo",
		12,
		lspSelfImprovementObservation(valueexecution.SelfImprovementStatusRegressed),
	)
	if observation.Decision != SelfImprovementProvenanceVisible || observation.Reason != "SELF_IMPROVEMENT_STATUS_BOUND" {
		t.Fatalf("observation=%#v, want visible regression", observation)
	}
	if observation.CandidateStatus != valueexecution.SelfImprovementStatusRegressed || !observation.NonAuthorizing {
		t.Fatalf("observation=%#v, want non-authorizing regression status", observation)
	}
	if err := ValidateSelfImprovementProvenance(observation); err != nil {
		t.Fatalf("visible observation should validate: %v", err)
	}
}

func TestObserveSelfImprovementProvenancePreservesUnknown(t *testing.T) {
	observation := ObserveSelfImprovementProvenance(
		"examples/language-runtime-binding/typed-chain.gooo",
		12,
		lspSelfImprovementObservation(valueexecution.SelfImprovementStatusUnknown),
	)
	if observation.Decision != SelfImprovementProvenanceUnknown || observation.Reason != "METRIC_REGRESSED" {
		t.Fatalf("observation=%#v, want UNKNOWN evidence without promotion", observation)
	}
	if err := ValidateSelfImprovementProvenance(observation); err != nil {
		t.Fatalf("unknown observation should validate: %v", err)
	}
}

func TestObserveSelfImprovementProvenanceRejectsMalformedContext(t *testing.T) {
	observation := ObserveSelfImprovementProvenance("", -1, valueexecution.ExecutedSelfImprovementObservation{
		Status: valueexecution.SelfImprovementStatusImproved,
		Candidate: valueexecution.SelfImprovementObservation{
			OriginTransition: provenance.OriginChainTransitionChanged,
		},
	})
	if observation.Decision != SelfImprovementProvenanceUnknown || observation.Reason != "MISSING_DOCUMENT_URI" {
		t.Fatalf("observation=%#v, want UNKNOWN malformed context", observation)
	}
}
