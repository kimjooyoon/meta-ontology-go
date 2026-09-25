package semanticbinding

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func canonicalFixture(records []fixtureRecord) string {
	parts := make([]string, 0, len(records))
	for _, record := range records {
		if record.Directive == "obligation" {
			parts = append(parts, fmt.Sprintf("obligation|%s|%s|%s", record.ID, record.Subject, record.Pressure))
			continue
		}
		parts = append(parts, fmt.Sprintf("bind|%s", record.ID))
	}
	return strings.Join(parts, "\n")
}
func loadFixture(t *testing.T, name string) ([]byte, fixtureExpectation) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", name+".go"))
	if err != nil {
		t.Fatal(err)
	}
	wantBytes, err := os.ReadFile(filepath.Join("testdata", name+".want.json"))
	if err != nil {
		t.Fatal(err)
	}
	var want fixtureExpectation
	if err := json.Unmarshal(wantBytes, &want); err != nil {
		t.Fatalf("decode %s.want.json: %v", name, err)
	}
	return source, want
}
