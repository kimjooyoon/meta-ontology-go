package languageprofile

import (
	"fmt"
	"reflect"
	"strings"
)

func Validate(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema {
		return fmt.Errorf("PROFILE_SCHEMA_UNKNOWN")
	}
	if receipt.Decision != "PASS" && receipt.Decision != "FAIL_CLOSED" {
		return fmt.Errorf("PROFILE_DECISION_UNKNOWN")
	}
	if receipt.Decision == "PASS" && receipt.Resolution != RunnerScopedResolution {
		return fmt.Errorf("PROFILE_RESOLUTION_UNKNOWN")
	}
	if receipt.Decision == "FAIL_CLOSED" && receipt.Resolution != "EXACT" && receipt.Resolution != "LOWER_RESOLUTION" {
		return fmt.Errorf("PROFILE_RESOLUTION_UNKNOWN")
	}
	if receipt.Effects.RepositoryWrites != 0 || receipt.Effects.MutationAuthority {
		return fmt.Errorf("PROFILE_EFFECT_BOUNDARY_VIOLATED")
	}
	if receipt.Digest != receiptDigest(receipt) || !strings.HasPrefix(receipt.SourceDigest, "sha256:") {
		return fmt.Errorf("PROFILE_DIGEST_INVALID")
	}
	if receipt.Runner.GoVersion == "" || receipt.Runner.OS == "" || receipt.Runner.Architecture == "" ||
		!reflect.DeepEqual(receipt.NotClaimed, DefaultNonClaims()) {
		return fmt.Errorf("PROFILE_CONTEXT_UNKNOWN")
	}
	if receipt.Decision == "FAIL_CLOSED" {
		if receipt.Reason == "" {
			return fmt.Errorf("PROFILE_FAILURE_REASON_UNKNOWN")
		}
		return nil
	}
	return validateSuccess(receipt)
}

func validateSuccess(receipt Receipt) error {
	if receipt.Reason != "LANGUAGE_PROFILE_OBSERVED" || receipt.ProfiledEntry.Activity == "" ||
		receipt.SemanticDigest == "" || receipt.Summary.SamplesRequested < 1 ||
		receipt.Summary.SamplesRequested > MaximumSamples || !reflect.DeepEqual(receipt.Summary,
		summarize(receipt.Summary.SamplesRequested, receipt.Samples)) {
		return fmt.Errorf("PROFILE_SUCCESS_INVALID")
	}
	if receipt.Summary.SamplesObserved != receipt.Summary.SamplesRequested ||
		receipt.Summary.SuccessfulExecutions != receipt.Summary.SamplesRequested ||
		receipt.Summary.ExecutionDigestVariants != 1 ||
		receipt.Summary.WallObservations != receipt.Summary.SamplesRequested ||
		receipt.Summary.AllocationObservations != receipt.Summary.SamplesRequested {
		return fmt.Errorf("PROFILE_SUMMARY_INVALID")
	}
	for index, sample := range receipt.Samples {
		if sample.Sequence != index+1 || sample.Decision != "PASS" ||
			!strings.HasPrefix(sample.ExecutionDigest, "sha256:") ||
			sample.WallNanoseconds <= 0 || sample.TotalAllocBytes == 0 {
			return fmt.Errorf("PROFILE_SAMPLE_INVALID")
		}
	}
	return nil
}
