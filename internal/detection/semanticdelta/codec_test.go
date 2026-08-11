package semanticdelta

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func sampleRequest() Request {
	return Request{
		Allowed: Scope{
			IDs:        []string{"billing://activity/pay-order"},
			Prefixes:   []string{"billing://entity/"},
			Predicates: []string{"gooo:invokes"},
		},
		Delta: Delta{
			AddedFacts: []Fact{{Subject: "billing://activity/pay-order", Predicate: "gooo:invokes", Object: "fraud://activity/check"}},
			AddedNodes: []Node{{ID: "fraud://activity/check", Kind: "Activity"}},
		},
	}
}

func TestJSONAndTextRoundTripToCanonicalBytes(t *testing.T) {
	request := sampleRequest()
	jsonFirst, err := EncodeJSON(request)
	if err != nil {
		t.Fatal(err)
	}
	jsonSecond, err := EncodeJSON(request)
	if err != nil || !bytes.Equal(jsonFirst, jsonSecond) {
		t.Fatalf("JSON encoding is not deterministic: %q != %q (%v)", jsonFirst, jsonSecond, err)
	}
	decodedJSON, err := DecodeJSON(jsonFirst)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decodedJSON, mustNormalizeRequest(t, request)) {
		t.Fatalf("JSON round trip = %#v", decodedJSON)
	}
	textFirst, err := EncodeText(request)
	if err != nil {
		t.Fatal(err)
	}
	decodedText, err := DecodeText(textFirst)
	if err != nil {
		t.Fatal(err)
	}
	textSecond, err := EncodeText(decodedText)
	if err != nil || !bytes.Equal(textFirst, textSecond) {
		t.Fatalf("text round trip is not canonical: %q != %q (%v)", textFirst, textSecond, err)
	}
}

func TestDecodeRejectsUnknownFieldsAndMultipleJSONValues(t *testing.T) {
	if _, err := DecodeJSON([]byte(`{"version":"semanticdelta/v1","allowed":{},"delta":{},"extra":true}`)); err == nil {
		t.Fatal("unknown JSON field was accepted")
	}
	if _, err := DecodeJSON([]byte(`{"version":"semanticdelta/v1","allowed":{},"delta":{}} {}`)); err == nil {
		t.Fatal("multiple JSON values were accepted")
	}
}

func TestDecodeSelectsTextAndRequiresVersion(t *testing.T) {
	text := "# comment\nversion semanticdelta/v1\nscope id billing://activity/pay-order\n"
	request, err := Decode([]byte(text))
	if err != nil || len(request.Allowed.IDs) != 1 {
		t.Fatalf("text detection = %#v, %v", request, err)
	}
	if _, err := DecodeText([]byte("scope id billing://activity/pay-order\n")); err == nil {
		t.Fatal("text without version was accepted")
	}
	if !strings.HasPrefix(string(mustEncodeText(t, request)), "version\tsemanticdelta/v1\n") {
		t.Fatal("canonical text did not include version")
	}
}

func mustNormalizeRequest(t *testing.T, request Request) Request {
	t.Helper()
	normalized, err := request.Normalized()
	if err != nil {
		t.Fatal(err)
	}
	return normalized
}

func mustEncodeText(t *testing.T, request Request) []byte {
	t.Helper()
	encoded, err := EncodeText(request)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
