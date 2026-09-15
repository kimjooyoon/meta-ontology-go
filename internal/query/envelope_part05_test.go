package query

import (
	"encoding/json"
	"testing"
)

func withTarget(request Request, target ID) Request {
	request.Target = target
	return request
}
func authorityLabel(metadata EnvelopeMetadata, view string) AuthorityLabel {
	for _, label := range metadata.AuthorityLabels {
		if label.View == view {
			return label
		}
	}
	return AuthorityLabel{}
}
func mustMarshal(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
func (response Response) CanonicalDigestValue() string {
	digest, err := response.CanonicalDigest()
	if err != nil {
		return ""
	}
	return digest
}
