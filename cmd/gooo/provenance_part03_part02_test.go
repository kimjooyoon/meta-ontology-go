package main

import (
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"os"
	"testing"
)

func (fixture provenanceCLIFixture) publish(t *testing.T, records []provenance.Evidence) ([]byte, int, string) {
	t.Helper()
	data, err := json.Marshal(records)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixture.evidencePath, data, 0o640); err != nil {
		t.Fatal(err)
	}
	return fixture.publishRaw(t)
}
func (fixture provenanceCLIFixture) publishRaw(t *testing.T) ([]byte, int, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run([]string{"provenance", "publish", "--json", fixture.sourcePath, "--store", fixture.storePath, "--evidence", fixture.evidencePath}, &stdout, &stderr)
	return stdout.Bytes(), code, stderr.String()
}
