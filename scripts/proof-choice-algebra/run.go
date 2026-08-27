package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/proofchoicealgebra"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/proofchoicejudge"
)

func run(mode, source, receipt, output, root, beforePath, afterPath, expect, baselinePath string) error {
	if output == "" {
		return fmt.Errorf("output is required")
	}
	if mode == "artifact" {
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		artifact, err := proofchoicealgebra.CanonicalArtifact(source, data)
		if err != nil {
			return err
		}
		return os.WriteFile(output, artifact, 0o644)
	}
	baseline, err := readOptional(baselinePath)
	if err != nil {
		return err
	}
	var result any
	switch mode {
	case "produce":
		result = produce(source, root, beforePath, afterPath, baseline)
	case "judge":
		result = judge(source, receipt, beforePath, afterPath, baseline)
	case "receipt-only":
		data, err := os.ReadFile(receipt)
		if err != nil {
			return err
		}
		result = proofchoicejudge.Judge(data)
	default:
		return fmt.Errorf("unknown mode %q", mode)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(output, append(data, '\n'), 0o644); err != nil {
		return err
	}
	decision := decisionOf(result)
	if expect != "" && decision != expect {
		return fmt.Errorf("decision = %s, want %s", decision, expect)
	}
	fmt.Printf("proof-choice algebra: mode=%s decision=%s\n", mode, decision)
	return nil
}
