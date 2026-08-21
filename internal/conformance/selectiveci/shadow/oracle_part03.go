package shadow

// Evaluate applies the contract's fixed precedence. It performs no process,
// filesystem, network, or argv execution side effect.
func Evaluate(c Case) Result {
	digest := caseDigest(c)
	inputs, err := decodeFiles(c.Files)
	if err != nil {
		return fallback(StageInput, err.Error(), digest)
	}
	if err := validateSnapshots(inputs.base, inputs.head, inputs.planner); err != nil {
		return fallback(StageSnapshot, err.Error(), digest)
	}
	if err := validateRegistry(inputs); err != nil {
		return fallback(StageRegistry, err.Error(), digest)
	}
	if err := validatePlan(inputs.base, inputs.head, inputs.planner); err != nil {
		return fallback(StagePlan, err.Error(), digest)
	}
	if err := validatePlanProofBinding(inputs); err != nil {
		return fallback(StagePlanProof, err.Error(), digest)
	}
	if inputs.proof.Status != "VERIFIED" || inputs.proof.Fallback != "NONE" || inputs.proof.ProofDigest != proofDigest(inputs.proof) {
		if inputs.proof.Status == "UNKNOWN" {
			return fallback(StageProofUnknown, "proof is UNKNOWN", digest)
		}
		return fallback(StageProofFail, "proof is not verified", digest)
	}
	if !validLaneFacts(inputs.lane) || inputs.lane.Schema != LaneSchema || inputs.lane.CanonicalDigest != laneDigest(inputs.lane) || inputs.lane.Decision == "UNKNOWN" {
		return fallback(StageLaneUnknown, "lane is UNKNOWN or stale", digest)
	}
	if inputs.lane.Decision == "INELIGIBLE" {
		return fallback(StageLaneIneligible, "lane is INELIGIBLE", digest)
	}
	if inputs.lane.Decision != "ELIGIBLE" || inputs.lane.Reason != "ELIGIBLE" {
		return fallback(StageLaneUnknown, "lane decision is malformed", digest)
	}

	selected, guards, work, argv := normalizedSelection(inputs.planner)
	return Result{
		Status: ShadowSelective, Stage: StageSelective, Reason: "all bindings verified",
		SelectedCommandIDs: selected, SelectedGuardIDs: guards, SelectedWorkIDs: work,
		SelectedArgv: argv, ExecutionAuthorized: false, CanonicalDigest: digest,
	}
}
func fallback(stage, reason, digest string) Result {
	return Result{Status: FullSuiteFallback, Stage: stage, Reason: reason,
		SelectedCommandIDs: []string{}, SelectedGuardIDs: []string{}, SelectedWorkIDs: []string{},
		SelectedArgv: map[string][]string{}, ExecutionAuthorized: false, CanonicalDigest: digest}
}
