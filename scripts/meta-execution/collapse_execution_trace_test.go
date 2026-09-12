package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

// The trace is diagnostic test output, never canonical operation evidence.
func executeCollapseWithTestTrace(t *testing.T, workspace, gitDir, metricsPath string, plan generation.Plan, action generation.Action) (operationMaterialization, *operationError) {
	t.Helper()
	manifest := generation.BuildExecutionManifest(plan)
	sequence := 0
	for index, step := range manifest.Steps {
		if step.ActionIndicatorID == action.IndicatorID {
			sequence = index + 1
			break
		}
	}
	if sequence == 0 {
		t.Fatal("native collapse trace action is absent from the derived manifest")
	}
	var output bytes.Buffer
	defer func() {
		for event := range strings.SplitSeq(strings.TrimSpace(output.String()), "\n") {
			if event != "" {
				t.Logf("native-collapse-test-trace=%s", event)
			}
		}
		if os.Getenv("CI") == "true" {
			summary, err := renderNativeCollapseCostSummary(output.Bytes())
			if err == nil {
				err = appendNativeCollapseCostSummary(os.Getenv("RUNNER_TEMP"), os.Getenv("GITHUB_STEP_SUMMARY"), summary)
			}
			if err != nil {
				t.Logf("native-collapse-cost-summary unavailable (diagnostic only): %v", err)
			}
		}
	}()
	trace := newMetaExecutionTrace(plan, manifest, action, sequence, newMetaExecutionTraceStateWithWriter(&output))
	trace.emitActionEntered()
	materialized, failure := executeCollapse(workspace, gitDir, metricsPath, plan, action, trace)
	trace.emitActionReturned(materialized, failure)
	return materialized, failure
}

