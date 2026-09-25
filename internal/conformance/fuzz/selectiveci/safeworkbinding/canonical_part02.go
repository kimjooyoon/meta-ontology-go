package safeworkbinding

func decisionSpelling(value Decision) ([]byte, bool) {
	switch value {
	case DecisionPass:
		return []byte("PASS"), true
	case DecisionUnknown:
		return []byte("UNKNOWN"), true
	case DecisionFailClosed:
		return []byte("FAIL_CLOSED"), true
	default:
		return nil, false
	}
}
func reasonSpelling(value Reason) ([]byte, bool) {
	spellings := [...]string{
		"NONE", "REQUIRED_INPUT_MISSING", "INVALID_UTF8", "BOM_FORBIDDEN",
		"INVALID_JSON", "TRAILING_VALUE", "DUPLICATE_KEY", "UNKNOWN_FIELD",
		"NULL_VALUE", "EMPTY_VALUE", "INVALID_SCHEMA", "INVALID_STABLE_ID",
		"INVALID_DIGEST", "BINDING_DIGEST_MISMATCH",
	}
	if int(value) >= len(spellings) {
		return nil, false
	}
	return []byte(spellings[value]), true
}
func enforcementEffectSpelling(value EnforcementEffect) ([]byte, bool) {
	if value != EnforcementEffectNoEffect {
		return nil, false
	}
	return []byte("NO_EFFECT"), true
}
func decisionField(name string, value Decision) (frameField, bool) {
	spelling, ok := decisionSpelling(value)
	if !ok {
		return frameField{}, false
	}
	return encodeEnumField(name, spelling), true
}
func reasonField(name string, value Reason) (frameField, bool) {
	spelling, ok := reasonSpelling(value)
	if !ok {
		return frameField{}, false
	}
	return encodeEnumField(name, spelling), true
}
func enforcementEffectField(name string, value EnforcementEffect) (frameField, bool) {
	spelling, ok := enforcementEffectSpelling(value)
	if !ok {
		return frameField{}, false
	}
	return encodeEnumField(name, spelling), true
}
func reasonListField(name string, values []Reason) (frameField, bool) {
	spellings := make([][]byte, len(values))
	for i, value := range values {
		spelling, ok := reasonSpelling(value)
		if !ok {
			return frameField{}, false
		}
		spellings[i] = spelling
	}
	return encodeListField(name, spellings), true
}
