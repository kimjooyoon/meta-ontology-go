package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestSelectiveCIShadowUsageRejectsUnknownDuplicateMissingAndTrailing(t *testing.T) {
	fixture := newShadowFixture(t)
	cases := []struct {
		name string
		args []string
	}{
		{name: "unknown flag", args: append(fixture.args(), "--unknown", "value")},
		{name: "duplicate flag", args: append(fixture.args(), "--lane-input", "lane.json")},
		{name: "missing value", args: []string{"shadow", "--base-snapshot"}},
		{name: "trailing positional", args: append(fixture.args(), "trailing")},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := runSelectiveCI(test.args, fixture.reader(), &stdout, &stderr); code != exitUsage || stdout.Len() != 0 || !strings.Contains(stderr.String(), "usage: gooo selective-ci shadow") {
				t.Fatalf("usage result = code:%d stdout:%q stderr:%q", code, stdout.String(), stderr.String())
			}
		})
	}

	for _, name := range []string{"base_snapshot", "head_snapshot", "plan_input", "evidence_input", "lane_input"} {
		t.Run("missing "+name, func(t *testing.T) {
			missing := newShadowFixture(t)
			delete(missing.files, name+".json")
			var stdout, stderr bytes.Buffer
			if code := runSelectiveCI(missing.args(), missing.reader(), &stdout, &stderr); code != exitUsage || stdout.Len() != 0 || !strings.Contains(stderr.String(), "cli.usage") {
				t.Fatalf("missing result = code:%d stdout:%q stderr:%q", code, stdout.String(), stderr.String())
			}
		})
	}
}
