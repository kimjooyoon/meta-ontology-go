package transformationeffect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect/workspace"
)

func runAction(box *workspace.Sandbox, opts Options, plan generation.Plan, action generation.Action, check bool) ([]byte, error) {
	executor := map[sourcepolicy.Operation]string{
		sourcepolicy.OperationCollapseAssign: "scripts/refactor-metrics",
		sourcepolicy.OperationSplitGo:        "scripts/source-splitter",
		sourcepolicy.OperationSplitGooo:      "bootstrap/source-repacker",
	}[action.Operation]
	if executor == "" || action.Executor != executor || action.Evaluator != executor+":check" {
		return nil, fmt.Errorf("unbound executor for %s", action.Operation)
	}
	args := []string{"run", "./" + executor, "-root", box.Root, "-metrics", opts.MetricsPath,
		"-sha", opts.ExpectedSHA, "-subject", action.Subject}
	if action.Operation == sourcepolicy.OperationCollapseAssign {
		args = append(args, "-base-sha", plan.BaseSHA)
	}
	if check {
		args = append(args, "-check")
	}
	output, err := workspace.RunCombined(box.Root, os.Environ(), "go", args...)
	if err != nil {
		return nil, fmt.Errorf("%s check=%t: %w: %s", executor, check, err, output)
	}
	return output, nil
}

func freshMetrics(box *workspace.Sandbox, expected string) (linecaps.LineMetricsReport, []byte, error) {
	args := []string{"run", "./scripts/line-metrics", "-root", box.Root, "-storage-root", box.Root, "-json"}
	command := exec.Command("go", args...)
	command.Dir = box.Root
	command.Env = append(os.Environ(), "METRICS_COMMIT_SHA="+expected)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		return linecaps.LineMetricsReport{}, nil, fmt.Errorf("remeasure sandbox: %w: %s", err, stderr.Bytes())
	}
	var report linecaps.LineMetricsReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		return report, nil, err
	}
	if report.CommitSHA != expected {
		return report, nil, fmt.Errorf("remeasured SHA is not exact")
	}
	return report, stdout.Bytes(), nil
}

func residualActionable(report linecaps.LineMetricsReport, action generation.Action) int {
	count := 0
	for _, indicator := range report.Meta.Indicators {
		if indicator.MetricID == action.MetricID && indicator.Subject == action.Subject &&
			indicator.Operation == action.Operation && indicator.Applicability == sourcepolicy.ApplicabilityApplicable &&
			!indicator.Satisfied {
			count++
		}
	}
	return count
}
