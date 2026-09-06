package generation

func receiptPlanKnown(plan Plan) bool {
	if plan.SchemaVersion != SchemaVersion || !validSHA(plan.BaseSHA) ||
		!validSHA(plan.HeadSHA) || plan.PromotionAuthorized ||
		plan.ReplayProof != ProofCoherence ||
		plan.RequestedK != requestedK ||
		plan.MinimumIndependent != minimumIndependent ||
		!validDigest(plan.InputDigest) || !validDigest(plan.RegistryDigest) {
		return false
	}
	if validatePlanRefutedEvidence(plan) != nil {
		return false
	}
	unsigned := plan
	unsigned.PlanDigest, unsigned.ReplayDigest = "", ""
	if !validDigest(plan.PlanDigest) ||
		plan.PlanDigest != digestJSON(unsigned) ||
		plan.ReplayDigest != digestPair(plan.InputDigest, plan.PlanDigest) ||
		plan.RegistryDigest != digestJSON(plan.Registry) {
		return false
	}
	selected, valid := selectedActionIndex(plan)
	if !valid || !planApplicabilityKnown(plan, selected) {
		return false
	}
	switch plan.Decision {
	case DecisionFixedPoint:
		return plan.Reason == ReasonExactFixedPoint && len(plan.Selected) == 0 && len(plan.RefutedIndicatorIDs) == 0 && len(plan.Counterexamples) == 0
	case DecisionPlan:
		return plan.Reason == ReasonIndependentActions &&
			uint32(len(plan.Selected)) >= minimumIndependent &&
			uint32(len(plan.Selected)) <= requestedK
	case DecisionUnknown, DecisionRejected:
		return len(plan.Selected) == 0
	default:
		return false
	}
}

func selectedActionIndex(plan Plan) (map[string]Action, bool) {
	bindings, valid := registryIndex(plan.Registry)
	if !valid {
		return nil, false
	}
	result := make(map[string]Action, len(plan.Selected))
	for _, action := range plan.Selected {
		binding, exists := bindings[action.Operation]
		if !exists || !actionMatchesBinding(action, binding) ||
			!validActionIndicatorID(action.IndicatorID) ||
			action.MetricID == "" || action.Subject == "" ||
			!validActionApplicability(action) {
			return nil, false
		}
		if _, duplicate := result[action.IndicatorID]; duplicate {
			return nil, false
		}
		result[action.IndicatorID] = action
	}
	return result, true
}

func actionMatchesBinding(action Action, binding Binding) bool {
	return action.SubjectKind == binding.InputSubjectKind &&
		action.InputSubjectKind == binding.InputSubjectKind &&
		action.InputContractSourceDigest == binding.InputContractSourceDigest &&
		action.InputContractSemanticDigest == binding.InputContractSemanticDigest &&
		action.IndependenceGroupID == binding.IndependenceGroupID &&
		action.Activity == binding.Activity &&
		action.Output == binding.Output &&
		action.ProofChoice == binding.ProofChoice &&
		action.Executor == binding.Executor &&
		action.Evaluator == binding.Evaluator &&
		action.ReceiptRequired && binding.ReceiptRequired &&
		action.Priority == binding.Priority &&
		digestJSON(action.RequiredIndicatorIDs) ==
			digestJSON(binding.RequiredIndicatorIDs)
}
