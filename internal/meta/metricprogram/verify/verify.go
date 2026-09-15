package verify

import (
	"fmt"
	"reflect"
)

func Verify(strategyPayload, strategyVerificationPayload, programPayload, source []byte) (Report, error) {
	var strategy strategyPlan
	if err := decodeExact(strategyPayload, &strategy); err != nil {
		return Report{}, fmt.Errorf("decode strategy: %w", err)
	}
	var strategyVerification strategyVerification
	if err := decodeExact(strategyVerificationPayload, &strategyVerification); err != nil {
		return Report{}, fmt.Errorf("decode strategy verification: %w", err)
	}
	var actual program
	if err := decodeExact(programPayload, &actual); err != nil {
		return Report{}, fmt.Errorf("decode program: %w", err)
	}
	expected, err := reconstruct(strategy, strategyVerification, source)
	if err != nil {
		return Report{}, err
	}
	if actual.Schema != programSchema || actual.RepositoryWorkspaceWrites || actual.PromotionAuthorized {
		return Report{}, fmt.Errorf("program boundary is invalid")
	}
	if !reflect.DeepEqual(actual, expected) {
		return Report{}, fmt.Errorf("program does not match independent reconstruction")
	}
	report := Report{
		Schema: reportSchema, SubjectSHA: actual.SubjectSHA, StrategyDigest: actual.StrategyDigest,
		ProgramDigest: actual.Digest, RegistryDigest: actual.RegistryDigest, SourceDigest: actual.SourceDigest,
		SemanticDigest: actual.SemanticDigest, BindingCount: len(actual.Bindings), OperationCount: len(actual.Operations),
		StepCount: len(actual.Steps), Status: "VERIFIED", RepositoryWorkspaceWrites: false, PromotionAuthorized: false,
	}
	report.Digest, err = valueDigest(report)
	if err != nil {
		return Report{}, err
	}
	return report, nil
}
