package bodypathstream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodytiming"
)

type fileRunRow struct {
	Sequence       uint64  `json:"sequence"`
	Status         string  `json:"status"`
	ResponseMS     float64 `json:"response_ms"`
	ResponseFile   string  `json:"response_file"`
	GenerationFile string  `json:"generation_file,omitempty"`
	RuntimeFile    string  `json:"runtime_file,omitempty"`
	NativeRuns     int     `json:"native_runs"`
	Passed         int     `json:"passed"`
	Total          int     `json:"total"`
	ArtifactReused bool    `json:"artifact_reused"`
}

// RunFilesCommand constructs and executes bounded repetitions from file inputs.
// It uses the stream's exact request evaluator, retained generator and executor;
// completed and failed observations are saved before the next construction.
func RunFilesCommand(ctx context.Context, name string, args []string, stdout, stderr io.Writer) int {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.SetOutput(stderr)
	source := f.String("source", "", "original Gooo source file")
	activity := f.String("activity", "", "source Integer -> Integer activity name")
	plan := f.String("path-plan", "", "full typed plan or source recipe JSON")
	cases := f.String("cases", "", "independent runtime expectations JSON")
	options := f.String("options", "", "optional stream options JSON, including CI context")
	model := f.String("model", "", "explicit local model.json; omit for deterministic construction")
	goBin := f.String("go-bin", "", "local Go 1.27.1 tool")
	out := f.String("out", "", "fresh output directory, or saved directory for --verify-timing")
	repeat := f.Int("repeat", 1, "bounded sequential requests (1..16)")
	timing := f.Bool("timing", false, "save bounded wall phases and original-file SHA256 bindings")
	verifyTiming := f.Bool("verify-timing", false, "check saved timing consistency and file bindings; requires only --out")
	f.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s --source file.gooo --activity name --path-plan recipe.json --cases cases.json --out fresh-directory [options]\n", name)
		f.PrintDefaults()
	}
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *verifyTiming {
		if f.NArg() != 0 || *out == "" || f.NFlag() != 2 {
			f.Usage()
			return 2
		}
		if err := verifyTimingDirectory(ctx, *out); err != nil {
			fmt.Fprintf(stderr, "%s: %v\n", name, err)
			return 1
		}
		fmt.Fprintln(stdout, "timing consistency and saved-file bindings: PASS")
		return 0
	}
	if f.NArg() != 0 || *source == "" || *activity == "" || *plan == "" || *cases == "" ||
		*out == "" || *repeat < 1 || *repeat > 16 {
		f.Usage()
		return 2
	}
	fail := func(err error) int {
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return 1
	}
	if ctx == nil {
		return fail(errors.New("context is required"))
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	inputs := make([][]byte, 4)
	for i, input := range []struct {
		path string
		max  int64
	}{{*source, 128 << 10}, {*plan, 256 << 10}, {*cases, 32 << 10}, {*options, 64 << 10}} {
		if i == 3 && input.path == "" {
			continue
		}
		b, err := readFileInput(input.path, input.max)
		if err != nil {
			return fail(err)
		}
		inputs[i] = b
	}
	var requestOptions bodycodegen.TypedPathOptions
	if inputs[3] != nil {
		if err := decodeFileOptions(inputs[3], &requestOptions); err != nil {
			return fail(fmt.Errorf("options: %w", err))
		}
	}
	request := Request{Schema: RequestSchema, CorrelationID: "run-1", Source: string(inputs[0]),
		Activity: *activity, Document: inputs[1], Options: requestOptions, ExecutionCases: inputs[2]}
	if _, err := json.Marshal(request); err != nil {
		return fail(fmt.Errorf("input JSON: %w", err))
	}
	generator, err := bodycodegen.NewTypedPathGenerator(*model)
	if err != nil {
		return fail(err)
	}
	if err := os.Mkdir(*out, 0755); err != nil {
		return fail(fmt.Errorf("create fresh output directory: %w", err))
	}
	for i, filename := range []string{"source.gooo", "recipe.json", "cases.json", "options.json"} {
		if inputs[i] != nil {
			if err := os.WriteFile(filepath.Join(*out, filename), inputs[i], 0644); err != nil {
				return fail(err)
			}
		}
	}
	if err := writeFileJSON(*out, "model-retention.json", generator.Info()); err != nil {
		return fail(err)
	}
	return runFileRequests(ctx, generator, request, *repeat, *goBin, *out, stdout, stderr, fail, *timing)
}

