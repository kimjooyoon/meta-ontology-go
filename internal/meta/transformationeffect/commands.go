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

type executorBindingError struct {
	Operation string
	FieldPath string
	Expected  string
	Observed  string
}

func (err *executorBindingError) Error() string {
	return fmt.Sprintf("executor binding mismatch: operation=%s field=%s expected=%s observed=%s",
		err.Operation, err.FieldPath, err.Expected, err.Observed)
}

func resolveActionBinding(plan generation.Plan, action generation.Action) (generation.Binding, error) {
	binding, ok := generation.BindingForOperation(plan.Registry, action.Operation)
	if !ok {
		return generation.Binding{}, &executorBindingError{
			Operation: string(action.Operation), FieldPath: "$.registry",
			Expected: "one valid operation binding", Observed: string(action.Operation),
		}
	}
	if action.Executor != binding.Executor {
		return generation.Binding{}, &executorBindingError{
			Operation: string(action.Operation), FieldPath: "$.selected.executor",
			Expected: binding.Executor, Observed: action.Executor,
		}
	}
	if action.Evaluator != binding.Evaluator {
		return generation.Binding{}, &executorBindingError{
			Operation: string(action.Operation), FieldPath: "$.selected.evaluator",
			Expected: binding.Evaluator, Observed: action.Evaluator,
		}
	}
	if action.SubjectKind != binding.InputSubjectKind {
		return generation.Binding{}, &executorBindingError{
			Operation: string(action.Operation), FieldPath: "$.selected.subject_kind",
			Expected: string(binding.InputSubjectKind), Observed: string(action.SubjectKind),
		}
	}
	if action.InputSubjectKind != binding.InputSubjectKind ||
		action.InputContractSourceDigest != binding.InputContractSourceDigest ||
		action.InputContractSemanticDigest != binding.InputContractSemanticDigest {
		return generation.Binding{}, &executorBindingError{
			Operation: string(action.Operation), FieldPath: "$.selected.input_contract",
			Expected: binding.InputContractSourceDigest + ":" + binding.InputContractSemanticDigest,
			Observed: action.InputContractSourceDigest + ":" + action.InputContractSemanticDigest,
		}
	}
	return binding, nil
}

func runAction(box *workspace.Sandbox, opts Options, plan generation.Plan, action generation.Action, check bool) ([]byte, error) {
	binding, err := resolveActionBinding(plan, action)
	if err != nil {
		return nil, err
	}
	executor := binding.Executor
	args := actionArguments(executor, box.Root, opts, plan, action, check)
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
	canonicalPayload, err := canonicalMetricsPayload(stdout.Bytes(), box.Root)
	if err != nil {
		return report, nil, err
	}
	return report, canonicalPayload, nil
}

func canonicalMetricsPayload(payload []byte, root string) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		return nil, err
	}
	for _, field := range []string{"root", "storage_root"} {
		raw, ok := fields[field]
		if !ok {
			continue
		}
		var value string
		if err := json.Unmarshal(raw, &value); err != nil || value != root {
			continue
		}
		fields[field] = json.RawMessage(`"<workspace>"`)
	}
	return json.Marshal(fields)
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
