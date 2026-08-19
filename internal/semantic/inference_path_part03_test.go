package semantic

import (
	"strings"
	"testing"
)

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
