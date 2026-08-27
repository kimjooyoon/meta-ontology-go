package languageartifactoracle

import "testing"

func TestSuiteCapturesSharedValidatorCounterexample(t *testing.T) {
	genuine := artifactFixture()
	forged := cloneArtifact(genuine)
	forged.Entry.Output.ID, forged.Events[3].Subject = "forged://payment", "forged://payment"
	forged.Digest = artifactDigest(forged)
	unknown := cloneArtifact(genuine)
	unknown.Decision, unknown.Digest = "UNKNOWN", ""
	unknown.Digest = artifactDigest(unknown)
	legacy := []byte(`{"decision":"PASS","summary":{"cases_satisfied":4,"cases_total":4}}`)
	report := Evaluate(Input{Contract: CanonicalContract(), HeadSHA: "head", Entry: "PayOrder",
		Filename: "examples/billing/main.gooo", UnsupportedFilename: "examples/language-artifact-oracle/unsupported.gooo",
		Source: []byte(sourceFixture), UnsupportedSource: []byte("unknown Thing\n"),
		Genuine: artifactJSON(genuine), Forged: artifactJSON(forged),
		UnknownDecision: artifactJSON(unknown), LegacyAcceptance: legacy,
		Independence: IndependenceEvidence{Schema: IndependenceSchema, ProducerDependencies: 0}})
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
}
