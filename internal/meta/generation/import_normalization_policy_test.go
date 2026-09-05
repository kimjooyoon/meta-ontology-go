package generation

import (
	"strings"
	"testing"
)

func TestImportNormalizationPolicyContractRegressionCohort(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{name: "exact-policy-is-bound-to-ir", source: string(operationInputContractSource)},
		{name: "missing-or-mismatched-policy-fails-closed", source: ""},
	}
	if len(cases) != 2 {
		t.Fatalf("import normalization contract cohort denominator=%d, want 2", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.source != "" {
				policy, err := ImportNormalizationPolicyEvidence()
				operation, operationErr := ExtractFunctionInputContractEvidence()
				if err != nil || operationErr != nil || policy.Activity != "NormalizeEligibleImportGroup" || policy.InputEntity != "FileInput" || policy.OutputEntity != "ImportNormalizationPolicy" || !policy.UsedInputFact || !policy.GeneratedOutputFact || policy.HeaderCapability != ImportHeaderNamedAliasCapability || policy.SourceDigest != operation.SourceDigest || policy.SemanticDigest != operation.SemanticDigest {
					t.Fatalf("policy evidence=%+v operation=%+v policy_err=%v operation_err=%v", policy, operation, err, operationErr)
				}
				return
			}
			base := string(operationInputContractSource)
			missing := strings.Replace(base, "activity NormalizeEligibleImportGroup(FileInput) -> ImportNormalizationPolicy\n", "", 1)
			if _, err := parseOperationInputContract([]byte(missing)); err == nil {
				t.Fatal("missing import normalization policy was accepted")
			}
			mismatched := strings.Replace(base, "activity NormalizeEligibleImportGroup(FileInput) -> ImportNormalizationPolicy", "activity NormalizeEligibleImportGroup(FunctionInput) -> ImportNormalizationPolicy", 1)
			if _, err := parseOperationInputContract([]byte(mismatched)); err == nil {
				t.Fatal("mismatched import normalization policy input was accepted")
			}
		})
	}
}
