package bodycodegen

import "go/token"

// Nil comparisons ask whether a value is present. Two present pointers retain
// their existing unsupported evaluator boundary; no heap identity is invented.
func recordBodyPresenceComparison(operator token.Token, left, right any) (bool, bool) {
	a, first := left.(recordBodyOptionalScalar)
	b, second := right.(recordBodyOptionalScalar)
	if !first || !second || a.Present && b.Present {
		return false, false
	}
	switch operator {
	case token.EQL:
		return a.Present == b.Present, true
	case token.NEQ:
		return a.Present != b.Present, true
	}
	return false, false
}
