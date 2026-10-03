package bodyexecution

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodytiming"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

type Observation struct {
	Schema                  string                             `json:"schema"`
	Stage                   string                             `json:"stage"`
	Failure                 string                             `json:"failure,omitempty"`
	OriginalSourceSHA256    string                             `json:"original_source_sha256"`
	SelectedSourceSHA256    string                             `json:"selected_source_sha256"`
	GeneratedSHA256         string                             `json:"generated_sha256"`
	ActivityID              string                             `json:"activity_id"`
	PlanSHA256              string                             `json:"plan_sha256"`
	ParentReceiptSHA256     string                             `json:"parent_receipt_sha256"`
	RuntimeSuiteSHA256      string                             `json:"runtime_suite_sha256"`
	CompilerSourceSHA       string                             `json:"declared_compiler_source_sha"`
	ProducerSourceSHA       string                             `json:"producer_source_sha"`
	GoToolSHA256            string                             `json:"go_tool_sha256"`
	GoVersion               string                             `json:"go_version"`
	ExecutableSHA256        string                             `json:"executable_sha256"`
	Toolchain               ProcessObservation                 `json:"toolchain"`
	Build                   ProcessObservation                 `json:"build"`
	Runs                    []ProcessObservation               `json:"runs"`
	Cases                   []bodycodegen.IRBodyFillCaseResult `json:"cases"`
	DeclaredCases           int                                `json:"declared_cases"`
	SelectionDisjointInputs int                                `json:"selection_disjoint_inputs"`
	ProjectionReplayed      bool                               `json:"projection_replayed"`
	RuntimeReplayed         bool                               `json:"runtime_replayed"`
	ElapsedNS               int64                              `json:"elapsed_ns"`
	Scope                   string                             `json:"scope"`
	Artifact                *ArtifactObservation               `json:"artifact,omitempty"`
	ToolchainReference      *ToolchainObservation              `json:"toolchain_reference,omitempty"`
}

type Result struct {
	Observation Observation `json:"observation"`
	// []byte uses base64 on the wire, preserving original whitespace and escaping.
	ParentReceipt       []byte                            `json:"parent_receipt_bytes"`
	CompletenessReceipt *completeness.CompletenessReceipt `json:"completeness_receipt"`
}

// Execute returns the immutable parent and a new runtime observation, including
// failed builds/runs. Original selection cases never become a holdout claim.
func Execute(ctx context.Context, filename string, source []byte, document pathplan.Document,
	prior bodycodegen.Result, parentReceipt []byte, cases []pathplan.TestCase, goBinary string) (Result, error) {
	return execute(ctx, filename, source, document, prior, parentReceipt, cases, goBinary, nil)
}

func initialResult(source []byte, prior bodycodegen.Result, parentReceipt []byte, cases []pathplan.TestCase) Result {
	result := Result{Observation: Observation{Schema: "gooo/typed-path-runtime-observation/v1", Stage: "VALIDATE",
		OriginalSourceSHA256: digest(source), SelectedSourceSHA256: prior.Report.SourceDigest,
		GeneratedSHA256: prior.Report.GeneratedDigest, ActivityID: prior.Report.ActivityID, PlanSHA256: prior.Report.PlanSHA256,
		CompilerSourceSHA: prior.Report.CompilerSourceSHA, ProducerSourceSHA: producerSourceSHA(),
		Runs: make([]ProcessObservation, 0, 2), Cases: make([]bodycodegen.IRBodyFillCaseResult, 0), DeclaredCases: len(cases),
		Scope: "Independent compiled execution of one source-replayed Integer -> Integer projection; caller-supplied finite expectations; parent model observations are not re-attested; no inference or provider requests."}}
	if len(parentReceipt) <= 1<<20 {
		result.ParentReceipt = append([]byte(nil), parentReceipt...)
	}
	result.Observation.ParentReceiptSHA256 = digest(parentReceipt)
	return result
}

