package bodyrefinement

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func policyInputs(observation Observation) bodyexecution.CompositionCases {
	raw, _ := json.Marshal(observation)
	return bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
		Cases: []bodyexecution.CompositionCase{{Inputs: map[string]json.RawMessage{"Decide": raw}}}}
}

func preparePolicy(ctx context.Context, options Options) (bodyexecution.Composition, error) {
	composition, err := bodyexecution.GenerateComposition(ctx, "policy.gooo", options.PolicySource, policyInputs(Observation{}), "")
	if err != nil {
		return composition, err
	}
	if len(composition.Plan.Activities) != 1 || composition.Plan.Activities[0].Name != "Decide" || composition.Plan.Activities[0].Assembling {
		return composition, fmt.Errorf("feedback policy requires one ordinary Gooo activity named Decide")
	}
	return composition, nil
}

func decide(ctx context.Context, options Options, policy bodyexecution.Composition, round *Round) error {
	runtime, err := bodyexecution.ExecuteComposition(ctx, "policy.gooo", options.PolicySource, policy,
		policyInputs(round.Observation), options.GoBinary)
	round.PolicyRuntime = runtime
	if err != nil {
		return fmt.Errorf("execute Gooo feedback policy: %w", err)
	}
	if len(runtime.Traces) != 1 || len(runtime.Traces[0].Deliveries) != 1 {
		return fmt.Errorf("feedback policy returned no unique decision")
	}
	decoder := json.NewDecoder(bytes.NewReader(runtime.Traces[0].Deliveries[0].Actual))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&round.Decision); err != nil {
		return fmt.Errorf("decode Gooo feedback decision: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("feedback decision contains trailing output")
	}
	d := round.Decision
	if d.Action == "STOP" && d.NextAttempts == round.Observation.Attempts {
		return nil
	}
	if d.Action == "INCORPORATE" && d.NextAttempts >= round.Observation.Attempts && d.NextAttempts <= options.MaxAttempts && round.Observation.Counterexamples > 0 {
		return nil
	}
	if d.Action != "CONTINUE" || d.NextAttempts <= round.Observation.Attempts || d.NextAttempts > options.MaxAttempts {
		return fmt.Errorf("feedback decision must stop at the current budget or increase within the caller's limit")
	}
	return nil
}
