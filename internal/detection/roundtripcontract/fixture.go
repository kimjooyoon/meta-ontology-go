package roundtripcontract

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
)

var (
	//go:embed testdata/minimal.gooo
	minimalDSL []byte
	//go:embed testdata/minimal.go
	minimalGo []byte
)

// Fixture is the smallest source pair used to measure the evidence contract.
type Fixture struct {
	DSL         []byte
	Go          []byte
	Measurement Measurement
	Artifacts   []ArtifactRef
}

// MinimalFixture returns three nodes, two facts, and three generated regions.
func MinimalFixture() Fixture {
	dsl := append([]byte(nil), minimalDSL...)
	goSource := append([]byte(nil), minimalGo...)
	return Fixture{
		DSL: dsl,
		Go:  goSource,
		Measurement: Measurement{
			Nodes: 3, Facts: 2, Regions: int64(bytes.Count(goSource, []byte("//gooo:generated:start"))),
			SourceBytes: int64(len(dsl) + len(goSource)), Checks: 4,
		},
		Artifacts: []ArtifactRef{
			{Stage: StageDSL, URI: "fixture://billing/main.gooo", Format: "gooo", Digest: digest(dsl)},
			{Stage: StageIR, URI: "fixture://billing/ir", Format: "semantic-ir", Digest: "fixture-ir-v1"},
			{Stage: StageGo, URI: "fixture://billing/generated.go", Format: "go", Digest: digest(goSource)},
			{Stage: StageLiftedIR, URI: "fixture://billing/lifted-ir", Format: "semantic-ir", Digest: "fixture-lifted-ir-v1"},
		},
	}
}

// MinimalHypothesis states the falsifiable equivalence claim for this lane.
func MinimalHypothesis() Hypothesis {
	return Hypothesis{
		ID:        "H-roundtrip-stable-id-v1",
		Statement: "presentation-only renames and ordering changes preserve semantic evidence, while stable-ID or fact changes are detected",
		Falsifier: "a presentation-only mutation changes canonical evidence, or a semantic mutation produces no finding",
	}
}

// MinimalScenarios returns pass, negative, and deferred evidence cases.
func MinimalScenarios() []Scenario {
	fixture := MinimalFixture()
	hypothesis := MinimalHypothesis()
	base := Evidence{Version: Version, Artifacts: fixture.Artifacts, Measurement: fixture.Measurement}
	return []Scenario{
		{CaseID: "presentation-rename", Hypothesis: hypothesis, Mutation: "rename Go display symbol without changing stable ID", Expected: OutcomePass, Evidence: withCase(base, "presentation-rename", hypothesis.ID, OutcomePass)},
		{CaseID: "stable-id-change", Hypothesis: hypothesis, Mutation: "change billing://entity/order to billing://entity/purchase", Expected: OutcomeFail, Evidence: withCase(base, "stable-id-change", hypothesis.ID, OutcomeFail)},
		{CaseID: "gooo-hosted-adapter", Hypothesis: hypothesis, Mutation: "run without a gooo-hosted Go→IR adapter", Expected: OutcomeDeferred, Evidence: withCase(base, "gooo-hosted-adapter", hypothesis.ID, OutcomeDeferred)},
	}
}

func withCase(base Evidence, caseID, hypothesisID string, outcome Outcome) Evidence {
	base.CaseID, base.HypothesisID, base.Outcome = caseID, hypothesisID, outcome
	if outcome == OutcomeFail {
		base.Findings = []Finding{{Rule: "round-trip", Path: "go-ir", Identity: "billing://entity/order", Detail: "stable semantic ID changed"}}
	}
	if outcome == OutcomeDeferred {
		base.DeferredReason = "gooo-hosted adapter is not implemented in this stage"
		base.Artifacts = withoutStage(base.Artifacts, StageLiftedIR)
	}
	return base
}

func withoutStage(artifacts []ArtifactRef, stage Stage) []ArtifactRef {
	result := make([]ArtifactRef, 0, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.Stage != stage {
			result = append(result, artifact)
		}
	}
	return result
}

func digest(source []byte) string {
	sum := sha256.Sum256(source)
	return hex.EncodeToString(sum[:])
}
