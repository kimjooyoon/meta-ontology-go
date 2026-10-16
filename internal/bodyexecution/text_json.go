package bodyexecution

import (
	"encoding/json"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
	"github.com/kimjooyoon/meta-ontology-go/internal/valuejson"
)

// Composition text is checked against the original JSON before accepting its
// decoded value. Syntax, byte limits and duplicate-field checks stay in decode.
func decodeCompositionTextData(raw []byte, output any, limit int) error {
	if len(raw) == 0 || len(raw) > limit || !utf8.Valid(raw) || !json.Valid(raw) {
		return decode(raw, output, limit)
	}
	if err := validateJSONTextEscapes(raw); err != nil {
		return err
	}
	return decode(raw, output, limit)
}

func decodeTextScalar(raw []byte) (string, error) {
	return decodeBoundedTextScalar(raw, 0)
}

func decodeBoundedTextScalar(raw []byte, bound textbound.UTF8) (string, error) {
	maximum, err := bound.Resolve(compositionTextMaxBytes)
	if err != nil {
		return "", err
	}
	return valuejson.DecodeText(raw, maximum)
}

func validateJSONTextEscapes(raw []byte) error {
	return valuejson.CheckEscapesInValidJSON(raw)
}
