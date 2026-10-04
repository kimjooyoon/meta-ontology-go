package bodypathstream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type fakeGenerator struct {
	generate func(context.Context, string) (bodycodegen.Result, error)
}

func (g fakeGenerator) Generate(ctx context.Context, _ string, source []byte, _ string,
	_ pathplan.Document, _ bodycodegen.TypedPathOptions) (bodycodegen.Result, error) {
	return g.generate(ctx, string(source))
}

func successfulModel() fakeGenerator {
	return fakeGenerator{generate: func(_ context.Context, source string) (bodycodegen.Result, error) {
		return bodycodegen.Result{Source: source}, nil
	}}
}

func requestLine(t testing.TB, id, source string) []byte {
	t.Helper()
	// Transport tests replace generation, while source parsing remains real.
	if source == "queued" || source == "quick" {
		fixture, err := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
		if err != nil {
			t.Fatal(err)
		}
		source = string(fixture) + "\n// " + source
	}
	raw, err := os.ReadFile("../../examples/body-codegen/typed-path-compound-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	request := Request{Schema: RequestSchema, CorrelationID: id, Source: source, Activity: "Combined", Document: raw}
	raw, err = json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func decodeResults(t *testing.T, raw []byte) []Result {
	t.Helper()
	var results []Result
	for line := range bytes.SplitSeq(bytes.TrimSpace(raw), []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var result Result
		if err := json.Unmarshal(line, &result); err != nil {
			t.Fatal(err)
		}
		results = append(results, result)
	}
	return results
}

func TestNativeStreamRejectsBadRecordThenConstructsFreshBody(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	valid := requestLine(t, "valid-ko-en", string(source))
	unknown := bytes.Replace(valid, []byte(`"activity":`), []byte(`"extra":1,"activity":`), 1)
	duplicate := bytes.Replace(valid, []byte(`"schema":`), []byte(`"SCHEMA":"x","schema":`), 1)
	badBody := requestLine(t, "bad-body", strings.Replace(string(source), "input + 2", "input + 99", 1))
	options := bytes.Replace(valid, []byte(`"activity":`), []byte(`"options":{"feedback_unfixed":true},"activity":`), 1)
	raw := bytes.Join([][]byte{unknown, duplicate, badBody, options, []byte("{"),
		bytes.Repeat([]byte{'x'}, MaxRecordBytes+1), valid}, []byte{'\n'})
	var out bytes.Buffer
	if err = Run(context.Background(), g, io.NopCloser(bytes.NewReader(raw)), asWriteCloser(&out), 3); err != nil {
		t.Fatal(err)
	}
	results := decodeResults(t, out.Bytes())
	sort.Slice(results, func(i, j int) bool { return results[i].Sequence < results[j].Sequence })
	if len(results) != 7 {
		t.Fatal("lost record")
	}
	for _, v := range results[:6] {
		if v.Status != "rejected" || v.Response != nil {
			t.Fatal("invalid record accepted")
		}
	}
	if results[2].Failure == nil || results[2].Failure.Search.Selection.ModelCalls != 0 {
		t.Fatal("lost source failure")
	}
	v := results[6]
	if v.Status != "completed" || v.Response == nil || !v.Response.Report.TypecheckPassed ||
		v.Response.Report.BodyPaths.FunctionalCompleteness != 100 || v.Response.Report.RepositoryWrites != 0 {
		t.Fatal("next native construction inherited failure")
	}
}

func TestNativeStreamBoundsActiveRequestsAndCancelsGeneration(t *testing.T) {
	var input bytes.Buffer
	for i := range 100 {
		input.Write(requestLine(t, fmt.Sprint(i), "queued"))
		input.WriteByte('\n')
	}
	reader := &countingReadCloser{reader: bytes.NewReader(input.Bytes())}
	var active, peak atomic.Int32
	entered := make(chan struct{}, 2)
	g := fakeGenerator{generate: func(ctx context.Context, _ string) (bodycodegen.Result, error) {
		n := active.Add(1)
		for old := peak.Load(); n > old && !peak.CompareAndSwap(old, n); old = peak.Load() {
		}
		defer active.Add(-1)
		entered <- struct{}{}
		<-ctx.Done()
		return bodycodegen.Result{}, ctx.Err()
	}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Run(ctx, g, reader, asWriteCloser(io.Discard), 2) }()
	for range 2 {
		select {
		case <-entered:
		case <-time.After(time.Second):
			t.Fatal("workers did not start")
		}
	}
	if reader.consumed.Load() >= int64(input.Len()) || peak.Load() > 2 {
		t.Fatal("unbounded producer or workers")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) || active.Load() != 0 {
			t.Fatal("generation outlived stream")
		}
	case <-time.After(time.Second):
		t.Fatal("generation cancellation deadlocked")
	}
}

func TestCanonicalStreamKeysAndDepth(t *testing.T) {
	for _, raw := range []string{`{"schema":"ok","options":{"ci":{"status":"UNKNOWN"}}}`,
		`{"rows":[{"intent":"한국어 / English"}],"max_attempts":4}`} {
		if err := canonicalKeys([]byte(raw)); err != nil {
			t.Fatal(err)
		}
	}
	for _, raw := range []string{`{"schema":"ok","SCHEMA":"overwrite"}`, `{"options":{"CI":null}}`,
		strings.Repeat("[", 65) + "0" + strings.Repeat("]", 65)} {
		if err := canonicalKeys([]byte(raw)); err == nil {
			t.Fatal("case alias or depth accepted")
		}
	}
}
func TestRunWritesResultsBeforeInputEOF(t *testing.T) {
	reader := newGatedReader(append(requestLine(t, "first", "quick"), '\n'))
	writer := &observingWriter{written: make(chan struct{}, 1)}
	done := make(chan error, 1)
	go func() {
		done <- Run(context.Background(), successfulModel(), reader, writer, 2)
	}()

	select {
	case <-writer.written:
		// The input reader is still blocked waiting for the next line here.
	case <-time.After(time.Second):
		_ = reader.Close()
		t.Fatal("first result waited for input EOF")
	}
	_ = reader.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("stream did not finish after input close")
	}
	results := decodeResults(t, writer.Bytes())
	if len(results) != 1 || results[0].CorrelationID != "first" {
		t.Fatalf("unexpected output: %+v", results)
	}
}

