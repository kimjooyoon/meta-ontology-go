package main

import (
	"fmt"
	"os"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
)

func run(value options) error {
	if value.output == "" {
		return fmt.Errorf("output is required")
	}
	switch value.mode {
	case "generate":
		if value.metrics == "" || value.intervention == "" || value.interventionVerification == "" || value.repository == "" || value.subjectSHA == "" {
			return fmt.Errorf("generation inputs, repository, and subject-sha are required")
		}
		plan, err := strategy.Generate(os.DirFS(value.root), value.metrics, value.intervention, value.interventionVerification, value.repository, value.subjectSHA)
		if err != nil {
			return err
		}
		return writeJSON(value.output, plan)
	case "verify":
		if value.metrics == "" || value.intervention == "" || value.interventionVerification == "" || value.plan == "" {
			return fmt.Errorf("verification inputs and plan are required")
		}
		plan, err := artifact.ReadJSON[strategy.Plan](value.plan)
		if err != nil {
			return err
		}
		receipt, err := strategyverify.Replay(os.DirFS(value.root), value.metrics, value.intervention, value.interventionVerification, plan)
		if err != nil {
			return err
		}
		return writeJSON(value.output, receipt)
	case "proposal-contract":
		return writeProposalContract(value)
	default:
		return fmt.Errorf("unsupported metric strategy mode %q", value.mode)
	}
}
