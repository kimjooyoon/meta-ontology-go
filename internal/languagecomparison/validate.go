package languagecomparison

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
)

func Validate(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema || receipt.ContractID != ContractID || receipt.Scope != comparisonScope {
		return fmt.Errorf("LANGUAGE_COMPARISON_IDENTITY_INVALID")
	}
	if receipt.Decision != "PASS" && receipt.Decision != "FAIL_CLOSED" {
		return fmt.Errorf("LANGUAGE_COMPARISON_DECISION_INVALID")
	}
	if receipt.Decision == "PASS" && receipt.Resolution != "RUNNER_SCOPED" {
		return fmt.Errorf("LANGUAGE_COMPARISON_RESOLUTION_INVALID")
	}
	if receipt.Decision == "FAIL_CLOSED" && receipt.Resolution != "EXACT" {
		return fmt.Errorf("LANGUAGE_COMPARISON_RESOLUTION_INVALID")
	}
	if !validSubject(receipt.SubjectSHA) || !validDigest(receipt.ExecutableDigest) ||
		!validDigest(receipt.GoooSourceDigest) || !validDigest(receipt.GoSourceDigest) ||
		receipt.Runner.GoVersion == "" || receipt.Runner.OS == "" || receipt.Runner.Arch == "" ||
		receipt.Runner.CPUs < 1 || strings.TrimSpace(receipt.Runner.Label) == "" ||
		!reflect.DeepEqual(receipt.NotClaimed, defaultNonClaims()) {
		return fmt.Errorf("LANGUAGE_COMPARISON_CONTEXT_INVALID")
	}
	if receipt.Effects.RepositoryWrites != 0 || receipt.Effects.MutationAuthority ||
		receipt.Digest != receiptDigest(receipt) {
		return fmt.Errorf("LANGUAGE_COMPARISON_EFFECT_OR_DIGEST_INVALID")
	}
	if receipt.Decision == "FAIL_CLOSED" {
		if receipt.Reason == "" {
			return fmt.Errorf("LANGUAGE_COMPARISON_FAILURE_REASON_UNKNOWN")
		}
		return nil
	}
	return validateSuccess(receipt)
}

func validateSuccess(receipt Receipt) error {
	if receipt.Reason != "EQUIVALENT_DECLARATION_SIGNATURE_OBSERVED" || receipt.Entry == "" ||
		len(receipt.Samples) < 1 || len(receipt.Samples) > MaximumSamples ||
		!reflect.DeepEqual(receipt.Summary, summarize(receipt.Summary.SamplesRequested, receipt.Samples)) {
		return fmt.Errorf("LANGUAGE_COMPARISON_SUMMARY_INVALID")
	}
	if receipt.Summary.SamplesRequested != SamplesPerLanguage || receipt.Summary.SamplesRequested != len(receipt.Samples) ||
		receipt.Summary.SamplesObserved != len(receipt.Samples) ||
		receipt.Summary.EquivalentOutputSamples != len(receipt.Samples) ||
		receipt.Summary.GoooOutputDigestVariants != 1 || receipt.Summary.GoOutputDigestVariants != 1 {
		return fmt.Errorf("LANGUAGE_COMPARISON_EQUIVALENCE_INVALID")
	}
	for index, sample := range receipt.Samples {
		if sample.Sequence != index+1 || sample.GoooDecision != "PASS" || sample.GoDecision != "PASS" ||
			!validDigest(sample.GoooOutputDigest) || sample.GoooOutputDigest != sample.GoOutputDigest ||
			sample.Gooo.WallNanoseconds <= 0 || sample.Go.WallNanoseconds <= 0 ||
			sample.Gooo.TotalAllocBytes == 0 || sample.Go.TotalAllocBytes == 0 ||
			(sample.FirstMeasured != "gooo" && sample.FirstMeasured != "go") {
			return fmt.Errorf("LANGUAGE_COMPARISON_SAMPLE_INVALID")
		}
	}
	return nil
}

func receiptDigest(receipt Receipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
}

func validSubject(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}
