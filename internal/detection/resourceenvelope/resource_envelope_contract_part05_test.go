package resourceenvelope

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func readCorpus(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "resource-envelope-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func decodeStrict(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return io.ErrUnexpectedEOF
		}
		return err
	}
	return nil
}
func contractByName(t *testing.T, corpus contractCorpus, name string) contractCase {
	t.Helper()
	for _, testCase := range corpus.Cases {
		if testCase.Name == name {
			return testCase
		}
	}
	t.Fatalf("missing contract case %q", name)
	return contractCase{}
}
