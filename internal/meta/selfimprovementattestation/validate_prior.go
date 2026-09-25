package selfimprovementattestation

import (
	"fmt"
	"reflect"
)

func validatePrior(request Request) error {
	receipt := request.TransportReceipt
	if receipt.Schema != transportSchema || receipt.MetricID != metricID {
		return fmt.Errorf("prior transport schema or metric mismatch")
	}
	if receipt.Contract.ObligationTotal != 8 || len(receipt.Obligations) != 8 {
		return fmt.Errorf("prior fixed denominator mismatch")
	}
	want := Metrics{8, 7, 1, 0, 1, 8750, 0}
	if !reflect.DeepEqual(receipt.Metrics, want) {
		return fmt.Errorf("prior metrics are not the 7/8 lower-resolution state")
	}
	if receipt.Decision != "OBSERVED" || receipt.Resolution != "LOWER_RESOLUTION" {
		return fmt.Errorf("prior decision is not lower-resolution OBSERVED")
	}
	if receipt.Reason != "PRODUCER_ATTESTATION_UNKNOWN" {
		return fmt.Errorf("prior reason does not identify producer attestation")
	}
	if len(receipt.OpenObligationIDs) != 1 || receipt.OpenObligationIDs[0] != attestationID {
		return fmt.Errorf("prior open obligation set mismatch")
	}
	if err := validatePriorBindings(receipt); err != nil {
		return err
	}
	return validatePriorObligations(receipt.Obligations)
}

func validatePriorBindings(receipt TransportReceipt) error {
	producer, transport := receipt.Producer, receipt.Transport
	if producer.Contract != receipt.Contract {
		return fmt.Errorf("producer contract mismatch")
	}
	if producer.SubjectSHA != receipt.SubjectSHA || producer.CheckoutSHA != receipt.SubjectSHA {
		return fmt.Errorf("logical subject binding mismatch")
	}
	if producer.RunID != transport.ProducerRunID || producer.RunAttempt != transport.ProducerRunAttempt {
		return fmt.Errorf("producer run binding mismatch")
	}
	if producer.ArtifactName != transport.ArtifactName {
		return fmt.Errorf("producer artifact name mismatch")
	}
	if producer.RepositoryURI != "https://github.com/"+transport.Repository {
		return fmt.Errorf("producer repository binding mismatch")
	}
	if producer.LogicalSubject.Digest != receipt.SourceObservationDigest {
		return fmt.Errorf("logical observation digest mismatch")
	}
	return nil
}
