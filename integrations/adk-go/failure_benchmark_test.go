package adkcheck

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func benchmarkNoCompilerFixture(b *testing.B) (string, string) {
	b.Helper()
	root, err := filepath.EvalSymlinks(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source.gooo"), []byte("package p\nnamespace p\n"), 0600); err != nil {
		b.Fatal(err)
	}
	binary := filepath.Join(root, "fixture.sh")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nprintf '{\"status\":\"ok\"}'\n"), 0700); err != nil {
		b.Fatal(err)
	}
	return binary, root
}

// These two benchmarks use the same fixed shell fixture. They include process
// startup and are observations of the bridge boundary, not Gooo performance.
func BenchmarkBridgeNoCompilerProcess(b *testing.B) {
	binary, root := benchmarkNoCompilerFixture(b)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		result, err := check(context.Background(), binary, root, time.Second, CheckArgs{Path: "source.gooo"})
		if err != nil || result.ExitCode != 0 {
			b.Fatalf("bridge fixture: result=%#v err=%v", result, err)
		}
	}
}

func BenchmarkDirectNoCompilerProcess(b *testing.B) {
	binary, root := benchmarkNoCompilerFixture(b)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		cmd := exec.Command(binary, "check", "--semantic", "--json", "source.gooo")
		cmd.Dir = root
		output, err := cmd.Output()
		if err != nil || !json.Valid(output) {
			b.Fatalf("direct fixture: output=%q err=%v", output, err)
		}
	}
}

func BenchmarkBridgeRejectedQuery(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_, err := discover(context.Background(), "unused", "unused", time.Second, DiscoverArgs{Query: ""})
		if err == nil {
			b.Fatal("invalid query was accepted")
		}
	}
}
