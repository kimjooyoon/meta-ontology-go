package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type artifactProvenance = generation.ArtifactProvenance

func readArtifactProvenance(
	filename, baseSHA, headSHA string,
) (artifactProvenance, error) {
	payload, err := os.ReadFile(filename)
	if err != nil {
		return artifactProvenance{}, fmt.Errorf("read artifact provenance: %w", err)
	}
	var provenance artifactProvenance
	if err := json.Unmarshal(payload, &provenance); err != nil {
		return artifactProvenance{}, fmt.Errorf("decode artifact provenance: %w", err)
	}
	if !artifactProvenanceBound(provenance, baseSHA, headSHA) {
		return artifactProvenance{}, fmt.Errorf("artifact provenance is not bound to CI identity")
	}
	return provenance, nil
}

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
		if len(digest) != 64 {
			return false
		}
	}
	return true
}
