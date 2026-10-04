// Package bodypathstream runs bounded native Gooo construction requests over a
// retained optional model. Mutable construction state belongs to each request.
package bodypathstream

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodytiming"
)

const (
	RequestSchema  = "gooo/native-body-stream-request/v1"
	ResultSchema   = "gooo/native-body-stream-result/v1"
	MaxRecordBytes = 1024 * 1024
	MaxWorkers     = 8
	readerBuffer   = 4 * 1024
)

// Generator is implemented by bodycodegen.TypedPathGenerator. This also makes
// queue and cancellation behavior testable without loading extra models.
type Generator interface {
	Generate(context.Context, string, []byte, string, pathplan.Document, bodycodegen.TypedPathOptions) (bodycodegen.Result, error)
}

type Request struct {
	Schema         string                       `json:"schema"`
	CorrelationID  string                       `json:"correlation_id"`
	Source         string                       `json:"source"`
	Activity       string                       `json:"activity"`
	Document       json.RawMessage              `json:"document,omitempty"`
	Options        bodycodegen.TypedPathOptions `json:"options"`
	ExecutionCases json.RawMessage              `json:"execution_cases,omitempty"`
}

type Result struct {
	Schema        string                       `json:"schema"`
	Sequence      uint64                       `json:"sequence"`
	CorrelationID string                       `json:"correlation_id,omitempty"`
	Status        string                       `json:"status"`
	Response      *bodycodegen.Result          `json:"response,omitempty"`
	Failure       *bodycodegen.BodyPathReceipt `json:"failure_receipt,omitempty"`
	Error         string                       `json:"error,omitempty"`
	Execution     *bodyexecution.Result        `json:"execution,omitempty"`
}

type record struct {
	sequence uint64
	raw      []byte
	tooLarge bool
}

// Run reads one JSON object per line and writes one result line per input
// record. The model is already loaded and is shared read-only. Exactly
// workers worker goroutines are created; both channels are bounded by workers.
// On cancellation it closes input and output to interrupt pending I/O. Both
// Close methods must promptly unblock a concurrent Read or Write.
func Run(parent context.Context, model Generator, input io.ReadCloser, output io.WriteCloser, workers int) error {
	return run(parent, model, input, output, workers, nil)
}

type executionSettings struct {
	owner    *bodyexecution.Executor
	goBinary string
}

// RunWithExecution constructs each request, then immediately observes its native
// behavior. One owned artifact slot is shared by the bounded construction workers.
// Each request supplies current expectations; all source checks and runs are fresh.
func RunWithExecution(parent context.Context, model Generator, input io.ReadCloser, output io.WriteCloser,
	workers int, goBinary string) (err error) {
	owner := bodyexecution.NewExecutor()
	defer func() { err = errors.Join(err, owner.Close()) }()
	return run(parent, model, input, output, workers, &executionSettings{owner, goBinary})
}

func run(parent context.Context, model Generator, input io.ReadCloser, output io.WriteCloser,
	workers int, execution *executionSettings) error {
	if parent == nil {
		return errors.New("context is required")
	}
	if model == nil {
		return errors.New("generator is required")
	}
	if input == nil {
		return errors.New("input is required")
	}
	if output == nil {
		return errors.New("output is required")
	}
	if workers < 1 || workers > MaxWorkers {
		return fmt.Errorf("workers must be between 1 and %d", MaxWorkers)
	}

	ctx, cancel := context.WithCancel(parent)
	closeIOOnce := sync.Once{}
	closeIO := func() {
		closeIOOnce.Do(func() {
			_ = output.Close()
			_ = input.Close()
		})
	}
	watcherStop := make(chan struct{})
	watcherDone := make(chan struct{})
	var lifecycleMu sync.Mutex
	normalCompletion := false
	go func() {
		defer close(watcherDone)
		select {
		case <-ctx.Done():
			lifecycleMu.Lock()
			shouldClose := !normalCompletion
			lifecycleMu.Unlock()
			if shouldClose {
				closeIO()
			}
		case <-watcherStop:
		}
	}()
	var watcherStopOnce sync.Once
	stopWatcher := func() {
		watcherStopOnce.Do(func() { close(watcherStop) })
		<-watcherDone
	}
	defer func() {
		stopWatcher()
		cancel()
	}()
	jobs := make(chan record, workers)
	results := make(chan Result, workers)
	producerDone := make(chan error, 1)

	var workerGroup sync.WaitGroup
	workerGroup.Add(workers)
	for range workers {
		go func() {
			defer workerGroup.Done()
			worker(ctx, model, jobs, results, execution)
		}()
	}
	go func() {
		workerGroup.Wait()
		close(results)
	}()
	go func() {
		defer close(jobs)
		producerDone <- produce(ctx, input, jobs)
	}()

	encoder := json.NewEncoder(output)
	for {
		select {
		case <-ctx.Done():
			closeIO()
			<-watcherDone
			drain(results)
			<-producerDone
			return ctx.Err()
		case result, ok := <-results:
			if !ok {
				producerErr := <-producerDone
				lifecycleMu.Lock()
				normalCompletion = true
				lifecycleMu.Unlock()
				stopWatcher()
				if err := parent.Err(); err != nil {
					return err
				}
				return producerErr
			}
			if err := encoder.Encode(result); err != nil {
				cancel()
				closeIO()
				<-watcherDone
				drain(results)
				<-producerDone
				if parentErr := parent.Err(); parentErr != nil {
					return parentErr
				}
				return fmt.Errorf("write result: %w", err)
			}
		}
	}
}