func renderNativeCollapseCostSummary(raw []byte) ([]byte, error) {
	if len(raw) == 0 || len(raw) > 64*1024 {
		return nil, fmt.Errorf("native trace is empty or exceeds the diagnostic report limit")
	}
	var rows bytes.Buffer
	events, measured, unknown := 0, 0, 0
	escape := strings.NewReplacer("|", "\\|", "\n", " ", "\r", " ", "\t", " ")
	for line := range strings.SplitSeq(strings.TrimSpace(string(raw)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var event struct {
			Schema        string `json:"schema"`
			EventSequence int    `json:"event_sequence"`
			Boundary      string `json:"boundary"`
			CommandKind   string `json:"command_kind"`
			Pass          string `json:"pass"`
			Cost          struct {
				State     string `json:"state"`
				ElapsedNS *int64 `json:"elapsed_ns"`
			} `json:"cost"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, err
		}
		if event.Schema != "gooo/meta-execution-driver-boundary/v1" {
			return nil, fmt.Errorf("native trace schema is absent or unsupported")
		}
		events++
		if !strings.HasSuffix(event.Boundary, "_RETURNED") {
			continue
		}
		value := "UNKNOWN"
		if event.Cost.State == "OBSERVED" {
			if event.Cost.ElapsedNS == nil || *event.Cost.ElapsedNS < 0 {
				return nil, fmt.Errorf("observed native boundary has no non-negative integer duration")
			}
			value = fmt.Sprintf("%d", *event.Cost.ElapsedNS)
			measured++
		} else {
			unknown++
		}
		fmt.Fprintf(&rows, "| %d | %s | %s | %s |\n", event.EventSequence, escape.Replace(event.CommandKind), escape.Replace(event.Pass), value)
	}
	if events == 0 {
		return nil, fmt.Errorf("native trace contains no events")
	}
	var summary bytes.Buffer
	summary.WriteString("\n### Native collapse materializer costs (diagnostic)\n\n")
	summary.WriteString("These are existing source-bound native trace observations, not a quality score. ")
	summary.WriteString("Action and child-command intervals overlap; do not add them into a total. ")
	summary.WriteString("Source authenticity: UNVERIFIED. Semantic admission: UNASSESSED. Improvement: UNKNOWN.\n\n")
	fmt.Fprintf(&summary, "Trace events: %d. Measured returned intervals: %d. Unknown returned intervals: %d. Raw trace bytes: %d.\n\n", events, measured, unknown, len(raw))
	fmt.Fprintf(&summary, "Raw trace SHA-256: `sha256:%x`\n\n", sha256.Sum256(raw))
	summary.WriteString("| Event | Boundary kind | Pass | Elapsed ns |\n| --- | --- | --- | --- |\n")
	summary.Write(rows.Bytes())
	summary.WriteString("\n<details><summary>Exact source-bound native trace</summary>\n\n````jsonl\n")
	summary.Write(raw)
	if raw[len(raw)-1] != '\n' {
		summary.WriteByte('\n')
	}
	summary.WriteString("````\n\n</details>\n")
	return summary.Bytes(), nil
}

func appendNativeCollapseCostSummary(runnerTemp, target string, payload []byte) error {
	if !filepath.IsAbs(runnerTemp) || !filepath.IsAbs(target) || len(payload) == 0 || len(payload) > 64*1024 {
		return fmt.Errorf("native summary requires a bounded report and absolute runner paths")
	}
	root, err := filepath.EvalSymlinks(runnerTemp)
	if err != nil {
		return err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(target))
	if err != nil {
		return err
	}
	if parent != filepath.Join(root, "_runner_file_commands") || !strings.HasPrefix(filepath.Base(target), "step_summary_") {
		return fmt.Errorf("native summary target is not the runner step-summary command file")
	}
	if info, err := os.Lstat(target); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("native summary target is not a regular file")
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return err
	}
	if info.Size() > (1<<20)-int64(len(payload)) {
		_ = file.Close()
		return fmt.Errorf("native summary would exceed the GitHub step-summary limit")
	}
	_, err = file.Write(payload)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func TestNativeCollapseCostSummaryPreservesTraceAndUnknown(t *testing.T) {
	raw := []byte(`{"schema":"gooo/meta-execution-driver-boundary/v1","event_sequence":1,"boundary":"COMMAND_RETURNED","command_kind":"verifier","pass":"first","cost":{"state":"OBSERVED","elapsed_ns":7},"input_contract_source_digest":"synthetic-source"}
{"schema":"gooo/meta-execution-driver-boundary/v1","event_sequence":2,"boundary":"ACTION_RETURNED","command_kind":"action","pass":"selected","cost":{"state":"UNKNOWN"}}
`)
	rendered, err := renderNativeCollapseCostSummary(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		"| 1 | verifier | first | 7 |", "| 2 | action | selected | UNKNOWN |",
		"Trace events: 2.", "Measured returned intervals: 1.", "Unknown returned intervals: 1.",
		fmt.Sprintf("sha256:%x", sha256.Sum256(raw)), string(raw),
		"do not add them into a total", "Improvement: UNKNOWN",
	} {
		if !bytes.Contains(rendered, []byte(wanted)) {
			t.Fatalf("native cost summary lost %q", wanted)
		}
	}
	replayed, err := renderNativeCollapseCostSummary(raw)
	if err != nil || !bytes.Equal(rendered, replayed) {
		t.Fatal("same native trace did not render deterministically")
	}
}

func TestNativeCollapseCostSummaryRejectsUnsupportedObservations(t *testing.T) {
	for name, raw := range map[string]string{
		"empty":           "",
		"malformed":       "{",
		"null":            "null",
		"schema":          `{"schema":"unknown"}`,
		"missing-time":    `{"schema":"gooo/meta-execution-driver-boundary/v1","boundary":"COMMAND_RETURNED","cost":{"state":"OBSERVED"}}`,
		"negative-time":   `{"schema":"gooo/meta-execution-driver-boundary/v1","boundary":"COMMAND_RETURNED","cost":{"state":"OBSERVED","elapsed_ns":-1}}`,
		"fractional-time": `{"schema":"gooo/meta-execution-driver-boundary/v1","boundary":"COMMAND_RETURNED","cost":{"state":"OBSERVED","elapsed_ns":0.5}}`,
		"oversized":       strings.Repeat("x", 64*1024+1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := renderNativeCollapseCostSummary([]byte(raw)); err == nil {
				t.Fatal("unsupported native cost observation was published")
			}
		})
	}
	t.Run("zero-is-observed", func(t *testing.T) {
		raw := []byte(`{"schema":"gooo/meta-execution-driver-boundary/v1","event_sequence":1,"boundary":"COMMAND_RETURNED","command_kind":"verifier","pass":"first","cost":{"state":"OBSERVED","elapsed_ns":0}}`)
		rendered, err := renderNativeCollapseCostSummary(raw)
		if err != nil || !bytes.Contains(rendered, []byte("| 1 | verifier | first | 0 |")) {
			t.Fatalf("observed zero was not preserved: %v", err)
		}
	})
}

func TestNativeCollapseCostSummaryOnlyAppendsToRunnerOutput(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "_runner_file_commands")
	if err := os.Mkdir(directory, 0700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(directory, "step_summary_native")
	if err := os.WriteFile(target, []byte("existing\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := appendNativeCollapseCostSummary(root, target, []byte("observed\n")); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(target)
	if err != nil || string(raw) != "existing\nobserved\nobserved\n" {
		t.Fatalf("native summary did not preserve prior content: %q %v", raw, err)
	}
	for name, invalid := range map[string]string{
		"relative":                  "step_summary_relative",
		"outside-command-directory": filepath.Join(root, "step_summary_outside"),
		"different-command":         filepath.Join(directory, "set_env_not_a_summary"),
	} {
		t.Run(name, func(t *testing.T) {
			if err := appendNativeCollapseCostSummary(root, invalid, []byte("no\n")); err == nil {
				t.Fatal("native summary escaped its runner output boundary")
			}
		})
	}
	t.Run("oversized-report", func(t *testing.T) {
		if err := appendNativeCollapseCostSummary(root, target, bytes.Repeat([]byte("x"), 64*1024+1)); err == nil {
			t.Fatal("unbounded native summary was accepted")
		}
	})
	t.Run("step-limit", func(t *testing.T) {
		full := filepath.Join(directory, "step_summary_full")
		if err := os.WriteFile(full, bytes.Repeat([]byte("x"), 1<<20), 0600); err != nil {
			t.Fatal(err)
		}
		if err := appendNativeCollapseCostSummary(root, full, []byte("no\n")); err == nil {
			t.Fatal("native summary exceeded the step limit")
		}
	})
	t.Run("symlink", func(t *testing.T) {
		link := filepath.Join(directory, "step_summary_link")
		if err := os.Symlink(target, link); err != nil {
			t.Skipf("symlink creation unavailable: %v", err)
		}
		if err := appendNativeCollapseCostSummary(root, link, []byte("no\n")); err == nil {
			t.Fatal("native summary followed a symlink")
		}
	})
}
