package lsp

import "strings"

type sourceParser struct {
	uri    string
	source string
	tokens []parserToken
	index  int
	result ParseResult
}

func (p *sourceParser) parse() ParseResult {
	p.parseHeader(tokenPackage, "package", symbolFile)
	p.parseHeader(tokenNamespace, "namespace", symbolNamespace)
	for !p.at(tokenEOF) {
		switch p.current().kind {
		case tokenEntity:
			p.parseEntity()
		case tokenActivity:
			p.parseActivity()
		default:
			p.error("parse.unexpected-token", "expected entity or activity declaration")
			p.advance()
		}
	}
	return p.result
}

func (p *sourceParser) parseHeader(kind tokenKind, label string, symbol SymbolKind) {
	start := p.current().start
	if !p.accept(kind) {
		p.error("parse.expected-"+label, "expected "+label+" declaration")
		return
	}
	name := p.expect(tokenIdentifier, "parse.expected-identifier", "expected "+label+" name")
	end := name.end
	if end == 0 {
		end = p.tokens[p.index-1].end
	}
	p.result.Symbols = append(p.result.Symbols, Symbol{Name: name.value, Kind: symbol, Detail: label, Range: sourceRange(p.source, start, end), SelectionRange: sourceRange(p.source, name.start, name.end)})
}

func (p *sourceParser) parseEntity() {
	start := p.advance().start
	name := p.expect(tokenIdentifier, "parse.expected-identifier", "expected entity name")
	p.expect(tokenID, "parse.expected-id", "expected id")
	id := p.expect(tokenString, "parse.expected-string", "expected quoted semantic identifier")
	end := id.end
	if end == 0 {
		end = name.end
	}
	p.result.Symbols = append(p.result.Symbols, Symbol{Name: name.value, ID: id.value, Kind: symbolClass, Detail: id.value, Range: sourceRange(p.source, start, end), SelectionRange: sourceRange(p.source, name.start, name.end)})
}

func (p *sourceParser) parseActivity() {
	start := p.advance().start
	name := p.expect(tokenIdentifier, "parse.expected-identifier", "expected activity name")
	p.expect(tokenLParen, "parse.expected-left-paren", "expected (")
	var inputs []string
	for !p.at(tokenRParen) && !p.at(tokenEOF) {
		input := p.expect(tokenIdentifier, "parse.expected-identifier", "expected activity parameter")
		if input.value != "" {
			inputs = append(inputs, input.value)
			p.result.References = append(p.result.References, Reference{Name: input.value, Range: sourceRange(p.source, input.start, input.end)})
		}
		if !p.accept(tokenComma) {
			break
		}
	}
	p.expect(tokenRParen, "parse.expected-right-paren", "expected )")
	p.expect(tokenArrow, "parse.expected-arrow", "expected ->")
	result := p.expect(tokenIdentifier, "parse.expected-result", "expected activity result")
	if result.value != "" {
		p.result.References = append(p.result.References, Reference{Name: result.value, Range: sourceRange(p.source, result.start, result.end)})
	}
	detail := "activity " + name.value + "(" + strings.Join(inputs, ", ") + ") -> " + result.value
	end := result.end
	if end == 0 {
		end = name.end
	}
	p.result.Symbols = append(p.result.Symbols, Symbol{Name: name.value, Kind: symbolFunction, Detail: detail, Range: sourceRange(p.source, start, end), SelectionRange: sourceRange(p.source, name.start, name.end)})
}

func (p *sourceParser) expect(kind tokenKind, code, message string) parserToken {
	if p.at(kind) {
		return p.advance()
	}
	p.error(code, message)
	return parserToken{}
}

func (p *sourceParser) accept(kind tokenKind) bool {
	if p.at(kind) {
		p.advance()
		return true
	}
	return false
}

func (p *sourceParser) at(kind tokenKind) bool { return p.current().kind == kind }

func (p *sourceParser) current() parserToken { return p.tokens[p.index] }

func (p *sourceParser) advance() parserToken {
	token := p.current()
	if p.index < len(p.tokens)-1 {
		p.index++
	}
	return token
}

func (p *sourceParser) error(code, message string) {
	token := p.current()
	p.result.Diagnostics = append(p.result.Diagnostics, Diagnostic{Range: sourceRange(p.source, token.start, token.end), Severity: DiagnosticError, Code: code, Source: "gooo", Message: message})
}

func sourceRange(source string, start, end int) Range {
	return Range{Start: offsetPosition(source, start), End: offsetPosition(source, end)}
}