func runFileRequests(ctx context.Context, generator Generator, request Request, repeat int,
	goBin, out string, stdout, stderr io.Writer, fail func(error) int, timingFlags ...bool) (code int) {
	owner := bodyexecution.NewExecutor()
	defer func() {
		if err := owner.Close(); err != nil {
			code = fail(err)
		}
	}()
	// CLI stdout is a closeable file. Cancellation interrupts a blocked pipe;
	// ordinary completion preserves ownership of the caller's output.
	closed := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		for _, output := range []io.Writer{stdout, stderr} {
			if c, ok := output.(io.Closer); ok {
				_ = c.Close()
			}
		}
		close(closed)
	})
	defer func() {
		if !stop() {
			<-closed
		}
	}()
	rows := make([]fileRunRow, 0, repeat)
	timingRows := make([]fileTimingRow, 0, repeat)
	encoder := json.NewEncoder(stdout)
	for i := range repeat {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		request.CorrelationID = fmt.Sprintf("run-%d", i+1)
		raw, err := json.Marshal(request)
		if err != nil {
			return fail(err)
		}
		requestCtx := ctx
		var recorder *bodytiming.Recorder
		if len(timingFlags) > 0 && timingFlags[0] {
			recorder = bodytiming.NewRecorder()
			requestCtx = bodytiming.WithRecorder(ctx, recorder)
		}
		started := time.Now()
		result := evaluateWithExecution(requestCtx, generator, record{sequence: uint64(i + 1), raw: raw},
			&executionSettings{owner: owner, goBinary: goBin})
		responseNS := time.Since(started).Nanoseconds()
		row := fileRunRow{Sequence: result.Sequence, Status: result.Status,
			ResponseMS: float64(responseNS) / float64(time.Millisecond)}
		phase := bodytiming.Start(requestCtx, "artifact_save")
		err = saveFileResult(out, result, &row)
		if err == nil && recorder != nil {
			err = os.WriteFile(filepath.Join(out, fmt.Sprintf("run-%d-request.json", result.Sequence)), raw, 0644)
		}
		phase.End(err == nil)
		if err != nil {
			return fail(err)
		}
		if recorder != nil {
			timingRow, err := saveFileTiming(requestCtx, out, raw, result, responseNS, recorder)
			if err != nil {
				return fail(err)
			}
			timingRows = append(timingRows, timingRow)
			if err := writeFileJSON(out, "timing-summary.json", fileTimingSummary{
				Schema: "gooo/body-path-file-timing-summary/v1", Scope: timingScope, Rows: timingRows,
			}); err != nil {
				return fail(err)
			}
			fmt.Fprintf(stderr, "%s timing: generation %s, source replay %s, native %s, save %s\n",
				request.CorrelationID, phaseDuration(timingRow.PhaseMS, "generation"),
				phaseDuration(timingRow.PhaseMS, "source_replay"),
				phaseDuration(timingRow.PhaseMS, "native_run_1", "native_run_2"),
				phaseDuration(timingRow.PhaseMS, "artifact_save"))
		}
		rows = append(rows, row)
		if err := writeFileJSON(out, "summary.json", struct {
			Schema string       `json:"schema"`
			Rows   []fileRunRow `json:"rows"`
		}{"gooo/body-path-file-run/v1", rows}); err != nil {
			return fail(err)
		}
		if err := encoder.Encode(result); err != nil {
			return fail(err)
		}
		fmt.Fprintf(stderr, "%s: %s, finite expectations %s, artifact reused=%t, response %.3fms\n",
			request.CorrelationID, row.Status, finiteExpectationDisplay(result, row.Passed), row.ArtifactReused, row.ResponseMS)
		if result.Status != "completed" {
			fmt.Fprintf(stderr, "%s: %s\n", request.CorrelationID, result.Error)
			code = 1
		}
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	return code
}

