package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ExecuteGeneratedBatch builds one generated judge and executes the supplied
// cases in order. It copies the case list without rebinding evidence digests.
// An execution error returns only the successfully decoded prefix; no result
// is invented for the failed case or the remaining cases.
func ExecuteGeneratedBatch(ctx context.Context, judgeSource []byte, inputs []Case) ([]DecisionResult, error) {
	if len(inputs) == 0 {
		return nil, errors.New("generated judge batch requires at least one case")
	}
	inputs = append([]Case(nil), inputs...)
	work, err := os.MkdirTemp("", "gooo-policy-judge-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(work)
	sourcePath := filepath.Join(work, "judge.go")
	if err := os.WriteFile(sourcePath, judgeSource, 0o600); err != nil {
		return nil, err
	}
	binaryPath := filepath.Join(work, "judge")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	environment := append(os.Environ(), "GO111MODULE=off", "GOTOOLCHAIN=go1.27.0")
	build := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binaryPath, sourcePath)
	build.Dir, build.Env = work, environment
	if output, err := build.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("build generated judge: %w: %s", err, strings.TrimSpace(string(output)))
	}
	results := make([]DecisionResult, 0, len(inputs))
	for _, input := range inputs {
		result, err := executeGeneratedBinary(ctx, binaryPath, environment, input)
		if err != nil {
			return results, fmt.Errorf("generated judge case %q: %w", input.ID, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func executeGeneratedBinary(ctx context.Context, binaryPath string, environment []string, input Case) (DecisionResult, error) {
	payload, err := json.Marshal(generatedJudgeInput{
		ID:                           input.ID,
		ProducerAvailable:            input.ProducerAvailable,
		ConsumerAvailable:            input.ConsumerAvailable,
		ObservedSourceDigest:         input.ObservedSourceDigest,
		ObservedArtifactSourceDigest: input.ObservedArtifactSourceDigest,
		ObservedGeneratedJudgeDigest: input.ObservedGeneratedJudgeDigest,
		ObservedIndependentDigest:    input.ObservedIndependentDigest,
		UpperDecision:                input.UpperDecision,
	})
	if err != nil {
		return DecisionResult{}, err
	}
	command := exec.CommandContext(ctx, binaryPath)
	command.Dir, command.Env = filepath.Dir(binaryPath), environment
	command.Stdin = bytes.NewReader(payload)
	output, err := command.CombinedOutput()
	if err != nil {
		return DecisionResult{}, fmt.Errorf("execute generated judge: %w: %s", err, strings.TrimSpace(string(output)))
	}
	var result DecisionResult
	if err := decodeStrictJSON(output, &result); err != nil {
		return DecisionResult{}, fmt.Errorf("decode generated judge: %w", err)
	}
	if result.CaseID != input.ID {
		return DecisionResult{}, errors.New("generated judge changed case identity")
	}
	return result, nil
}
