package main

import (
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposal"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
)

func writeProposalContract(value options) error {
	if value.plan == "" || value.replayPlan == "" || value.strategyVerification == "" || value.repository == "" || value.subjectSHA == "" {
		return fmt.Errorf("plan, replay-plan, strategy-verification, repository, and subject-sha are required")
	}
	first, err := artifact.ReadJSON[strategy.Plan](value.plan)
	if err != nil { return err }
	replay, err := artifact.ReadJSON[strategy.Plan](value.replayPlan)
	if err != nil { return err }
	receipt, err := artifact.ReadJSON[strategyverify.Receipt](value.strategyVerification)
	if err != nil { return err }
	report, err := proposal.Evaluate(value.repository, value.subjectSHA, first, replay, receipt)
	if err != nil { return err }
	if err := proposal.Validate(report); err != nil { return err }
	return writeJSON(value.output, report)
}
