package verify

import (
	"fmt"
	"io/fs"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
)

// RefusalReport authenticates a diagnostic against supplied bytes and repository
// contents. It does not establish upstream execution or readiness conformance.
type RefusalReport struct {
	Schema                    string                            `json:"schema"`
	Status                    string                            `json:"status"`
	Origin                    unresolvedpolicy.OperationOrigin  `json:"origin"`
	Policy                    unresolvedpolicy.ProjectionOrigin `json:"policy"`
	RefusalPayloadDigest      string                            `json:"refusal_payload_digest"`
	UpstreamVerification      string                            `json:"upstream_verification"`
	ReadinessCredit           bool                              `json:"readiness_credit"`
	RepositoryWorkspaceWrites bool                              `json:"repository_workspace_writes"`
	PromotionAuthorized       bool                              `json:"promotion_authorized"`
}

func VerifyOperationRefusal(repositoryFS fs.FS, repository, subjectSHA string,
	strategyPayload, verificationPayload, refusalPayload []byte) (RefusalReport, error) {
	var strategy strategyPlan
	if err := decodeExact(strategyPayload, &strategy); err != nil {
		return RefusalReport{}, fmt.Errorf("refusal strategy: %w", err)
	}
	var verification strategyVerification
	if err := decodeExact(verificationPayload, &verification); err != nil {
		return RefusalReport{}, fmt.Errorf("refusal verification input: %w", err)
	}
	var refusal unresolvedpolicy.OperationRefusal
	if err := decodeExact(refusalPayload, &refusal); err != nil {
		return RefusalReport{}, fmt.Errorf("refusal document: %w", err)
	}
	if err := verifyRefusalInputs(strategy, verification, repository, subjectSHA); err != nil {
		return RefusalReport{}, err
	}
	if err := verifyRefusalSources(repositoryFS, refusal); err != nil {
		return RefusalReport{}, err
	}
	origin, err := reconstructRefusalOrigin(strategy, strategyPayload, verificationPayload)
	if err != nil {
		return RefusalReport{}, err
	}
	if refusal.Origin != origin || !reflect.DeepEqual(refusal, expectedRefusal(origin)) {
		return RefusalReport{}, fmt.Errorf("refusal does not match independent source-origin reconstruction")
	}
	return RefusalReport{
		Schema: "gooo/source-operation-refusal-verification/v1", Status: "VERIFIED_SOURCE_ORIGIN_REFUSAL",
		Origin: origin, Policy: refusal.Policy, RefusalPayloadDigest: bytesDigest(refusalPayload),
		UpstreamVerification: "UNASSESSED",
	}, nil
}

func reconstructRefusalOrigin(strategy strategyPlan, strategyPayload, verificationPayload []byte) (
	unresolvedpolicy.OperationOrigin, error) {
	declared := make(map[string]bool)
	for _, operation := range metricprogram.CanonicalOperations() {
		declared[operation.ID] = true
	}
	for _, binding := range strategy.Bindings {
		absent, err := refusalSourceAbsent(binding.MetaOperation, declared)
		if err != nil {
			return unresolvedpolicy.OperationOrigin{}, err
		}
		if !absent {
			continue
		}
		registry := metricprogram.RegistrySourceOrigin()
		return unresolvedpolicy.OperationOrigin{
			Repository: strategy.Repository, SubjectSHA: strategy.SubjectSHA,
			IndicatorID: binding.IndicatorID, RequestedOperation: binding.MetaOperation,
			Registry: unresolvedpolicy.ProjectionOrigin{
				SourcePath: registry.SourcePath, SourceDigest: registry.SourceDigest,
				ProjectionPath: registry.ProjectionPath, ProjectionDigest: registry.ProjectionDigest,
			},
			Inputs: unresolvedpolicy.InputDigests{StrategyPayloadDigest: bytesDigest(strategyPayload),
				VerificationPayloadDigest: bytesDigest(verificationPayload)},
		}, nil
	}
	return unresolvedpolicy.OperationOrigin{}, fmt.Errorf("refusal has no operation absent from the source registry")
}

func refusalSourceAbsent(operation string, declared map[string]bool) (bool, error) {
	if !declared[operation] {
		return true, nil
	}
	if _, supported := findOperation(operation); !supported {
		return false, fmt.Errorf("COMPATIBILITY_KERNEL_UNSUPPORTED: %q", operation)
	}
	return false, nil
}
