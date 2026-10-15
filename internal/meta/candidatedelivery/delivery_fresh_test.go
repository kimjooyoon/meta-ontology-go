package candidatedelivery

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	transport "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

func TestDeliverySelectionFresh1613(t *testing.T) {
	bundle, _ := bundle1613(t)
	located, _ := selected1613(t, bundle)
	cases := []struct {
		name, state string
		mutate      func(*Run, *[]Artifact, *int)
	}{
		{"exact-run", "LOCATED", func(*Run, *[]Artifact, *int) {}},
		{"missing-role", "UNKNOWN", func(_ *Run, a *[]Artifact, n *int) { *a = []Artifact{}; *n = 0 }},
		{"duplicate-role", "REFUTED", func(_ *Run, a *[]Artifact, n *int) { *a = append(*a, (*a)[0]); *n = 2 }},
		{"wrong-run", "REFUTED", func(r *Run, _ *[]Artifact, _ *int) { r.ID++ }},
		{"missing-attempt", "UNKNOWN", func(r *Run, _ *[]Artifact, _ *int) { r.Attempt = 0 }},
		{"wrong-attempt", "REFUTED", func(r *Run, _ *[]Artifact, _ *int) { r.Attempt++ }},
		{"expired", "UNKNOWN", func(_ *Run, a *[]Artifact, _ *int) { (*a)[0].Expired = true }},
		{"incomplete-population", "UNKNOWN", func(_ *Run, _ *[]Artifact, n *int) { *n = 2 }},
		{"missing-event", "UNKNOWN", func(r *Run, _ *[]Artifact, _ *int) { r.Event = "" }},
		{"artifact-run-mismatch", "REFUTED", func(_ *Run, a *[]Artifact, _ *int) { (*a)[0].Run.ID++ }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			run, artifacts, total := located.Run, []Artifact{located.Artifact}, 1
			c.mutate(&run, &artifacts, &total)
			raw, _ := json.Marshal(run)
			list, _ := json.Marshal(map[string]any{"total_count": total, "artifacts": artifacts})
			r := Select(raw, list, located.Expected)
			assertState1613(t, r, c.state)
		})
	}
}

func TestDeliveryArchiveBindingFresh1613(t *testing.T) {
	base, manifest := bundle1613(t)
	cases := []struct {
		name, state string
		mutate      func(*Bundle, *Manifest)
	}{
		{"cross-origin", "CLOSED", func(*Bundle, *Manifest) {}},
		{"same-origin-push", "CLOSED", func(_ *Bundle, m *Manifest) {
			m.Event = "push"
			m.WorkflowSHA = m.SubjectSHA
			m.RunHeadSHA = m.SubjectSHA
		}},
		{"subject-contradiction", "REFUTED", func(_ *Bundle, m *Manifest) { m.SubjectSHA = strings.Repeat("c", 40) }},
		{"workflow-contradiction", "REFUTED", func(_ *Bundle, m *Manifest) { m.WorkflowSHA = strings.Repeat("c", 40) }},
		{"workflow-missing", "UNKNOWN", func(_ *Bundle, m *Manifest) { m.WorkflowSHA = "" }},
		{"candidate-missing", "UNKNOWN", func(b *Bundle, _ *Manifest) { b.Candidate = nil }},
		{"manifest-missing", "UNKNOWN", func(b *Bundle, _ *Manifest) { b.Manifest = nil }},
		{"transport-origin-missing", "UNKNOWN", missingTransport1613},
		{"payload-bytes-changed", "REFUTED", func(b *Bundle, _ *Manifest) { b.Candidate = append(b.Candidate, ' ') }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b, m := base, manifest
			c.mutate(&b, &m)
			if c.name != "manifest-missing" {
				m.Digest = ""
				m.Digest = digestJSON(m)
				b.Manifest, _ = json.Marshal(m)
			}
			r, archive := selected1613(t, b)
			if c.name == "same-origin-push" {
				r.Run.Event, r.Run.Head, r.Artifact.Run.Head = m.Event, m.RunHeadSHA, m.RunHeadSHA
				r = closeReceipt(r, r.State, r.Reason)
			}
			r, got := Consume(r, archive)
			assertState1613(t, r, c.state)
			if r.State == "CLOSED" && (r.SubjectSHA != manifest.SubjectSHA || len(got.Candidate) == 0 || Metadata(r).SubjectSHA != r.SubjectSHA) {
				t.Fatal("lost logical subject")
			}
			if r.State != "CLOSED" && len(got.Candidate) != 0 {
				t.Fatal("unbound delivery exposed a candidate")
			}
		})
	}
}

func missingTransport1613(b *Bundle, m *Manifest) {
	var old transport.Report
	_ = json.Unmarshal(b.Transport, &old)
	metadata := old.Transport
	metadata.ProducerEvent = ""
	raw, _ := json.Marshal(metadata)
	producer, _ := json.Marshal(old.Producer)
	report := transport.Evaluate(os.DirFS("../../.."), "examples/self-improvement/transport.gooo", repo1613,
		old.Transport.ProducerRunID, observation1613(), producer, raw, old.ActualArchiveDigest)
	b.Transport, _ = json.Marshal(report)
	m.TransportDigest = Digest(b.Transport)
}

func TestDeliveryArchiveTamperFresh1613(t *testing.T) {
	b, _ := bundle1613(t)
	r, archive := selected1613(t, b)
	archive[len(archive)/2] ^= 1
	got, _ := Consume(r, archive)
	assertState1613(t, got, "REFUTED")
	r, _ = Consume(r, nil)
	assertState1613(t, r, "UNKNOWN")
}

func assertState1613(t *testing.T, r Receipt, state string) {
	t.Helper()
	if r.State != state {
		t.Fatalf("want %s, got %s/%s", state, r.State, r.Reason)
	}
	if state == "UNKNOWN" && (r.Unknown == nil || r.Unknown.Stage == "" || r.Unknown.Step == "" || r.Unknown.Reason == "" || r.Unknown.UnknownClass == "" || r.Unknown.NextOperation == "" || len(r.Unknown.BlockedBy) == 0) {
		t.Fatal("UNKNOWN lost causal fields")
	}
}
