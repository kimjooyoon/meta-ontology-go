package extractor

import (
	"strings"
	"testing"
)

func TestCallbackCounterexamplesSurviveIncompleteStreams(t *testing.T) {
	failed := "{\"Action\":\"fail\",\"Test\":\"TestSubject\"}\n"
	passed := strings.Replace(failed, "\"fail\"", "\"pass\"", 1)
	unknown := strings.Replace(failed, "\"fail\"", "\"future\"", 1)
	cases := []struct {
		name, raw, state string
	}{
		{"failure then malformed JSON", failed + "{", "REFUTED"},
		{"failure then unknown action", failed + unknown, "REFUTED"},
		{"failure then duplicate pass", failed + passed, "REFUTED"},
		{"pass then duplicate failure", passed + failed, "REFUTED"},
		{"other failure then malformed", strings.Replace(failed, "TestSubject", "TestOther", 1) + passed + "{", "REFUTED"},
		{"malformed without failure", passed + "{", "UNKNOWN"},
		{"duplicate pass only", passed + passed, "UNKNOWN"},
		{"unknown action without failure", passed + unknown, "UNKNOWN"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			events, err := callbackPackageTestEvents([]byte(tc.raw), "TestSubject")
			if err == nil {
				t.Fatal("incomplete or contradictory event stream was accepted")
			}
			assertCallbackCounterexampleClaim(t, events, tc.state)
		})
	}
}

func TestCallbackCompleteEventsRemainUsable(t *testing.T) {
	raw := "{\"Action\":\"run\",\"Test\":\"TestSubject\"}\n" +
		"{\"Action\":\"pass\",\"Test\":\"TestSubject\"}\n"
	events, err := callbackPackageTestEvents([]byte(raw), "TestSubject")
	if err != nil || len(events) != 1 || events[0].Name != "TestSubject" || events[0].Action != "pass" {
		t.Fatalf("complete events=%+v error=%v", events, err)
	}
}
