package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageReceiptTransportPreservesJSONBytes(t *testing.T) {
	value := json.RawMessage(`{"source":"기원 <origin>","value":9007199254740993,"observation":{"status":"PARTIAL","dropped":47}}`)
	var plain, compressed bytes.Buffer
	if err := writePackageReceiptJSON(&plain, value, ""); err != nil {
		t.Fatal(err)
	}
	if err := writePackageReceiptJSON(&compressed, value, "gzip"); err != nil {
		t.Fatal(err)
	}
	raw, err := decodePackageReceiptTransport(compressed.Bytes(), 4096)
	if err != nil || !bytes.Equal(raw, plain.Bytes()) || workspaceDigest(raw) != workspaceDigest(plain.Bytes()) {
		t.Fatal("transport changed JSON bytes or identity", err)
	}
	path := filepath.Join(t.TempDir(), "opaque-receipt")
	if err := os.WriteFile(path, compressed.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	read, err := readPackageReceipt(path)
	if err != nil || !bytes.Equal(read, raw) {
		t.Fatal("receipt reader inferred encoding from extension", err)
	}
	read, err = readPackageInputBytes(OSFileReader{}, path, true)
	if err != nil || !bytes.Equal(read, raw) {
		t.Fatal("origin/construction reader lost JSON bytes", err)
	}
}

func TestPackageReceiptTransportRejectsIncompleteStreams(t *testing.T) {
	var compressed bytes.Buffer
	if err := writePackageReceiptJSON(&compressed, map[string]string{"status": "PARTIAL"}, "gzip"); err != nil {
		t.Fatal(err)
	}
	raw := compressed.Bytes()
	corrupt := append([]byte{}, raw...)
	corrupt[len(corrupt)-5] ^= 1
	for name, data := range map[string][]byte{
		"truncated": raw[:len(raw)-1], "checksum": corrupt,
		"trailing": append(append([]byte{}, raw...), 'x'),
		"multiple": append(append([]byte{}, raw...), raw...),
		"empty":    {}, "bad-header": {0x1f, 0x8b, 0x00},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decodePackageReceiptTransport(data, 4096); err == nil {
				t.Fatal("invalid transport accepted")
			}
		})
	}
	var large bytes.Buffer
	z, _ := gzip.NewWriterLevel(&large, gzip.BestSpeed)
	_, _ = z.Write(bytes.Repeat([]byte("a"), 8192))
	_ = z.Close()
	if _, err := decodePackageReceiptTransport(large.Bytes(), 1024); err == nil {
		t.Fatal("expanded receipt exceeded its independent bound")
	}
	oversized := append([]byte{0x1f, 0x8b}, make([]byte, packageReceiptGzipMaxBytes)...)
	if _, err := decodePackageReceiptTransport(oversized, packageReceiptMaxBytes); err == nil {
		t.Fatal("compressed bound not enforced")
	}
}

func TestPackageReceiptTransportReportsWriterFailures(t *testing.T) {
	for _, compression := range []string{"", "gzip"} {
		if err := writePackageReceiptJSON(receiptBrokenWriter{}, map[string]int{"value": 1}, compression); err == nil {
			t.Fatal("output failure became success")
		}
	}
	w := &packageReceiptLimitWriter{output: io.Discard, limit: 4}
	if n, err := w.Write([]byte("abc")); err != nil || n != 3 {
		t.Fatal(n, err)
	}
	if n, err := w.Write([]byte("de")); err == nil || n != 0 || w.written != 3 {
		t.Fatal("partial budget overflow accepted", n, err)
	}
}

func TestPackageReceiptWriteFailureKeepsExecutionCause(t *testing.T) {
	for _, compression := range []string{"", "gzip"} {
		var stderr bytes.Buffer
		cause := errors.New("original execution failed")
		receipt := packageExecutionReceipt{Schema: "gooo/workspace-body-execution-receipt/v1", Decision: "FAIL_CLOSED", Error: cause.Error()}
		code := writePackageExecution(receipt, cause, true, compression, receiptBrokenWriter{}, &stderr)
		original := bytes.Index(stderr.Bytes(), []byte(cause.Error()))
		transport := bytes.Index(stderr.Bytes(), []byte("write receipt:"))
		if code != exitFailure || original < 0 || transport <= original {
			t.Fatal("publication failure replaced original failure", code, stderr.String())
		}
	}
}

