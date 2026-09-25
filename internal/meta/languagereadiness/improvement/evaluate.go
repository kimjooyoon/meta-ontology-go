package improvement

// Evaluate proves improvement only from comparable integer snapshots.
func Evaluate(before, after Snapshot) Transition {
	result := newTransition(before, after)
	beforeInspection := inspect(before)
	if beforeInspection.reason != "" {
		return reject(result, "BEFORE_"+beforeInspection.reason)
	}
	afterInspection := inspect(after)
	if afterInspection.reason != "" {
		return reject(result, "AFTER_"+afterInspection.reason)
	}
	if reason := comparisonReason(before, after, beforeInspection, afterInspection); reason != "" {
		return reject(result, reason)
	}
	quantify(&result, before, after, beforeInspection, afterInspection)
	classify(&result)
	return seal(result)
}

func newTransition(before, after Snapshot) Transition {
	return Transition{
		Schema:            TransitionSchema,
		Decision:          LowerResolution,
		ReasonCode:        "SNAPSHOT_NOT_COMPARABLE",
		ContractSchema:    before.ContractSchema,
		RegistryDigest:    before.RegistryDigest,
		BeforeCompleted:   before.Completed,
		AfterCompleted:    after.Completed,
		Total:             before.Total,
		BeforeBasisPoints: before.BasisPoints,
		AfterBasisPoints:  after.BasisPoints,
		Proofs:            proofs(false, false, false, false),
	}
}

func reject(result Transition, reason string) Transition {
	result.ReasonCode = reason
	return seal(result)
}
