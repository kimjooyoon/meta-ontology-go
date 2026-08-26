package safeworkbinding

import (
	"strings"
)

func (p *jsonParser) parseString() (string, bool) {
	if p.offset >= len(p.data) || p.data[p.offset] != '"' {
		return "", false
	}
	p.offset++
	var decoded strings.Builder
	for p.offset < len(p.data) {
		current := p.data[p.offset]
		p.offset++
		switch current {
		case '"':
			return decoded.String(), true
		case '\\':
			if p.offset >= len(p.data) {
				return "", false
			}
			escaped := p.data[p.offset]
			p.offset++
			switch escaped {
			case '"', '\\', '/':
				decoded.WriteByte(escaped)
			case 'b':
				decoded.WriteByte('\b')
			case 'f':
				decoded.WriteByte('\f')
			case 'n':
				decoded.WriteByte('\n')
			case 'r':
				decoded.WriteByte('\r')
			case 't':
				decoded.WriteByte('\t')
			case 'u':
				scalar, ok := p.parseUnicodeEscape()
				if !ok {
					return "", false
				}
				decoded.WriteRune(scalar)
			default:
				return "", false
			}
		default:
			if current < 0x20 {
				return "", false
			}
			decoded.WriteByte(current)
		}
	}
	return "", false
}
