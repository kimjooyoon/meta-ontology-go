package languagedelivery

func observeRule(rule EvidenceRule, decoded decodedEvidence) (int, string) {
	switch rule.Kind {
	case EvidenceJourney:
		for _, item := range decoded.Journey.Journeys {
			if item.ID == rule.ID && item.Samples > 0 && item.Successful == item.Samples && item.OutputReplay && item.EnvelopePassed {
				return 1, "USER_JOURNEY_EXACT"
			}
		}
	case EvidenceIndicator:
		for _, item := range decoded.Journey.Indicators {
			if item.ID == rule.ID && item.Satisfied {
				return 1, "USER_INDICATOR_EXACT"
			}
		}
	case EvidenceSurface:
		for _, item := range decoded.Conformance.Surfaces {
			if item.ID == rule.ID && item.Status == "SATISFIED" {
				return 1, "CONFORMANCE_SURFACE_EXACT"
			}
		}
	case EvidenceReadiness:
		for _, item := range decoded.Readiness.Snapshot.Obligations {
			if item.ID == rule.ID && item.Status == "SATISFIED" {
				return 1, "READINESS_OBLIGATION_EXACT"
			}
		}
	case EvidenceLSPCounter, EvidenceConformance, EvidenceRelease:
		return observeCounter(rule, decoded)
	}
	return 0, "REQUIRED_EVIDENCE_NOT_SATISFIED"
}

func observeCounter(rule EvidenceRule, decoded decodedEvidence) (int, string) {
	switch rule.Kind {
	case EvidenceLSPCounter:
		if rule.Counter == "diagnostic_paths" {
			return decoded.LSP.Summary.DiagnosticPaths, "LSP_COUNTER_OBSERVED"
		}
		if rule.Counter == "navigation_paths" {
			return decoded.LSP.Summary.NavigationPaths, "LSP_COUNTER_OBSERVED"
		}
	case EvidenceConformance:
		return decoded.Conformance.Summary.SurfacesSatisfied, "CONFORMANCE_COUNTER_OBSERVED"
	case EvidenceRelease:
		if decoded.Release.Summary.NativeSmokes < decoded.Release.Summary.PlatformReceipts {
			return decoded.Release.Summary.NativeSmokes, "RELEASE_SMOKE_COUNTER_OBSERVED"
		}
		return decoded.Release.Summary.PlatformReceipts, "RELEASE_PLATFORM_COUNTER_OBSERVED"
	case EvidenceExecution:
		switch rule.Counter {
		case "source_executions":
			return decoded.Execution.Summary.SourceExecutions, "SOURCE_EXECUTIONS_OBSERVED"
		case "deterministic_replays":
			return decoded.Execution.Summary.DeterministicReplays, "SOURCE_EXECUTION_REPLAYS_OBSERVED"
		case "diagnostic_rejections":
			return decoded.Execution.Summary.DiagnosticRejections, "SOURCE_EXECUTION_DIAGNOSTICS_OBSERVED"
		}
	}
	return 0, "COUNTER_UNKNOWN"
}
