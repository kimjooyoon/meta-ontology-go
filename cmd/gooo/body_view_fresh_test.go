package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/viewlens"
)

const advisorySource1807 = `package advisory1807
namespace advisory1807
entity Integer id "urn:gooo:type:integer"
activity Balance(원금 Integer, deduction Integer) -> Integer id "urn:advisory1807:balance" computes ` + "`" + `
let remaining = 원금 - deduction
if remaining < 0 { return 0 } else { return remaining }
` + "`" + `
`

func TestCanonicalViewIsolationFresh1807(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "http://127.0.0.1:1/must-not-call")
	const filename = "advisory1807.gooo"
	canonical, err := bodycodegen.InspectTypedValueOrigins(filename, []byte(advisorySource1807), "Balance")
	if err != nil || canonical.Status != "RESOLVED" {
		t.Fatalf("source inspection: %v %+v", err, canonical)
	}
	want, _ := json.Marshal(canonical)
	for _, item := range []struct{ name, hint, view, actor string }{
		{"default", "", "auto", "kimjooyoon"}, {"full", "", "full", "kimjooyoon"},
		{"broken-json", "{", "inputs", "kimjooyoon"}, {"unknown-view", "", "missing", "kimjooyoon"},
		{"bad-identity", "", "inputs", "bad\nactor"},
		{"hint-bound", strings.Repeat("x", 4097), "inputs", "kimjooyoon"},
	} {
		t.Run(item.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runBodyView([]string{"--activity", "Balance", "--audience", "person", "--actor", item.actor,
				"--hint", item.hint, "--view", item.view, filename}, fixtureReader{source: advisorySource1807}, &stdout, &stderr)
			if code != exitOK || stderr.Len() != 0 {
				t.Fatalf("hint changed canonical success: code=%d %s %s", code, &stdout, &stderr)
			}
			var got bodyViewOutput
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(got.Canonical)
			if !bytes.Equal(raw, want) || got.Projection.CanonicalPointer != "#/canonical" {
				t.Fatal("hint changed or copied canonical origins")
			}
			if got.Selection.Choice.ViewID == viewlens.NamedInputs && len(got.Projection.InputRefs) != 2 {
				t.Fatalf("named references: %+v", got.Projection)
			}
		})
	}
}

func TestViewSourceErrorsRemainFresh1808(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runBodyView([]string{"--activity", "Missing", "--hint", "{", "broken1808.gooo"},
		fixtureReader{source: "package broken1808\nactivity {"}, &stdout, &stderr)
	if code != exitFailure || !strings.Contains(stdout.String(), "SOURCE_INSPECTION_FAILED") {
		t.Fatalf("source error lost: code=%d %s %s", code, &stdout, &stderr)
	}
}
