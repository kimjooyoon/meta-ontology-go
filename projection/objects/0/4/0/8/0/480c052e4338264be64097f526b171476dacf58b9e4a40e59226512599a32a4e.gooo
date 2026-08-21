package metarecognition

import (
	"strings"
	"testing"
)

func TestReplayRejectsSchemaAndDuplicates(t *testing.T) {
	caseJSON := `{"id":"case-x","subject":"SEMANTIC_BINDING","source":"case.go","roots":[],"commands":[],"paths":[]}`
	valid := `{"schema":"` + SchemaVersion + `","cases":[` + caseJSON + `]}`
	unknown := strings.Replace(valid, `,"cases"`, `,"extra":true,"cases"`, 1)
	duplicateCase := `{"schema":"` + SchemaVersion + `","cases":[` + caseJSON + `,` + caseJSON + `]}`
	duplicateCommand := `{"schema":"` + SchemaVersion + `","cases":[{"id":"case-x","subject":"SEMANTIC_BINDING","roots":[],"commands":["cmd","cmd"],"paths":[]}]}`
	duplicatePath := `{"schema":"` + SchemaVersion + `","cases":[{"id":"case-x","subject":"SEMANTIC_BINDING","roots":[],"commands":[],"paths":["path","path"]}]}`
	duplicateField := `{"schema":"` + SchemaVersion + `","schema":"` + SchemaVersion + `","cases":[]}`
	for name, data := range map[string]string{"unknown": unknown, "case": duplicateCase, "command": duplicateCommand, "path": duplicatePath, "field": duplicateField} {
		if _, err := DecodeReplayJSON([]byte(data)); err == nil {
			t.Errorf("DecodeReplayJSON accepted %s duplicate/invalid input", name)
		}
	}
	if _, err := DecodeReplayJSON([]byte(valid)); err != nil {
		t.Fatal(err)
	}
	invalidPaths := map[string][2]string{
		"relative-root":  {"workspace/fixture", "/workspace/fixture/case.go"},
		"escaped-source": {"/workspace/fixture", "/workspace/fixture/../secret.go"},
		"ambiguous-root": {"/workspace/./fixture", "/workspace/fixture/case.go"},
		"outside-root":   {"/workspace/fixture", "/workspace/fixtures/case.go"},
	}
	for name, values := range invalidPaths {
		if _, err := canonicalRootRelativePath(values[0], values[1]); err == nil {
			t.Errorf("canonicalRootRelativePath accepted %s", name)
		}
	}
}
