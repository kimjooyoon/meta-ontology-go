package semantic

import (
	"reflect"
	"strings"
	"testing"
)

func inferenceTestDigest(value string) string { return StableHashString("inference-test/" + value) }

func inferenceEdgeFixture(kind InferenceKind, suffix string) InferenceEdge {
	semantic := inferenceTestDigest("semantic/" + suffix)
	before := SnapshotDigests{Source: inferenceTestDigest("source-before/" + suffix), Semantic: semantic}
	after := SnapshotDigests{Source: inferenceTestDigest("source-after/" + suffix), Semantic: semantic}
	controls := InferenceControls{}
	authority := AuthorityBinding{}
	phase := PhasePlacement{Ordinal: 1}
	switch kind {
	case InferenceAuthoritativeDeclaration:
		phase.Phase, authority.Layer, authority.Effect = PhaseDeclaration, AuthoritySource, AuthorityDeclare
	case InferenceDeterministicDerivation:
		phase.Phase, authority.Layer, authority.Effect = PhaseDerivation, AuthoritySemantic, AuthorityDerive
	case InferenceDerivedProjection:
		phase.Phase, authority.Layer, authority.Effect = PhaseProjection, AuthorityDerived, AuthorityProject
		controls.Profile = ProfileBinding{ID: "gooo.test.profile.v1", Version: "1", Digest: inferenceTestDigest("profile")}
	case InferenceObservationCandidate:
		phase.Phase, authority.Layer, authority.Effect = PhaseObservation, AuthorityCandidate, AuthorityObserve
		controls.CatalogDigest = inferenceTestDigest("catalog")
	case InferenceAcceptedLift:
		phase.Phase, authority.Layer, authority.Effect = PhaseLift, AuthoritySemantic, AuthorityLift
		controls.PolicyDigest = inferenceTestDigest("policy")
	case InferenceIndependentVerification:
		phase.Phase, authority.Layer, authority.Effect = PhaseVerification, AuthorityVerification, AuthorityVerify
		controls.PolicyDigest = inferenceTestDigest("policy")
	}
	evidenceID := MustIdentity("inference-test://evidence/" + suffix)
	edge := InferenceEdge{
		InferenceRecord: InferenceRecord{
			RecordID:  MustIdentity("inference-test://record/" + suffix),
			SubjectID: MustIdentity("inference-test://subject/" + suffix),
			ObjectID:  MustIdentity("inference-test://object/" + suffix),
			Rule: RuleBinding{
				ID: MustIdentity("inference-test://rule/v1"), Version: "1", Digest: inferenceTestDigest("rule"),
			},
			Phase: phase, Before: before, After: after, Authority: authority, Controls: controls,
			Evidence: []EvidenceReference{{ID: evidenceID, Digest: inferenceTestDigest("payload/" + suffix)}},
		},
	}
	if kind == InferenceAuthoritativeDeclaration {
		edge.SourceRoots = []ID{MustIdentity("inference-test://source/root/" + suffix)}
	}
	if kind == InferenceAcceptedLift {
		edge.AcceptanceReceipt = evidenceID
	}
	edge.Kind = kind
	return edge
}

func inferenceEvidenceFixture(edge InferenceEdge) InferenceEvidence {
	ref := edge.Evidence[0]
	return InferenceEvidence{
		ID: ref.ID, Digest: ref.Digest, Before: edge.Before, After: edge.After, Controls: edge.Controls,
		SourceBacked: edge.Kind == InferenceAcceptedLift,
		Independent:  edge.Kind == InferenceIndependentVerification,
	}
}

func inferenceBundle(edge InferenceEdge) InferencePathV1 {
	return InferencePathV1{
		Version: InferencePathSchemaVersion, Edges: []InferenceEdge{edge},
		Evidence: []InferenceEvidence{inferenceEvidenceFixture(edge)},
	}
}

func assertInferencePathRejected(t *testing.T, path InferencePathV1) {
	t.Helper()
	if err := path.Validate(); err == nil {
		t.Fatal("invalid inference path was accepted")
	}
}

func TestInferencePathKindsAreClosedAndEvidenceBound(t *testing.T) {
	kinds := []InferenceKind{
		InferenceAuthoritativeDeclaration, InferenceDeterministicDerivation, InferenceDerivedProjection,
		InferenceObservationCandidate, InferenceAcceptedLift, InferenceIndependentVerification,
	}
	for _, kind := range kinds {
		t.Run(string(kind), func(t *testing.T) {
			edge := inferenceEdgeFixture(kind, strings.ToLower(string(kind)))
			if err := inferenceBundle(edge).Validate(); err != nil {
				t.Fatalf("valid edge rejected: %v", err)
			}
		})
	}
	if InferenceKind(NoSemanticDelta).Valid() {
		t.Fatal("semantic-change claim kind crossed into the inference-kind sum")
	}
	if InferenceKind("UNKNOWN").Valid() {
		t.Fatal("unknown inference kind was accepted")
	}
}

