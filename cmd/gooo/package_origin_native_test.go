package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type originConsumerCase struct {
	name, record, kind   string
	offset, limit, total int
	labels               string
	counts               []int
	partial              bool
}

func originReaderRoot(t *testing.T) string {
	t.Helper()
	if root := os.Getenv("GOOO_ORIGIN_INPUT_FIXTURES"); root != "" {
		return root
	}
	return filepath.Join("..", "..", "examples", "native-origin-reader")
}

func TestOriginConsumers9719(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	rows := []originConsumerCase{
		{"elements", "construction9533", "elements", 2, 7, 24, "NO_READ_IN_WINDOW ONE_READ NO_READ_IN_WINDOW NO_READ_IN_WINDOW NO_READ_IN_WINDOW NO_READ_IN_WINDOW MULTIPLE_READS", []int{0, 1, 0, 0, 0, 0, 2}, false},
		{"empty-lists", "construction9547", "constructions", 0, 32, 4, "EMPTY_LIST COMPLETED_LIST EMPTY_LIST COMPLETED_LIST", []int{0, 1, 0, 1}, false},
		{"reads", "construction9533", "reads", 9, 8, 24, "EXTERNAL_INPUT CONSTRUCTED_ELEMENT EXTERNAL_INPUT EXTERNAL_INPUT CONSTRUCTED_ELEMENT CONSTRUCTED_ELEMENT CONSTRUCTED_ELEMENT EXTERNAL_INPUT", []int{1, 1, 1, 1, 1, 1, 1, 1}, false},
		{"links", "construction9533", "links", 5, 4, 12, "FIRST_ELEMENT_LINK LATER_ELEMENT_LINK LATER_ELEMENT_LINK FIRST_ELEMENT_LINK", []int{1, 1, 1, 1}, false},
		{"interrupted", "construction9551", "elements", 0, 32, 5, "ONE_READ INTERRUPTED ONE_READ ONE_READ NO_READ_IN_WINDOW", []int{1, 1, 1, 1, 0}, true},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) { runOriginConsumer9719(t, row) })
	}
}

func runOriginConsumer9719(t *testing.T, row originConsumerCase) {
	t.Helper()
	root := originReaderRoot(t)
	path := filepath.Join(root, row.record+"-runtime.json")
	var out, stderr bytes.Buffer
	args := []string{"package", "execute", "--json", "--origin-receipt", path, "--origin-kind", row.kind, "--origin-offset", strconv.Itoa(row.offset), "--origin-limit", strconv.Itoa(row.limit), filepath.Join(root, row.kind, row.kind+".workspace.json")}
	if code := run(args, &out, &stderr); code != exitOK {
		t.Fatal(code, out.String(), stderr.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	assertOriginConsumer9719(t, receipt, row, path)
	if dir := os.Getenv("GOOO_ORIGIN_INPUT_RESULTS"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "consumer9719-"+row.name+".json"), out.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func assertOriginConsumer9719(t *testing.T, receipt packageExecutionReceipt, row originConsumerCase, path string) {
	t.Helper()
	origin := receipt.OriginInput
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if origin == nil || origin.TotalRows != row.total || origin.ReturnedRows != len(row.counts) || origin.ReceiptSHA256 != workspaceDigest(raw) || origin.ProducerExecutions != 0 || origin.ProducerModelCalls != 0 {
		t.Fatalf("origin page mismatch: %+v", origin)
	}
	if origin.Query.Offset != row.offset || origin.Query.Limit != row.limit || origin.Query.Kind != row.kind {
		t.Fatal("query was lost", origin.Query)
	}
	more := row.offset+len(row.counts) < row.total
	if origin.Truncated != more || more && (origin.NextOffset == nil || *origin.NextOffset != row.offset+len(row.counts)) || !more && origin.NextOffset != nil {
		t.Fatal("page boundary differs", origin)
	}
	if (origin.Observation.Status == "PARTIAL") != row.partial {
		t.Fatal("producer partial state was lost", origin.Observation)
	}
	if receipt.Decision != "OBSERVED" || receipt.Result == nil {
		t.Fatal("record-only inputs acquired expected values", receipt.Decision)
	}
	runtime := receipt.Result.Runtime
	if runtime.FiniteTotal != 0 || runtime.FinitePassed != 0 || runtime.ModelCalls != 0 || !runtime.RuntimeReplayed || len(runtime.Traces) != len(row.counts) {
		t.Fatal("consumer scope differs")
	}
	labels := strings.Fields(row.labels)
	for i, trace := range runtime.Traces {
		var actual struct {
			Category string `json:"category"`
			Count    int    `json:"count"`
		}
		if len(trace.Deliveries) != 1 {
			t.Fatal("unexpected delivery count")
		}
		if err := json.Unmarshal(trace.Deliveries[0].Actual, &actual); err != nil || actual.Category != labels[i] || actual.Count != row.counts[i] {
			t.Fatal(i, actual, err)
		}
	}
}

func TestOriginElementWindows9743(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	rows := []originConsumerCase{
		{"window9743", "construction9533", "elements", 6, 5, 24, "NO_READ_IN_WINDOW NO_READ_IN_WINDOW MULTIPLE_READS MULTIPLE_READS NO_READ_IN_WINDOW", []int{0, 0, 2, 2, 0}, false},
		{"fault-window9743", "construction9551", "elements", 1, 2, 5, "INTERRUPTED ONE_READ", []int{1, 1}, true},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) { runOriginConsumer9719(t, row) })
	}
}
