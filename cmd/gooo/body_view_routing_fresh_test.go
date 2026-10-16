package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestAdvisoryCapabilitiesRoutingFresh1819(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"body-view", "--capabilities"}, &stdout, &stderr); code != exitOK {
		t.Fatalf("capabilities route: %d %s", code, &stderr)
	}
	var capabilities struct {
		Authority        string `json:"authority"`
		AutomaticStorage bool   `json:"automatic_storage"`
		HintLimit        int    `json:"hint_byte_limit"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &capabilities); err != nil || capabilities.Authority != "advisory" ||
		capabilities.AutomaticStorage || capabilities.HintLimit != 4096 {
		t.Fatalf("machine contract: %v %s", err, &stdout)
	}
	stdout.Reset()
	if code := run([]string{"help", "body-view"}, &stdout, &stderr); code != exitOK ||
		!strings.Contains(stdout.String(), "Selection counts") && !strings.Contains(stdout.String(), "Counts describe requests") {
		t.Fatalf("help route: %d %s %s", code, &stdout, &stderr)
	}
	stdout.Reset()
	if code := run([]string{"view-lens", "read", "--state-dir", t.TempDir(), "--digest", "invalid"},
		&stdout, &stderr); code != exitOK || !strings.Contains(stdout.String(), "ARCHIVE_REQUEST_INVALID") {
		t.Fatalf("detached route: %d %s %s", code, &stdout, &stderr)
	}
}
