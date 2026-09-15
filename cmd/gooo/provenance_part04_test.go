package main

import (
	"encoding/json"
	"os"
	"testing"
)

func decodeProvenanceResponse(t *testing.T, data []byte) provenancePublishResponse {
	t.Helper()
	var response provenancePublishResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("decode provenance response %q: %v", data, err)
	}
	return response
}
func assertCanonicalProvenanceResponse(t *testing.T, response provenancePublishResponse) {
	t.Helper()
	want := response.CanonicalDigest
	response.CanonicalDigest = ""
	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256Digest(payload); got != want {
		t.Fatalf("canonical response digest = %s, want %s", got, want)
	}
}
func decodeManifestFixture(t *testing.T, path string) map[string]any {
	t.Helper()
	data := mustReadProvenanceFile(t, path+".manifest")
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest
}
func mustReadProvenanceFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
