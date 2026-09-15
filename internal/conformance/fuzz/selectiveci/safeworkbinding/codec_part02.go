package safeworkbinding

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
