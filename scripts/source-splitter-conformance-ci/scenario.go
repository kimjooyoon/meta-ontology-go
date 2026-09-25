package main

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

func evaluateScenario(
	name string,
	contractRaw []byte,
	evidenceRaw []byte,
	required []string,
	proofChoice string,
) (scenario, error) {
	evaluation, err := transformationeffect.EvaluateSplitGo(
		contractRaw, evidenceRaw, required, proofChoice,
	)
	if err != nil {
		return scenario{}, fmt.Errorf("evaluate %s: %w", name, err)
	}
	if err := transformationeffect.ValidateSplitGoEvaluation(evaluation); err != nil {
		return scenario{}, fmt.Errorf("validate %s: %w", name, err)
	}
	raw, err := json.Marshal(evaluation)
	if err != nil {
		return scenario{}, fmt.Errorf("marshal %s: %w", name, err)
	}
	stats, err := projectEvaluation(raw, required)
	if err != nil {
		return scenario{}, fmt.Errorf("project %s: %w", name, err)
	}
	decision := decisionBlock
	if stats.passCount == len(required) {
		decision = decisionPass
	}
	return scenario{
		Name: name, Decision: decision, Resolution: stats.resolution,
		PassCount: stats.passCount, FailCount: stats.failCount,
		UnknownCount: stats.unknownCount, Denominator: len(required),
		Evaluation: raw,
	}, nil
}
