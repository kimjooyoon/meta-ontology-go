package syntax

import "github.com/kimjooyoon/meta-ontology-go/internal/textbound"

const DiagExpectedFieldByteBound DiagnosticCode = "parse.expected-field-byte-bound"

func (p *Parser) parseFieldTextBound(end Position) (textbound.UTF8, Position, bool) {
	if !p.atEntityFieldsWord("max_utf8_bytes") {
		return 0, end, true
	}
	p.advance()
	value := p.expectString()
	if value.Span.IsEmpty() {
		return 0, end, false
	}
	bound, err := textbound.Parse(value.Name)
	if err != nil {
		p.error(DiagExpectedFieldByteBound, value.Span,
			"max_utf8_bytes requires a canonical decimal string in 0..4294967295")
		return 0, end, false
	}
	return bound, value.Span.End, true
}
