package policycompilation

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

//go:embed revision-operation.gooo
var policyRevisionOperationContract []byte

func PolicyRevisionOperationContract() []byte {
	return append([]byte(nil), policyRevisionOperationContract...)
}

// BindPolicyRevisionOperation admits the pinned native ABI, not arbitrary
// computes programs. Raw source identity stays distinct from semantic identity.
func BindPolicyRevisionOperation(source []byte) (PolicyRevisionOperationBinding, error) {
	binding, err := readPolicyRevisionOperation(source)
	if err != nil {
		return PolicyRevisionOperationBinding{}, err
	}
	pinned, err := readPolicyRevisionOperation(policyRevisionOperationContract)
	if err != nil {
		return PolicyRevisionOperationBinding{}, fmt.Errorf("embedded revision operation ABI: %w", err)
	}
	if binding.ContractSemanticDigest != pinned.ContractSemanticDigest {
		return PolicyRevisionOperationBinding{}, fmt.Errorf("policy revision operation semantic ABI mismatch")
	}
	binding.NativeContractSourceDigest = pinned.ContractSourceDigest
	return binding, nil
}

func readPolicyRevisionOperation(source []byte) (PolicyRevisionOperationBinding, error) {
	file, diagnostics := syntax.ParseFile("revision-operation.gooo", string(source))
	if file == nil || diagnostics.HasErrors() {
		return PolicyRevisionOperationBinding{}, fmt.Errorf("policy revision operation cannot be parsed")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return PolicyRevisionOperationBinding{}, fmt.Errorf("lower policy revision operation: %w", err)
	}
	if ir.Namespace != "policyrevisionoperation" || len(ir.Graph.Nodes()) != 4 {
		return PolicyRevisionOperationBinding{}, fmt.Errorf("policy revision operation node inventory mismatch")
	}
	policy, policyFound := ir.Graph.NodeByName(ir.Namespace, "PolicySource")
	request, requestFound := ir.Graph.NodeByName(ir.Namespace, "RevisionRequest")
	observation, observationFound := ir.Graph.NodeByName(ir.Namespace, "RevisionObservation")
	activity, activityFound := ir.Graph.NodeByName(ir.Namespace, "ObservePolicyDecisionRevision")
	if !policyFound || !requestFound || !observationFound || !activityFound ||
		policy.Kind != semantic.Entity || request.Kind != semantic.Entity ||
		observation.Kind != semantic.Entity || activity.Kind != semantic.Activity ||
		policy.ID.String() != "gooo://policy-revision-operation/policy-source" ||
		request.ID.String() != "gooo://policy-revision-operation/request" ||
		observation.ID.String() != "gooo://policy-revision-operation/observation" ||
		activity.ValueProgram != PolicyRevisionOperationProgram {
		return PolicyRevisionOperationBinding{}, fmt.Errorf("policy revision operation native ABI mismatch")
	}
	usedPolicy := ir.Graph.HasFact(semantic.FactKey{Subject: activity.ID, Predicate: semantic.Used, Object: policy.ID})
	usedRequest := ir.Graph.HasFact(semantic.FactKey{Subject: activity.ID, Predicate: semantic.Used, Object: request.ID})
	generated := ir.Graph.HasFact(semantic.FactKey{Subject: observation.ID, Predicate: semantic.WasGeneratedBy, Object: activity.ID})
	if !usedPolicy || !usedRequest || !generated {
		return PolicyRevisionOperationBinding{}, fmt.Errorf("policy revision operation relation mismatch")
	}
	return PolicyRevisionOperationBinding{
		ContractSourceDigest: DigestBytes(source), ContractSemanticDigest: ir.StableHash(),
		ActivityID: activity.ID.String(), Program: activity.ValueProgram,
		PolicySourceEntityID: policy.ID.String(), RequestEntityID: request.ID.String(),
		ObservationEntityID: observation.ID.String(), UsedPolicySource: usedPolicy,
		UsedRevisionRequest: usedRequest, GeneratedObservation: generated,
	}, nil
}

// ObserveGoooPolicyDecisionRevision is an explicit single-request route.
// It neither changes legacy inventory selection nor grants adoption authority.
func ObserveGoooPolicyDecisionRevision(ctx context.Context, operationSource []byte, filename string, source []byte, expectedPackage, expectedNamespace string, requestBytes []byte) (PolicyRevisionOperationObservation, error) {
	binding, err := BindPolicyRevisionOperation(operationSource)
	if err != nil {
		return PolicyRevisionOperationObservation{}, err
	}
	request, err := DecodePolicyRevisionObservationRequest(requestBytes)
	if err != nil {
		return PolicyRevisionOperationObservation{}, fmt.Errorf("revision operation request: %w", err)
	}
	if DigestBytes(source) != request.ExpectedSourceDigest {
		return PolicyRevisionOperationObservation{}, fmt.Errorf("revision operation source digest mismatch")
	}
	result := PolicyRevisionOperationObservation{
		Schema: PolicyRevisionOperationSchema, Binding: binding,
		PolicySourceDigest: DigestBytes(source), RequestArtifactDigest: DigestBytes(requestBytes),
	}
	result.NativeWorkerInvocations++
	observation, observationError := ObservePolicyDecisionRevision(ctx, filename, source, expectedPackage, expectedNamespace, request)
	if observation.Schema != "" {
		observation.RequestArtifactDigest = result.RequestArtifactDigest
		result.Observation = &observation
	}
	return result, observationError
}
