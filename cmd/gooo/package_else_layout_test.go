package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestElseLayoutNativeProgramAndGoooSourceReader(t *testing.T) {
	for _, test := range []struct {
		name, cases, workspace string
		count                  int
	}{
		{"program", "../../examples/else-layout/cases.json", "../../examples/else-layout/gooo.workspace.json", 8},
		{"source reader", "../../examples/else-layout/source-control-case.json",
			"../../examples/source-control-reader/gooo.workspace.json", 1},
		{"named syntax policy", "../../internal/meta/bodysurface/policy-cases.json",
			"../../internal/meta/bodysurface/gooo.workspace.json", 16},
	} {
		t.Run(test.name, func(t *testing.T) {
			var out, diagnostics bytes.Buffer
			code := run([]string{"package", "execute", "--json", "--cases", test.cases, test.workspace}, &out, &diagnostics)
			var receipt packageExecutionReceipt
			if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
				t.Fatal("else layout integration", code, err, receipt.Error, diagnostics.String())
			}
			if r := receipt.Result.Runtime; receipt.Decision != "PASS" || r.FiniteTotal != test.count ||
				r.FinitePassed != test.count || r.ModelCalls != 0 {
				t.Fatal("else layout outcome", receipt.Decision, r.FiniteTotal, r.FinitePassed, r.ModelCalls)
			}
		})
	}
}
