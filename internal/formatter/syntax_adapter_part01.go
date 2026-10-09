package formatter

// SyntaxAdapter adapts the current surface syntax AST to the formatter's
// parser-neutral document. The formatter deliberately supports only the
// nominal entity/activity declarations, including explicit activity IDs.
type SyntaxAdapter struct{}
