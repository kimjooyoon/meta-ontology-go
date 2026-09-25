package query

import (
	"fmt"
	"strconv"
	"strings"
)

func normalizeDatalogTerm(term DatalogTerm) (DatalogTerm, error) {
	if term.Variable != "" && term.Constant != "" {
		return DatalogTerm{}, datalogError("term cannot be both variable and constant")
	}
	if term.Variable != "" {
		name := strings.TrimPrefix(strings.TrimSpace(term.Variable), "?")
		if !validDatalogIdentifier(name) {
			return DatalogTerm{}, datalogError("variable %q is invalid", term.Variable)
		}
		return DatalogTerm{Variable: "?" + name}, nil
	}
	if term.Constant == "" {
		return DatalogTerm{}, datalogError("term is empty")
	}
	id, err := ParseID(term.Constant.String())
	if err != nil {
		return DatalogTerm{}, datalogError("constant: %v", err)
	}
	return DatalogTerm{Constant: id}, nil
}
func normalizeDatalogPredicate(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if relation, err := ParseRelation(Relation(raw)); err == nil {
		return string(relation), nil
	}
	if !validDatalogIdentifier(raw) {
		return "", datalogError("predicate %q is invalid", raw)
	}
	return raw, nil
}
func validDatalogIdentifier(value string) bool {
	if value == "" || !((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z') || value[0] == '_') {
		return false
	}
	for _, character := range value[1:] {
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_') {
			return false
		}
	}
	return true
}
func validDatalogRuleID(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if !(character == '/' || character == ':' || character == '.' || character == '-' ||
			(character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_') {
			return false
		}
	}
	return true
}
func datalogError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidDatalogQuery, fmt.Sprintf(format, args...))
}
func datalogAtomCanonical(atom DatalogAtom) string {
	return strconv.Quote(atom.Predicate) + "\x00" +
		datalogTermCanonical(atom.Subject) + "\x00" + datalogTermCanonical(atom.Object)
}
