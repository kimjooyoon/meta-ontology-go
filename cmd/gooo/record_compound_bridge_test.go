package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoooReadsCompoundFieldOrigins(t *testing.T) {
	raw := readCompoundRuntimeFixture(t)
	path := filepath.Join(t.TempDir(), "recorded-runtime.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	args := []string{"package", "execute", "--json", "--origin-receipt", path,
		"--origin-shape", "field-local-columns", "--origin-offset", "9", "--origin-limit", "4",
		"../../examples/local-field-bridge/trace.workspace.json"}
	code := run(args, &out, &diagnostics)
	if dir := os.Getenv("GOOO_COMPOUND_READER_OUTPUT"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "bridge-receipt.json"), out.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err)
	}
	if receipt.Decision != "OBSERVED" || receipt.Result.Runtime.FiniteTotal != 0 || receipt.Result.Runtime.ModelCalls != 0 ||
		receipt.OriginInput.ProducerExecutions != 0 || len(receipt.Result.Runtime.Traces) != 4 {
		t.Fatal("reader execution scope")
	}
	wanted := []string{"count;step", "step", "title", "untouched"}
	for i, trace := range receipt.Result.Runtime.Traces {
		var actual struct {
			Status, Roots string
			RootCount     int `json:"root_count"`
		}
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil {
			t.Fatal("reader output", i)
		}
		roots := ""
		for field := range strings.SplitSeq(wanted[i], ";") {
			roots += "urn:gooo:compound:summarize|input0|urn:gooo:compound:" + field + ";"
		}
		if actual.Status != "SCOPED_FIELD_LOCAL_ORIGINS" || actual.Roots != roots || actual.RootCount != len(strings.Split(wanted[i], ";")) {
			t.Fatal("Gooo compound lineage", i, actual, roots)
		}
	}
	t.Log("four new Gooo lineage comparisons match; prior count and local-carried step remain distinct; producer executions 0, model calls 0")
}

func readCompoundRuntimeFixture(t *testing.T) []byte {
	t.Helper()
	f, err := os.Open("../../examples/compound-updates/recorded-fields.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(io.LimitReader(z, 4<<20))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
