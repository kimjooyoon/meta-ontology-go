package safeworkbinding

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
