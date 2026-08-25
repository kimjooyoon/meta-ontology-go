package symbolicinvocationusecase

import "slices"

const (
	reasonSatisfied       = "SYMBOLIC_INVOCATION_USECASE_OBSERVED"
	reasonDecisionUnknown = "SYMBOLIC_INVOCATION_USECASE_DECISION_UNKNOWN"
	reasonSubjectMismatch = "SYMBOLIC_INVOCATION_USECASE_SUBJECT_MISMATCH"
	reasonEvidenceInvalid = "SYMBOLIC_INVOCATION_USECASE_EVIDENCE_INVALID"
	reasonLinkMismatch    = "SYMBOLIC_INVOCATION_USECASE_LINK_MISMATCH"
	reasonEffectsObserved = "SYMBOLIC_INVOCATION_USECASE_EFFECTS_OBSERVED"
)

func Evaluate(input Input) (Report, error) {
	if err := input.Contract.Validate(); err != nil {
		return Report{}, err
	}
	value, reason, resolution := collectFacts(input)
	indicators := buildIndicators(input.Contract, value)
	if reason != "" {
		for index := range indicators {
			indicators[index].Satisfied = false
		}
	}
	coordinates := countIndicators(indicators)
	report := Report{
		Schema: "gooo/symbolic-invocation-usecase-report/v1", SubjectSHA: input.SubjectSHA,
		MetricID: input.Contract.MetricID, Decision: "PASS", Resolution: "EXACT", Reason: reasonSatisfied,
		Summary: Summary{
			Coordinates: coordinates, UserDecisions: value.UserDecisions,
			AcceptedInstances: value.AcceptedInstances, RejectedInstances: value.RejectedInstances,
			DeterministicReplays: value.DeterministicReplays, Unknowns: value.Unknowns,
			Source: value.Source, Producer: value.Producer, Resources: value.Resources, Effects: value.Effects,
		},
		Indicators: indicators, Views: buildViews(indicators), PromotionCreditBPS: 0,
		RepositoryWrites: value.Effects.RepositoryWrites, MutationAuthority: value.Effects.MutationAuthority,
		NotClaimed: CanonicalNonClaims(),
	}
	if reason != "" {
		report.Decision, report.Resolution, report.Reason = "FAIL_CLOSED", resolution, reason
	} else if coordinates.Satisfied != coordinates.Total {
		report.Decision, report.Resolution, report.Reason = "FAIL_CLOSED", "INVARIANT_ONLY", reasonEvidenceInvalid
	}
	return sealReport(report), nil
}

