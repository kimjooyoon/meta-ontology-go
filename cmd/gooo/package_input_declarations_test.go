package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestInputDeclarationGoooCases(t *testing.T) {
	var out, diagnostics bytes.Buffer
	code := run([]string{"package", "execute", "--json", "--cases", "../../examples/source-control-reader/input-cases.json",
		"../../examples/source-control-reader/inputs.workspace.json"}, &out, &diagnostics)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || code != exitOK || receipt.Result == nil {
		t.Fatal("input declaration reader", code, err, diagnostics.String())
	}
	r := receipt.Result.Runtime
	if receipt.Decision != "PASS" || r.FiniteTotal != 16 || r.FinitePassed != 16 || r.ModelCalls != 0 {
		t.Fatal("declared input outcomes", receipt.Decision, r.FiniteTotal, r.FinitePassed, r.ModelCalls)
	}
}
