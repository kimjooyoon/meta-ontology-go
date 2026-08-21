package safeworkbinding

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
