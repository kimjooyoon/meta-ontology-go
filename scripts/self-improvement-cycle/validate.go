package main

type validation struct {
	Schemas, Context, Heads, Contract                    bool
	MetricSchema, MetricRootException, MetricRoots       bool
	MetricObservations, MetricWitnesses, MetricSemantics bool
	States, Links, Ledger, Digests                        bool
}

func validateInputs(in inputs, opts options) validation {
	plan, execution := in.Plan.Value, in.Execution.Value
	receipts, provenance := in.Receipts.Value, in.Provenance.Value
	contract := in.Contract.Value
	base := plan.BaseSHA
	metricSchema, metricRootException, metricRoots, metricObservations, metricWitnesses, metricSemantics := validateMetrics(in.Metrics.Value)
	return validation{
		Schemas: plan.SchemaVersion == "gooo/self-improvement-generation/v6" && execution.SchemaVersion == "gooo/meta-operation-execution/v6" &&
			receipts.SchemaVersion == "gooo/meta-operation-receipt-report/v2" &&
			provenance.SchemaVersion == "gooo/meta-artifact-provenance/v1" && contract.Schema == "gooo/self-improvement-contract/v1" &&
			in.Metrics.Value.Meta.Schema == "gooo/indicator-report/v3",
		Context: opts.runID > 0 && (opts.branch == "dev" || opts.branch == "main") &&
			(opts.conclusion == "success" || opts.conclusion == "failure"),
		Heads: validSHA(opts.headSHA) && in.Metrics.Value.CommitSHA == opts.headSHA && plan.HeadSHA == opts.headSHA &&
			execution.HeadSHA == opts.headSHA &&
			receipts.HeadSHA == opts.headSHA && provenance.HeadSHA == opts.headSHA &&
			contract.CommitSHA == opts.headSHA && validSHA(base) && execution.BaseSHA == base &&
			receipts.BaseSHA == base && provenance.BaseSHA == base,
		Contract: contract.Status == "PASS" && !contract.PromotionAuthorized &&
			validDigest(contract.SourceSHA256) && validDigest(contract.SemanticHash) && validDigest(contract.RegistryDigest) &&
			validContractIndicators(contract) && validContractCoverage(contract),
		MetricSchema: metricSchema, MetricRootException: metricRootException,
		MetricRoots: metricRoots, MetricObservations: metricObservations,
		MetricWitnesses: metricWitnesses, MetricSemantics: metricSemantics,
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

func validateMetrics(metrics metricsDocument) (schema, rootExceptionValid, roots, observations, witnesses, semantics bool) {
	binding := metricsProjection(metrics)
	schema = metrics.Meta.Schema == "gooo/indicator-report/v3"
	rootExceptionValid = metrics.Meta.Policy.ExemptProjectRootTopology
	roots = validMetricRoot(binding.LogicalRoot) && validMetricRoot(binding.StorageRoot)
	observations = true
	expected, exemptions := metricExpectations(binding), 0
	for _, indicator := range metrics.Meta.Indicators {
		if indicator.Subject != "." {
			continue
		}
		if value, ok := expected[indicator.MetricID]; ok {
			if value != indicator.Value || indicator.Applicability != "APPLICABLE" ||
				indicator.Blocking || indicator.Decision != "PASS" {
				observations = false
			}
			delete(expected, indicator.MetricID)
		} else if rootException(indicator, binding.StorageRoot) {
			exemptions++
		}
	}
	observations = observations && len(expected) == 0
	rootExceptionValid = rootExceptionValid && exemptions == 2
	witnessDigest, witnessCount := metricWitnessBinding(metrics, binding)
	witnesses = binding.RootTopologyExempt == metrics.Meta.Policy.ExemptProjectRootTopology &&
		binding.RootWitnessDigest == witnessDigest && binding.RootWitnessCount == witnessCount &&
		witnessCount == 10 && validDigest(witnessDigest)
	canonical := binding
	canonical.SemanticDigest = ""
	semantics = validDigest(binding.SemanticDigest) && binding.SemanticDigest == digestJSON(canonical)
	return
}
