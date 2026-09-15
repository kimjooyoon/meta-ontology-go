package main

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type artifactProvenance = generation.ArtifactProvenance

func artifactProvenanceBound(
	provenance artifactProvenance, baseSHA, headSHA string,
) bool {
	ledger := strings.TrimPrefix(provenance.IndicatorDecisionLedgerDigest, "sha256:")
	return provenance.SchemaVersion == generation.ArtifactProvenanceSchemaVersion &&
		provenance.BaseSHA == baseSHA && provenance.HeadSHA == headSHA &&
		provenance.Decision == generation.ArtifactProvenanceDecisionBound &&
		provenance.Reason == generation.ArtifactProvenanceReasonBound &&
		provenance.Summary == (generation.ArtifactProvenanceSummary{Pass: 4}) &&
		!provenance.PromotionAuthorized && len(provenance.Indicators) == 4 &&
		len(ledger) == 64 && provenance.IndicatorDecisionLedgerCount >= 0 &&
		artifactProvenanceDigestsKnown(provenance)
}

func artifactProvenanceDigestsKnown(provenance artifactProvenance) bool {
	for _, digest := range []string{provenance.PlanDigest,
		provenance.ExecutionManifestDigest, provenance.ReceiptReportDigest,
		provenance.InputDigest, provenance.EnvelopeDigest, provenance.ReplayDigest} {
		if !validDigest(digest) {
			return false
		}
	}
	return true
}