func collectFacts(input Input) (facts, string, string) {
	receipt, artifact, observation := input.ProducerReceipt, input.ProducerArtifact, input.Observation
	if !validSHA(input.SubjectSHA) {
		return facts{Unknowns: 1}, reasonDecisionUnknown, "LOWER_RESOLUTION"
	}
	if unknownTop(receipt.Decision, receipt.Resolution) || unknownTop(artifact.Decision, artifact.Resolution) ||
		unknownTop(observation.Decision, observation.Resolution) {
		return facts{Unknowns: 1}, reasonDecisionUnknown, "LOWER_RESOLUTION"
	}
	if receipt.Decision != "PASS" || receipt.Resolution != "EXACT" || artifact.Decision != "PASS" ||
		artifact.Resolution != "SYMBOLIC_ONLY" || observation.Decision != "PASS" || observation.Resolution != "EXACT" {
		return facts{}, reasonEvidenceInvalid, "INVARIANT_ONLY"
	}
	if receipt.SubjectSHA != input.SubjectSHA || observation.SubjectSHA != input.SubjectSHA {
		return facts{}, reasonSubjectMismatch, "INVARIANT_ONLY"
	}
	if receipt.Schema != "gooo/symbolic-invocation-schema-receipt/v1" ||
		receipt.Reason != "EXTERNAL_SCHEMA_VALIDATION_OBSERVED" ||
		artifact.Schema != "gooo/symbolic-invocation-schema-artifact/v1" ||
		artifact.Reason != "SYMBOLIC_INVOCATION_SCHEMA_EMITTED" || artifact.Kind != "symbolic-invocation-schema" ||
		observation.Schema != "gooo/symbolic-invocation-usecase-observation/v1" ||
		observation.Reason != "EXTERNAL_USER_VALIDATION_REPLAYED" {
		return facts{}, reasonEvidenceInvalid, "INVARIANT_ONLY"
	}
	if !validDigest(receipt.Compiler.BinaryDigest) || !validDigest(receipt.Artifact.Digest) ||
		!validDigest(receipt.Artifact.JSONSchemaDigest) || !validDigest(receipt.Validation.ToolDigest) ||
		!validDigest(artifact.Digest) || !validDigest(observation.ArtifactDigest) ||
		!validDigest(observation.JSONSchemaDigest) || !validDigest(observation.ToolDigest) {
		return facts{}, reasonEvidenceInvalid, "INVARIANT_ONLY"
	}
	if receipt.Artifact.ArtifactSchema != artifact.Schema || receipt.Artifact.Digest != artifact.Digest ||
		receipt.Artifact.JSONSchemaDigest != observation.JSONSchemaDigest ||
		receipt.Validation.ToolDigest != observation.ToolDigest || artifact.Digest != observation.ArtifactDigest {
		return facts{}, reasonLinkMismatch, "INVARIANT_ONLY"
	}
	contract := input.Contract
	if receipt.Compiler.GoVersion != contract.ExpectedGoVersion ||
		receipt.Compiler.RegisteredEmitters != contract.ExpectedRegisteredEmitters ||
		artifact.Extensions.RegisteredEmitters != contract.ExpectedRegisteredEmitters ||
		!slices.Equal(artifact.Extensions.Kinds, []string{"operation-interface", "operation-manifest", "symbolic-invocation-schema"}) ||
		receipt.Compiler.BinaryBytes < 1 {
		return facts{}, reasonEvidenceInvalid, "INVARIANT_ONLY"
	}
	wantSource := SourceCoordinate{
		GoooFiles: contract.ExpectedGoooFiles, GoFiles: contract.ExpectedGoFiles,
		GoooLines: contract.ExpectedGoooLines, Files: contract.ExpectedFiles, Directories: contract.ExpectedDirectories,
	}
	if receipt.Source != wantSource || receipt.Artifact.Kind != artifact.Kind ||
		receipt.Artifact.JSONSchemaDialect != "https://json-schema.org/draft/2020-12/schema" ||
		receipt.Validation.Tool != contract.ExpectedValidator ||
		receipt.Validation.AcceptedInstances != contract.ExpectedAcceptedInstances ||
		receipt.Validation.RejectedInstances != contract.ExpectedRejectedInstances ||
		receipt.DeterministicReplays != contract.ExpectedDeterministicReplays ||
		!canonicalNonClaims(receipt.NotClaimed) || len(receipt.NotClaimed) != contract.ExpectedNonClaims {
		return facts{}, reasonEvidenceInvalid, "INVARIANT_ONLY"
	}
	if observation.AcceptedInstances != contract.ExpectedAcceptedInstances ||
		observation.RejectedInstances != contract.ExpectedRejectedInstances ||
		!validResources(receipt.Resources, contract.ExpectedResourceSamples) {
		return facts{}, reasonEvidenceInvalid, "INVARIANT_ONLY"
	}
	effects := Effects{
		RepositoryWrites:  receipt.Effects.RepositoryWrites + artifact.Effects.RepositoryWrites + observation.Effects.RepositoryWrites,
		MutationAuthority: receipt.Effects.MutationAuthority || artifact.Effects.MutationAuthority || observation.Effects.MutationAuthority,
	}
	value := facts{
		UserDecisions:     observation.AcceptedInstances + observation.RejectedInstances,
		AcceptedInstances: observation.AcceptedInstances, RejectedInstances: observation.RejectedInstances,
		DeterministicReplays: receipt.DeterministicReplays, Source: receipt.Source, Effects: effects,
		Producer: ProducerBinding{
			ReceiptSchema: receipt.Schema, ArtifactSchema: artifact.Schema, ArtifactDigest: artifact.Digest,
			JSONSchemaDigest: receipt.Artifact.JSONSchemaDigest, Validator: receipt.Validation.Tool,
			ValidatorDigest: receipt.Validation.ToolDigest, CompilerBinaryBytes: receipt.Compiler.BinaryBytes,
			CompilerBinaryDigest: receipt.Compiler.BinaryDigest, RegisteredEmitters: receipt.Compiler.RegisteredEmitters,
		},
		Resources: ResourceObservation{
			Mode: "RUNNER_SCOPED_NONDETERMINISTIC", MeasurementReplayAuthority: false,
			Samples: receipt.Resources.SampleCount, MaxWallMS: receipt.Resources.MaxWallMS, MaxRSSKiB: receipt.Resources.MaxRSSKiB,
		},
	}
	mutationAuthorities := 0
	if effects.MutationAuthority {
		mutationAuthorities = 1
	}
	if effects.RepositoryWrites != contract.ExpectedRepositoryWrites || mutationAuthorities != contract.ExpectedMutationAuthorities {
		return value, reasonEffectsObserved, "INVARIANT_ONLY"
	}
	return value, "", "EXACT"
}

