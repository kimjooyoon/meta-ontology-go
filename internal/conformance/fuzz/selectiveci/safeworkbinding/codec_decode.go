package safeworkbinding

func DecodeJSON(data []byte) (SafeWorkBinding, ParseResult) {
	root, reason := parseDocument(data)
	if reason != ReasonNone {
		return SafeWorkBinding{}, completeResult(resultForReason(reason))
	}
	binding, reason := validateEnvelope(root)
	if reason != ReasonNone {
		return SafeWorkBinding{}, completeResult(resultForReason(reason))
	}
	computed := bindingDigest(binding)
	result := completeResult(resultForPass())
	result.ReplayDigest = replayDigest(computed, result.ResultDigest)
	return binding, result
}

func resultForReason(reason Reason) ParseResult {
	if reason < ReasonRequiredInputMissing || reason > ReasonBindingDigestMismatch {
		return ParseResult{}
	}
	result := ParseResult{
		Decision:          DecisionFailClosed,
		Reason:            reason,
		Faults:            []Reason{reason},
		EnforcementEffect: EnforcementEffectNoEffect,
	}
	if reason == ReasonRequiredInputMissing {
		result.Decision = DecisionUnknown
		result.FullSuiteRequired = true
	}
	return result
}

func resultForPass() ParseResult {
	return ParseResult{
		Decision:          DecisionPass,
		Reason:            ReasonNone,
		Faults:            []Reason{},
		EnforcementEffect: EnforcementEffectNoEffect,
	}
}

func completeResult(result ParseResult) ParseResult {
	digest, _ := resultDigest(result)
	result.ResultDigest = digest
	result.ReplayDigest = ""
	return result
}
