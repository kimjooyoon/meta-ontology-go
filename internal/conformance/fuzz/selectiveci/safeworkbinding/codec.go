package safeworkbinding

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"
)

type jsonValueKind uint8

const (
	jsonNullValue jsonValueKind = iota
	jsonBoolValue
	jsonNumberValue
	jsonStringValue
	jsonArrayValue
	jsonObjectValue
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

func parseDocument(data []byte) (jsonValue, Reason) {
	if len(data) >= 3 && data[0] == 0xef && data[1] == 0xbb && data[2] == 0xbf {
		return jsonValue{}, ReasonBOMForbidden
	}
	if !utf8.Valid(data) {
		return jsonValue{}, ReasonInvalidUTF8
	}
	p := jsonParser{data: data}
	root, ok := p.parseValue()
	if !ok {
		return jsonValue{}, ReasonInvalidJSON
	}
	trailing := false
	for {
		p.skipSpace()
		if p.offset == len(p.data) {
			break
		}
		trailing = true
		if _, ok = p.parseValue(); !ok {
			return jsonValue{}, ReasonInvalidJSON
		}
	}
	if p.duplicate {
		return jsonValue{}, ReasonDuplicateKey
	}
	if trailing {
		return jsonValue{}, ReasonTrailingValue
	}
	if root.kind != jsonObjectValue {
		return jsonValue{}, ReasonInvalidSchema
	}
	return root, ReasonNone
}
func (p *jsonParser) parseValue() (jsonValue, bool) {
	p.skipSpace()
	if p.offset == len(p.data) {
		return jsonValue{}, false
	}
	switch p.data[p.offset] {
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case '"':
		text, ok := p.parseString()
		return jsonValue{kind: jsonStringValue, text: text}, ok
	case 'n':
		return jsonValue{kind: jsonNullValue}, p.parseLiteral("null")
	case 't':
		return jsonValue{kind: jsonBoolValue, text: "true"}, p.parseLiteral("true")
	case 'f':
		return jsonValue{kind: jsonBoolValue, text: "false"}, p.parseLiteral("false")
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if p.parseNumber() {
			return jsonValue{kind: jsonNumberValue}, true
		}
	}
	return jsonValue{}, false
}
func (p *jsonParser) parseObject() (jsonValue, bool) {
	decoder := json.NewDecoder(bytes.NewReader(p.data[p.offset:]))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return jsonValue{}, false
	}
	object := make(map[string]jsonValue)
	for decoder.More() {
		keyToken, err := decoder.Token()
		key, ok := keyToken.(string)
		if err != nil || !ok {
			return jsonValue{}, false
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return jsonValue{}, false
		}
		child := jsonParser{data: raw}
		value, ok := child.parseValue()
		if !ok {
			return jsonValue{}, false
		}
		if child.duplicate {
			p.duplicate = true
		}
		if _, exists := object[key]; exists {
			p.duplicate = true
		}
		object[key] = value
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') {
		return jsonValue{}, false
	}
	p.offset += int(decoder.InputOffset())
	return jsonValue{kind: jsonObjectValue, object: object}, true
}
func (p *jsonParser) parseArray() (jsonValue, bool) {
	decoder := json.NewDecoder(bytes.NewReader(p.data[p.offset:]))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('[') {
		return jsonValue{}, false
	}
	for decoder.More() {
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return jsonValue{}, false
		}
		child := jsonParser{data: raw}
		if _, ok := child.parseValue(); !ok {
			return jsonValue{}, false
		}
		if child.duplicate {
			p.duplicate = true
		}
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim(']') {
		return jsonValue{}, false
	}
	p.offset += int(decoder.InputOffset())
	return jsonValue{kind: jsonArrayValue}, true
}
func (p *jsonParser) parseString() (string, bool) {
	p.skipSpace()
	if p.offset == len(p.data) || p.data[p.offset] != '"' {
		return "", false
	}
	decoder := json.NewDecoder(bytes.NewReader(p.data[p.offset:]))
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return "", false
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return "", false
	}
	wantReplacement := bytes.Count(raw, []byte{0xef, 0xbf, 0xbd})
	for i := 0; i+5 < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		if raw[i+1] == '\\' {
			i++
			continue
		}
		if raw[i+1] == 'u' {
			value, _ := strconv.ParseUint(string(raw[i+2:i+6]), 16, 16)
			if value == 0xfffd {
				wantReplacement++
			}
			i += 5
		}
	}
	if strings.Count(text, string(utf8.RuneError)) != wantReplacement {
		return "", false
	}
	p.offset += int(decoder.InputOffset())
	return text, true
}
func (p *jsonParser) parseUnicodeEscape() (rune, bool) {
	if p.offset+4 > len(p.data) {
		return 0, false
	}
	value, err := strconv.ParseUint(string(p.data[p.offset:p.offset+4]), 16, 16)
	if err != nil {
		return 0, false
	}
	p.offset += 4
	return rune(value), true
}
func (p *jsonParser) parseNumber() bool {
	p.skipSpace()
	decoder := json.NewDecoder(bytes.NewReader(p.data[p.offset:]))
	decoder.UseNumber()
	var number json.Number
	if err := decoder.Decode(&number); err != nil {
		return false
	}
	end := p.offset + int(decoder.InputOffset())
	if end < len(p.data) && p.data[end] >= '0' && p.data[end] <= '9' &&
		(number[0] == '0' || number[0] == '-') {
		return false
	}
	p.offset = end
	return true
}
func (p *jsonParser) parseLiteral(literal string) bool {
	if !bytes.HasPrefix(p.data[p.offset:], []byte(literal)) {
		return false
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
func hexValue(value byte) (byte, bool) {
	index := bytes.IndexByte([]byte("0123456789abcdefABCDEF"), value)
	if index < 0 {
		return 0, false
	}
	if index >= 16 {
		return byte(index - 6), true
	}
	return byte(index), true
}