func unknownTop(decision, resolution string) bool {
	knownDecision := decision == "PASS" || decision == "FAIL_CLOSED"
	knownResolution := resolution == "EXACT" || resolution == "SYMBOLIC_ONLY" ||
		resolution == "LOWER_RESOLUTION" || resolution == "INVARIANT_ONLY"
	return !knownDecision || !knownResolution
}

func validResources(value ResourceEvidence, expected int) bool {
	if value.SampleCount != expected || len(value.Samples) != expected || value.MaxWallMS < 1 || value.MaxRSSKiB < 1 {
		return false
	}
	maxWall, maxRSS := 0, 0
	for index, sample := range value.Samples {
		if sample.Sequence != index+1 || sample.WallMS < 1 || sample.RSSKiB < 1 {
			return false
		}
		maxWall = max(maxWall, sample.WallMS)
		maxRSS = max(maxRSS, sample.RSSKiB)
	}
	return maxWall == value.MaxWallMS && maxRSS == value.MaxRSSKiB
}

func buildIndicators(contract Contract, value facts) []Indicator {
	mutationAuthorities := 0
	if value.Effects.MutationAuthority {
		mutationAuthorities = 1
	}
	return []Indicator{
		indicator("user.validation-decisions", "OUTCOME", "COHERENCE", "sum-external-user-decisions", value.UserDecisions, contract.ExpectedAcceptedInstances+contract.ExpectedRejectedInstances),
		indicator("user.accepted-instances", "DRIVER", "FOUNDATION", "count-externally-accepted-instances", value.AcceptedInstances, contract.ExpectedAcceptedInstances),
		indicator("user.rejected-instances", "DRIVER", "REGRESSION", "count-externally-rejected-instances", value.RejectedInstances, contract.ExpectedRejectedInstances),
		indicator("guardrail.deterministic-replays", "GUARDRAIL", "FOUNDATION", "count-producer-replays", value.DeterministicReplays, contract.ExpectedDeterministicReplays),
		indicator("guardrail.repository-writes", "GUARDRAIL", "FOUNDATION", "sum-cross-boundary-writes", value.Effects.RepositoryWrites, contract.ExpectedRepositoryWrites),
		indicator("guardrail.mutation-authorities", "GUARDRAIL", "COHERENCE", "join-cross-boundary-authority", mutationAuthorities, contract.ExpectedMutationAuthorities),
	}
}

func indicator(id, class, proof, operation string, observed, expected int) Indicator {
	return Indicator{ID: id, Class: class, ProofChoice: proof, MetaOperation: operation,
		Observed: observed, Expected: expected, Satisfied: observed == expected}
}

func countIndicators(indicators []Indicator) Counter {
	result := Counter{Total: len(indicators)}
	for _, indicator := range indicators {
		if indicator.Satisfied {
			result.Satisfied++
		}
	}
	if result.Total > 0 {
		result.BasisPoints = result.Satisfied * 10000 / result.Total
	}
	return result
}

func buildViews(indicators []Indicator) []View {
	return []View{
		buildView("USER", "USER_VISIBLE", indicators[:3]),
		buildView("TOOL_AUTHOR", "TOOL_CONTRACT", indicators[:4]),
		buildView("GOVERNOR", "FULL_RECEIPT", indicators),
	}
}

func buildView(audience, resolution string, indicators []Indicator) View {
	view := View{Audience: audience, Resolution: resolution, Total: len(indicators)}
	for _, indicator := range indicators {
		view.IndicatorIDs = append(view.IndicatorIDs, indicator.ID)
		if indicator.Satisfied {
			view.Satisfied++
		}
	}
	if view.Total > 0 {
		view.BasisPoints = view.Satisfied * 10000 / view.Total
	}
	return view
}
