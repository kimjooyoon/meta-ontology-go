package syntax

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func (p *Parser) parseActivityIdentity(activity *ActivityDecl) {
	if !p.at(TokenID) {
		return
	}
	p.advance()
	activity.IDPresent = true
	id := p.expect(TokenString, "quoted activity identity", DiagExpectedString)
	activity.ID, activity.IDSpan = id.Value, id.Span
	if !id.Span.IsEmpty() {
		activity.Span.End = id.Span.End
	}
}

func formatActivityIdentity(output *strings.Builder, activity *ActivityDecl) error {
	if !activity.IDPresent && activity.ID == "" {
		return nil
	}
	if !utf8.ValidString(activity.ID) {
		return fmt.Errorf("activity id is not valid UTF-8")
	}
	fmt.Fprintf(output, " id %s", quoteString(activity.ID))
	return nil
}
