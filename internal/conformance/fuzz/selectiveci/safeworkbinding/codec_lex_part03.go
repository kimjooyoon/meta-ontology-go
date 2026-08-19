package safeworkbinding

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
