package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/valuejson"

// TextEscapePolicy keeps the existing entry point while both execution paths
// consume the same byte-preserved Gooo projection in the value transport layer.
func TextEscapePolicy(unit, next int64) string { return valuejson.TextEscapePolicy(unit, next) }
