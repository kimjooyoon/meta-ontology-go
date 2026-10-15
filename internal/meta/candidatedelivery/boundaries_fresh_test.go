package candidatedelivery

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDeliveryPartialMetadataFresh1614(t *testing.T) {
	b, _ := bundle1613(t)
	located, _ := selected1613(t, b)
	raw, _ := json.Marshal(located.Run)
	for _, field := range []string{"expired", "digest", "size_in_bytes", "workflow_run"} {
		t.Run(field, func(t *testing.T) {
			a, _ := json.Marshal(located.Artifact)
			var fields map[string]any
			_ = json.Unmarshal(a, &fields)
			delete(fields, field)
			list, _ := json.Marshal(map[string]any{"total_count": 1, "artifacts": []any{fields}})
			assertState1613(t, Select(raw, list, located.Expected), "UNKNOWN")
		})
	}
}

func TestDeliveryZipBoundariesFresh1614(t *testing.T) {
	b, _ := bundle1613(t)
	located, _ := selected1613(t, b)
	for _, kind := range []string{"duplicate-member", "member-count", "member-size", "invalid-zip"} {
		t.Run(kind, func(t *testing.T) {
			var buf bytes.Buffer
			z := zip.NewWriter(&buf)
			count := 1
			if kind == "duplicate-member" {
				count = 2
			}
			if kind == "member-count" {
				count = 129
			}
			for range count {
				entry, err := z.Create("output/first.json")
				if err != nil {
					t.Fatal(err)
				}
				body := []byte("candidate")
				if kind == "member-size" {
					body = []byte(strings.Repeat("x", (8<<20)+1))
				}
				if _, err := entry.Write(body); err != nil {
					t.Fatal(err)
				}
			}
			if err := z.Close(); err != nil {
				t.Fatal(err)
			}
			archive := buf.Bytes()
			if kind == "invalid-zip" {
				archive = []byte("invalid")
			}
			r := located
			r.Artifact.Digest, r.Artifact.Size = Digest(archive), int64(len(archive))
			r = closeReceipt(r, r.State, r.Reason)
			got, _ := Consume(r, archive)
			assertState1613(t, got, "REFUTED")
		})
	}
}

func TestDeliveryLabelIsNotSubjectFresh1614(t *testing.T) {
	b, _ := bundle1613(t)
	r, archive := selected1613(t, b)
	r.Artifact.Name = "self-improvement-nonexecuting-candidate-" + strings.Repeat("d", 40)
	r = closeReceipt(r, r.State, r.Reason)
	got, _ := Consume(r, archive)
	assertState1613(t, got, "CLOSED")
	if got.SubjectSHA != strings.Repeat("1", 40) {
		t.Fatal("label suffix replaced declared subject")
	}
}
