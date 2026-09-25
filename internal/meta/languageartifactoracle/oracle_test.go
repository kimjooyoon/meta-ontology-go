package languageartifactoracle

import "testing"

func TestIndependentOracleBoundaries(t *testing.T) {
	genuine := artifactFixture()
	forged := cloneArtifact(genuine)
	forged.Entry.Output.ID = "forged://entity/payment"
	forged.Events[3].Subject = "forged://entity/payment"
	forged.Digest = artifactDigest(forged)
	unknown := cloneArtifact(genuine)
	unknown.Decision = "UNKNOWN"
	unknown.Digest = artifactDigest(unknown)
	cases := []struct {
		name, source, reason, resolution string
		artifact                         sourceArtifact
	}{
		{"genuine", sourceFixture, "ARTIFACT_SOURCE_PROJECTION_EXACT", "EXACT", genuine},
		{"forged", sourceFixture, "ARTIFACT_SOURCE_PROJECTION_MISMATCH", "INVARIANT_ONLY", forged},
		{"unknown", sourceFixture, "ARTIFACT_DECISION_UNKNOWN", "LOWER_RESOLUTION", unknown},
		{"unsupported", "unknown Thing\n", "ORACLE_SOURCE_PROJECTION_UNKNOWN", "LOWER_RESOLUTION", genuine},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result := evaluateArtifact([]byte(test.source), artifactJSON(test.artifact),
				"examples/billing/main.gooo", "PayOrder")
			if result.Reason != test.reason || result.Resolution != test.resolution {
				t.Fatalf("got %s/%s want %s/%s", result.Reason, result.Resolution, test.reason, test.resolution)
			}
		})
	}
}
