package selfimprovementloop

import (
	"encoding/json"
	"fmt"
)

func DecodeGraph(data []byte) (Graph, error) {
	var graph Graph
	if err := json.Unmarshal(data, &graph); err != nil {
		return Graph{}, fmt.Errorf("decode released semantic graph: %w", err)
	}
	return graph, nil
}

func DecodeInput(data []byte) (Input, error) {
	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		return Input{}, fmt.Errorf("decode loop input: %w", err)
	}
	return input, nil
}

func Run(graph Graph, input Input) (Artifacts, error) {
	report, err := Evaluate(graph, input)
	if err != nil {
		return Artifacts{}, err
	}
	return BuildArtifacts(input, report), nil
}
