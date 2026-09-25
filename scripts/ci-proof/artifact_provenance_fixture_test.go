package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func artifactProvenanceFixture(baseSHA, headSHA string) artifactProvenance {
	report := sourcepolicy.Report{
		Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{},
	}
	plan := generation.Build(baseSHA, headSHA, report)
	return generation.BindArtifactProvenance(
		plan, generation.BuildExecutionManifest(plan), generation.VerifyReceipts(plan, nil))
}
