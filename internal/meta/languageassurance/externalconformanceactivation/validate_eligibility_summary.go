package externalconformanceactivation

func validEligibilitySummary(s eligibilitySummary) bool {
	return s.AssuranceDenominator == 12 && s.BeforeOperating == 11 && s.ProjectedOperating == 12 &&
		s.OfficialOperating == 11 && s.BeforeCoverageBPS == 9166 && s.ProjectedCoverageBPS == 10000 &&
		s.OfficialCoverageBPS == 9166 && s.EvidenceTotal == 7 && s.EvidenceExact == 7 &&
		s.ParentCompleted == 6 && s.ParentTotal == 8 && s.ParentCoverageBPS == 7500 &&
		s.ParentKnownFailures == 2 && s.CapabilityCompleted == 10 && s.CapabilityTotal == 10 &&
		s.CapabilityCoverageBPS == 10000 && s.CapabilityOutcomes == 3 && s.CapabilityOutcomeTotal == 3 &&
		s.CapabilitySuitePassed == 15 && s.CapabilitySuiteTotal == 15 && s.CapabilitySuiteCoverageBPS == 10000 &&
		s.ExternalExecutions == 4 && s.EligiblePaths == 1 && s.UnknownPaths == 0 &&
		s.RepositoryWrites == 0 && s.ExternalRepositoryWrites == 0 && s.OfficialMutations == 0 &&
		s.Promotions == 0 && s.IndicatorCompleted == 18 && s.IndicatorTotal == 18 &&
		s.IndicatorCoverageBPS == 10000 && s.DriverCompleted == 7 && s.DriverTotal == 7 &&
		s.OutcomeCompleted == 4 && s.OutcomeTotal == 4 && s.GuardrailCompleted == 7 && s.GuardrailTotal == 7
}