// Case observations and replay are separate facts. A first run can supply all
// finite outputs even when the second run fails; the display preserves both.
func finiteExpectationDisplay(result Result, passed int) string {
	if result.Execution == nil {
		return "unobserved"
	}
	o := result.Execution.Observation
	observed := len(o.Cases)
	if observed == 0 {
		if o.DeclaredCases > 0 {
			return fmt.Sprintf("unobserved (%d declared)", o.DeclaredCases)
		}
		return "unobserved"
	}
	if o.RuntimeReplayed && observed == o.DeclaredCases {
		return fmt.Sprintf("%d/%d", passed, observed)
	}
	detail := ""
	if observed != o.DeclaredCases {
		detail = fmt.Sprintf("%d declared", o.DeclaredCases)
	}
	if !o.RuntimeReplayed {
		if detail != "" {
			detail += "; "
		}
		detail += "replay incomplete"
	}
	return fmt.Sprintf("observed %d/%d (%s)", passed, observed, detail)
}

// A missing phase is unobserved, even when another member of the group ran.
func phaseDuration(phases map[string]float64, names ...string) string {
	total := 0.0
	for _, name := range names {
		value, observed := phases[name]
		if !observed {
			return "unobserved"
		}
		total += value
	}
	return fmt.Sprintf("%.3fms", total)
}

func saveFileResult(out string, result Result, row *fileRunRow) error {
	stem := fmt.Sprintf("run-%d", result.Sequence)
	row.ResponseFile = stem + "-response.json"
	if err := writeFileJSON(out, row.ResponseFile, result); err != nil {
		return err
	}
	if result.Response != nil {
		row.GenerationFile = stem + "-generation.json"
		if err := writeFileJSON(out, row.GenerationFile, result.Response); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, stem+"-generated.go"), []byte(result.Response.Source), 0644); err != nil {
			return err
		}
	}
	if result.Execution != nil {
		row.RuntimeFile = stem + "-runtime.json"
		if err := writeFileJSON(out, row.RuntimeFile, result.Execution); err != nil {
			return err
		}
		o := result.Execution.Observation
		row.Total = o.DeclaredCases
		for _, run := range o.Runs {
			if run.Started {
				row.NativeRuns++
			}
		}
		for _, c := range o.Cases {
			if c.Passed {
				row.Passed++
			}
		}
		row.ArtifactReused = o.Artifact != nil && o.Artifact.Reused
	}
	return nil
}

func writeFileJSON(out, filename string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, filename), append(b, '\n'), 0644)
}

func readFileInput(filename string, limit int64) ([]byte, error) {
	before, err := os.Stat(filename)
	if err != nil || !before.Mode().IsRegular() {
		return nil, fmt.Errorf("%s: input must be a regular file", filename)
	}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > limit {
		return nil, fmt.Errorf("%s: input must be nonempty and within %d bytes", filename, limit)
	}
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || len(b) == 0 || int64(len(b)) > limit || !utf8.Valid(b) {
		return nil, fmt.Errorf("%s: invalid input size, read or UTF-8", filename)
	}
	return b, nil
}

func decodeFileOptions(raw []byte, target *bodycodegen.TypedPathOptions) error {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return errors.New("options must be an object")
	}
	if err := canonicalKeys(raw); err != nil {
		return err
	}
	if err := decision.RejectDuplicateJSONKeys(raw); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		return errors.New("options have trailing JSON content")
	}
	return nil
}
