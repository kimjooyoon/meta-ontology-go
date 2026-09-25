package syntax

import (
	"bytes"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestSupportedEntityFieldsMalformedInputPublishesNoPartialFields(t *testing.T) {
	support := supportedEntityFields()
	seeds := []string{
		`package p namespace n entity E id "urn:e" fields { field id "urn:f" type string required one }`,
		`package p namespace n entity E id "urn:e" fields { field F "urn:f" type string required one }`,
		`package p namespace n entity E id "urn:e" fields field F id "urn:f" type string required one`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" string string required one }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string required }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string one }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string required manyx }`,
		`package p namespace n entity E id "urn:e" fields { field field id "urn:f" type string required one }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type type string required one }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string required one garbage }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string required one`,
		`package p namespace n entity E id "urn:e" fields { field Good id "urn:good" type string required one field Bad id "urn:bad" type string required }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type @ required one }`,
	}
	for _, source := range seeds {
		first, firstDiagnostics := ParseWithEntityFieldsSupport(source, support)
		second, secondDiagnostics := ParseWithEntityFieldsSupport(source, support)
		if !reflect.DeepEqual(first, second) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) || len(firstDiagnostics) == 0 {
			t.Fatalf("malformed replay changed: %#v %#v / %#v %#v", first, firstDiagnostics, second, secondDiagnostics)
		}
		if first != nil && len(first.Declarations) != 0 && first.Declarations[0].(*EntityDecl).Fields != nil {
			t.Fatalf("malformed source published fields: %#v", first.Declarations[0].(*EntityDecl).Fields)
		}
		formatted, formatDiagnostics, err := FormatSourceWithEntityFieldsSupport("bad.gooo", source, support)
		if formatted != "" || len(formatDiagnostics) == 0 || err == nil {
			t.Fatalf("malformed format wrote output: %q %#v %v", formatted, formatDiagnostics, err)
		}
	}
}
func TestEntityFieldsCanonicalReplayFreshProcess(t *testing.T) {
	const helper = "GOOO_ENTITY_FIELDS_REPLAY_HELPER"
	if os.Getenv(helper) == "1" {
		source := `package p namespace n entity E id "urn:e" fields { field F id "urn:f" type "gooo:string" optional many }`
		formatted, diagnostics, err := FormatSourceWithEntityFieldsSupport("fresh.gooo", source, supportedEntityFields())
		if err != nil || len(diagnostics) != 0 {
			t.Fatalf("fresh helper format = %q, %#v, %v", formatted, diagnostics, err)
		}
		_, _ = os.Stdout.WriteString(formatted)
		return
	}
	run := func() []byte {
		command := exec.Command(os.Args[0], "-test.run", "^TestEntityFieldsCanonicalReplayFreshProcess$")
		command.Env = append(os.Environ(), helper+"=1")
		output, err := command.Output()
		if err != nil {
			t.Fatalf("fresh replay helper failed: %v", err)
		}
		return output
	}
	first, second := run(), run()
	if !bytes.Equal(first, second) {
		t.Fatalf("fresh-process canonical replay changed: %q != %q", first, second)
	}
}
