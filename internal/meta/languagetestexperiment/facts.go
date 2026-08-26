package languagetestexperiment

import (
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagetest"
)

type facts struct {
	Receipts, DeclaredTests, ExecutedTests, PassedTests             int
	SourceCoherence, ReceiptDigestVariants, ExecutionDigestVariants int
	Go127Runtimes, AssertionRejections, MissingTestRejections       int
	NonClaims, Unknowns, RepositoryWrites                           int
	MutationAuthority                                               bool
}

func collectFacts(input Input) (facts, string) {
	var value facts
	receipts := []languagetest.Receipt{
		input.First.Receipt, input.Replay.Receipt, input.AssertionFailure, input.Missing,
	}
	for _, receipt := range receipts {
		if receipt.Decision != languagetest.DecisionPass && receipt.Decision != languagetest.DecisionFailClosed {
			value.Unknowns++
		}
	}
	if value.Unknowns > 0 {
		return value, "LANGUAGE_TEST_DECISION_UNKNOWN"
	}
	for _, receipt := range receipts {
		if err := languagetest.Validate(receipt); err != nil {
			return value, "LANGUAGE_TEST_RECEIPT_INVALID"
		}
		value.RepositoryWrites += receipt.Effects.RepositoryWrites
		value.MutationAuthority = value.MutationAuthority || receipt.Effects.MutationAuthority
	}
	if input.SubjectSHA == "" || !strings.HasPrefix(input.ExecutableDigest, "sha256:") {
		return value, "LANGUAGE_TEST_SUBJECT_UNBOUND"
	}
	positives := []Observation{input.First, input.Replay}
	receiptDigests, executionDigests := map[string]struct{}{}, map[string]struct{}{}
	for _, observation := range positives {
		if observation.Receipt.Decision != languagetest.DecisionPass {
			return value, "LANGUAGE_TEST_POSITIVE_RECEIPT_REJECTED"
		}
		value.Receipts++
		value.DeclaredTests += observation.Receipt.Summary.Declared
		value.ExecutedTests += observation.Receipt.Summary.Executed
		value.PassedTests += observation.Receipt.Summary.Passed
		receiptDigests[observation.Receipt.Digest] = struct{}{}
		for _, testCase := range observation.Receipt.Cases {
			executionDigests[testCase.ExecutionDigest] = struct{}{}
		}
		if strings.HasPrefix(observation.Runtime, "go1.27.") {
			value.Go127Runtimes++
		}
	}
	if input.First.Receipt.SourceDigest != "" && input.First.Receipt.SourceDigest == input.Replay.Receipt.SourceDigest {
		value.SourceCoherence = 2
	}
	value.ReceiptDigestVariants = len(receiptDigests)
	value.ExecutionDigestVariants = len(executionDigests)
	if input.AssertionFailure.Reason == "LANGUAGE_TEST_ASSERTION_FAILED" {
		value.AssertionRejections = 1
	}
	if input.Missing.Reason == "LANGUAGE_TESTS_MISSING" {
		value.MissingTestRejections = 1
	}
	if slices.Equal(input.First.Receipt.NonClaims, input.Replay.Receipt.NonClaims) {
		value.NonClaims = len(input.First.Receipt.NonClaims)
	}
	return value, ""
}
