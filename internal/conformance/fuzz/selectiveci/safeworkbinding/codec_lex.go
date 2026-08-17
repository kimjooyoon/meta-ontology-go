package safeworkbinding

import "strings"

type jsonValueKind uint8

const (
	jsonNullValue   jsonValueKind = 0
	jsonBoolValue   jsonValueKind = 1
	jsonNumberValue jsonValueKind = 2
	jsonStringValue jsonValueKind = 3
	jsonArrayValue  jsonValueKind = 4
	jsonObjectValue jsonValueKind = 5
)

type jsonValue struct {
	kind   jsonValueKind
	text   string
	object map[string]jsonValue
}

type jsonParser struct {
	data      []byte
	offset    int
	duplicate bool
}

func hexValue(value byte) (byte, bool) {
	switch {
	case value >= '0' && value <= '9':
		return value - '0', true
	case value >= 'a' && value <= 'f':
		return value - 'a' + 10, true
	case value >= 'A' && value <= 'F':
		return value - 'A' + 10, true
	default:
		return 0, false
	}
}

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

func (p *jsonParser) parseUnicodeEscape() (rune, bool) {
	if p.offset+4 > len(p.data) {
		return 0, false
	}
	var value rune
	for index := 0; index < 4; index++ {
		digit, ok := hexValue(p.data[p.offset+index])
		if !ok {
			return 0, false
		}
		value = value*16 + rune(digit)
	}
	p.offset += 4
	if value < 0xD800 || value > 0xDFFF {
		return value, true
	}
	if value >= 0xDC00 || p.offset+6 > len(p.data) {
		return 0, false
	}
	if p.data[p.offset] != '\\' || p.data[p.offset+1] != 'u' {
		return 0, false
	}
	p.offset += 2
	var low rune
	for index := 0; index < 4; index++ {
		digit, ok := hexValue(p.data[p.offset+index])
		if !ok {
			return 0, false
		}
		low = low*16 + rune(digit)
	}
	if low < 0xDC00 || low > 0xDFFF {
		return 0, false
	}
	p.offset += 4
	return 0x10000 + (value-0xD800)*0x400 + (low - 0xDC00), true
}

func (p *jsonParser) parseNumber() bool {
	if p.offset >= len(p.data) {
		return false
	}
	if p.data[p.offset] == '-' {
		p.offset++
		if p.offset >= len(p.data) {
			return false
		}
	}
	if p.data[p.offset] == '0' {
		p.offset++
		if p.offset < len(p.data) && p.data[p.offset] >= '0' && p.data[p.offset] <= '9' {
			return false
		}
	} else if p.data[p.offset] >= '1' && p.data[p.offset] <= '9' {
		for p.offset < len(p.data) && p.data[p.offset] >= '0' && p.data[p.offset] <= '9' {
			p.offset++
		}
	} else {
		return false
	}
	if p.offset < len(p.data) && p.data[p.offset] == '.' {
		p.offset++
		start := p.offset
		for p.offset < len(p.data) && p.data[p.offset] >= '0' && p.data[p.offset] <= '9' {
			p.offset++
		}
		if start == p.offset {
			return false
		}
	}
	if p.offset < len(p.data) && (p.data[p.offset] == 'e' || p.data[p.offset] == 'E') {
		p.offset++
		if p.offset < len(p.data) && (p.data[p.offset] == '+' || p.data[p.offset] == '-') {
			p.offset++
		}
		start := p.offset
		for p.offset < len(p.data) && p.data[p.offset] >= '0' && p.data[p.offset] <= '9' {
			p.offset++
		}
		if start == p.offset {
			return false
		}
	}
	return true
}

func (p *jsonParser) parseLiteral(literal string) bool {
	switch literal {
	case "null", "true", "false":
	default:
		return false
	}
	if p.offset+len(literal) > len(p.data) {
		return false
	}
	for index := range literal {
		if p.data[p.offset+index] != literal[index] {
			return false
		}
	}
	p.offset += len(literal)
	return true
}

func (p *jsonParser) skipSpace() {
	for p.offset < len(p.data) {
		switch p.data[p.offset] {
		case ' ', '\t', '\n', '\r':
			p.offset++
		default:
			return
		}
	}
}
