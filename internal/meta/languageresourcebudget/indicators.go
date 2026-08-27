package languageresourcebudget

func buildIndicators(input Input, summaries []ResourceSummary, complete, semanticErr bool, violations int) []Indicator {
	values := make([]Indicator, 0, ExpectedIndicators)
	limits := input.Contract.Limits
	for index, summary := range summaries {
		spec := input.Contract.Operations[index]
		values = append(values,
			resourceIndicator(spec, "wall-time", summary.WallMaxNS, limits.WallTimeMS*1000000, complete && summary.WallMaxNS > 0 && summary.WallMaxNS <= limits.WallTimeMS*1000000),
			resourceIndicator(spec, "peak-rss", summary.PeakRSSMaxKiB, limits.PeakRSSKiB, complete && summary.PeakRSSMaxKiB > 0 && summary.PeakRSSMaxKiB <= limits.PeakRSSKiB),
			resourceIndicator(spec, "receipt-bytes", summary.ReceiptMax, limits.ReceiptBytes, complete && summary.ReceiptMax >= 0 && summary.ReceiptMax <= limits.ReceiptBytes),
			resourceIndicator(spec, "generated-bytes", summary.GeneratedMax, limits.GeneratedBytes, complete && summary.GeneratedMax >= 0 && summary.GeneratedMax <= limits.GeneratedBytes),
		)
	}
	values = append(values,
		metaIndicator("semantic.source-files", "DRIVER", "FOUNDATION", "observe-source-file-set", "INPUT", "source-identity", int64(input.Producer.SourceFileCount), int64(len(input.Contract.SourcePaths)), input.Producer.SourceFileCount == len(input.Contract.SourcePaths)),
		metaIndicator("semantic.go-files", "DRIVER", "FOUNDATION", "exclude-go-definition-files", "INPUT", "go-source-boundary", int64(input.Producer.GoFiles), 0, input.Producer.GoFiles == 0),
		metaIndicator("semantic.source-digest", "DRIVER", "FOUNDATION", "bind-source-content-digest", "INPUT", "source-digest", boolInt(contentDigest(input.Producer.SourceDigest)), 1, contentDigest(input.Producer.SourceDigest)),
		metaIndicator("semantic.source-receipt", "OUTCOME", "FOUNDATION", "verify-source-receipt", "REDUCE", "source-receipt", boolInt(!semanticErr), 1, !semanticErr),
		metaIndicator("semantic.artifact-replay", "OUTCOME", "COHERENCE", "compare-semantic-artifact-replay", "REDUCE", "artifact-replay", boolInt(!semanticErr), 1, !semanticErr),
		metaIndicator("semantic.artifact-digest", "DRIVER", "COHERENCE", "bind-artifact-content-digest", "REDUCE", "artifact-digest", boolInt(!semanticErr), 1, !semanticErr),
		metaIndicator("guardrail.effects", "GUARDRAIL", "REGRESSION", "verify-structured-write-set", "REDUCE", "effect-boundary", boolInt(writeSetTransitionOnly(input) == "DISCHARGED"), 1, writeSetTransitionOnly(input) == "DISCHARGED"),
		metaIndicator("guardrail.producer-binding", "GUARDRAIL", "FOUNDATION", "bind-producer-consumer-metadata", "REDUCE", "producer-consumer-binding", boolInt(complete), 1, complete),
		metaIndicator("guardrail.fixed-sample-set", "GUARDRAIL", "REGRESSION", "require-fixed-resource-sample-set", "REDUCE", "sample-cardinality", int64(len(input.Observations)), int64(len(input.Contract.Operations)*input.Contract.SamplesPerOp), complete),
		metaIndicator("guardrail.claim-chain", "GUARDRAIL", "REGRESSION", "preserve-claim-transition-chain", "REDUCE", "claim-transition", boolInt(complete && !semanticErr), 1, complete && !semanticErr),
	)
	_ = violations
	return values
}

func resourceIndicator(spec Operation, metric string, observed, expected int64, satisfied bool) Indicator {
	return metaIndicator("resource."+spec.ID+"."+metric, "DRIVER", spec.ProofChoice, spec.MetaOperation+"/"+metric, spec.Stage, spec.Step, observed, expected, satisfied)
}

func metaIndicator(id, class, proof, operation, stage, step string, observed, expected int64, satisfied bool) Indicator {
	return Indicator{ID: id, Class: class, ProofChoice: proof, Producer: Producer, Consumer: Consumer,
		MetaOperation: operation, Stage: stage, Step: step, Reason: "RESOURCE_CLAIM_EVALUATED",
		Observed: observed, Expected: expected, Satisfied: satisfied}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
