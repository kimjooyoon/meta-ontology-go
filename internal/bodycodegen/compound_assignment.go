package bodycodegen

import "go/token"

// CompoundAssignmentOperator is the source profile shared by generation,
// deterministic evaluation, recipe construction and native origin observation.
func CompoundAssignmentOperator(operation token.Token) (token.Token, bool) {
	switch operation {
	case token.ADD_ASSIGN:
		return token.ADD, true
	case token.SUB_ASSIGN:
		return token.SUB, true
	case token.MUL_ASSIGN:
		return token.MUL, true
	case token.QUO_ASSIGN:
		return token.QUO, true
	}
	return token.ILLEGAL, false
}
