package sourcerepack

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
)

func TestRepackBlockedObservationPreservesFields(t *testing.T) {
	cases := []struct {
		reason string
		want   string
	}{
		{"no-movable-declaration", `source-repacker: blocked subject="fixture.go" reason="no-movable-declaration"`},
		{"no-build-domain-destination", `source-repacker: blocked subject="fixture.go" reason="no-build-domain-destination"`},
		{"requires-multiple-or-extraction", `source-repacker: blocked subject="fixture.go" reason="requires-multiple-or-extraction"`},
		{"import-conflict", `source-repacker: blocked subject="fixture.go" reason="import-conflict"`},
		{"destination-capacity", `source-repacker: blocked subject="fixture.go" reason="destination-capacity"`},
	}
	for _, test := range cases {
		t.Run(test.reason, func(t *testing.T) {
			blocked := repackBlocked{Subject: "fixture.go", Reason: test.reason}
			if got := blocked.observation(); got != test.want {
				t.Fatalf("observation differs: %q", got)
			}
			if blocked.Error() != "fixture.go: "+test.reason || !errors.Is(blocked, errRepackBlocked) {
				t.Fatal("write-mode error identity or meaning changed")
			}
			retained, ok := errors.AsType[repackBlocked](fmt.Errorf("plan: %w", blocked))
			if !ok || retained != blocked {
				t.Fatal("wrapped planning outcome lost its fields")
			}
		})
	}
}

func TestRepackDiagnosticPatternBoundary(t *testing.T) {
	// Literal pattern from actions/setup-go b7ad1dad31e06c5925ef5d2fc7ad053ef454303e.
	// This checks the observed formatter boundary, not the GitHub runner itself.
	matcher := regexp.MustCompile(`^\s*(.+\.go):(?:(\d+):(\d+):)? (.*)`)
	old := "source-repacker: blocked subject=fixture.go detail=fixture.go: destination-capacity"
	if !matcher.MatchString(old) {
		t.Fatal("original false error annotation is not represented")
	}
	blocked := repackBlocked{Subject: "folder with spaces/fixture.go", Reason: "destination-capacity"}
	if matcher.MatchString(blocked.observation()) {
		t.Fatal("planning observation has a compiler diagnostic shape")
	}
	groups := matcher.FindStringSubmatch("fixture.go:3:2: undefined: input")
	if len(groups) != 5 || groups[1] != "fixture.go" || groups[2] != "3" || groups[3] != "2" {
		t.Fatal("actual compiler diagnostic was not retained")
	}
}
