package bodycodegen

import "fmt"

// SourceLiteralByteRange maps decoded source edits to the original Gooo literal.
// It shares the escape/Unicode mapping used by source cursor references.
func SourceLiteralByteRange(raw, value string, start, end int) (int, int, error) {
	if start < 0 || end < start || end > len(value) {
		return 0, 0, fmt.Errorf("decoded source edit is outside the literal")
	}
	first, err := sourceLiteralPosition(raw, value, start, false)
	if err != nil {
		return 0, 0, err
	}
	last, err := sourceLiteralPosition(raw, value, end, false)
	return first, last, err
}
