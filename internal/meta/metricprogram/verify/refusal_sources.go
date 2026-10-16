package verify

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
)

func verifyRefusalSources(repository fs.FS, refusal unresolvedpolicy.OperationRefusal) error {
	if err := metricprogram.VerifyCompiledRegistry(); err != nil {
		return fmt.Errorf("refusal registry projection unavailable: %w", err)
	}
	if err := metricprogram.VerifyRegistrySource(repository, metricprogram.RegistrySourceOrigin()); err != nil {
		return fmt.Errorf("refusal registry source: %w", err)
	}
	if refusal.Policy != unresolvedpolicy.Origin() {
		return fmt.Errorf("refusal policy origin differs from compiled declaration")
	}
	return unresolvedpolicy.VerifySource(repository)
}

func verifyRefusalInputs(strategy strategyPlan, verification strategyVerification, repository, subjectSHA string) error {
	if repository == "" || len(subjectSHA) != 40 ||
		strategy.Repository != repository || strategy.SubjectSHA != subjectSHA ||
		strategy.Schema != strategySchema || verification.Schema != strategyVerificationSchema {
		return fmt.Errorf("refusal input identity is not exact")
	}
	for _, value := range subjectSHA {
		if !(value >= '0' && value <= '9' || value >= 'a' && value <= 'f') {
			return fmt.Errorf("refusal subject is not a commit SHA")
		}
	}
	if strategy.ExecutionPolicy != "READ_ONLY_ARTIFACT_SYNTHESIS" || strategy.RepositoryWorkspaceWrites ||
		strategy.PromotionAuthorized || verification.RepositoryWorkspaceWrites || verification.PromotionAuthorized {
		return fmt.Errorf("refusal input execution boundary is invalid")
	}
	if verification.PlanDigest != strategy.Digest || !validDigest(strategy.Digest) || !validDigest(verification.Digest) ||
		verification.Status != "VERIFIED" || verification.BindingCount != len(strategy.Bindings) ||
		verification.CandidateCount != len(strategy.Candidates) || strategy.Input.IndicatorCount != len(strategy.Bindings) ||
		verification.SourceMetricsDigest != strategy.Input.SourceMetricsDigest ||
		verification.InterventionDigest != strategy.Input.InterventionDigest {
		return fmt.Errorf("refusal input digest or cardinality binding is invalid")
	}
	if !slices.IsSortedFunc(strategy.Bindings, func(a, b strategyBinding) int {
		return strings.Compare(a.IndicatorID, b.IndicatorID)
	}) {
		return fmt.Errorf("refusal input bindings are not canonical")
	}
	return nil
}

// This compatibility kernel pins the diagnostic meaning independently of the
// producer's generated functions. Source membership is checked separately.
func expectedRefusal(origin unresolvedpolicy.OperationOrigin) unresolvedpolicy.OperationRefusal {
	registry, inputs := origin.Registry, origin.Inputs
	recorded := origin.Repository != "" && origin.SubjectSHA != "" && origin.IndicatorID != "" &&
		origin.RequestedOperation != "" && registry.SourcePath != "" && registry.SourceDigest != "" &&
		registry.ProjectionPath != "" && registry.ProjectionDigest != "" &&
		inputs.StrategyPayloadDigest != "" && inputs.VerificationPayloadDigest != ""
	reason, next := "OPERATION_ORIGIN_REQUIRED", "CAPTURE_OPERATION_ORIGIN"
	if recorded {
		reason, next = "OPERATION_NOT_IN_SOURCE_REGISTRY", "DECLARE_SOURCE_OPERATION"
	}
	return unresolvedpolicy.OperationRefusal{
		Schema: "gooo/unresolved-source-operation/v1", Origin: origin, Policy: unresolvedpolicy.Origin(),
		OriginRecorded: recorded, Resolution: "invariant_only", Reason: reason, NextOperation: next,
	}
}
