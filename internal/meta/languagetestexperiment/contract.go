package languagetestexperiment

import "fmt"

const ContractSchema = "gooo/language-test-experiment-contract/v1"

type Contract struct {
	Schema                            string `json:"schema"`
	ID                                string `json:"id"`
	ExpectedReceipts                  int    `json:"expected_receipts"`
	ExpectedDeclaredTests             int    `json:"expected_declared_tests"`
	ExpectedExecutedTests             int    `json:"expected_executed_tests"`
	ExpectedPassedTests               int    `json:"expected_passed_tests"`
	ExpectedSourceCoherence           int    `json:"expected_source_coherence"`
	ExpectedReceiptDigestVariants     int    `json:"expected_receipt_digest_variants"`
	ExpectedExecutionDigestVariants   int    `json:"expected_execution_digest_variants"`
	ExpectedGo127Runtimes             int    `json:"expected_go127_runtimes"`
	ExpectedAssertionRejections       int    `json:"expected_assertion_rejections"`
	ExpectedMissingTestRejections     int    `json:"expected_missing_test_rejections"`
	ExpectedNonClaims                 int    `json:"expected_non_claims"`
	ExpectedRepositoryWrites          int    `json:"expected_repository_writes"`
	ExpectedMutationAuthority         bool   `json:"expected_mutation_authority"`
}

func (contract Contract) Validate() error {
	if contract.Schema != ContractSchema || contract.ID == "" {
		return fmt.Errorf("LANGUAGE_TEST_CONTRACT_IDENTITY_INVALID")
	}
	values := []int{
		contract.ExpectedReceipts, contract.ExpectedDeclaredTests, contract.ExpectedExecutedTests,
		contract.ExpectedPassedTests, contract.ExpectedSourceCoherence,
		contract.ExpectedReceiptDigestVariants, contract.ExpectedExecutionDigestVariants,
		contract.ExpectedGo127Runtimes, contract.ExpectedAssertionRejections,
		contract.ExpectedMissingTestRejections, contract.ExpectedNonClaims,
	}
	for _, value := range values {
		if value <= 0 {
			return fmt.Errorf("LANGUAGE_TEST_CONTRACT_COUNTER_INVALID")
		}
	}
	if contract.ExpectedRepositoryWrites != 0 || contract.ExpectedMutationAuthority {
		return fmt.Errorf("LANGUAGE_TEST_CONTRACT_EFFECT_INVALID")
	}
	return nil
}
