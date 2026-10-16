package executioncontinuation

import (
	"fmt"
	"testing"
)

func TestWrappedFailureFresh1623(t *testing.T) {
	for _, c := range []struct {
		name          string
		err           error
		state, reason string
	}{
		{"direct", refuted("ORIGIN_CONFLICT"), "REFUTED", "ORIGIN_CONFLICT"},
		{"wrapped", fmt.Errorf("read: %w", unknown("SOURCE_MEMBER_MISSING")), "UNKNOWN", "SOURCE_MEMBER_MISSING"},
		{"untyped", fmt.Errorf("unavailable"), "UNKNOWN", "PROTOCOL_OBSERVATION_UNAVAILABLE"},
		{"nil", nil, "UNKNOWN", "PROTOCOL_OBSERVATION_UNAVAILABLE"},
	} {
		t.Run(c.name, func(t *testing.T) {
			state, reason := errorState(c.err)
			if state != c.state || reason != c.reason {
				t.Fatal("typed failure origin lost", state, reason)
			}
		})
	}
}

func TestPromotedRunIdentityFresh1623(t *testing.T) {
	r := run1617("consumer", 9162301, "0123456789012345678901234567890123456789")
	if r.ID != 9162301 || r.Attempt != 1 || r.Repository.Name != "kimjooyoon/meta-ontology-go" ||
		r.Path != ".github/workflows/self-improvement-execution-consumer.yml" || r.Branch != "dev" {
		t.Fatal("promoted literal changed origin", r)
	}
}
