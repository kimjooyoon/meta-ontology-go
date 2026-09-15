package selfimprovementattestation

import "fmt"

func validatePriorObligations(obligations []Obligation) error {
	verified, unknown := 0, 0
	for _, obligation := range obligations {
		switch obligation.Status {
		case "VERIFIED":
			verified++
		case "UNKNOWN":
			unknown++
			if obligation.ID != attestationID || obligation.Reason != "PRODUCER_ATTESTATION_UNKNOWN" {
				return fmt.Errorf("unexpected UNKNOWN obligation %q", obligation.ID)
			}
		default:
			return fmt.Errorf("unexpected prior obligation status %q", obligation.Status)
		}
	}
	if verified != 7 || unknown != 1 {
		return fmt.Errorf("prior obligation status totals mismatch")
	}
	return nil
}

func setAttestation(obligations []Obligation, status, reason, evidence string) []Obligation {
	result := append([]Obligation(nil), obligations...)
	for index := range result {
		if result[index].ID != attestationID {
			continue
		}
		result[index].Status = status
		result[index].Reason = reason
		result[index].EvidenceDigest = evidence
	}
	return result
}
