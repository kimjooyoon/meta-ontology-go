package safeworkbinding

import "unicode/utf8"

func parseDocument(data []byte) (jsonValue, Reason) {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return jsonValue{}, ReasonBOMForbidden
	}
	if !utf8.Valid(data) {
		return jsonValue{}, ReasonInvalidUTF8
	}
	parser := jsonParser{data: data}
	first, ok := parser.parseValue()
	if !ok {
		return jsonValue{}, ReasonInvalidJSON
	}
	parser.skipSpace()
	trailing := false
	for parser.offset < len(parser.data) {
		trailing = true
		if _, ok := parser.parseValue(); !ok {
			return jsonValue{}, ReasonInvalidJSON
		}
		parser.skipSpace()
	}
	if parser.duplicate {
		return jsonValue{}, ReasonDuplicateKey
	}
	if trailing {
		return jsonValue{}, ReasonTrailingValue
	}
	if first.kind != jsonObjectValue {
		return jsonValue{}, ReasonInvalidSchema
	}
	return first, ReasonNone
}

func (p *jsonParser) parseValue() (jsonValue, bool) {
	p.skipSpace()
	if p.offset >= len(p.data) {
		return jsonValue{}, false
	}
	switch p.data[p.offset] {
	case '"':
		text, ok := p.parseString()
		if !ok {
			return jsonValue{}, false
		}
		return jsonValue{kind: jsonStringValue, text: text}, true
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case 'n':
		if !p.parseLiteral("null") {
			return jsonValue{}, false
		}
		return jsonValue{kind: jsonNullValue}, true
	case 't':
		if !p.parseLiteral("true") {
			return jsonValue{}, false
		}
		return jsonValue{kind: jsonBoolValue}, true
	case 'f':
		if !p.parseLiteral("false") {
			return jsonValue{}, false
		}
		return jsonValue{kind: jsonBoolValue}, true
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if !p.parseNumber() {
			return jsonValue{}, false
		}
		return jsonValue{kind: jsonNumberValue}, true
	default:
		return jsonValue{}, false
	}
}

func (p *jsonParser) parseObject() (jsonValue, bool) {
	p.offset++
	value := jsonValue{kind: jsonObjectValue, object: make(map[string]jsonValue)}
	seen := make(map[string]struct{})
	p.skipSpace()
	if p.offset < len(p.data) && p.data[p.offset] == '}' {
		p.offset++
		return value, true
	}
	for {
		p.skipSpace()
		key, ok := p.parseString()
		if !ok {
			return jsonValue{}, false
		}
		p.skipSpace()
		if p.offset >= len(p.data) || p.data[p.offset] != ':' {
			return jsonValue{}, false
		}
		p.offset++
		member, ok := p.parseValue()
		if !ok {
			return jsonValue{}, false
		}
		if _, exists := seen[key]; exists {
			p.duplicate = true
		} else {
			seen[key] = struct{}{}
			value.object[key] = member
		}
		p.skipSpace()
		if p.offset >= len(p.data) {
			return jsonValue{}, false
		}
		switch p.data[p.offset] {
		case ',':
			p.offset++
		case '}':
			p.offset++
			return value, true
		default:
			return jsonValue{}, false
		}
	}
}

func (p *jsonParser) parseArray() (jsonValue, bool) {
	p.offset++
	value := jsonValue{kind: jsonArrayValue}
	p.skipSpace()
	if p.offset < len(p.data) && p.data[p.offset] == ']' {
		p.offset++
		return value, true
	}
	for {
		if _, ok := p.parseValue(); !ok {
			return jsonValue{}, false
		}
		p.skipSpace()
		if p.offset >= len(p.data) {
			return jsonValue{}, false
		}
		switch p.data[p.offset] {
		case ',':
			p.offset++
		case ']':
			p.offset++
			return value, true
		default:
			return jsonValue{}, false
		}
	}
}