func TestPackageSavedWriteFailureKeepsInputCause(t *testing.T) {
	for _, resume := range []bool{false, true} {
		var stderr bytes.Buffer
		path := filepath.Join(t.TempDir(), "missing-workspace")
		args := []string{"--json", "--receipt-compression", "gzip", "--receipt", "unused", "--cases", "unused", path}
		if resume {
			args = append(args, "--assembly-policy-workspace", "unused")
		}
		code := runPackageSaved(args, OSFileReader{}, receiptBrokenWriter{}, &stderr, resume)
		original := bytes.Index(stderr.Bytes(), []byte("execution:"))
		transport := bytes.Index(stderr.Bytes(), []byte("write receipt:"))
		if code != exitFailure || original < 0 || transport <= original || !bytes.Contains(stderr.Bytes(), []byte("missing-workspace")) {
			t.Fatal("saved operation replaced its original input failure", resume, code, stderr.String())
		}
	}
}

type receiptBrokenWriter struct{}

func (receiptBrokenWriter) Write([]byte) (int, error) { return 0, errors.New("writer unavailable") }

func TestPackageReceiptCompressionRequiresExplicitJSON(t *testing.T) {
	for _, operation := range []string{"execute", "replay", "resume"} {
		for _, options := range [][]string{{"--receipt-compression", "gzip"}, {"--json", "--receipt-compression", "zip"}, {"--json", "--receipt-compression", "gzip", "--receipt-compression", "gzip"}} {
			args := append([]string{"package", operation, "--cases", "absent", "workspace"}, options...)
			if operation != "execute" {
				args = append(args, "--receipt", "absent")
			}
			if operation == "resume" {
				args = append(args, "--assembly-policy-workspace", "absent")
			}
			var stdout, stderr bytes.Buffer
			if code := run(args, &stdout, &stderr); code != exitUsage || stdout.Len() != 0 {
				t.Fatal("invalid transport launched work", operation, code)
			}
		}
	}
}

func TestPackageReceiptTransportIsDiscoverableWithoutExecution(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := runVersion([]string{"--build", "--json"}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var report struct {
		Transport packageReceiptTransportProfile `json:"package_receipt_transport"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	want := packageReceiptTransportProfile{"gzip", true, 2097152, 33554432, "original_json_bytes"}
	if report.Transport != want || report.Transport != currentPackageReceiptTransport() {
		t.Fatal("transport discovery differs from enforced limits", report)
	}
	if versionBuildMetadata(nil, "unobserved").CompilerSourceSHA != "UNBOUND_LOCAL_SOURCE" {
		t.Fatal("transport profile manufactured build identity")
	}
}

func TestPackageCompressedReplayKeepsTransportFailureFirst(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "constant-helpers")
	var output, stderr, compressed bytes.Buffer
	if err := writePackageReceiptJSON(&compressed, map[string]string{"status": "PARTIAL"}, "gzip"); err != nil {
		t.Fatal(err)
	}
	raw := compressed.Bytes()
	raw[len(raw)-5] ^= 1
	path := filepath.Join(t.TempDir(), "broken.json.gz")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	code := run([]string{"package", "replay", "--json", "--receipt", path,
		"--cases", filepath.Join(root, "cases.json"), "--go", filepath.Join(t.TempDir(), "must-not-run"),
		filepath.Join(root, "gooo.workspace.json")}, &output, &stderr)
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if code != exitFailure || receipt.Decision != "FAIL_CLOSED" || receipt.Result != nil ||
		receipt.ReplayedFrom != "" || receipt.CasesDigest != "" || !bytes.Contains([]byte(receipt.Error), []byte("decode compressed package receipt")) {
		t.Fatal("transport failure lost priority or admitted execution", code, receipt)
	}
}

func TestPackageCompressedReceiptExecutesAndReplays(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "constant-helpers")
	args := []string{"package", "execute", "--json", "--receipt-compression", "gzip", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "gooo.workspace.json")}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	plain, err := decodePackageReceiptTransport(stdout.Bytes(), packageReceiptMaxBytes)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "execution.json.gz")
	if err := os.WriteFile(path, stdout.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	args[1] = "replay"
	args = append(args, "--receipt", path)
	stdout.Reset()
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	replayed, err := decodePackageReceiptTransport(stdout.Bytes(), packageReceiptMaxBytes)
	if err != nil {
		t.Fatal(err)
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(replayed, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.ReplayedFrom != workspaceDigest(plain) || receipt.Result == nil || receipt.Decision != "PASS" {
		t.Fatal("compressed parent identity differs")
	}
	runtime := receipt.Result.Runtime
	if !runtime.RuntimeReplayed || runtime.ModelCalls != 0 || runtime.FiniteTotal == 0 || runtime.FinitePassed != runtime.FiniteTotal {
		t.Fatal("compressed replay lost finite values")
	}
}
