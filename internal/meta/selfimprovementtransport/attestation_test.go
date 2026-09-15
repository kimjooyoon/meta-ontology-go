package selfimprovementtransport

import "testing"

func TestVerifiedAttestationClosesEHT8(t *testing.T) {
	_, _, _, metadata, archiveDigest := fixture(t)
	metadata.Attestation = Attestation{
		Status: "VERIFIED", Digest: digestBytes([]byte("attestation")), ProducerIdentity: "github-actions",
	}
	report := evaluateFixture(t, metadata, archiveDigest)
	if err := ValidateReport(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionPass || report.Metrics.VerifiedTotal != 8 ||
		report.Metrics.CoverageBasisPoints != 10000 {
		t.Fatalf("unexpected exact receipt: %+v", report)
	}
}