func TestSemanticChangeClaimsAreClosedAndTotal(t *testing.T) {
	makeClaim := func(kind SemanticChangeKind, suffix string) SemanticChangeClaim {
		edge := inferenceEdgeFixture(InferenceDeterministicDerivation, suffix)
		if kind == SemanticDelta {
			edge.After.Semantic = inferenceTestDigest("changed/" + suffix)
			edge.Authority.Effect = AuthorityDelta
		} else {
			edge.Authority.Effect = AuthorityNoChange
		}
		return SemanticChangeClaim{InferenceRecord: edge.InferenceRecord, Kind: kind, CanonicalDelta: func() string {
			if kind == SemanticDelta {
				return "field\t" + suffix + "\tchanged"
			}
			return ""
		}(), DeltaDigest: func() string {
			if kind == SemanticDelta {
				return StableHashString("field\t" + suffix + "\tchanged")
			}
			return ""
		}()}
	}
	for _, kind := range []SemanticChangeKind{SemanticDelta, NoSemanticDelta} {
		t.Run(string(kind), func(t *testing.T) {
			claim := makeClaim(kind, strings.ToLower(string(kind)))
			p := InferencePathV1{
				Version: InferencePathSchemaVersion, Claims: []SemanticChangeClaim{claim},
				Evidence: []InferenceEvidence{inferenceEvidenceFixture(InferenceEdge{InferenceRecord: claim.InferenceRecord})},
			}
			if err := p.Validate(); err != nil {
				t.Fatalf("valid claim rejected: %v", err)
			}
		})
	}
	cases := []struct {
		name string
		edit func(*SemanticChangeClaim)
	}{
		{"unknown kind", func(c *SemanticChangeClaim) { c.Kind = "UNKNOWN" }},
		{"delta equal snapshots", func(c *SemanticChangeClaim) {
			c.Kind = SemanticDelta
			c.CanonicalDelta = "x"
			c.DeltaDigest = StableHashString("x")
		}},
		{"no delta unequal snapshots", func(c *SemanticChangeClaim) {
			c.InferenceRecord.After.Semantic = inferenceTestDigest("other")
			c.Authority.Effect = AuthorityNoChange
		}},
		{"delta missing canonical delta", func(c *SemanticChangeClaim) {
			c.Kind = SemanticDelta
			c.InferenceRecord.After.Semantic = inferenceTestDigest("changed")
			c.DeltaDigest = ""
		}},
		{"delta digest mismatch", func(c *SemanticChangeClaim) {
			c.Kind = SemanticDelta
			c.InferenceRecord.After.Semantic = inferenceTestDigest("changed")
			c.CanonicalDelta = "x"
			c.DeltaDigest = inferenceTestDigest("wrong")
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			claim := makeClaim(NoSemanticDelta, "invalid-"+strings.ReplaceAll(test.name, " ", "-"))
			test.edit(&claim)
			assertInferencePathRejected(t, InferencePathV1{
				Version: InferencePathSchemaVersion, Claims: []SemanticChangeClaim{claim},
				Evidence: []InferenceEvidence{inferenceEvidenceFixture(InferenceEdge{InferenceRecord: claim.InferenceRecord})},
			})
		})
	}
}

func TestInferencePathCanonicalReplayAndNoMutation(t *testing.T) {
	first := inferenceEdgeFixture(InferenceDeterministicDerivation, "canonical-first")
	second := inferenceEdgeFixture(InferenceDerivedProjection, "canonical-second")
	left := InferencePathV1{
		Version: InferencePathSchemaVersion, Edges: []InferenceEdge{first, second},
		Evidence: []InferenceEvidence{inferenceEvidenceFixture(first), inferenceEvidenceFixture(second)},
	}
	right := InferencePathV1{
		Version: InferencePathSchemaVersion, Edges: []InferenceEdge{second, first},
		Evidence: []InferenceEvidence{inferenceEvidenceFixture(second), inferenceEvidenceFixture(first)},
	}
	leftBefore, rightBefore := left, right
	if left.Canonical() != right.Canonical() || left.StableHash() != right.StableHash() {
		t.Fatal("canonical replay changed with insertion order")
	}
	if !reflect.DeepEqual(left, leftBefore) || !reflect.DeepEqual(right, rightBefore) {
		t.Fatal("canonicalization mutated the input record sets")
	}
	empty := InferencePathV1{Version: InferencePathSchemaVersion}
	canonical := empty.Canonical()
	for _, marker := range []string{"edges\t0", "claims\t0", "evidence-records\t0"} {
		if !strings.Contains(canonical, marker) {
			t.Fatalf("canonical empty set omitted explicit marker %q: %s", marker, canonical)
		}
	}
	if left.Canonical() != left.Canonical() {
		t.Fatal("two clean canonical runs diverged")
	}
}
