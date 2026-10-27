package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectionOwnershipObservationDoesNotSelectGeneratedOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "evidence.json")
	raw := `{"schema":"gooo.repository-projection-evidence.v1","source_sha":"head","subjects":[
{"indicator":"source.line-cap-debt","logical":"generated.go","value":100,"limit":75,"applicability":"NOT_APPLICABLE","applicability_reason":"GENERATED_SOURCE_OWNER_REQUIRED"},
{"indicator":"source.line-cap-debt","logical":"handwritten.go","value":100,"limit":75}
]}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := loadSubjects(path, "head")
	if err != nil || len(got) != 1 || got[0].Logical != "handwritten.go" || got[0].Value != 100 {
		t.Fatal(got, err)
	}
}
