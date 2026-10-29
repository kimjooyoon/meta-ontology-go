package extractor

import "testing"

func TestRecordOriginReaderRenderedCapacity(t *testing.T) {
	requireNamedCapacities(t, "../../../..", []string{
		"internal/bodycodegen#NewFixedProgram",
		"internal/bodycodegen#FixedProgram.Evaluate",
		"internal/bodycodegen#EvaluateFixedValue",
		"examples/record-origin-reader#main",
		"examples/record-origin-reader#run",
		"examples/record-origin-reader#query",
		"examples/record-origin-reader#cursorProgress",
		"examples/record-origin-reader#tuple",
		"examples/record-origin-reader#project",
		"examples/record-origin-reader#projectNodes",
		"examples/record-origin-reader#projectTables",
		"examples/record-origin-reader#sha",
	})
}
