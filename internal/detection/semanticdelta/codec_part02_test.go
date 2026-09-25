package semanticdelta

import (
	"testing"
)

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
