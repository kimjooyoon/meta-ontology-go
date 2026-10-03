package bodypathstream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func executionLine(t *testing.T, id string, cases string) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	var request Request
	if err := json.Unmarshal(requestLine(t, id, string(source)), &request); err != nil {
		t.Fatal(err)
	}
	request.ExecutionCases = json.RawMessage(cases)
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return append(raw, '\n')
}

func TestCommandExecutesBeforeEOFAndUsesNewCases(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	input, upstream := io.Pipe()
	output, downstream := io.Pipe()
	defer upstream.Close()
	defer output.Close()
	var diagnostics bytes.Buffer
	done := make(chan int, 1)
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	go func() {
		done <- RunCommand(ctx, "worker", []string{"--execute", "--go-bin", tool}, input, downstream, &diagnostics)
	}()
	decoder := json.NewDecoder(output)
	for i, suite := range []string{
		`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":2,"expected":15}]}`,
		`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":3,"expected":999}]}`,
	} {
		line := executionLine(t, "current", suite)
		go func() { _, _ = upstream.Write(line) }()
		var result Result
		if err := decoder.Decode(&result); err != nil {
			t.Fatal(err)
		}
		if result.Status != "completed" || result.Response == nil || result.Execution == nil {
			t.Fatal(result.Error)
		}
		o := result.Execution.Observation
		if !o.RuntimeReplayed || len(o.Runs) != 2 || len(o.Cases) != 1 || o.Artifact == nil || o.Artifact.Reused != (i == 1) || o.Build.Started != (i == 0) {
			t.Fatalf("lost immediate native execution: %+v", o)
		}
		if o.Cases[0].Actual != int64(15+5*i) || o.Cases[0].Passed != (i == 0) {
			t.Fatal("old cases reused", o.Cases)
		}
		parent, _ := json.Marshal(result.Response.Report.CompletenessReceipt)
		if !bytes.Equal(parent, result.Execution.ParentReceipt) {
			t.Fatal("generation parent changed")
		}
	}
	_ = upstream.Close()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatal(code, diagnostics.String())
		}
	case <-ctx.Done():
		t.Fatal("execution stream deadlocked")
	}
}

func TestStreamExecutionCasesRequireOptInAndValidSuite(t *testing.T) {
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	for _, suite := range []string{`null`, `{}`, `{"schema":"gooo/body-runtime-cases/v1","cases":[]}`} {
		line := executionLine(t, "invalid", suite)
		result := evaluateWithExecution(context.Background(), g, record{sequence: 1, raw: bytes.TrimSpace(line)}, &executionSettings{})
		if result.Status != "rejected" || result.Response != nil || result.Execution != nil {
			t.Fatal("invalid suite reached generation")
		}
	}
	line := executionLine(t, "disabled", `{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":2,"expected":15}]}`)
	result := evaluate(context.Background(), g, record{sequence: 1, raw: bytes.TrimSpace(line)})
	if result.Status != "rejected" || result.Response != nil {
		t.Fatal("execution enabled implicitly")
	}
}

func TestStreamRetainsGenerationWhenNativeSetupFails(t *testing.T) {
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	line := executionLine(t, "missing-tool", `{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":2,"expected":15}]}`)
	if err := RunWithExecution(context.Background(), g, io.NopCloser(bytes.NewReader(line)), asWriteCloser(&out), 2, filepath.Join(t.TempDir(), "missing")); err != nil {
		t.Fatal(err)
	}
	results := decodeResults(t, out.Bytes())
	if len(results) != 1 || results[0].Status != "execution_failed" || results[0].Response == nil || results[0].Execution == nil || results[0].Execution.CompletenessReceipt == nil {
		t.Fatal("lost failure observation", results)
	}
}

func TestStreamParallelConstructionKeepsCurrentExecutionInputs(t *testing.T) {
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	var input, output bytes.Buffer
	for i := range 4 {
		input.Write(executionLine(t, fmt.Sprint(i), fmt.Sprintf(`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":%d,"expected":%d}]}`, i, 5*i+5)))
	}
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	if err := RunWithExecution(context.Background(), g, io.NopCloser(&input), asWriteCloser(&output), 4, tool); err != nil {
		t.Fatal(err)
	}
	results, builds := decodeResults(t, output.Bytes()), 0
	if len(results) != 4 {
		t.Fatal("lost concurrent result")
	}
	for _, r := range results {
		if r.Status != "completed" || r.Execution == nil || len(r.Execution.Observation.Cases) != 1 || !r.Execution.Observation.Cases[0].Passed {
			t.Fatal("input mixed across workers", r.Error)
		}
		if r.Execution.Observation.Build.Completed {
			builds++
		}
	}
	if builds != 1 {
		t.Fatal("same artifact rebuilt concurrently", builds)
	}
}
