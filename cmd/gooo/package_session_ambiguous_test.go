package main

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

func TestPackageSessionAmbiguousRequestPreservesFollowingClose(t *testing.T) {
	prefix := `{"schema":"gooo/workspace-session-request/v1","id":"ambiguous",`
	cases := map[string]string{
		"operation": prefix + `"operation":"execute","operation":"close"}`,
		"id":        prefix + `"id":"other","operation":"close"}`,
		"schema":    prefix + `"schema":"gooo/workspace-session-request/v1","operation":"close"}`,
		"arguments": prefix + `"operation":"close","arguments":[],"arguments":[]}`,
		"case":      prefix + `"operation":"execute","Operation":"close"}`,
		"escape":    prefix + `"operation":"execute","oper\u0061tion":"close"}`,
	}
	finish := `{"schema":"gooo/workspace-session-request/v1","id":"finish","operation":"close"}`
	for name, frame := range cases {
		var output, diagnostics bytes.Buffer
		code := runWithInput([]string{"package", "session"},
			bytes.NewBufferString(frame+"\n"+finish+"\n"), &output, &diagnostics)
		decoder := json.NewDecoder(&output)
		var invalid, closed packageSessionResponse
		first, second := decoder.Decode(&invalid), decoder.Decode(&closed)
		extra := decoder.Decode(new(any))
		if code != exitOK || first != nil || second != nil || extra != io.EOF ||
			invalid.ExitCode != exitUsage || invalid.Error == "" || invalid.Receipt != nil || invalid.Closed ||
			!closed.Closed || closed.ID != "finish" || closed.ExitCode != exitOK {
			t.Errorf("ambiguous %s frame changed the following lifecycle command: exit=%d first=%v second=%v\n%s",
				name, code, first, second, diagnostics.String())
		}
	}
}
