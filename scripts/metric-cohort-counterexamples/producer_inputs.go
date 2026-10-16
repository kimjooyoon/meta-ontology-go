package main

import (
	"encoding/json"
	"fmt"
	"strings"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
)

// These are adversarial resealed inputs, never new upstream verification authority.
func producerInputs(f fixture, id string) (co.SourceInputs, error) {
	inputs := f.Inputs
	var plan metricstrategy.Plan
	if err := json.Unmarshal(inputs.Strategy, &plan); err != nil {
		return inputs, err
	}
	index := -1
	for i, binding := range plan.Bindings {
		if binding.Family == "FOUNDATION" && strings.HasPrefix(binding.IndicatorID, "gooo.concept.operation.") {
			index = i
			break
		}
	}
	if index < 0 {
		return inputs, fmt.Errorf("original producer has no foundation concept binding for the declared cases")
	}
	row := plan.Bindings[index]
	switch id {
	case "missing-observed-binding":
		plan.Bindings = append(plan.Bindings[:index], plan.Bindings[index+1:]...)
	case "replaced-observed-binding":
		plan.Bindings[index].IndicatorID = "gooo.concept.unresolved-operation.native-replaced.v1"
	case "extra-unknown-observation":
		row.IndicatorID = "gooo.concept.unresolved-operation.native-extra.v1"
		plan.Bindings = append(plan.Bindings, row)
	case "ambiguous-binding":
		plan.Bindings = append(plan.Bindings, row)
	default:
		return inputs, fmt.Errorf("undeclared producer mutation: %s", id)
	}
	if err := repartitionPlan(&plan); err != nil {
		return inputs, err
	}
	var verification strategyverify.Receipt
	if err := json.Unmarshal(inputs.StrategyVerification, &verification); err != nil {
		return inputs, err
	}
	verification.PlanDigest, verification.BindingCount = plan.Digest, len(plan.Bindings)
	verification.Digest = ""
	digest, err := artifact.Digest(verification)
	if err != nil {
		return inputs, err
	}
	verification.Digest = digest
	inputs.Strategy, err = json.Marshal(plan)
	if err == nil {
		inputs.StrategyVerification, err = json.Marshal(verification)
	}
	return inputs, err
}
