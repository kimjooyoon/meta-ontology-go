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
				header, headerErr := ImportHeaderNormalizationPolicyEvidence()
				operation, operationErr := ExtractFunctionInputContractEvidence()
				if err != nil || headerErr != nil || operationErr != nil || policy.Activity != "NormalizeEligibleImportGroup" || header.Activity != "NormalizeNamedAliasImportHeader" || policy.InputEntity != "FileInput" || header.InputEntity != "FileInput" || policy.OutputEntity != "ImportNormalizationPolicy" || header.OutputEntity != "ImportNormalizationPolicy" || !policy.UsedInputFact || !policy.GeneratedOutputFact || !header.UsedInputFact || !header.GeneratedOutputFact || policy.HeaderCapability != ImportHeaderPlainCapability || header.HeaderCapability != ImportHeaderNamedAliasCapability || policy.SourceDigest != operation.SourceDigest || policy.SemanticDigest != operation.SemanticDigest || header.SourceDigest != operation.SourceDigest || header.SemanticDigest != operation.SemanticDigest {
					t.Fatalf("policy evidence=%+v header=%+v operation=%+v policy_err=%v header_err=%v operation_err=%v", policy, header, operation, err, headerErr, operationErr)
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
			missingHeader := strings.Replace(base, "activity NormalizeNamedAliasImportHeader(FileInput) -> ImportNormalizationPolicy\n", "", 1)
			if _, err := parseOperationInputContract([]byte(missingHeader)); err == nil {
				t.Fatal("missing named-alias header policy was accepted")
			}
			mismatchedHeader := strings.Replace(base, "activity NormalizeNamedAliasImportHeader(FileInput) -> ImportNormalizationPolicy", "activity NormalizeNamedAliasImportHeader(FunctionInput) -> ImportNormalizationPolicy", 1)
			if _, err := parseOperationInputContract([]byte(mismatchedHeader)); err == nil {
				t.Fatal("mismatched named-alias header policy input was accepted")
			}
		})
	}
}
