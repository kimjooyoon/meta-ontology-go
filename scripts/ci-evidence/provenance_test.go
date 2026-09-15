package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestEvidenceRejectsUnboundArtifactProvenance(t *testing.T) {
	bundle := validEvidence()
	bundle.ArtifactProvenance.HeadSHA = bundle.BaseSHA
	if err := validateEvidence(bundle); err == nil {
		t.Fatal("unbound artifact provenance was accepted")
	}
}

func TestOldEvidenceSchemaFailsClosed(t *testing.T) {
	bundle := validEvidence()
	bundle.Schema = "gooo/ci-evidence/v2"
	if err := validateEvidence(bundle); err == nil {
		t.Fatal("old evidence schema was accepted")
	}
}

func artifactProvenanceFixture(baseSHA, headSHA string) artifactProvenance {
	report := sourcepolicy.Report{
		Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{},
	}
	plan := generation.Build(baseSHA, headSHA, report)
	return generation.BindArtifactProvenance(
		plan, generation.BuildExecutionManifest(plan), generation.VerifyReceipts(plan, nil))
}
