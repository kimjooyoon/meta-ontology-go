package userjourneyscorecard

func (s *inspection) buildReport(head string, contractRaw, upstreamRaw, profileRaw []byte) Report {
	indicators := s.indicators()
	satisfied, failures := 0, []string{}
	for _, indicator := range indicators {
		if indicator.Satisfied {
			satisfied++
		} else {
			failures = append(failures, indicator.ID)
		}
	}
	report := Report{
		Schema: "gooo/user-journey-scorecard/v1", Status: "PASS", Decision: "PASS",
		Resolution: "EXACT", Reason: "USER_JOURNEY_SCORECARD_EXACT",
		Source: EvidenceSource{ExpectedHeadSHA: head, Runner: s.profile.Runner,
			Executable: s.profile.Executable, SourcePath: s.profile.SourcePath, SourceDigest: s.profile.SourceDigest},
		Summary: Summary{
			Coordinates: Counter{Satisfied: satisfied, Total: 15, BasisPoints: satisfied * 10000 / 15},
			Functional: FunctionalSummary{UpstreamCases: s.upstream.Summary.Satisfied,
				PositiveJourneys: s.journeysPassed, OutputReplays: s.outputReplays,
				StructuredOutputs:  s.upstream.Summary.StructuredOutputs,
				LanguageOperations: s.upstream.Summary.LanguageOperations, DeclaredCommands: s.upstream.Summary.DeclaredCommands},
			Resources: ResourceSummary{SamplesObserved: s.samplesObserved,
				SamplesExpected: len(s.contract.Journeys) * s.contract.SamplesPerJourney,
				EnvelopesPassed: s.envelopesPassed, WallViolations: s.wallViolations,
				RSSViolations: s.rssViolations, BinarySizeBytes: s.profile.Executable.SizeBytes,
				BinarySizeLimit: s.contract.BinarySizeLimit, BinarySizeViolations: s.binaryViolations},
			Meta:    MetaSummary{Bindings: s.metaBindings, Unknowns: s.unknowns},
			Effects: EffectSummary{RepositoryWrites: s.repositoryWrites, MutationAuthority: s.upstream.MutationAuthorized}},
		Journeys: s.stats, Indicators: indicators, Views: buildViews(indicators), Proofs: proofs(indicators), Failures: failures,
		NotClaimed:     []string{"overall language completeness", "cross-run performance improvement", "cross-run memory improvement"},
		ContractDigest: digestBytes(contractRaw), UpstreamDigest: digestBytes(upstreamRaw), ProfileDigest: digestBytes(profileRaw),
		ResourceObservationMode: "RUNNER_SCOPED_NONDETERMINISTIC", ResourceMeasurementReplayAuthority: false,
		RepositoryWrites: s.repositoryWrites, MutationAuthority: false,
	}
	if s.lowerResolution {
		report.Status, report.Decision, report.Resolution, report.Reason = "FAIL_CLOSED", "FAIL_CLOSED", "LOWER_RESOLUTION", "USER_JOURNEY_EVIDENCE_UNKNOWN"
	} else if satisfied != 15 {
		report.Status, report.Decision, report.Resolution, report.Reason = "FAIL_CLOSED", "FAIL_CLOSED", "INVARIANT_ONLY", "USER_JOURNEY_CONTRACT_NOT_SATISFIED"
	}
	return report
}

func expectedContractBase() Contract {
	source := "examples/billing/main.gooo"
	return Contract{Schema: "gooo/user-journey-scorecard-contract/v1", Version: 2,
		SamplesPerJourney: 5, WallMSLimit: 5000, MaxRSSKiBLimit: 262144,
		BinarySizeLimit: 33554432, Source: source, Journeys: []JourneyDefinition{
			{ID: "version-text", Operation: "VERSION_TEXT", Arguments: []string{"version"}, ProofChoice: "FOUNDATION", MetaOperation: "measure-version-text-resource"},
			{ID: "version-json", Operation: "VERSION_JSON", Arguments: []string{"version", "--json"}, ProofChoice: "FOUNDATION", MetaOperation: "measure-version-json-resource"},
			{ID: "check-text", Operation: "CHECK_TEXT", Arguments: []string{"check", source}, ProofChoice: "COHERENCE", MetaOperation: "measure-syntax-check-resource"},
			{ID: "check-json", Operation: "CHECK_JSON", Arguments: []string{"check", "--json", source}, ProofChoice: "COHERENCE", MetaOperation: "measure-structured-check-resource"},
			{ID: "roundtrip-json", Operation: "ROUNDTRIP_JSON", Arguments: []string{"roundtrip", "--json", source}, ProofChoice: "COHERENCE", MetaOperation: "measure-roundtrip-resource"},
			{ID: "semantic-check", Operation: "SEMANTIC_CHECK", Arguments: []string{"check", "--semantic", source}, ProofChoice: "COHERENCE", MetaOperation: "measure-semantic-check-resource"},
			{ID: "run-source", Operation: "RUN_SOURCE", Arguments: []string{"run", "--json", "--entry", "PayOrder", source}, ProofChoice: "COHERENCE", MetaOperation: "measure-source-execution-resource"},
		}}
}
