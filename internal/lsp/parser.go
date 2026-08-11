package lsp

import (
	"context"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Parser is the small integration seam between the LSP and a .gooo parser.
type Parser interface {
	Parse(uri, source string) ParseResult
}

// DocumentParser is a descriptive alias for Parser.
type DocumentParser = Parser

// ParserFunc adapts a function to Parser.
type ParserFunc func(uri, source string) ParseResult

func (f ParserFunc) Parse(uri, source string) ParseResult { return f(uri, source) }

// ContextParser can stop parsing when the request's context is canceled.
type ContextParser interface {
	ParseContext(ctx context.Context, uri, source string) (ParseResult, error)
}

// ContextParserFunc adapts a cancellable parser function.
type ContextParserFunc func(ctx context.Context, uri, source string) (ParseResult, error)

func (f ContextParserFunc) ParseContext(ctx context.Context, uri, source string) (ParseResult, error) {
	return f(ctx, uri, source)
}

func (f ContextParserFunc) Parse(uri, source string) ParseResult {
	result, _ := f(context.Background(), uri, source)
	return result
}

type ParseResult struct {
	Symbols     []Symbol
	References  []Reference
	Diagnostics []Diagnostic
}

type Symbol struct {
	Name           string
	ID             string
	Kind           SymbolKind
	Detail         string
	Range          Range
	SelectionRange Range
}

type Reference struct {
	Name  string
	Range Range
}

// SyntaxParser is a compact standard-library parser for the .gooo editor view.
type SyntaxParser struct{}

func (SyntaxParser) Parse(uri, source string) ParseResult {
	parser := sourceParser{uri: uri, source: source}
	parser.tokens, parser.result.Diagnostics = lexSource(source)
	return parser.parse()
}

type tokenKind uint8

const (
	tokenEOF tokenKind = iota
	tokenIdentifier
	tokenString
	tokenPackage
	tokenNamespace
	tokenEntity
	tokenID
	tokenActivity
	tokenLParen
	tokenRParen
	tokenComma
	tokenArrow
)

type parserToken struct {
	kind       tokenKind
	text       string
	value      string
	start, end int
}

type sourceLexer struct {
	source      string
	offset      int
	tokens      []parserToken
	diagnostics []Diagnostic
}

func lexSource(source string) ([]parserToken, []Diagnostic) {
	lexer := &sourceLexer{source: source}
	for lexer.offset < len(source) {
		if lexer.skipSpaceOrComment() {
			continue
		}
		lexer.lexToken()
	}
	lexer.tokens = append(lexer.tokens, parserToken{kind: tokenEOF, start: len(source), end: len(source)})
	return lexer.tokens, lexer.diagnostics
}

func (l *sourceLexer) skipSpaceOrComment() bool {
	if strings.ContainsRune(" \t\r\n", rune(l.source[l.offset])) {
		l.offset++
		return true
	}
	if strings.HasPrefix(l.source[l.offset:], "//") {
		l.offset += 2
		for l.offset < len(l.source) && l.source[l.offset] != '\n' {
			l.offset++
		}
		return true
	}
	return false
}

func (l *sourceLexer) lexToken() {
	start := l.offset
	runeValue, size := utf8.DecodeRuneInString(l.source[start:])
	switch {
	case isIdentifierStart(runeValue):
		l.lexIdentifier(start)
	case runeValue == '"':
		l.lexString(start)
	case runeValue == '(':
		l.emit(tokenLParen, start, start+size, "(")
	case runeValue == ')':
		l.emit(tokenRParen, start, start+size, ")")
	case runeValue == ',':
		l.emit(tokenComma, start, start+size, ",")
	case runeValue == '-' && strings.HasPrefix(l.source[start:], "->"):
		l.emit(tokenArrow, start, start+2, "->")
	default:
		l.offset += size
		l.diagnostics = append(l.diagnostics, Diagnostic{Range: sourceRange(l.source, start, l.offset), Severity: DiagnosticError, Code: "lex.unexpected-character", Source: "gooo", Message: "unexpected character"})
	}
}

func (l *sourceLexer) lexIdentifier(start int) {
	_, firstSize := utf8.DecodeRuneInString(l.source[start:])
	l.offset += firstSize
	for l.offset < len(l.source) {
		runeValue, size := utf8.DecodeRuneInString(l.source[l.offset:])
		if !isIdentifierPart(runeValue) {
			break
		}
		l.offset += size
	}
	text := l.source[start:l.offset]
	l.emit(identifierKind(text), start, l.offset, text)
}

func (l *sourceLexer) lexString(start int) {
	l.offset++
	closed := false
	for l.offset < len(l.source) {
		if l.source[l.offset] == '\\' {
			l.offset += 1
			if l.offset < len(l.source) {
				_, size := utf8.DecodeRuneInString(l.source[l.offset:])
				l.offset += size
			}
			continue
		}
		if l.source[l.offset] == '"' {
			l.offset++
			closed = true
			break
		}
		_, size := utf8.DecodeRuneInString(l.source[l.offset:])
		l.offset += size
	}
	if !closed {
		l.diagnostics = append(l.diagnostics, Diagnostic{Range: sourceRange(l.source, start, l.offset), Severity: DiagnosticError, Code: "lex.unterminated-string", Source: "gooo", Message: "unterminated string"})
	}
	value := l.source[start:l.offset]
	if decoded, err := strconv.Unquote(value); err == nil {
		value = decoded
	}
	l.emit(tokenString, start, l.offset, value)
}

func (l *sourceLexer) emit(kind tokenKind, start, end int, value string) {
	l.offset = end
	l.tokens = append(l.tokens, parserToken{kind: kind, text: l.source[start:end], value: value, start: start, end: end})
}

func identifierKind(text string) tokenKind {
	switch text {
	case "package":
		return tokenPackage
	case "namespace":
		return tokenNamespace
	case "entity":
		return tokenEntity
	case "id":
		return tokenID
	case "activity":
		return tokenActivity
	default:
		return tokenIdentifier
	}
}

func isIdentifierStart(value rune) bool {
	return value == '_' || unicode.IsLetter(value)
}

func isIdentifierPart(value rune) bool {
	return isIdentifierStart(value) || unicode.IsDigit(value)
}
