package languagediagnosticprovenance

import (
	"go/parser"
	"go/scanner"
	"go/token"
	"strings"
)

func syntaxObservation(fixture string) (Observation, bool) {
	source, code := syntaxFixture(fixture)
	if source == "" {
		return Observation{}, false
	}
	files := token.NewFileSet()
	_, parseError := parser.ParseFile(files, "generated.go", source, parser.AllErrors)
	if parseError == nil {
		return Observation{}, false
	}
	var sourceFile *token.File
	files.Iterate(func(file *token.File) bool {
		sourceFile = file
		return false
	})
	offset := strings.Index(source, "@")
	if sourceFile == nil || offset < 0 {
		return Observation{}, false
	}
	position := sourceFile.Pos(offset)
	physical := tokenPosition(files.PositionFor(position, false))
	logical := tokenPosition(files.PositionFor(position, true))
	message, ordered := orderedScannerMessage(parseError)
	return Observation{
		Origin: "GO", Stage: "PARSE", Code: code, Message: message,
		Hardness: "NOT_APPLICABLE", Physical: oneByteSpan(physical),
		Logical: oneByteSpan(logical),
	}, ordered
}

func syntaxFixture(fixture string) (string, string) {
	switch fixture {
	case "physical-error":
		return "package p\nvar _ = @\n", "GO_PARSE_PHYSICAL"
	case "line-directive":
		return "//line model.gooo:40\npackage p\nvar _ = @\n", "GO_PARSE_LINE_DIRECTIVE"
	case "multiple-errors":
		return "package p\nvar _ = @\nvar _ = @\n", "GO_PARSE_ORDERED"
	default:
		return "", ""
	}
}

func tokenPosition(position token.Position) Position {
	return Position{
		Filename: position.Filename, Offset: position.Offset,
		Line: position.Line, Column: position.Column,
	}
}

func orderedScannerMessage(parseError error) (string, bool) {
	errors, ok := parseError.(scanner.ErrorList)
	if !ok || len(errors) == 0 {
		return parseError.Error(), false
	}
	errors.Sort()
	messages := make([]string, 0, len(errors))
	for _, item := range errors {
		messages = append(messages, item.Msg)
	}
	return strings.Join(messages, " | "), true
}
