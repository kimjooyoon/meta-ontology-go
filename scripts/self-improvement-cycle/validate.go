package main

type validation struct {
	Schemas, Context, Heads, Contract bool
	States, Links, Ledger, Digests    bool
}

func validateInputs(in inputs, opts options) validation {
	plan, execution := in.Plan.Value, in.Execution.Value
	receipts, provenance := in.Receipts.Value, in.Provenance.Value
	contract := in.Contract.Value
	base := plan.BaseSHA
	return validation{
		Schemas: plan.SchemaVersion == "gooo/self-improvement-generation/v6" &&
			execution.SchemaVersion == "gooo/meta-operation-execution/v6" &&
			receipts.SchemaVersion == "gooo/meta-operation-receipt-report/v2" &&
			provenance.SchemaVersion == "gooo/meta-artifact-provenance/v1" &&
			contract.Schema == "gooo/self-improvement-contract/v1",
		Context: opts.runID > 0 && (opts.branch == "dev" || opts.branch == "main") &&
			(opts.conclusion == "success" || opts.conclusion == "failure"),
		Heads: validSHA(opts.headSHA) && in.Metrics.Value.CommitSHA == opts.headSHA &&
			plan.HeadSHA == opts.headSHA && execution.HeadSHA == opts.headSHA &&
			receipts.HeadSHA == opts.headSHA && provenance.HeadSHA == opts.headSHA &&
			contract.CommitSHA == opts.headSHA && validSHA(base) &&
			execution.BaseSHA == base && receipts.BaseSHA == base && provenance.BaseSHA == base,
		Contract: contract.Status == "PASS" && !contract.PromotionAuthorized &&
			validDigest(contract.SourceSHA256) && validDigest(contract.SemanticHash) &&
			validDigest(contract.RegistryDigest) && validContractIndicators(contract) &&
			validContractCoverage(contract),
		States: plan.Decision != "" && plan.Reason != "" &&
			execution.Decision != "" && execution.Reason != "" &&
			receipts.Decision != "" && receipts.Reason != "" &&
			provenance.Decision == "BOUND" && provenance.Reason != "" &&
			!plan.PromotionAuthorized && !execution.PromotionAuthorized &&
			!receipts.PromotionAuthorized && !provenance.PromotionAuthorized,
		Links: execution.PlanDigest == plan.PlanDigest &&
			receipts.PlanDigest == plan.PlanDigest && provenance.PlanDigest == plan.PlanDigest &&
			provenance.ExecutionManifestDigest == execution.ManifestDigest &&
			provenance.ReceiptReportDigest == receipts.ReportDigest,
		Ledger: execution.IndicatorDecisionLedgerCount > 0 &&
			execution.IndicatorDecisionLedgerCount == receipts.IndicatorDecisionLedgerCount &&
			execution.IndicatorDecisionLedgerCount == provenance.IndicatorDecisionLedgerCount &&
			execution.IndicatorDecisionLedgerDigest == receipts.IndicatorDecisionLedgerDigest &&
			execution.IndicatorDecisionLedgerDigest == provenance.IndicatorDecisionLedgerDigest,
		Digests: validDigest(plan.PlanDigest) && validDigest(execution.ManifestDigest) &&
			validDigest(receipts.ReportDigest) && validDigest(provenance.EnvelopeDigest) &&
			validLedgerDigest(execution.IndicatorDecisionLedgerDigest) &&
			validFileDigests(in),
	}
}

func validContractIndicators(contract contractDocument) bool {
	counts := map[string]int{}
	for _, indicator := range contract.Indicators {
		if indicator.Verdict != "PASS" {
			return false
		}
		counts[indicator.Route]++
	}
	return len(contract.Indicators) == 7 &&
		counts["FOUNDATION"] == 3 && counts["COHERENCE"] == 3 &&
		counts["REGRESSION"] == 1
}

func validContractCoverage(contract contractDocument) bool {
	if len(contract.ExecutorCoverage) != 3 {
		return false
	}
	for _, coverage := range contract.ExecutorCoverage {
		if !coverage.Covered {
			return false
		}
	}
	return true
}
