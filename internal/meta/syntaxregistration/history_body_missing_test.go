package syntaxregistration

import "testing"

func TestHistoryDeclarationWithoutBodyProducesNoEdits(t *testing.T) {
	source := parseHistoryCapacityFixture(t, []byte("package verticalsliceclosureshadow\n"+
		"func expectedBoundarySpecsForVersion(version int) []boundarySpec\n"))
	if err := pinPreviousSyntaxTarget(source, 201, 1013); err == nil || len(source.edits) != 0 {
		t.Fatal("body-free declaration produced edits", err)
	}
}
