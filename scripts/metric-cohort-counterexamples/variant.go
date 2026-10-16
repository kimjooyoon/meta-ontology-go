package main

import "fmt"

func buildVariant(f fixture, row countercase) (variant, error) {
	v := variant{Receipt: f.Raw, Inputs: f.Inputs}
	var err error
	switch row.Kind {
	case "PRODUCER_INPUT_MUTATION":
		inputs, err := producerInputs(f, row.ID)
		if err != nil {
			return v, err
		}
		v, err = produce(inputs)
		if err != nil {
			v.ProducerError = err.Error()
		}
		return v, nil
	case "RESEALED_RECEIPT_MUTATION":
		v.Receipt, err = mutatedReceipt(f, row.ID)
	case "SOURCE_INPUT_MUTATION":
		v.Inputs, err = sourceMutation(v.Inputs, row.ID)
	case "MISSING_INPUT":
		switch row.ID {
		case "missing-receipt":
			v.Receipt = nil
		case "missing-producer-input":
			v.Inputs.Strategy = nil
		default:
			err = fmt.Errorf("undeclared missing input: %s", row.ID)
		}
	default:
		err = fmt.Errorf("undeclared counterexample kind: %s", row.Kind)
	}
	return v, err
}
