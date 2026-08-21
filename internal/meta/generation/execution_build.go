package generation

func BuildExecutionManifest(plan Plan) ExecutionManifest {
	manifest := ExecutionManifest{
		SchemaVersion: ExecutionManifestSchemaVersion,
		BaseSHA:       plan.BaseSHA, HeadSHA: plan.HeadSHA,
		PlanDigest: plan.PlanDigest,
	}
	if !receiptPlanKnown(plan) {
		manifest.Decision = ExecutionDecisionUnknown
		manifest.Reason = ExecutionReasonInvalidPlan
		return finishExecutionManifest(manifest)
	}
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
		Operation:           action.Operation,
		IndependenceGroupID: action.IndependenceGroupID,
		ProofChoice:         action.ProofChoice,
		Executor:            action.Executor, Evaluator: action.Evaluator,
		RequiredIndicatorIDs: append([]string{}, action.RequiredIndicatorIDs...),
		ReceiptRequired:      action.ReceiptRequired, Priority: action.Priority,
		WorkspaceMode: WorkspaceModeDisposable,
		WriteBoundary: WriteBoundarySandboxOnly,
	}
}