func TestRunCancellationUnblocksAWaitingReader(t *testing.T) {
	reader, writer := io.Pipe()
	defer writer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Run(ctx, successfulModel(), reader, asWriteCloser(io.Discard), 1) }()
	time.Sleep(30 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not close the blocked reader")
	}
}

func TestRunOutputErrorCancelsAndClosesInput(t *testing.T) {
	reader, writer := io.Pipe()
	defer writer.Close()
	input := append(requestLine(t, "first", "quick"), '\n')
	writeDone := make(chan error, 1)
	go func() { _, err := writer.Write(input); writeDone <- err }()
	done := make(chan error, 1)
	go func() { done <- Run(context.Background(), successfulModel(), reader, failingWriter{}, 1) }()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "writer failed") {
			t.Fatalf("Run error = %v, want writer failure", err)
		}
	case <-time.After(time.Second):
		t.Fatal("writer failure deadlocked the producer")
	}
	select {
	case <-writeDone:
	case <-time.After(time.Second):
		t.Fatal("closing input did not unblock the upstream writer")
	}
}

func TestRunCancellationClosesBlockedOutputWriter(t *testing.T) {
	inputBytes := append(requestLine(t, "cancel-write", "quick"), '\n')
	input := &closeCountingReader{ReadCloser: io.NopCloser(bytes.NewReader(inputBytes))}
	output := newBlockingWriter()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Run(ctx, successfulModel(), input, output, 1) }()
	select {
	case <-output.entered:
	case <-time.After(time.Second):
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
		}
		t.Fatal("stream did not enter blocked output write")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run error = %v, want context cancellation", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not close blocked output writer")
	}
	if input.closeCount.Load() != 1 || output.closeCount.Load() != 1 {
		t.Fatalf("close counts: input=%d output=%d, want exactly one each", input.closeCount.Load(), output.closeCount.Load())
	}
}

func TestRunCancellationReleasesOutputBackpressureAndReader(t *testing.T) {
	reader, upstream := io.Pipe()
	input := &closeCountingReader{ReadCloser: reader}
	output := newBlockingWriter()
	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- Run(ctx, successfulModel(), input, output, 2) }()

	line := append(requestLine(t, "backpressure", "queued"), '\n')
	feedDone := make(chan error, 1)
	go func() {
		for range 4096 {
			if _, err := upstream.Write(line); err != nil {
				feedDone <- err
				return
			}
		}
		feedDone <- upstream.Close()
	}()
	select {
	case <-output.entered:
	case <-time.After(time.Second):
		cancel()
		_ = upstream.Close()
		select {
		case <-runDone:
		case <-time.After(time.Second):
		}
		select {
		case <-feedDone:
		case <-time.After(time.Second):
		}
		t.Fatal("stream did not reach its blocked result write")
	}
	select {
	case err := <-feedDone:
		cancel()
		select {
		case <-runDone:
		case <-time.After(time.Second):
		}
		t.Fatalf("upstream feed escaped bounded backpressure before cancellation: %v", err)
	case <-time.After(40 * time.Millisecond):
	}
	cancel()
	select {
	case err := <-runDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run error = %v, want context cancellation", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not release blocked result output")
	}
	select {
	case <-feedDone:
	case <-time.After(time.Second):
		t.Fatal("closing the input reader did not release the upstream writer")
	}
	_ = upstream.Close()
	if input.closeCount.Load() != 1 || output.closeCount.Load() != 1 {
		t.Fatalf("close counts: input=%d output=%d, want exactly one each", input.closeCount.Load(), output.closeCount.Load())
	}
}

