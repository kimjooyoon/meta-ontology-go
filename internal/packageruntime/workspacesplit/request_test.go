package workspacesplit

import (
	"encoding/json"
	"testing"
)

func TestRequestKeepsExplicitScopeAndCanonicalAddresses(t *testing.T) {
	request := Request{Workspace: "examples/a/workspace.json", Package: "a", Source: "main.gooo",
		Into: "part.gooo", DiscoveryRoot: "examples/a", Cases: "examples/a/cases.json", LineLimit: 75}
	if err := request.Validate(); err != nil {
		t.Fatal(err)
	}
	if request.Subject() != "examples/a/main.gooo" {
		t.Fatal(request.Subject())
	}
	for _, mutate := range []func(*Request){
		func(r *Request) { r.Workspace = "../workspace.json" }, func(r *Request) { r.Source = "./main.gooo" },
		func(r *Request) { r.Into = r.Source }, func(r *Request) { r.DiscoveryRoot = "unrelated" },
		func(r *Request) { r.LineLimit = 0 }, func(r *Request) { r.Cases = "/tmp/cases.json" },
	} {
		bad := request
		mutate(&bad)
		if bad.Validate() == nil {
			t.Fatal("ambiguous input accepted", bad)
		}
	}
	raw, _ := json.Marshal(map[string]any{"schema": "gooo/workspace-split-inputs/v1", "requests": []Request{request}})
	decoded, err := DecodeRequests(raw)
	if err != nil || len(decoded) != 1 || decoded[0] != request {
		t.Fatal(decoded, err)
	}
	for _, raw := range []string{`{"schema":"gooo/workspace-split-inputs/v1","requests":[]}`,
		`{"schema":"wrong","requests":[]}`, `{"schema":"a","schema":"b"}`} {
		if _, err := DecodeRequests([]byte(raw)); err == nil {
			t.Fatal("ambiguous input envelope accepted")
		}
	}
}
