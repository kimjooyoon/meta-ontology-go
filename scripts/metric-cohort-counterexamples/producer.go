package main

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
)

func produce(inputs co.SourceInputs) (variant, error) {
	v := variant{Inputs: inputs, ProducerAttempted: true}
	program, source, err := metricprogram.Compile(inputs.Strategy, inputs.StrategyVerification)
	if err != nil {
		return v, err
	}
	v.Inputs.ProgramSource = source
	v.Inputs.Program, err = json.Marshal(program)
	if err != nil {
		return v, err
	}
	verification, err := programverify.Verify(inputs.Strategy, inputs.StrategyVerification, v.Inputs.Program, source)
	if err != nil {
		return v, err
	}
	v.Inputs.ProgramVerification, err = json.Marshal(verification)
	if err != nil {
		return v, err
	}
	receipt, err := co.Build(inputs.Strategy, inputs.StrategyVerification, inputs.Intervention,
		v.Inputs.Program, source, v.Inputs.ProgramVerification)
	if err != nil {
		return v, err
	}
	v.Producer = &receipt
	v.Receipt, err = json.Marshal(receipt)
	return v, err
}
