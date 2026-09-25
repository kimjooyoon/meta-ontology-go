package selfimprovementobservation

import (
	"encoding/json"
	"fmt"
	"os"
)

type Document[T any] struct {
	Value      T
	FileDigest string
}

type Inputs struct {
	Report          Document[SourceReport]
	Counterexamples Document[CounterexampleReport]
	Contract        Document[ContractReport]
}

func LoadInputs(reportPath, counterexamplePath, contractPath string) (Inputs, error) {
	var in Inputs
	var err error
	if in.Report, err = decodeDocument[SourceReport](reportPath); err != nil {
		return in, fmt.Errorf("language report: %w", err)
	}
	if in.Counterexamples, err = decodeDocument[CounterexampleReport](counterexamplePath); err != nil {
		return in, fmt.Errorf("counterexamples: %w", err)
	}
	if in.Contract, err = decodeDocument[ContractReport](contractPath); err != nil {
		return in, fmt.Errorf("Gooo contract: %w", err)
	}
	return in, nil
}

func decodeDocument[T any](path string) (Document[T], error) {
	var result Document[T]
	data, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(data, &result.Value); err != nil {
		return result, err
	}
	result.FileDigest = digestBytes(data)
	return result, nil
}
