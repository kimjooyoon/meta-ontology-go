package safeworkbinding

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
