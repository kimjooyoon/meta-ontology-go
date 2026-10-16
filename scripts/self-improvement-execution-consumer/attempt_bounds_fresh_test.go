package main

import (
	"strconv"
	"testing"

	continuation "github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation"
)

func TestAttemptEnvironmentBoundsFresh1651(t *testing.T) {
	maximum := "2147483647"
	if strconv.IntSize == 64 {
		maximum = "9223372036854775807"
	}
	for _, item := range []struct{ name, raw, expected string }{
		{"first", "1", "1"}, {"second", "2", "2"}, {"zero", "0", "0"},
		{"negative", "-1", "0"}, {"empty", "", "0"}, {"malformed", "1.5", "0"},
		{"space", " 1", "0"}, {"positive-sign", "+1", "1"}, {"leading-zero", "01", "1"},
		{"native-maximum", maximum, maximum}, {"signed-overflow", "9223372036854775808", "0"},
		{"unsigned-overflow", "18446744073709551615", "0"},
	} {
		t.Run(item.name, func(t *testing.T) {
			t.Setenv("ATTEMPT_BOUNDARY_FIXTURE", item.raw)
			if got := strconv.Itoa(envAttempt("ATTEMPT_BOUNDARY_FIXTURE")); got != item.expected {
				t.Fatalf("attempt = %s, expected %s", got, item.expected)
			}
		})
	}
}

func TestSourceAttemptEnvironmentFresh1653(t *testing.T) {
	t.Setenv("SOURCE_RUN_ID", "38840774191")
	for _, raw := range []string{"1", "9223372036854775808"} {
		t.Run(raw, func(t *testing.T) {
			t.Setenv("SOURCE_RUN_ATTEMPT", raw)
			envelope := envelopeFromEnvironment()
			expected := 0
			if raw == "1" {
				expected = 1
			}
			if envelope.Attempt != expected || envelope.RunID != 38840774191 {
				t.Fatal("attempt bounds changed the source run identity", envelope.Attempt, envelope.RunID)
			}
		})
	}
}

func TestOverflowAttemptInvocationFresh1655(t *testing.T) {
	t.Setenv("GITHUB_RUN_ATTEMPT", "9223372036854775808")
	prepared := continuation.Prepared{Metadata: continuation.Metadata{ConsumerRunID: 1655001, ConsumerAttempt: 1}}
	if err := currentInvocation(prepared, options{run: 1655001, attempt: envAttempt("GITHUB_RUN_ATTEMPT")}); err == nil {
		t.Fatal("an out-of-range environment value matched the prepared first attempt")
	}
}
