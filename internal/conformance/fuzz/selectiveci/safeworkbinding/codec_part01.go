package safeworkbinding

import (
	"unicode/utf8"
)

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
