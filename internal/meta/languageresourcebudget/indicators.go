package languageresourcebudget

func buildIndicators(input Input, summaries []ResourceSummary, complete bool, semantic Semantic, violations int) []Indicator {
	values := make([]Indicator, 0, ExpectedIndicators+3)
	limits := input.Contract.Limits
	byOperation := make(map[string]ResourceSummary, len(summaries))
	for _, summary := range summaries {
		byOperation[summary.Operation] = summary
	}
	for _, spec := range input.Contract.Operations {
		summary := byOperation[spec.ID]
		values = append(values,
			resourceIndicator(spec, "wall-time", summary.WallMaxNS, limits.WallTimeMS*1000000, summary, true),
			resourceIndicator(spec, "peak-rss", summary.PeakRSSMaxKiB, limits.PeakRSSKiB, summary, true),
			resourceIndicator(spec, "receipt-bytes", summary.ReceiptMax, limits.ReceiptBytes, summary, spec.Output == "RECEIPT"),
			resourceIndicator(spec, "generated-bytes", summary.GeneratedMax, limits.GeneratedBytes, summary, spec.Output == "GENERATED"),
		)
	}
	semanticStatus := semanticIndicatorStatus(semantic)
	values = append(values,
		metaIndicatorStatus("semantic.source-files", "DRIVER", "FOUNDATION", "observe-source-file-set", "INPUT", "source-identity", int64(input.Producer.SourceFileCount), int64(len(input.Contract.SourcePaths)), statusFromBool(input.Producer.SourceFileCount == len(input.Contract.SourcePaths), semanticStatus), "SOURCE_FILE_SET_EVALUATED"),
		metaIndicatorStatus("semantic.go-files", "DRIVER", "FOUNDATION", "exclude-go-definition-files", "INPUT", "go-source-boundary", int64(input.Producer.GoFiles), 0, statusFromBool(input.Producer.GoFiles == 0, semanticStatus), "GO_SOURCE_BOUNDARY_EVALUATED"),
		metaIndicatorStatus("semantic.source-digest", "DRIVER", "FOUNDATION", "bind-source-content-digest", "INPUT", "source-digest", boolInt(contentDigest(input.Producer.SourceDigest)), 1, statusFromBool(contentDigest(input.Producer.SourceDigest), semanticStatus), "SOURCE_DIGEST_BINDING_EVALUATED"),
		metaIndicatorStatus("semantic.source-receipt", "OUTCOME", "FOUNDATION", "verify-source-receipt", "REDUCE", "source-receipt", boolInt(semanticStatus == "SATISFIED"), 1, semanticStatus, semantic.Reason),
		metaIndicatorStatus("semantic.artifact-replay", "OUTCOME", "COHERENCE", "compare-semantic-artifact-replay", "REDUCE", "artifact-replay", boolInt(semanticStatus == "SATISFIED"), 1, semanticStatus, semantic.Reason),
		metaIndicatorStatus("semantic.artifact-digest", "DRIVER", "COHERENCE", "bind-artifact-content-digest", "REDUCE", "artifact-digest", boolInt(semanticStatus == "SATISFIED"), 1, semanticStatus, semantic.Reason),
		metaIndicatorStatus("guardrail.effects", "GUARDRAIL", "REGRESSION", "verify-structured-write-set", "REDUCE", "effect-boundary", boolInt(writeSetTransitionOnly(input) == "DISCHARGED"), 1, effectIndicatorStatus(input), "NET_REPOSITORY_STATE_UNCHANGED"),
		metaIndicatorStatus("guardrail.producer-binding", "GUARDRAIL", "FOUNDATION", "bind-producer-consumer-metadata", "REDUCE", "producer-consumer-binding", boolInt(allPresentObservationMetadata(input)), 1, statusFromBool(allPresentObservationMetadata(input), "UNKNOWN"), "OBSERVATION_METADATA_BINDING_EVALUATED"),
		metaIndicatorStatus("guardrail.fixed-sample-set", "GUARDRAIL", "REGRESSION", "require-fixed-resource-sample-set", "REDUCE", "sample-cardinality", int64(len(input.Observations)), int64(len(input.Contract.Operations)*input.Contract.SamplesPerOp), statusFromBool(complete, "UNKNOWN"), "RESOURCE_SAMPLE_CARDINALITY_EVALUATED"),
		metaIndicatorStatus("guardrail.claim-chain", "GUARDRAIL", "REGRESSION", "preserve-claim-transition-chain", "REDUCE", "claim-transition", boolInt(semanticStatus == "SATISFIED" && complete), 1, statusFromBool(semanticStatus == "SATISFIED" && complete, "UNKNOWN"), "CLAIM_TRANSITION_CHAIN_EVALUATED"),
	)
	_ = violations
	return values
}

