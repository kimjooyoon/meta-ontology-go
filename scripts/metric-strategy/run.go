package main

import (
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
)

func run(value options) error {
	if value.metrics == "" || value.intervention == "" || value.interventionVerification == "" || value.output == "" {
		return fmt.Errorf("metrics, intervention, intervention-verification, and output are required")
	}
	switch value.mode {
	case "generate":
		if value.repository == "" || value.subjectSHA == "" {
			return fmt.Errorf("repository and subject-sha are required for generation")
		}
		plan, err := strategy.Generate(value.metrics, value.intervention, value.interventionVerification, value.repository, value.subjectSHA)
		if err != nil {
			return err
		}
		return writeJSON(value.output, plan)
	case "verify":
		if value.plan == "" {
			return fmt.Errorf("plan is required for verification")
		}
		plan, err := artifact.ReadJSON[strategy.Plan](value.plan)
		if err != nil {
			return err
		}
		receipt, err := strategyverify.Replay(value.metrics, value.intervention, value.interventionVerification, plan)
		if err != nil {
			return err
		}
		return writeJSON(value.output, receipt)
	default:
		return fmt.Errorf("unsupported metric strategy mode %q", value.mode)
	}
}