func TestRunPreservesCompletedResultsWhenInputReadFails(t *testing.T) {
	reader := &readFailureReader{first: append(requestLine(t, "before-error", "quick"), '\n')}
	var output bytes.Buffer
	err := Run(context.Background(), successfulModel(), reader, asWriteCloser(&output), 1)
	if err == nil || !strings.Contains(err.Error(), "upstream read failed") {
		t.Fatalf("Run error = %v, want upstream read failure", err)
	}
	results := decodeResults(t, output.Bytes())
	if len(results) != 1 || results[0].Sequence != 1 || results[0].CorrelationID != "before-error" || results[0].Status != "completed" {
		t.Fatalf("completed result was lost when a later read failed: %+v", results)
	}
}

type observingWriter struct {
	mu      sync.Mutex
	buffer  bytes.Buffer
	written chan struct{}
}

func (writer *observingWriter) Close() error { return nil }

func (writer *observingWriter) Write(value []byte) (int, error) {
	writer.mu.Lock()
	n, err := writer.buffer.Write(value)
	writer.mu.Unlock()
	select {
	case writer.written <- struct{}{}:
	default:
	}
	return n, err
}

func (writer *observingWriter) Bytes() []byte {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	return append([]byte(nil), writer.buffer.Bytes()...)
}

type gatedReader struct {
	first   []byte
	read    bool
	release chan struct{}
	once    sync.Once
}

func newGatedReader(first []byte) *gatedReader {
	return &gatedReader{first: first, release: make(chan struct{})}
}

func (reader *gatedReader) Read(buffer []byte) (int, error) {
	if !reader.read {
		reader.read = true
		return copy(buffer, reader.first), nil
	}
	<-reader.release
	return 0, io.EOF
}

func (reader *gatedReader) Close() error {
	reader.once.Do(func() { close(reader.release) })
	return nil
}

type countingReadCloser struct {
	reader   *bytes.Reader
	consumed atomic.Int64
}

func (reader *countingReadCloser) Read(buffer []byte) (int, error) {
	n, err := reader.reader.Read(buffer)
	reader.consumed.Add(int64(n))
	return n, err
}

func (reader *countingReadCloser) Close() error { return nil }

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("writer failed") }
func (failingWriter) Close() error              { return nil }

type closeCountingReader struct {
	io.ReadCloser
	closeCount atomic.Int32
}

func (reader *closeCountingReader) Close() error {
	reader.closeCount.Add(1)
	return reader.ReadCloser.Close()
}

type closeCountingWriter struct {
	io.Writer
	closeCount atomic.Int32
}

func (writer *closeCountingWriter) Close() error {
	writer.closeCount.Add(1)
	return nil
}

type blockingWriter struct {
	entered    chan struct{}
	closed     chan struct{}
	enterOnce  sync.Once
	closeOnce  sync.Once
	closeCount atomic.Int32
}

func newBlockingWriter() *blockingWriter {
	return &blockingWriter{entered: make(chan struct{}), closed: make(chan struct{})}
}

func (writer *blockingWriter) Write([]byte) (int, error) {
	writer.enterOnce.Do(func() { close(writer.entered) })
	<-writer.closed
	return 0, errors.New("write interrupted by close")
}

func (writer *blockingWriter) Close() error {
	writer.closeOnce.Do(func() {
		writer.closeCount.Add(1)
		close(writer.closed)
	})
	return nil
}

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

func asWriteCloser(writer io.Writer) io.WriteCloser { return nopWriteCloser{Writer: writer} }

type readFailureReader struct {
	first []byte
	read  bool
}

func (reader *readFailureReader) Read(buffer []byte) (int, error) {
	if !reader.read {
		reader.read = true
		return copy(buffer, reader.first), nil
	}
	return 0, errors.New("upstream read failed")
}

func (*readFailureReader) Close() error { return nil }
