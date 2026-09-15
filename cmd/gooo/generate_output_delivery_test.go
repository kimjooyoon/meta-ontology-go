package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicdiscovery"
)

const generateDeliveryRuntimeSource = "package normaloutput\nnamespace normaloutput\nentity Item id \"gooo://normal-output/item\"\n"

type generateDeliveryWriter struct {
	bytes.Buffer
	calls  int
	failAt int
}

func (writer *generateDeliveryWriter) Write(data []byte) (int, error) {
	writer.calls++
	if writer.calls == writer.failAt {
		return 0, os.ErrClosed
	}
	return writer.Buffer.Write(data)
}

func TestGenerateOutputDelivery(t *testing.T) {
	cases := []struct {
		name     string
		jsonMode bool
		rejected bool
	}{
		{"human-denied", false, true},
		{"json-denied", true, true},
		{"human-success", false, false},
		{"json-success", true, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			runGenerateDeliveryPublicCase(t, test.jsonMode, test.rejected)
		})
	}
}

func runGenerateDeliveryPublicCase(t *testing.T, jsonMode, rejected bool) {
	t.Helper()
	sourcePath := filepath.Join(t.TempDir(), "normal.gooo")
	original := []byte(generateDeliveryRuntimeSource)
	if err := os.WriteFile(sourcePath, original, 0600); err != nil {
		t.Fatal(err)
	}
	outputDir := filepath.Join(t.TempDir(), "generated")
	args := []string{sourcePath, "--out", outputDir}
	if jsonMode {
		args = append(args, "--json")
	}
	var baseline, stderr bytes.Buffer
	if code := runGenerate(args, OSFileReader{}, SyntaxSourceParser{}, &baseline, &stderr); code != exitOK {
		t.Fatalf("normal generation baseline exit=%d stderr=%s", code, stderr.String())
	}
	expected := snapshotGenerateDeliveryFiles(t, outputDir)
	writer := &generateDeliveryWriter{}
	wantCode := exitOK
	if rejected {
		writer.failAt, wantCode = 1, exitFailure
	}
	stderr.Reset()
	code := runGenerate(args, OSFileReader{}, SyntaxSourceParser{}, writer, &stderr)
	actual := snapshotGenerateDeliveryFiles(t, outputDir)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatal("result delivery changed already-generated file bytes")
	}
	after, err := os.ReadFile(sourcePath)
	if err != nil || !bytes.Equal(original, after) {
		t.Fatalf("result delivery changed Gooo source: %v", err)
	}
	t.Logf("human-output runtime input: sha256:%x", sha256.Sum256(original))
	if code != wantCode || writer.calls != 1 {
		t.Fatalf("public generation delivery exit=%d want=%d writes=%d want=1 stderr=%s",
			code, wantCode, writer.calls, stderr.String())
	}
	if !rejected {
		requireGenerateDeliveryPublicOutput(t, outputDir, jsonMode, writer.Bytes())
	}
}

func snapshotGenerateDeliveryFiles(t *testing.T, directory string) map[string][]byte {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("generated files disappeared: %v", err)
	}
	names := []string{}
	files := map[string][]byte{}
	for _, entry := range entries {
		names = append(names, entry.Name())
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		files[entry.Name()] = data
	}
	want := []string{generatedFileName, generatedManifestFileName}
	sort.Strings(want)
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("generated files=%v want=%v", names, want)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), generatedFileName, files[generatedFileName], 0); err != nil {
		t.Fatalf("generated Go syntax is invalid: %v", err)
	}
	return files
}

func requireGenerateDeliveryPublicOutput(t *testing.T, directory string, jsonMode bool, got []byte) {
	t.Helper()
	if !jsonMode {
		want := fmt.Sprintf("generated: %s\n", filepath.Join(directory, generatedFileName))
		if string(got) != want {
			t.Fatalf("human summary=%q want=%q", got, want)
		}
		return
	}
	var report map[string]any
	if err := json.Unmarshal(got, &report); err != nil || report["command"] != "generate" || report["status"] != "ok" {
		t.Fatalf("normal JSON report differs: %s error=%v", got, err)
	}
}

func TestGenerateConditionalOutputDelivery(t *testing.T) {
	cases := []struct {
		name      string
		failAt    int
		discovery bool
		wantCode  int
		wantCalls int
	}{
		{"observation-denied", 2, true, exitFailure, 2},
		{"candidate-denied", 3, true, exitFailure, 3},
		{"all-messages-success", 0, true, exitOK, 3},
		{"discovery-absent-success", 0, false, exitOK, 1},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			var discovery *publicdiscovery.Result
			want := []string{fmt.Sprintf("generated: %s\n", filepath.Join(directory, generatedFileName))}
			if test.discovery {
				discovery = &publicdiscovery.Result{CandidatePath: filepath.Join(directory, "candidate.json")}
				discovery.Report.MachineReportPath = filepath.Join(directory, "observation.json")
				discovery.Report.Decision = "UNKNOWN"
				discovery.Report.CandidatesEmitted = 1
				want = append(want, fmt.Sprintf("observation: %s (%s)\n", discovery.Report.MachineReportPath, discovery.Report.Decision))
				want = append(want, fmt.Sprintf("candidate: %s\n", discovery.CandidatePath))
			}
			writer := &generateDeliveryWriter{failAt: test.failAt}
			code := reportGenerateSuccess(generateOptions{outputDir: directory}, generateInput{}, generateArtifacts{}, discovery, false, writer)
			delivered := len(want)
			if test.failAt > 0 {
				delivered = test.failAt - 1
			}
			if code != test.wantCode || writer.calls != test.wantCalls ||
				writer.String() != joinGenerateDeliveryMessages(want[:delivered]) {
				t.Fatalf("conditional delivery exit=%d want=%d calls=%d want=%d bytes=%q",
					code, test.wantCode, writer.calls, test.wantCalls, writer.String())
			}
		})
	}
}

func joinGenerateDeliveryMessages(messages []string) string {
	var result bytes.Buffer
	for _, message := range messages {
		result.WriteString(message)
	}
	return result.String()
}
