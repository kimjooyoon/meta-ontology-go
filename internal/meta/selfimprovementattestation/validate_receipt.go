package selfimprovementattestation

import (
	"fmt"
	"reflect"
)

func ValidateReceipt(receipt ResolutionReceipt) error {
	if receipt.Schema != resolutionSchema || receipt.Metaprogram != metaprogram || receipt.MetricID != metricID {
		return fmt.Errorf("resolution identity mismatch")
	}
	if receipt.Authority != (Authority{}) {
		return fmt.Errorf("resolution granted authority")
	}
	if receipt.PriorMetrics != (Metrics{8, 7, 1, 0, 1, 8750, 0}) {
		return fmt.Errorf("prior fixed metric mismatch")
	}
	switch receipt.Decision {
	case "OBSERVED":
		if receipt.Metrics != (Metrics{8, 8, 0, 0, 0, 10000, 0}) || len(receipt.OpenObligationIDs) != 0 {
			return fmt.Errorf("observed resolution did not close 8/8")
		}
		if len(receipt.ClaimTransitions) != 1 || receipt.ClaimTransitions[0].After != "DISCHARGED" {
			return fmt.Errorf("observed claim was not discharged")
		}
	case "UNKNOWN":
		if !reflect.DeepEqual(receipt.Metrics, receipt.PriorMetrics) || receipt.Reason != "PRODUCER_ATTESTATION_UNAVAILABLE" {
			return fmt.Errorf("unknown resolution changed the metric")
		}
	case "FAIL_CLOSED":
		if receipt.Metrics.FalseTotal != 1 || receipt.Metrics.FalsePromotionCount != 0 {
			return fmt.Errorf("fail-closed metric mismatch")
		}
	default:
		return fmt.Errorf("unknown resolution decision %q", receipt.Decision)
	}
	copy := receipt
	copy.Digest = ""
	want, err := digestValue(copy)
	if err != nil {
		return err
	}
	if receipt.Digest != want {
		return fmt.Errorf("resolution digest mismatch")
	}
	return nil
}
