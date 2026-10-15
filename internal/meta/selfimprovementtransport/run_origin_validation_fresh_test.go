package selfimprovementtransport

import "testing"

func TestOriginReceiptValidationFresh1611(t *testing.T) {
	_, _, _, metadata, archive := fixture(t)
	report := evaluateFixture(t, metadata, archive)
	if err := CheckReadOnly(report); err != nil {
		t.Fatal(err)
	}
	changed := report
	changed.Provenance.RunOrigin.PolicyDigest = digestBytes([]byte("different policy"))
	changed.Digest = reportDigest(changed)
	if ValidateReport(changed) == nil {
		t.Fatal("consumer accepted another policy origin")
	}
	changed = report
	changed.Transport.ProducerEvent = ""
	changed.Provenance.RunOrigin = resolveRunOrigin(evaluationInput{
		producer: changed.Producer, metadata: changed.Transport,
	})
	changed.Digest = reportDigest(changed)
	if ValidateReport(changed) == nil {
		t.Fatal("closed resolution accepted an unresolved run origin")
	}
}