func resourceIndicator(spec Operation, metric string, observed, expected int64, summary ResourceSummary, applicable bool) Indicator {
	if !applicable {
		return indicator(spec, metric, observed, expected, "NOT_APPLICABLE", "METRIC_NOT_APPLICABLE_FOR_OUTPUT_KIND")
	}
	status, reason := "SATISFIED", "RESOURCE_ENVELOPE_WITHIN_LIMIT"
	if summary.MissingSamples > 0 {
		status, reason = "UNKNOWN", "RESOURCE_SAMPLE_MISSING"
	} else if summary.InvalidSamples > 0 {
		status, reason = "UNKNOWN", "RESOURCE_SAMPLE_INVALID"
	} else if observed > expected {
		status, reason = "REFUTED", "RESOURCE_BUDGET_EXCEEDED"
	} else if summary.Samples == 0 {
		status, reason = "UNKNOWN", "RESOURCE_SAMPLE_MISSING"
	}
	return indicator(spec, metric, observed, expected, status, reason)
}

func indicator(spec Operation, metric string, observed, expected int64, status, reason string) Indicator {
	return Indicator{ID: "resource." + spec.ID + "." + metric, Class: "DRIVER", ProofChoice: spec.ProofChoice, Producer: Producer, Consumer: Consumer,
		MetaOperation: spec.MetaOperation + "/" + metric, Stage: spec.Stage, Step: spec.Step, Reason: reason, Observed: observed, Expected: expected, Status: status, Satisfied: status == "SATISFIED"}
}

func metaIndicatorStatus(id, class, proof, operation, stage, step string, observed, expected int64, status, reason string) Indicator {
	return Indicator{ID: id, Class: class, ProofChoice: proof, Producer: Producer, Consumer: Consumer,
		MetaOperation: operation, Stage: stage, Step: step, Reason: reason, Observed: observed, Expected: expected, Status: status, Satisfied: status == "SATISFIED"}
}

func semanticIndicatorStatus(value Semantic) string {
	switch value.ClaimState {
	case "DISCHARGED":
		return "SATISFIED"
	case "OPEN":
		return "UNKNOWN"
	case "REFUTED":
		return "REFUTED"
	default:
		return "UNKNOWN"
	}
}

func effectIndicatorStatus(input Input) string {
	switch writeSetTransitionOnly(input) {
	case "DISCHARGED":
		return "SATISFIED"
	case "REFUTED":
		return "REFUTED"
	default:
		return "UNKNOWN"
	}
}

func statusFromBool(value bool, unknown string) string {
	if value {
		return "SATISFIED"
	}
	return unknown
}

func allPresentObservationMetadata(input Input) bool {
	if len(input.Observations) == 0 {
		return false
	}
	for _, value := range input.Observations {
		if value.Schema != ObservationSchema || value.SubjectSHA != input.ExpectedHead || value.Producer != Producer || value.Consumer != Consumer || value.Reason != "RUNNER_RESOURCE_OBSERVED" || !contentDigest(value.OutputDigest) || !contentDigest(value.SourceRawDigest) || !contentDigest(value.SourceSemanticDigest) || !contentDigest(value.EntryDigest) || !contentDigest(value.TargetDigest) {
			return false
		}
	}
	return true
}

func boolInt(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
