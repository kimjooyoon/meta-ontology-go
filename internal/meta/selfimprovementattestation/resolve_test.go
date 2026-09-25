package selfimprovementattestation

import "testing"

func TestResolveDischargesProducerAttestation(t *testing.T) {
	receipt, err := Resolve(validRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateReceipt(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "OBSERVED" || receipt.Metrics.VerifiedTotal != 8 || receipt.Metrics.OpenTotal != 0 {
		t.Fatalf("unexpected exact resolution: %+v", receipt.Metrics)
	}
	if receipt.Views[0].CoverageBasisPoints != 8750 || receipt.Views[1].CoverageBasisPoints != 10000 {
		t.Fatalf("reader resolutions were collapsed: %+v", receipt.Views)
	}
}

func TestResolveKeepsUnavailableAttestationUnknown(t *testing.T) {
	request := validRequest()
	request.VerifierExitCode = 1
	request.Verification = nil
	receipt, err := Resolve(request)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateReceipt(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "UNKNOWN" || receipt.Metrics.CoverageBasisPoints != 8750 {
		t.Fatalf("unavailable attestation was promoted: %+v", receipt)
	}
}
