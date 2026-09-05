package transformationeffect

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const replayDiagnosticSchema = "gooo/transformation-effect-replay-diagnostic/v1"

type replayDivergence struct {
	Stage    string
	Step     string
	Path     string
	Expected string
	Observed string
}

func (err *replayDivergence) Error() string {
	return fmt.Sprintf(
		"meta artifact replay diverged: stage=%s step=%s field_path=%s expected=%s observed=%s",
		err.Stage, err.Step, err.Path, err.Expected, err.Observed,
	)
}

func compareReplay(step string, expected, observed any) error {
	left, err := json.Marshal(expected)
	if err != nil {
		return fmt.Errorf("marshal expected replay %s: %w", step, err)
	}
	right, err := json.Marshal(observed)
	if err != nil {
		return fmt.Errorf("marshal observed replay %s: %w", step, err)
	}
	if string(left) == string(right) {
		return nil
	}
	var leftValue, rightValue any
	if err := json.Unmarshal(left, &leftValue); err != nil {
		return fmt.Errorf("decode expected replay %s: %w", step, err)
	}
	if err := json.Unmarshal(right, &rightValue); err != nil {
		return fmt.Errorf("decode observed replay %s: %w", step, err)
	}
	path, expectedValue, observedValue := firstReplayDifference("$", leftValue, rightValue)
	return &replayDivergence{Stage: "validate-inputs", Step: step, Path: path,
		Expected: expectedValue, Observed: observedValue}
}

func firstReplayDifference(path string, expected, observed any) (string, string, string) {
	return firstReplayDifferenceAt(path, expected, true, observed, true)
}

func firstReplayDifferenceAt(path string, expected any, expectedPresent bool, observed any, observedPresent bool) (string, string, string) {
	if !expectedPresent || !observedPresent {
		if !expectedPresent && !observedPresent {
			return "", "", ""
		}
		return path, replayPresenceValue(expectedPresent, expected), replayPresenceValue(observedPresent, observed)
	}
	if expected == nil || observed == nil {
		if expected == nil && observed == nil {
			return "", "", ""
		}
		return path, replayValue(expected), replayValue(observed)
	}
	switch left := expected.(type) {
	case map[string]any:
		right, ok := observed.(map[string]any)
		if !ok {
			return path, replayValue(expected), replayValue(observed)
		}
	capacity := max(len(left), len(right))
	keys := make([]string, 0, capacity)
	seen := make(map[string]bool, capacity)
		for key := range left {
			seen[key] = true
			keys = append(keys, key)
		}
		for key := range right {
			if !seen[key] {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)
		for _, key := range keys {
			leftValue, leftPresent := left[key]
			rightValue, rightPresent := right[key]
			child, expectedValue, observedValue := firstReplayDifferenceAt(
				path+"."+key, leftValue, leftPresent, rightValue, rightPresent,
			)
			if child != "" {
				return child, expectedValue, observedValue
			}
		}
		return "", "", ""
	case []any:
		right, ok := observed.([]any)
		if !ok || len(left) != len(right) {
			return path, replayValue(expected), replayValue(observed)
		}
		for index := range left {
			child, expectedValue, observedValue := firstReplayDifferenceAt(
				fmt.Sprintf("%s[%d]", path, index), left[index], true, right[index], true,
			)
			if child != "" {
				return child, expectedValue, observedValue
			}
		}
		return "", "", ""
	default:
		if replayValue(expected) == replayValue(observed) {
			return "", "", ""
		}
		return path, replayValue(expected), replayValue(observed)
	}
}

func replayPresenceValue(present bool, value any) string {
	if !present {
		return "<missing>"
	}
	return replayValue(value)
}

func replayValue(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("<marshal-error:%v>", err)
	}
	return string(payload)
}

type ReplayDiagnostic struct {
	Schema        string   `json:"schema"`
	Decision      string   `json:"decision"`
	Resolution    string   `json:"resolution"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class,omitempty"`
	NextOperation string   `json:"next_operation,omitempty"`
	BlockedBy     []string `json:"blocked_by"`
	FieldPath     string   `json:"field_path"`
	Expected      string   `json:"expected,omitempty"`
	Observed      string   `json:"observed,omitempty"`
	ExpectedHash  string   `json:"expected_sha256,omitempty"`
	ObservedHash  string   `json:"observed_sha256,omitempty"`
}

func WriteReplayDiagnostic(outputPath string, cause error) error {
	if outputPath == "" {
		return nil
	}
	diagnostic := ReplayDiagnostic{Schema: replayDiagnosticSchema, Decision: "UNKNOWN",
		Resolution: "LOWER_RESOLUTION", Stage: "validate-inputs",
		Step: "validate-artifact-set", Reason: "META_ARTIFACT_VALIDATION_UNCATALOGED",
		UnknownClass: "UNCATALOGED_CAUSE", NextOperation: "report-counterexample",
		BlockedBy: []string{}}
	if divergence, ok := errors.AsType[*replayDivergence](cause); ok {
		diagnostic.Decision = "REFUTED"
		diagnostic.Resolution = "EXACT"
		diagnostic.Step = divergence.Step
		diagnostic.Reason = "META_ARTIFACT_REPLAY_DIVERGED"
		diagnostic.UnknownClass = ""
		diagnostic.NextOperation = "report-counterexample"
		diagnostic.FieldPath = divergence.Path
		diagnostic.Expected = divergence.Expected
		diagnostic.Observed = divergence.Observed
		diagnostic.ExpectedHash = replayDigest(divergence.Expected)
		diagnostic.ObservedHash = replayDigest(divergence.Observed)
	} else if binding, ok := errors.AsType[*executorBindingError](cause); ok {
		diagnostic.Decision = "REFUTED"
		diagnostic.Resolution = "EXACT"
		diagnostic.Step = "bind-executor"
		diagnostic.Reason = "META_OPERATION_EXECUTOR_UNBOUND"
		diagnostic.UnknownClass = ""
		diagnostic.NextOperation = "report-counterexample"
		diagnostic.FieldPath = binding.FieldPath
		diagnostic.Expected = binding.Expected
		diagnostic.Observed = binding.Observed
		diagnostic.ExpectedHash = replayDigest(binding.Expected)
		diagnostic.ObservedHash = replayDigest(binding.Observed)
	}
	payload, err := json.MarshalIndent(diagnostic, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(filepath.Dir(outputPath), "replay-diagnostic.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}

func replayDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}
