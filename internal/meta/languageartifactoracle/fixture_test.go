package languageartifactoracle

import (
	"encoding/json"
	"strings"
)

const sourceFixture = "package billing\nnamespace billing\n\n" +
	"entity Order id \"billing://entity/order\"\n" +
	"entity PaymentMethod id \"billing://entity/payment-method\"\n" +
	"entity Payment id \"billing://entity/payment\"\n\n" +
	"activity PayOrder(Order, PaymentMethod) -> Payment\n"

func artifactFixture() sourceArtifact {
	sourceDigest := digestBytes([]byte(sourceFixture))
	semanticDigest := "sha256:" + strings.Repeat("a", 64)
	artifact := sourceArtifact{Schema: SourceArtifactSchema, Decision: "PASS",
		Reason: "SOURCE_ACTIVITY_EXECUTED", Resolution: "EXACT",
		Filename: "examples/billing/main.gooo", SourceDigest: sourceDigest,
		SemanticDigest: semanticDigest, Entry: artifactEntry{Package: "billing", Namespace: "billing",
			Activity: "PayOrder", Inputs: []artifactBinding{{"Order", "billing://entity/order"},
				{"PaymentMethod", "billing://entity/payment-method"}},
			Output: artifactBinding{"Payment", "billing://entity/payment"}},
		Events: []artifactEvent{{1, "SOURCE_PARSED", sourceDigest},
			{2, "SEMANTIC_LOWERED", semanticDigest}, {3, "ACTIVITY_INVOKED", "PayOrder"},
			{4, "ENTITY_PRODUCED", "billing://entity/payment"}},
		Diagnostics: []artifactDiagnostic{}, Effects: artifactEffects{}}
	artifact.Digest = artifactDigest(artifact)
	return artifact
}

func artifactJSON(artifact sourceArtifact) []byte {
	raw, err := json.Marshal(artifact)
	if err != nil {
		panic(err)
	}
	return append(raw, '\n')
}

func cloneArtifact(artifact sourceArtifact) sourceArtifact {
	raw, err := json.Marshal(artifact)
	if err != nil {
		panic(err)
	}
	var clone sourceArtifact
	if err := json.Unmarshal(raw, &clone); err != nil {
		panic(err)
	}
	return clone
}
