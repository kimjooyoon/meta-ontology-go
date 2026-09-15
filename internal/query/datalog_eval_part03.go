package query

import (
	"strings"
)

func datalogTermValue(term DatalogTerm, binding datalogBinding) ID {
	if term.Variable == "" {
		return term.Constant
	}
	return binding[strings.TrimPrefix(term.Variable, "?")]
}
func datalogSupport(support []DatalogFact) []DatalogFactKey {
	keys := make([]DatalogFactKey, len(support))
	for index, fact := range support {
		keys[index] = fact.Key()
	}
	return keys
}
