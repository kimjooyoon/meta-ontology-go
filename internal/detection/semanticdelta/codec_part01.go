package semanticdelta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type jsonRequest Request

// EncodeJSON returns canonical, indented JSON terminated by one newline.
func EncodeJSON(request Request) ([]byte, error) {
	normalized, err := request.Normalized()
	if err != nil {
		return nil, err
	}
	encoded, err := json.MarshalIndent(jsonRequest(normalized), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode semanticdelta JSON: %w", err)
	}
	return append(encoded, '\n'), nil
}

// DecodeJSON parses one strict semanticdelta JSON object.
func DecodeJSON(data []byte) (Request, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var raw *jsonRequest
	if err := decoder.Decode(&raw); err != nil {
		return Request{}, fmt.Errorf("decode semanticdelta JSON: %w", err)
	}
	if raw == nil {
		return Request{}, fmt.Errorf("decode semanticdelta JSON: expected an object")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Request{}, fmt.Errorf("decode semanticdelta JSON: multiple values")
		}
		return Request{}, fmt.Errorf("decode semanticdelta JSON: %w", err)
	}
	return Request(*raw).Normalized()
}

// Decode accepts canonical JSON or the line-oriented text format.
func Decode(data []byte) (Request, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return Request{}, fmt.Errorf("decode semanticdelta: empty input")
	}
	if trimmed[0] == '{' {
		return DecodeJSON(trimmed)
	}
	return DecodeText(trimmed)
}

// Encode returns the requested canonical interchange format.
func Encode(request Request, format Format) ([]byte, error) {
	switch format {
	case FormatJSON, "":
		return EncodeJSON(request)
	case FormatText:
		return EncodeText(request)
	default:
		return nil, fmt.Errorf("unsupported semanticdelta format %q", format)
	}
}

type jsonReport Report
