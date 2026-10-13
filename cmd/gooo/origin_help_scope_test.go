package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRegisteredOriginHelpForms5791(t *testing.T) {
	for _, args := range [][]string{{"help", "package", "origins"}, {"package", "origins", "--help"}, {"--help", "package", "origins"}} {
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK || diagnostics.Len() != 0 || !strings.Contains(out.String(), packageOriginsUsage) {
			t.Fatal(args, code, out.String(), diagnostics.String())
		}
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"help", "unknown5791", "topic5791"}, &out, &diagnostics); code != exitUsage || !strings.Contains(diagnostics.String(), "usage: gooo help") {
		t.Fatal("unknown multi-word help changed", code, diagnostics.String())
	}
	t.Log("5791: three registered origin help forms resolve; unknown multi-word topic retains usage error")
}
