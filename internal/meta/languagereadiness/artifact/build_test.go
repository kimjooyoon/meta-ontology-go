package artifact

import "testing"

func TestBuildRejectsUnknownConceptDecision(t *testing.T) {
	raw := []byte(`{"schema":"gooo/language-concept-artifact/v1","decision":"FUTURE_DECISION"}`)
	_, err := Build(raw, "0000000000000000000000000000000000000000")
	if err == nil {
		t.Fatal("unknown concept decision produced readiness artifact")
	}
}
