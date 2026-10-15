package bodycodegen

import (
	"go/token"
	"strings"
)

// Gooo owns the layout permission; this adapter supplies lexical observations.
func elseSurfaceEdit(body string, tokens []bodySurfaceToken, index int) (bodySurfaceEdit, bool) {
	if index < 2 {
		return bodySurfaceEdit{}, false
	}
	current, previous, before := tokens[index], tokens[index-1], tokens[index-2]
	if before.end < 0 || before.end > current.start || current.start > len(body) {
		return bodySurfaceEdit{}, false
	}
	gap := body[before.end:current.start]
	if !JoinElse(current.kind == token.ELSE, previous.kind == token.SEMICOLON && previous.text == "\n",
		before.kind == token.RBRACE, strings.TrimSpace(gap) == "") {
		return bodySurfaceEdit{}, false
	}
	return bodySurfaceEdit{start: before.end, end: current.start, text: strings.ReplaceAll(gap, "\n", " ")}, true
}
