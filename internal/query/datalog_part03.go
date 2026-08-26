package query

import (
	"strings"
)

func (row DatalogRow) Value(name string) (ID, bool) {
	name = strings.TrimPrefix(strings.TrimSpace(name), "?")
	value, ok := row.Bindings[name]
	return value, ok
}

// DatalogResult is deterministic by construction. Complete is false when the
// result limit trimmed rows; no partial result is ever reported as complete.
type DatalogResult struct {
	Rows     []DatalogRow
	Derived  []DatalogFact
	Complete bool
}
