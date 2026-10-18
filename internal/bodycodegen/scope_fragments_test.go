package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type scopeFragmentRequest struct {
	Branch string   `json:"branch"`
	Paths  []string `json:"paths"`
}

// The independent oracle uses Go's JSON encoder and parser. It does not execute
// a second copy of the Gooo validation algorithm to decide expected acceptance.
func scopeFragmentExpected(t *testing.T, request scopeFragmentRequest, accepted bool) string {
	t.Helper()
	result := map[string]string{"status": "OUTSIDE_PROFILE", "governance": "", "golang": "", "markdown": ""}
	if accepted {
		encoded, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		var lines, cells []string
		for _, path := range request.Paths {
			lines = append(lines, fmt.Sprintf("\t%q,\n", path))
			cells = append(cells, "`"+path+"`")
		}
		result["status"], result["governance"] = "PROPOSAL", string(encoded)
		result["golang"] = fmt.Sprintf("branchScopeAllowlist[%q] = []string{\n%s}", request.Branch, strings.Join(lines, ""))
		result["markdown"] = "| `" + request.Branch + "` | " + strings.Join(cells, ", ") + " |"
		if _, err := format.Source([]byte("package p\nfunc f(){\n" + result["golang"] + "\n}")); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestScopeFragmentsSource(t *testing.T) {
	source, err := os.ReadFile("../../examples/scope-fragments/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate("scope.gooo", source, "Project")
	if err != nil {
		t.Fatal(err)
	}
	maximum := scopeFragmentRequest{Branch: "agent/" + strings.Repeat("b", 250)}
	for i := range 256 {
		maximum.Paths = append(maximum.Paths, strings.Repeat("a", 252)+fmt.Sprintf("/%03d", i))
	}
	controls := []struct {
		name     string
		request  scopeFragmentRequest
		accepted bool
	}{
		{"ordinary", scopeFragmentRequest{"agent/example", []string{"docs/a.md", "examples/b.gooo"}}, true},
		{"maximum", maximum, true},
		{"empty", scopeFragmentRequest{"agent/example", []string{}}, false},
		{"duplicate", scopeFragmentRequest{"agent/example", []string{"a", "a"}}, false},
		{"unsorted", scopeFragmentRequest{"agent/example", []string{"b", "a"}}, false},
		{"traversal", scopeFragmentRequest{"agent/example", []string{"a/../b"}}, false},
		{"empty-component", scopeFragmentRequest{"agent/example", []string{"a//b"}}, false},
		{"absolute", scopeFragmentRequest{"agent/example", []string{"/a"}}, false},
		{"trailing", scopeFragmentRequest{"agent/example", []string{"a/"}}, false},
		{"dot", scopeFragmentRequest{"agent/example", []string{"a/./b"}}, false},
		{"quoted", scopeFragmentRequest{"agent/example", []string{"a\"b"}}, false},
		{"unicode", scopeFragmentRequest{"agent/example", []string{"한글.gooo"}}, false},
		{"outside-branch", scopeFragmentRequest{"dev", []string{"a"}}, false},
		{"long-path", scopeFragmentRequest{"agent/example", []string{strings.Repeat("a", 257)}}, false},
		{"long-branch", scopeFragmentRequest{"agent/" + strings.Repeat("a", 251), []string{"a"}}, false},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			inputs, err := json.Marshal([]scopeFragmentRequest{control.request})
			if err != nil {
				t.Fatal(err)
			}
			cases := []assemblyspec.ValueCase{{Inputs: string(inputs), Expected: scopeFragmentExpected(t, control.request, control.accepted)}}
			observed, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Project", generated.Report.RecordTypes, cases)
			if err != nil || len(observed) != 1 || !observed[0].Passed {
				t.Fatal("source result differs from independent expectation", err, observed)
			}
		})
	}
}
