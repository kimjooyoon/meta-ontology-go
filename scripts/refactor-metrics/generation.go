package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func emitGenerationPlan(opts options, report sourcepolicy.Report) error {
	if opts.baseSHA == "" {
		return fmt.Errorf("base-sha is required when plan is set")
	}
	inputs := generation.OperationInputs{}
	if opts.workspaceSplits != "" {
		file, err := os.Open(opts.workspaceSplits)
		if err != nil {
			return err
		}
		defer file.Close()
		raw, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
		if err != nil {
			return err
		}
		requests, err := workspacesplit.DecodeRequests(raw)
		if err != nil {
			return err
		}
		report, inputs, err = generation.BindWorkspaceSplitInputs(report, requests)
		if err != nil {
			return err
		}
	}
	plan := generation.BuildWithOperationInputs(opts.baseSHA, opts.sha, report, inputs)
	payload, err := generation.Encode(plan)
	if err != nil {
		return fmt.Errorf("encode generation plan: %w", err)
	}
	if err := writeAtomic(opts.plan, payload, 0o600); err != nil {
		return fmt.Errorf("write generation plan: %w", err)
	}
	fmt.Printf("self-improvement: decision=%s reason=%s selected=%d replay=%s\n", plan.Decision, plan.Reason, len(plan.Selected), plan.ReplayDigest)
	if plan.Decision != generation.DecisionPlan && plan.Decision != generation.DecisionFixedPoint {
		fmt.Printf("self-improvement-diagnostic: stage=selection step=independent-pressure shortfall=%v unknown=%v unselected=%v\n", plan.Shortfall, plan.UnknownIndicatorIDs, plan.UnselectedIndicatorIDs)
		for _, indicator := range report.Actionable() {
			fmt.Printf("self-improvement-candidate: operation=%s subject=%s\n", indicator.Operation, indicator.Subject)
		}
		return fmt.Errorf("self-improvement generation failed closed: %s/%s", plan.Decision, plan.Reason)
	}
	return nil
}
