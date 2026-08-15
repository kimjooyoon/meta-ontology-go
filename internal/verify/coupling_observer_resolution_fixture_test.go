package verify

func hasCouplingFailure(evidence CouplingEvidence, reason string) bool {
	want := CouplingFailureCodePrefix + reason
	for _, failure := range evidence.Failures {
		if failure.Code != want {
			continue
		}
		if reason == "source-binding-mismatch" && (failure.Domain != CouplingDomainIntegrity || failure.Owner != couplingSurface().SemanticOwnerID || failure.Retry) {
			return false
		}
		if reason == "surface-unregistered" || reason == "ambiguous-origin" || reason == "surface-not-applicable" || reason == "no-changed-sites" {
			return failure.Domain == CouplingDomainDependency && failure.Owner == CouplingOwnerUnavailable && failure.Retry
		}
		return true
	}
	return false
}

func couplingResolutionFixtures() []couplingFixtureCase {
	return []couplingFixtureCase{
		{"unregistered changed site", "NO_DELTA", CouplingDecisionUnknown, "surface-unregistered", func(in *CouplingInput) {
			in.ChangedSites[0].Path = "internal/unknown/missing.go"
			in.ChangedSites[0].CodeSymbolID = ""
		}},
		{"not applicable only changed site", "NO_DELTA", CouplingDecisionUnknown, "surface-not-applicable", func(in *CouplingInput) {
			in.Registry.Surfaces[0].Applicability = CouplingNotApplicable
			in.Envelope.RegistryDigest = in.Registry.Digest()
		}},
		{"zero changed sites", "NO_DELTA", CouplingDecisionUnknown, "no-changed-sites", func(in *CouplingInput) { in.ChangedSites = nil }},
	}
}
