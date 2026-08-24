package assuranceeligibility

func validParent(value evidence) bool {
	report, suite := value.ParentReport, value.ParentSuite
	return report.Schema == "external-ecosystem-execution-report/v1" &&
		report.ContractVersion == "external-ecosystem-conformance/v1" && report.Decision == DecisionFailClosed &&
		report.Resolution == ResolutionExact && report.Reason == "EXECUTION_INVARIANT_VIOLATED" &&
		report.DenominatorVersion == "external-ecosystem-execution-denominator/v1" &&
		report.DenominatorDigest == ParentDenominator && report.Completed == 6 && report.Total == 8 &&
		report.BasisPoints == 7500 && report.ExternalExecutions == 2 && report.UnknownIndicators == 0 &&
		suite.Schema == "external-ecosystem-execution-suite/v1" && suite.Passed == 10 && suite.Total == 10 &&
		suite.UnknownExpected == 3 && suite.InvariantExpected == 6 && suite.Unresolved == 0
}

func validCapability(value evidence) bool {
	report, observation, suite := value.CapabilityReport, value.CapabilityObservation, value.CapabilitySuite
	parent := report.Parent
	return report.Schema == "gooo/external-capability-report/v1" &&
		report.Decision == "CAPABILITY_EXECUTABLE" && report.Resolution == ResolutionExact &&
		report.EnforcementEffect == EffectNone && report.Reason == "CAPABILITY_EXECUTION_EXACT_PARENT_FAIL_CLOSED" &&
		report.Completed == 10 && report.Total == 10 && report.BasisPoints == 10000 &&
		report.DriverCompleted == 4 && report.DriverTotal == 4 && report.OutcomeCompleted == 3 &&
		report.OutcomeTotal == 3 && report.GuardrailCompleted == 3 && report.GuardrailTotal == 3 &&
		report.ExternalExecutions == 4 && report.UnknownIndicators == 0 &&
		parent.Decision == DecisionFailClosed && parent.Resolution == ResolutionExact &&
		parent.Completed == 6 && parent.Total == 8 && parent.BasisPoints == 7500 &&
		observation.Schema == "gooo/external-capability-observation/v1" && observation.Available &&
		observation.ReplayExact && observation.ExternalExecutions == 4 &&
		report.ObservationDigest == observation.ObservationDigest &&
		suite.Schema == "gooo/external-capability-conformance/v1" &&
		suite.Decision == "CAPABILITY_EXECUTABLE" && suite.Resolution == ResolutionExact &&
		suite.Passed == 15 && suite.Total == 15 && suite.CoverageBPS == 10000 &&
		suite.ExactExpected == 1 && suite.UnknownExpected == 3 && suite.InvariantExpected == 11
}

func observe(report *Report, value evidence) {
	report.Summary.ParentCompleted = value.ParentReport.Completed
	report.Summary.ParentCoverageBPS = value.ParentReport.BasisPoints
	report.Summary.ParentKnownFailures = value.ParentReport.Total - value.ParentReport.Completed
	report.Summary.CapabilityCompleted = value.CapabilityReport.Completed
	report.Summary.CapabilityCoverageBPS = value.CapabilityReport.BasisPoints
	report.Summary.CapabilityOutcomes = value.CapabilityReport.OutcomeCompleted
	report.Summary.CapabilitySuitePassed = value.CapabilitySuite.Passed
	report.Summary.CapabilitySuiteCoverageBPS = value.CapabilitySuite.CoverageBPS
	report.Summary.ExternalExecutions = value.CapabilityReport.ExternalExecutions
	report.Summary.RepositoryWrites = value.ParentReport.RepositoryWrites +
		value.CapabilityReport.RepositoryWrites + value.CapabilityObservation.RepositoryWrites
	report.Summary.ExternalRepositoryWrites = value.CapabilityReport.ExternalRepositoryWrites +
		value.CapabilityObservation.ExternalRepositoryWrites
	report.Summary.OfficialMutations = value.ParentReport.OfficialMutationCount +
		value.ParentObservation.OfficialMutationCount + value.CapabilityReport.OfficialMutationCount +
		value.CapabilityObservation.OfficialMutationCount + value.CapabilitySuite.OfficialMutations
	report.Summary.Promotions = value.ParentReport.PromotionCount + value.ParentObservation.PromotionCount +
		value.CapabilityReport.PromotionCount + value.CapabilityObservation.PromotionCount +
		value.CapabilitySuite.PromotionCount
}
