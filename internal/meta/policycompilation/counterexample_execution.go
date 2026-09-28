package policycompilation

import (
	"bytes"
	"context"
	"errors"
)

// ObserveGoooPolicyCounterexample composes accepted derivation and execution.
// It never repairs evidence digests, changes caller expectations or adopts.
func ObserveGoooPolicyCounterexample(ctx context.Context, operationSource []byte, filename string, source []byte, expectedPackage, expectedNamespace string, rawInput []byte) (PolicyCounterexampleExecution, error) {
	var input PolicyCounterexampleExecutionInput
	trimmed := bytes.TrimSpace(rawInput)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return PolicyCounterexampleExecution{}, errors.New("counterexample execution requires a JSON object")
	}
	if err := decodeStrictJSON(rawInput, &input); err != nil {
		return PolicyCounterexampleExecution{}, err
	}
	if _, err := BindPolicyRevisionOperation(operationSource); err != nil {
		return PolicyCounterexampleExecution{}, err
	}
	proposal, err := ProposePolicyRevisionFromCounterexample(filename, source, expectedPackage, expectedNamespace, input.Counterexample)
	report := PolicyCounterexampleExecution{
		Schema: PolicyCounterexampleExecutionSchema, InputArtifactDigest: DigestBytes(rawInput), Proposal: proposal,
	}
	if err != nil || proposal.State != "PROPOSED" {
		return report, err
	}
	if len(input.Cases) == 0 {
		pending := revisionPending("EXECUTION_INPUT", "BIND_COUNTEREXAMPLE_CASE_PAIRS",
			"EXECUTION_CASE_PAIRS_MISSING", "SUPPLY_EXECUTION_CASE_PAIRS")
		report.Pending = &pending
		return report, errors.New(pending.Reason)
	}
	request, err := bindCounterexampleExecutionCases(input, proposal)
	if err != nil {
		return report, err
	}
	requestBytes, err := canonicalJSON(request)
	if err != nil {
		return report, err
	}
	report.RevisionRequestJSON = string(requestBytes)
	operation, err := ObserveGoooPolicyDecisionRevision(ctx, operationSource, filename,
		source, expectedPackage, expectedNamespace, requestBytes)
	if operation.Schema != "" {
		report.Operation = &operation
	}
	return report, err
}

func bindCounterexampleExecutionCases(input PolicyCounterexampleExecutionInput, proposal PolicyCounterexampleProposal) (PolicyRevisionObservationRequest, error) {
	if proposal.Revision == nil {
		return PolicyRevisionObservationRequest{}, errors.New("proposed counterexample omitted its revision")
	}
	revision := proposal.Revision
	request := PolicyRevisionObservationRequest{
		ExpectedSourceDigest: revision.ExpectedSourceDigest, Condition: revision.Condition,
		FromDecision: revision.FromDecision, ToDecision: revision.ToDecision,
		Cases: append([]PolicyRevisionCasePair(nil), input.Cases...),
	}
	if err := validateRevisionObservationRequest(request); err != nil {
		return PolicyRevisionObservationRequest{}, err
	}
	covered := false
	for _, pair := range request.Cases {
		if pair.Baseline.ValidatorExpectation != pair.Candidate.ValidatorExpectation {
			return PolicyRevisionObservationRequest{}, errors.New("candidate cannot change the caller expectation")
		}
		if pair.Baseline == input.Counterexample.Input {
			covered = true
		}
	}
	if !covered {
		return PolicyRevisionObservationRequest{}, errors.New("execution omitted the exact original counterexample case")
	}
	return request, nil
}
