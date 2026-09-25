package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestCapturedNativeJournalRejectsUnboundTerminals(t *testing.T) {
	first := capturedNativeJournal(t, "first.ndjson")
	replay := capturedNativeJournal(t, "replay.ndjson")
	lines := bytes.Split(bytes.TrimSpace(first), []byte("\n"))
	var terminal costEvent
	if err := json.Unmarshal(lines[len(lines)-1], &terminal); err != nil {
		t.Fatal(err)
	}
	foreignAttempt, err := json.Marshal(terminal)
	if err != nil {
		t.Fatal(err)
	}
	wrongHead := terminal
	wrongHead.Invocation = capturedReplayInvocation
	wrongHead.Sequence = 17
	wrongHead.Head = strings.Repeat("0", 40)
	wrongHeadJSON, err := json.Marshal(wrongHead)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		tail   []byte
		reason string
	}{
		{"foreign-attempt-terminal", foreignAttempt, "invalid or reused interval start"},
		{"wrong-head-terminal", wrongHeadJSON, "invalid or reused interval start"},
		{"truncated-terminal", []byte("{\"schema\":"), "invalid NDJSON event"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := bytes.Join([][]byte{bytes.TrimSpace(replay), tc.tail}, []byte("\n"))
			if _, err := readCostReport(bytes.NewReader(input)); err == nil || !strings.Contains(err.Error(), tc.reason) {
				t.Fatalf("unbound terminal was not rejected: %v", err)
			}
		})
	}
}
