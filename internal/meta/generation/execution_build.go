package generation

func BuildExecutionManifest(plan Plan) ExecutionManifest {
	manifest := ExecutionManifest{
		SchemaVersion: ExecutionManifestSchemaVersion,
		BaseSHA:       plan.BaseSHA, HeadSHA: plan.HeadSHA,
		PlanDigest:                plan.PlanDigest,
		NotApplicableIndicatorIDs: append([]string{}, plan.NotApplicableIndicatorIDs...),
	}
	if !receiptPlanKnown(plan) || validatePlanIndicatorDecisionLedger(plan) != nil {
		manifest.Decision = ExecutionDecisionUnknown
		manifest.Reason = ExecutionReasonInvalidPlan
		return finishExecutionManifest(manifest)
	}
	manifest.IndicatorDecisionLedgerDigest, manifest.IndicatorDecisionLedgerCount = planIndicatorDecisionLedgerProvenance(plan)
	switch plan.Decision {
	case DecisionFixedPoint:
		manifest.Decision = ExecutionDecisionFixedPoint
		manifest.Reason = ExecutionReasonExactFixedPoint
	case DecisionPlan:
		for _, action := range sortedSelectedActions(plan.Selected) {
			manifest.Steps = append(manifest.Steps, executionStepFor(action))
		}
		manifest.Decision = ExecutionDecisionProposed
		manifest.Reason = ExecutionReasonIndependentActions
	case DecisionRejected:
		manifest.Decision = ExecutionDecisionRejected
		manifest.Reason = ExecutionReasonPlanRejected
	default:
		manifest.Decision = ExecutionDecisionUnknown
		manifest.Reason = ExecutionReasonPlanNotExecutable
	}
	return finishExecutionManifest(manifest)
}

func executionStepFor(action Action) ExecutionStep {
	return ExecutionStep{
		ActionIndicatorID: action.IndicatorID,
		MetricID:          action.MetricID, Subject: action.Subject,
		SubjectKind: action.SubjectKind, InputSubjectKind: action.InputSubjectKind,
		InputContractSourceDigest:   action.InputContractSourceDigest,
		InputContractSemanticDigest: action.InputContractSemanticDigest,
		Applicability:               action.Applicability,
		ApplicabilityRule:           action.ApplicabilityRule, ApplicabilityReason: action.ApplicabilityReason,
		Blocking: action.Blocking, SourceIndicator: action.SourceIndicator,
		IndicatorOutcome:  action.IndicatorOutcome,
		MetricProofChoice: action.MetricProofChoice, MetricProducer: action.MetricProducer,
		MetricConsumer: action.MetricConsumer,
		Operation:      action.Operation,
		Activity:       action.Activity, Output: action.Output,
		IndependenceGroupID: action.IndependenceGroupID,
		ProofChoice:         action.ProofChoice,
		Executor:            action.Executor, Evaluator: action.Evaluator,
		RequiredIndicatorIDs: append([]string{}, action.RequiredIndicatorIDs...),
		ReceiptRequired:      action.ReceiptRequired, Priority: action.Priority,
		RegistrationRequest: cloneRegistrationRequest(action.RegistrationRequest),
		WorkspaceMode: WorkspaceModeDisposable,
		WriteBoundary: WriteBoundarySandboxOnly,
	}
}