func produce(ctx context.Context, input io.Reader, jobs chan<- record) error {
	reader := bufio.NewReaderSize(input, readerBuffer)
	var sequence uint64
	for {
		raw, tooLarge, err := readRecord(reader, MaxRecordBytes)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("read input: %w", err)
		}
		sequence++
		item := record{sequence: sequence, raw: raw, tooLarge: tooLarge}
		select {
		case <-ctx.Done():
			return nil
		case jobs <- item:
		}
	}
}

func readRecord(reader *bufio.Reader, limit int) ([]byte, bool, error) {
	line := make([]byte, 0, min(limit, readerBuffer))
	seen := false
	tooLarge := false
	for {
		fragment, err := reader.ReadSlice('\n')
		if len(fragment) != 0 {
			seen = true
		}
		newline := len(fragment) != 0 && fragment[len(fragment)-1] == '\n'
		if newline {
			fragment = fragment[:len(fragment)-1]
		}
		if !tooLarge {
			if len(fragment) > limit-len(line) {
				tooLarge = true
				line = nil
			} else {
				line = append(line, fragment...)
			}
		}
		if newline {
			return trimCR(line), tooLarge, nil
		}
		if err == bufio.ErrBufferFull {
			continue
		}
		if err == io.EOF {
			if !seen {
				return nil, false, io.EOF
			}
			return trimCR(line), tooLarge, nil
		}
		if err != nil {
			return nil, false, err
		}
	}
}

func trimCR(line []byte) []byte {
	if len(line) > 0 && line[len(line)-1] == '\r' {
		return line[:len(line)-1]
	}
	return line
}

func worker(ctx context.Context, model Generator, jobs <-chan record, results chan<- Result, execution *executionSettings) {
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-jobs:
			if !ok {
				return
			}
			if ctx.Err() != nil {
				return
			}
			result := evaluateWithExecution(ctx, model, item, execution)
			select {
			case <-ctx.Done():
				return
			case results <- result:
			}
		}
	}
}

func evaluate(ctx context.Context, model Generator, item record) Result {
	return evaluateWithExecution(ctx, model, item, nil)
}

func evaluateWithExecution(ctx context.Context, model Generator, item record, execution *executionSettings) Result {
	phase := bodytiming.Start(ctx, "request_decode")
	result := Result{Schema: ResultSchema, Sequence: item.sequence, Status: "rejected"}
	reject := func(err error) Result {
		phase.End(false)
		result.Error = err.Error()
		return result
	}
	request, decodeErr := decodeStreamRequest(item)
	if validCorrelationID(request.CorrelationID) {
		result.CorrelationID = request.CorrelationID
	}
	if decodeErr != nil {
		return reject(fmt.Errorf("decode request: %w", decodeErr))
	}
	result.CorrelationID = request.CorrelationID
	var cases []pathplan.TestCase
	if execution != nil {
		var err error
		cases, err = bodyexecution.DecodeCases(request.ExecutionCases)
		if err != nil {
			return reject(fmt.Errorf("execution_cases: %w", err))
		}
	} else if request.ExecutionCases != nil {
		return reject(errors.New("execution_cases requires --execute"))
	}
	phase.End(true)
	response, document, err := constructStreamResponse(ctx, model, request, execution != nil)
	if err != nil {
		if failure, ok := errors.AsType[*bodycodegen.BodyPathError](err); ok {
			result.Failure = failure.Receipt
		}
		return reject(fmt.Errorf("construction rejected: %w", err))
	}
	result.Status = "completed"
	result.Response = &response
	if execution != nil {
		phase = bodytiming.Start(ctx, "generation_parent_encode")
		parent, err := json.Marshal(response.Report.CompletenessReceipt)
		if err != nil {
			return reject(err)
		}
		phase.End(true)
		observed, err := execution.owner.Execute(ctx, "stream.gooo", []byte(request.Source),
			document, response, parent, cases, execution.goBinary)
		result.Execution = &observed
		if err != nil {
			result.Status, result.Error = "execution_failed", err.Error()
		}
	}
	return result
}

func validCorrelationID(value string) bool {
	if value == "" || len(value) > 128 || !utf8.ValidString(value) || strings.TrimSpace(value) == "" {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func drain(results <-chan Result) {
	for range results {
	}
}

// encoding/json matches struct fields without regard to case. The stream uses
// canonical ASCII keys throughout, so case aliases cannot overwrite a field.
func canonicalKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	type frame struct{ object, key bool }
	var stack []frame
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if delim, ok := token.(json.Delim); ok {
			if delim == '}' || delim == ']' {
				stack = stack[:len(stack)-1]
				continue
			}
			if len(stack) > 0 {
				stack[len(stack)-1].key = true
			}
			if len(stack) == 64 {
				return errors.New("request JSON depth exceeds 64")
			}
			stack = append(stack, frame{object: delim == '{', key: delim == '{'})
			continue
		}
		if len(stack) == 0 || !stack[len(stack)-1].object {
			continue
		}
		f := &stack[len(stack)-1]
		if f.key {
			key, ok := token.(string)
			if !ok || key == "" {
				return errors.New("invalid object key")
			}
			for _, c := range key {
				if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '_' {
					return errors.New("request keys must use canonical lowercase ASCII")
				}
			}
		}
		f.key = !f.key
	}
}