func execute(ctx context.Context, filename string, source []byte, document pathplan.Document,
	prior bodycodegen.Result, parentReceipt []byte, cases []pathplan.TestCase, goBinary string, owner *Executor) (Result, error) {
	start := time.Now()
	result := initialResult(source, prior, parentReceipt, cases)
	phase := bodytiming.Start(ctx, "parent_receipt_validate")
	finish := func(err error) (Result, error) {
		phase.End(err == nil)
		result.Observation.ElapsedNS = time.Since(start).Nanoseconds()
		if err != nil {
			result.Observation.Failure = err.Error()
		}
		receiptPhase := bodytiming.Start(ctx, "runtime_receipt")
		result.CompletenessReceipt = runtimeCompleteness(prior, result)
		receiptPhase.End(true)
		return result, err
	}
	if ctx == nil || len(cases) == 0 || len(cases) > 128 || !token.IsIdentifier(prior.Report.Activity) || token.Lookup(prior.Report.Activity).IsKeyword() {
		return finish(fmt.Errorf("runtime requires context, a declared activity and 1..128 finite cases"))
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	parent, err := completeness.Decode(parentReceipt)
	if err != nil {
		return finish(fmt.Errorf("original parent receipt is invalid"))
	}
	left, _ := json.Marshal(parent)
	encodedPrior, err := json.Marshal(prior.Report.CompletenessReceipt)
	if err != nil {
		return finish(fmt.Errorf("generation parent receipt cannot be encoded"))
	}
	// A live generator can retain typed scope structs; decoded JSON uses maps.
	// Validate both sides with the same exact-number decoder before comparing.
	expectedParent, err := completeness.Decode(encodedPrior)
	if err != nil {
		return finish(fmt.Errorf("generation parent receipt is invalid"))
	}
	right, _ := json.Marshal(expectedParent)
	if string(left) != string(right) {
		return finish(fmt.Errorf("parent receipt and generation observation differ"))
	}
	result.Observation.Stage = "SOURCE_REPLAY"
	phase.End(true)
	phase = bodytiming.Start(ctx, "source_replay")
	if err := bodycodegen.VerifyTypedPathProjection(ctx, filename, source, document, prior); err != nil {
		return finish(err)
	}
	r := &result.Observation
	phase.End(true)
	phase = bodytiming.Start(ctx, "runtime_suite_prepare")
	r.ParentReceiptSHA256 = digest(parentReceipt)
	r.ProjectionReplayed = true
	suite, _ := json.Marshal(cases)
	r.RuntimeSuiteSHA256 = digest(suite)
	for _, c := range cases {
		if !selectionObservedInput(document, prior, c.Input) {
			r.SelectionDisjointInputs++
		}
	}
	if goBinary == "" {
		goBinary = "go"
	}
	goBinary, err = exec.LookPath(goBinary)
	if err != nil {
		return finish(fmt.Errorf("Go tool is unavailable"))
	}
	goBinary, err = filepath.Abs(goBinary)
	if err != nil {
		return finish(err)
	}
	phase.End(true)
	phase = bodytiming.Start(ctx, "go_tool_hash")
	r.GoToolSHA256, err = fileDigest(goBinary)
	if err != nil {
		return finish(fmt.Errorf("cannot bind Go tool bytes"))
	}
	r.Stage = "TOOLCHAIN"
	phase.End(true)
	phase = bodytiming.Start(ctx, "toolchain_bind")
	if err := observeToolchain(ctx, goBinary, r, owner); err != nil {
		return finish(err)
	}
	phase.End(true)
	phase = bodytiming.Start(ctx, "executable_prepare")
	root, executable, release, err := executableFor(ctx, prior, goBinary, r, owner)
	defer release()
	if err != nil {
		return finish(err)
	}
	phase.End(true)
	inputs := make([]int64, len(cases))
	for i, c := range cases {
		inputs[i] = c.Input
	}
	input, _ := json.Marshal(inputs)
	var baseline []int64
	phaseNames := [2][3]string{
		{"executable_hash_1", "native_run_1", "runtime_decode_1"},
		{"executable_hash_2", "native_run_2", "runtime_decode_2"},
	}
	for run := range 2 {
		r.Stage = fmt.Sprintf("EXECUTE_%d", run+1)
		if owner != nil {
			phase = bodytiming.Start(ctx, phaseNames[run][0])
			current, err := fileDigest(executable)
			if err != nil || current != r.ExecutableSHA256 {
				return finish(fmt.Errorf("owned executable binding changed before execution"))
			}
			phase.End(true)
		}
		phase = bodytiming.Start(ctx, phaseNames[run][1])
		runCtx, stop := context.WithTimeout(ctx, 2*time.Second)
		output, observation, runErr := process(runCtx, root, executable, input)
		stop()
		r.Runs = append(r.Runs, observation)
		if runErr != nil {
			return finish(runErr)
		}
		phase.End(true)
		phase = bodytiming.Start(ctx, phaseNames[run][2])
		var values []int64
		if err := json.Unmarshal(output, &values); err != nil || len(values) != len(cases) {
			return finish(fmt.Errorf("runtime output count or int64 shape differs"))
		}
		if run == 0 {
			baseline = values
			for i, c := range cases {
				r.Cases = append(r.Cases, bodycodegen.IRBodyFillCaseResult{Input: c.Input, Expected: c.Expected, Actual: values[i], Passed: values[i] == c.Expected})
			}
		} else if !slices.Equal(values, baseline) {
			return finish(fmt.Errorf("compiled runtime replay differs"))
		}
		phase.End(true)
	}
	r.RuntimeReplayed = true
	r.Stage = "COMPLETE"
	return finish(nil)
}

// Source replay above validates every observation before it contributes to the
// effective selection suite. Added oracle inputs cannot become holdout claims.
func selectionObservedInput(document pathplan.Document, prior bodycodegen.Result, input int64) bool {
	if slices.ContainsFunc(document.TestCases, func(t pathplan.TestCase) bool { return t.Input == input }) {
		return true
	}
	if prior.Report.BodyPaths != nil && prior.Report.BodyPaths.Observation != nil {
		for _, round := range prior.Report.BodyPaths.Observation.Rounds {
			if round.Observation != nil && round.Observation.Input == input {
				return true
			}
		}
	}
	return false
}

func prepareWorkspace(root string, prior bodycodegen.Result) error {
	if err := os.Mkdir(filepath.Join(root, "projection"), 0700); err != nil {
		return err
	}
	file, err := parser.ParseFile(token.NewFileSet(), "generated.go", prior.Source, parser.PackageClauseOnly)
	if err != nil {
		return fmt.Errorf("cannot read replayed package name")
	}
	bridge := "GoooRuntimeInvoke"
	if bridge == prior.Report.Activity {
		bridge += "Body"
	}
	adapter := fmt.Sprintf("package %s\nfunc %s(input int64) int64 {return %s(input)}\n", file.Name.Name, bridge, prior.Report.Activity)
	main := fmt.Sprintf("package main\nimport(\"encoding/json\";\"os\";p \"gooo.observed.body/projection\")\nfunc main(){var in []int64;if json.NewDecoder(os.Stdin).Decode(&in)!=nil||len(in)>128{os.Exit(2)};out:=make([]int64,len(in));for i,v:=range in{out[i]=p.%s(v)};if json.NewEncoder(os.Stdout).Encode(out)!=nil{os.Exit(3)}}\n", bridge)
	files := map[string]string{"go.mod": "module gooo.observed.body\n\ngo 1.27.1\n", "main.go": main, "projection/generated.go": prior.Source, "projection/adapter.go": adapter}
	if file.Name.Name == "main" {
		delete(files, "projection/generated.go")
		delete(files, "projection/adapter.go")
		files["generated.go"], files["adapter.go"] = prior.Source, adapter
		files["main.go"] = strings.ReplaceAll(strings.Replace(main, `;p "gooo.observed.body/projection"`, "", 1), "p."+bridge, bridge)
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0600); err != nil {
			return err
		}
	}
	return nil
}

func digest(b []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(b)) }
func fileDigest(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || !info.Mode().IsRegular() || info.Size() > 256<<20 {
		return "", fmt.Errorf("invalid executable file")
	}
	h := sha256.New()
	_, e = io.Copy(h, f)
	return fmt.Sprintf("sha256:%x", h.Sum(nil)), e
}
