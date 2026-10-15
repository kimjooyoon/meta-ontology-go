package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSourceControlReaderControlledCases(t *testing.T) {
	var out, diagnostics bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", "../../examples/source-control-reader/cases.json",
		"../../examples/source-control-reader/gooo.workspace.json"}, &out, &diagnostics)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
		t.Fatal("source control reader", code, err, receipt.Error, diagnostics.String())
	}
	runtime := receipt.Result.Runtime
	if receipt.Decision != "PASS" || runtime.FiniteTotal != 15 || runtime.FinitePassed != 15 || runtime.ModelCalls != 0 {
		t.Fatal("source control outcomes", receipt.Decision, runtime.FiniteTotal, runtime.FinitePassed, runtime.ModelCalls)
	}
}
