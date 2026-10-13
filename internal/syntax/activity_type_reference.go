package syntax

import (
	"fmt"
	"strings"
)

// An activity signature uses a local entity name or one import alias and member.
// Resolution belongs to the package layer; parsing retains the complete spelling.
func (p *Parser) parseActivityTypeReference(subject string, code DiagnosticCode) NameRef {
	first := p.expectIdentifier(subject, code)
	if !p.at(TokenDot) {
		return first
	}
	p.advance()
	member := p.expectIdentifier(subject+" after import alias", code)
	return NameRef{Span: startSpan(p.filename, first.Span.Start, member.Span.End), Name: first.Name + "." + member.Name}
}

func validateActivityTypeReference(name, subject string) error {
	parts := strings.Split(name, ".")
	if len(parts) > 2 {
		return fmt.Errorf("%s must be an entity name or import alias and entity name", subject)
	}
	for _, part := range parts {
		if err := validateIdentifier(part, subject); err != nil {
			return err
		}
	}
	return nil
}
