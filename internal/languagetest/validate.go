package languagetest

import (
	"fmt"
	"slices"
	"strings"
)

func Validate(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema {
		return fmt.Errorf("LANGUAGE_TEST_SCHEMA_UNKNOWN")
	}
	if receipt.Decision != DecisionPass && receipt.Decision != DecisionFailClosed {
		return fmt.Errorf("LANGUAGE_TEST_DECISION_UNKNOWN")
	}
	if receipt.Resolution != ResolutionExact {
		return fmt.Errorf("LANGUAGE_TEST_RESOLUTION_UNKNOWN")
	}
	if receipt.Effects.RepositoryWrites != 0 || receipt.Effects.MutationAuthority {
		return fmt.Errorf("LANGUAGE_TEST_EFFECT_BOUNDARY_VIOLATED")
	}
	if !strings.HasPrefix(receipt.SourceDigest, "sha256:") || receipt.Digest != receiptDigest(receipt) {
		return fmt.Errorf("LANGUAGE_TEST_IDENTITY_INVALID")
	}
	if !slices.Equal(receipt.NonClaims, nonClaims()) {
		return fmt.Errorf("LANGUAGE_TEST_NON_CLAIMS_INVALID")
	}
	passed, failed, executed := 0, 0, 0
	for _, testCase := range receipt.Cases {
		if testCase.ExecutionDigest != "" {
			executed++
		}
		switch testCase.Decision {
		case DecisionPass:
			passed++
			if testCase.Reason != "LANGUAGE_TEST_ASSERTION_PASSED" || testCase.Expected != testCase.Observed {
				return fmt.Errorf("LANGUAGE_TEST_PASS_CASE_INVALID")
			}
		case DecisionFailClosed:
			failed++
		default:
			return fmt.Errorf("LANGUAGE_TEST_CASE_DECISION_UNKNOWN")
		}
	}
	if receipt.Summary != (Summary{Declared: len(receipt.Cases), Executed: executed, Passed: passed, Failed: failed}) {
		return fmt.Errorf("LANGUAGE_TEST_SUMMARY_INVALID")
	}
	if receipt.Decision == DecisionPass {
		if receipt.Reason != "LANGUAGE_TESTS_PASSED" || passed == 0 || failed != 0 || len(receipt.Diagnostics) != 0 {
			return fmt.Errorf("LANGUAGE_TEST_SUCCESS_INVALID")
		}
		return nil
	}
	if receipt.Reason == "LANGUAGE_TESTS_MISSING" {
		if len(receipt.Cases) != 0 || len(receipt.Diagnostics) != 1 {
			return fmt.Errorf("LANGUAGE_TEST_MISSING_INVALID")
		}
		return nil
	}
	if len(receipt.Diagnostics) == 0 || receipt.Diagnostics[0].Code != receipt.Reason {
		return fmt.Errorf("LANGUAGE_TEST_FAILURE_INVALID")
	}
	return nil
}
